package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Stock       int                `bson:"stock"`
	Price       float64            `bson:"price"`
	Type        string             `bson:"tipe"`
}
