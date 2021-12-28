package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/kede-awak/product-service/model/proto"
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

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID: %v\n", err),
		)
	}

	return &proto.CreateProductResponse{
		Product: &proto.Product{
			Id:          oid.Hex(),
			Name:        product.GetName(),
			Description: product.GetDescription(),
			Stock:       product.GetStock(),
			Price:       product.GetPrice(),
		},
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
