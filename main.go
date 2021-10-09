package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ------- MODELS ------ //

type User struct {
	ID 			primitive.ObjectID	`json:"_id,omitempty" bson:"_id,omitempty"`
	Name 		string 				`json:"name"`
	Email		string 				`json:"email"`
	Password	string 				`json:"password"`
}

type Post struct {
	ID 			primitive.ObjectID	`json:"_id,omitempty" bson:"_id,omitempty"`
	Caption		string				`json:"caption"`
	ImageURL	string				`json:"imageURL"`
	Created		primitive.Timestamp	`json:"created,omitempty"`
}

// --------- DB CONNECTION ----- //
var db *mongo.Database

func connectDb() {
	// clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	clientOptions := options.Client().ApplyURI("mongodb+srv://gunjan:gunjan321@namesclash.qtwfp.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	client, err := mongo.NewClient(clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	defer cancel()

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatal("Couldn't connect to the database:\n", err)
	}
	log.Println("Database Connected!")
	
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(databases)
	db = client.Database("officegram")
}

// ------- HANDLERS --------- //
// User Handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

    switch r.Method {
    case "GET":
		cursor, err := db.Collection("users").Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}

		defer cursor.Close(ctx)

		var users []User
		for cursor.Next(ctx) {
			var profile User
			cursor.Decode(&profile)
			users = append(users, profile)
		}

		json.NewEncoder(w).Encode(users)

    case "POST":
		var newUser User
		json.NewDecoder(r.Body).Decode(&newUser)

		res, err := db.Collection("users").InsertOne(ctx, newUser)
        if err != nil {
			log.Fatal(err)
		}
		
        json.NewEncoder(w).Encode(res)
		fmt.Println("User Created")

    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "not found"}`))
    }
}



func main() {
	fmt.Println("Welcome to Golang Rest API")

	connectDb()
	
	http.HandleFunc("/users", userHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))


}