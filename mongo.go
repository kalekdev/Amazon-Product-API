package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client          *mongo.Client
	countCollection *mongo.Collection
)

const MONGO_URI = "MONGO_CONNECTION_STRING"

type CountObject struct {
	Count int
}

func init() {
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(MONGO_URI))

	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	countCollection = client.Database("amazon").Collection("counters")
	fmt.Println("Successfully connected to Database.")
}

func incrementUsage(apiKey string) error {
	_, err := countCollection.UpdateOne(context.TODO(), bson.D{{"_id", apiKey}}, bson.D{{"$inc", bson.D{{"count", 1}}}})

	return err
}

func getUsage(apiKey string) (int, error) {
	var countObj CountObject
	err := countCollection.FindOne(context.TODO(), bson.D{{"_id", apiKey}}).Decode(&countObj)
	return countObj.Count, err
}
