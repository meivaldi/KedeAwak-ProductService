package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"desc"`
	Stock       int                `bson:"stock"`
	Price       float64            `bson:"price"`
}
