package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	connect = "mongodb+srv://neel:testpass1994@movies-3mimb.mongodb.net/test?retryWrites=true&w=majority" // Change to local mongo setup
	client  = GetClient()
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/movies", handleMovies).Methods("GET")
	router.HandleFunc("/movies/{name}", handleMovies).Methods("GET")
	router.HandleFunc("/addrating", handleAddRating).Methods("POST")
	router.HandleFunc("/addcomment", handleAddComments).Methods("POST")
	router.HandleFunc("/userdetails", handleUserDetails).Methods("POST")
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

func handleMovies(w http.ResponseWriter, r *http.Request) {
	response := GetMovies(mux.Vars(r)["name"])
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if response.Body == nil {
		json.NewEncoder(w).Encode("No movies found")
	} else {
		json.NewEncoder(w).Encode(response)
	}
}

func handleAddRating(w http.ResponseWriter, r *http.Request) {
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

	response := AddRating(reqBody)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	json.NewEncoder(w).Encode(response)
}

func handleAddComments(w http.ResponseWriter, r *http.Request) {
	var requestComment RequestComment
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}
	err = json.Unmarshal(body, &requestComment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}

	response := AddComments(requestComment)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	json.NewEncoder(w).Encode(response)
}

func handleUserDetails(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Invalid request")
		return
	}

	response := GetUser(user.UserName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	json.NewEncoder(w).Encode(response)
}
