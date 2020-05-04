package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	connect = "mongodb+srv://neel:testpass1994@movies-3mimb.mongodb.net/test?retryWrites=true&w=majority"
	client  = GetClient()
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/movies", GetMovies).Methods("GET")
	router.HandleFunc("/movies/{name}", GetMovies).Methods("GET")
	router.HandleFunc("/addrating", AddRating).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// GetClient connects to MongoDB
func GetClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI(connect)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

type Movie struct {
	Name    string  `json:"name"`
	Rating  float64 `json:"rating"`
	RatedBy float64 `json:"ratedBy"`
}

type User struct {
	UserName string     `json:"userName"`
	Rated    []Rated    `json:"rated,omitempty"`
	Comments []Comments `json:"comments,omitempty"`
}

type Rated struct {
	Movie  string  `json:"movie"`
	Rating float64 `json:"rating"`
}

type Comments struct {
	Movie   string `json:"movie"`
	Comment string `json:"comments"`
}

type RatingRequest struct {
	UserName string  `json:"userName"`
	Movie    string  `json:"movie"`
	Rating   float64 `json:"rating"`
}

// GetMovies returns for the movies searched or all movies
func GetMovies(w http.ResponseWriter, r *http.Request) {
	var movies []*Movie

	collection := client.Database("Movies").Collection("movies")
	query := primitive.M{
		"name": primitive.Regex{
			Pattern: mux.Vars(r)["name"],
			Options: "i",
		},
	}
	cur, err := collection.Find(context.TODO(), query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
		return
	}

	for cur.Next(context.TODO()) {
		var movie Movie
		err = cur.Decode(&movie)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("")
		}
		movies = append(movies, &movie)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(movies) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("No movies found")
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(movies)
	}
}

func AddRating(w http.ResponseWriter, r *http.Request) {
	var user User

	collection := client.Database("Movies").Collection("users")

	var reqBody RatingRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}

	query := primitive.M{
		"username": reqBody.UserName,
	}
	cur := collection.FindOne(context.TODO(), query)
	err = cur.Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not Found")
		return
	}

	rating := Rated{
		Movie:  reqBody.Movie,
		Rating: reqBody.Rating,
	}
	user.Rated = append(user.Rated, rating)

	_, err = collection.UpdateOne(
		context.TODO(),
		primitive.M{
			"username": reqBody.UserName,
		},
		primitive.D{
			{"$set", primitive.D{{"rated", user.Rated}}},
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
	}

	collection = client.Database("Movies").Collection("movies")
	var movie Movie
	query = primitive.M{
		"name": reqBody.Movie,
	}
	cur = collection.FindOne(context.TODO(), query)
	err = cur.Decode(&movie)

	newRating := ((movie.Rating*movie.RatedBy)+reqBody.Rating)/movie.RatedBy + 1

	newRating = math.Round(newRating*100) / 100

	fmt.Println(newRating)
	fmt.Println(movie)

	_, err = collection.UpdateOne(
		context.TODO(),
		primitive.M{
			"name": reqBody.Movie,
		},
		primitive.D{
			{"$set", primitive.D{{"rating", newRating}}},
			{"$set", primitive.D{{"ratedBy", movie.RatedBy + 1}}},
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("")
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Rating updated successfully")
	}
}
