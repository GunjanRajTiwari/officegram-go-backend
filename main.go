package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	Creator 	primitive.ObjectID	`json:"creator"`
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

	db = client.Database("officegram")
}


// ------- HANDLERS --------- //

// User Handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

    switch r.Method {
    case "GET":
		userId, _ := primitive.ObjectIDFromHex(r.URL.Path[7:])
		fmt.Println(userId)

		filter := bson.D{{"_id", userId}}
		var profile User

		err := db.Collection("users").FindOne(ctx, filter).Decode(&profile)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(profile)

    case "POST":
		var newUser User
		json.NewDecoder(r.Body).Decode(&newUser)

		hash := sha256.Sum256([]byte(newUser.Password))
		newUser.Password = hex.EncodeToString(hash[0:])

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

// Post Handlers
func postHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

    switch r.Method {
    case "GET":
		postId, _ := primitive.ObjectIDFromHex(r.URL.Path[7:])
		fmt.Println(postId)

		filter := bson.D{{"_id", postId}}
		var post Post

		err := db.Collection("posts").FindOne(ctx, filter).Decode(&post)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(post)

    case "POST":
		var newPost Post
		json.NewDecoder(r.Body).Decode(&newPost)

		res, err := db.Collection("posts").InsertOne(ctx, newPost)
        if err != nil {
			log.Fatal(err)
		}
		
        json.NewEncoder(w).Encode(res)
		fmt.Println("Post Created")

    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "not found"}`))
    }
}

// Users Post Handlers
func userPostHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

    switch r.Method {
    case "GET":
		userId, _ := primitive.ObjectIDFromHex(r.URL.Path[13:])

		filter := bson.D{{"creator", userId}}
		var posts []Post

		cursor, err := db.Collection("posts").Find(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var post Post
			cursor.Decode(&post)
			posts = append(posts, post)
		}

		json.NewEncoder(w).Encode(posts)

    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "not found"}`))
    }
}


// -------- START -----/
func main() {
	fmt.Println("Welcome to Golang Rest API")

	connectDb()
	
	http.HandleFunc("/users", userHandler)
	http.HandleFunc("/posts/users", userPostHandler)
	http.HandleFunc("/posts", postHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))

}