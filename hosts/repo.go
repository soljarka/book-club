package hosts

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		log.Fatal(err)
	}
}

func insertBookclub(bookclub *Bookclub) error {
	collection := client.Database("bookclubdb").Collection("bookclubs")
	_, err := collection.InsertOne(context.TODO(), bookclub)
	return err
}

func updateBookclub(bookclub *Bookclub) error {
	collection := client.Database("bookclubdb").Collection("bookclubs")
	filter := bson.D{{Key: "_id", Value: bookclub.ChatId}}
	_, err := collection.ReplaceOne(context.TODO(), filter, bookclub)
	return err
}

func loadBookclub(chatId int64) (*Bookclub, error) {
	var bookclub Bookclub
	collection := client.Database("bookclubdb").Collection("bookclubs")
	filter := bson.D{{Key: "_id", Value: chatId}}
	err := collection.FindOne(context.TODO(), filter).Decode(&bookclub)
	return &bookclub, err
}
