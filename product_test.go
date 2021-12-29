package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/kede-awak/product-service/model/proto"
	"google.golang.org/grpc"
)

func TestProductCreate(t *testing.T) {
	opts := grpc.WithInsecure()

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := proto.NewProductServiceClient(conn)

	req := &proto.CreateProductRequest{
		Product: &proto.Product{
			Name:        "Ipad Pro M1 11 Inch",
			Description: "Ipad Pro M1 11 Inch Wifi Only",
			Stock:       8,
			Price:       12000000,
		},
	}

	res, err := c.CreateProduct(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %v\n", res)
}

func TestProductRead(t *testing.T) {
	opts := grpc.WithInsecure()

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := proto.NewProductServiceClient(conn)
	req := &proto.ReadProductRequest{Id: "61cb2b463f79f2a6eeb96f94"}

	res, err := c.ReadProduct(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Res: %v\n", res)
}

func TestProductUpdate(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	opts := grpc.WithInsecure()

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := proto.NewProductServiceClient(conn)
	req := &proto.UpdateProductRequest{
		Product: &proto.Product{
			Id:          "61cb2b463f79f2a6eeb96f94",
			Name:        "Ipad Pro M1 12.9 Inch Cellular 128 GB",
			Description: "Ipad Pro M1 12.9 Inch Cellular 128GB di tokohapedia",
			Price:       14490000,
			Stock:       50,
		},
	}

	res, err := c.UpdateProduct(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Res: %v\n", res)
}

func TestProductDelete(t *testing.T) {
	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := proto.NewProductServiceClient(conn)
	req := &proto.DeleteProductRequest{
		Id: "61cc0cd0e40c3df9e41f74e1",
	}

	res, err := c.DeleteProduct(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Res: %v\n", res)
}
