package main

import (
	"context"
	"log"

	"github.com/mongodb/mongo-go-driver/mongo"
)

const (
	hosts          = "localhost:27017"
	database       = "test_db"
	username       = "tmt"
	password       = "123456789"
	collection     = "notification"
	replicaSetName = "mongo-rs"
)

type Notify struct {
	Receivers        []string
	Message          string
}

func main() {
	client, err := mongo.NewClient("mongodb://tmt:123456789@localhost:27017/test_db")
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(database).Collection(collection)
	receivers := []string{"123", "456"}
	res, err := collection.InsertOne(context.Background(), Notify{receivers, "Nhận được thông báo rồi này."})
	if err != nil { log.Fatal(err) }
	id := res.InsertedID
	log.Println(id)
}

