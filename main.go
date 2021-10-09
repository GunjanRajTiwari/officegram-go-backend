package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	Password	string 				`json:"-"`
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
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
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


// ------- HELPERS --------- //
func encrypt(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[0:])
}


// ------- HANDLERS --------- //

// User Handlers
func userHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

    switch r.Method {
    case "GET":
		userId, _ := primitive.ObjectIDFromHex(r.URL.Path[7:24+7])

		filter := bson.D{{"_id", userId}}
		var profile User

		err := db.Collection("users").FindOne(ctx, filter).Decode(&profile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
		}

		json.NewEncoder(w).Encode(profile)

    case "POST":
		var newUser User
		json.NewDecoder(r.Body).Decode(&newUser)

		newUser.Password = encrypt(newUser.Password)

		res, err := db.Collection("users").InsertOne(ctx, newUser)
        if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
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
		postId, _ := primitive.ObjectIDFromHex(r.URL.Path[7:24+7])
		fmt.Println(r.URL.Path[7:24+7])

		filter := bson.D{{"_id", postId}}
		var post Post

		err := db.Collection("posts").FindOne(ctx, filter).Decode(&post)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
		}

		json.NewEncoder(w).Encode(post)

    case "POST":
		var newPost Post
		json.NewDecoder(r.Body).Decode(&newPost)

		res, err := db.Collection("posts").InsertOne(ctx, newPost)
        if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
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
		var page int
		
		query := r.URL.Query()
		if val, ok := query["page"]; ok {
			page,_ = strconv.Atoi(val[0])
		}

		skip := int64(page*5)
		limit := int64(5)

		opts := options.FindOptions {
			Skip: &skip,
			Limit: &limit,
		}

		userId, _ := primitive.ObjectIDFromHex(r.URL.Path[13:24+13])

		filter := bson.D{{"creator", userId}}
		var posts []Post

		cursor, err := db.Collection("posts").Find(ctx, filter, &opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
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
	
	http.HandleFunc("/users/", userHandler)
	http.HandleFunc("/users", userHandler)

	http.HandleFunc("/posts/users/", userPostHandler)
	http.HandleFunc("/posts/users", userPostHandler)

	http.HandleFunc("/posts/", postHandler)
	http.HandleFunc("/posts", postHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "not found"}`))
	})

	log.Fatal(http.ListenAndServe(":8000", nil))

}