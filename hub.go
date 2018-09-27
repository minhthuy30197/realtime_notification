package main

import (
	"context"
	"log"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type Hub struct {
	// Registered clients.
	clients map[string]*ListClient
}

type ListClient struct {
	list map[*Client]bool
}
func newHub() *Hub {
	return &Hub{
		clients:    make(map[string]*ListClient),
	}
}

type IDELem struct {
	Data string `json:"data" bson:"_data"`
}
type FullDocumentElem struct {
	Receivers        []string `json:"id" bson:"receivers"`
	Message          string	  `json:"id" bson:"message"`
}
type NSELem struct {
	DB   string `json:"db" bson:"db"`
	Coll string `json:"coll" bson:"coll"`
}
type DocumentKeyElem struct {
	ID objectid.ObjectID `json:"id" bson:"_id"`
}
type CSElem struct {
	ID            IDELem           `json:"id" bson:"_id"`
	OperationType string           `json:"operationType" bson:"operationType"`
	ClusterTime   time.Time        `json:"clusterTime" bson:"clusterTime"`
	FullDocument  FullDocumentElem `json:"fullDocument" bson:"fullDocument"`
	NS            NSELem           `json:"ns" bson:"ns"`
	DocumentKey   DocumentKeyElem  `json:"documentKey" bson:"documentKey"`
}

const (
	hosts          = "localhost:27017"
	database       = "test_db"
	username       = "tmt"
	password       = "123456789"
	collection     = "notification"
	replicaSetName = "mongo-rs"
)

func (h *Hub) run() {
	client, err := mongo.NewClient("mongodb://tmt:123456789@localhost:27017/test_db")
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(database).Collection(collection)
	ctx := context.Background()

	var pipeline interface{} 

	cur, err := collection.Watch(ctx, pipeline)
	if err != nil {
		// Handle err
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		elem := CSElem{}

		if err := cur.Decode(&elem); err != nil {
			log.Fatal(err)
		}

		log.Println(elem)
		log.Println("---------------")

		// Gửi message cho các client 
		for userID := range h.clients {
			for _, id := range elem.FullDocument.Receivers {
				if userID == id {
					msg := []byte(elem.FullDocument.Message)
					for key := range h.clients[userID].list {
						key.send <- msg
					}
					break
				}
			}
		}
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}
