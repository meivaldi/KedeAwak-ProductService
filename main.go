package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kede-awak/product-service/model/entity"
	"github.com/kede-awak/product-service/model/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	proto.UnimplementedProductServiceServer
}

var collection *mongo.Collection

func (*server) CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.CreateProductResponse, error) {
	product := req.GetProduct()

	data := proto.Product{
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Stock:       product.GetStock(),
		Price:       product.GetPrice(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Server Error: %v\n", err),
		)
	}

	_, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID: %v\n", err),
		)
	}

	return &proto.CreateProductResponse{
		Product: &proto.Product{
			Name:        product.GetName(),
			Description: product.GetDescription(),
			Stock:       product.GetStock(),
			Price:       product.GetPrice(),
		},
	}, nil
}

func (*server) ReadProduct(ctx context.Context, req *proto.ReadProductRequest) (*proto.ReadProductResponse, error) {
	productId := req.GetId()
	oid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse product id: %v\n", err),
		)
	}

	data := &entity.Product{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find document with specified id: %v\n", err),
		)
	}

	product := &entity.Product{
		Id:          data.Id,
		Name:        data.Name,
		Description: data.Description,
		Stock:       data.Stock,
		Price:       data.Price,
	}
	jsonProduct, err := json.Marshal(product)
	if err != nil {
		log.Fatalf("Cannot convert data to json format: %v\n", err)
	}

	//redis connection block
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	errRedis := rdb.Set(ctx, "product", jsonProduct, 15*time.Second).Err()
	if errRedis != nil {
		log.Fatalf("Redis error: %v\n", errRedis)
	}
	//

	return &proto.ReadProductResponse{
		Product: &proto.Product{
			Name:        data.Name,
			Description: data.Description,
			Stock:       int32(data.Stock),
			Price:       float32(data.Price),
		},
	}, nil
}

func (*server) UpdateProduct(ctx context.Context, req *proto.UpdateProductRequest) (*proto.UpdateProductResponse, error) {
	product := req.GetProduct()
	oid, err := primitive.ObjectIDFromHex(product.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse product id: %v\n", err),
		)
	}

	data := &entity.Product{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find document with specified id: %v\n", err),
		)
	}

	data.Name = product.GetName()
	data.Description = product.GetDescription()
	data.Price = float64(product.GetPrice())
	data.Stock = int(product.GetStock())

	_, errUpdate := collection.ReplaceOne(context.Background(), filter, data)
	if errUpdate != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update data: %v\n", errUpdate),
		)
	}

	return &proto.UpdateProductResponse{
		Product: &proto.Product{
			Id:          data.Id.Hex(),
			Name:        data.Name,
			Description: data.Description,
			Price:       float32(data.Price),
			Stock:       int32(data.Stock),
		},
	}, nil
}

func (*server) DeleteProduct(ctx context.Context, req *proto.DeleteProductRequest) (*proto.DeleteProductResponse, error) {
	productId := req.GetId()
	oid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse product id: %v\n", err),
		)
	}

	data := &entity.Product{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find document with specified id: %v\n", err),
		)
	}

	errRes, errDelete := collection.DeleteOne(context.Background(), filter)
	if errDelete != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete data: %v\n", errDelete),
		)
	}

	if errRes.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find document with specified id: %v\n", errDelete),
		)
	}

	return &proto.DeleteProductResponse{
		Id: productId,
	}, nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.Println("Product service started...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection = client.Database("kede_awak").Collection("product")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	proto.RegisterProductServiceServer(s, &server{})

	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	log.Println("Stopping product service...")
	s.Stop()
	log.Println("Stopping listener...")
	lis.Close()
	log.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	log.Println("Product service stopped")
}
