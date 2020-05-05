package main

import (
	"context"
	"math"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResponseMovies describes the structure of response for Movies API
type ResponseMovies struct {
	Status int      `json:"status"`
	Body   []*Movie `json:"body"`
}

// Movie describes the structure of details for movies
type Movie struct {
	Name     string    `json:"name"`
	Rating   float64   `json:"rating"`
	RatedBy  float64   `json:"ratedBy"`
	Comments []Comment `json:"comments"`
}

// User described the structure of User data
type User struct {
	UserName string  `json:"userName"`
	Rated    []Rated `json:"rated,omitempty"`
}

// Rated describes the structure of Rated data
type Rated struct {
	Movie  string  `json:"movie"`
	Rating float64 `json:"rating"`
}

// ResponseUser describes the structure of Response of user data
type ResponseUser struct {
	Status int  `json:"status"`
	Body   User `json:"user"`
}

// Comment describes the structure of Comment data
type Comment struct {
	UserName string `json:"username"`
	Comment  string `json:"comment"`
}

// RatingRequest describes the structure of request for AddRating
type RatingRequest struct {
	UserName string  `json:"userName"`
	Movie    string  `json:"movie"`
	Rating   float64 `json:"rating"`
}

// ResponseRating describes the structure of Response sent for AddRating
type ResponseRating struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// RequestComment describes the structure of Request data of AddComment
type RequestComment struct {
	Movie   string  `json:"movieName"`
	Comment Comment `json:"comment"`
}

// ResponseComment describes the structure of Response sent for AddComment
type ResponseComment struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// GetMovies returns for the movies searched or all movies
func GetMovies(movieName string) ResponseMovies {
	var movies []*Movie
	var response ResponseMovies

	collection := client.Database("Movies").Collection("movies")
	query := primitive.M{
		"name": primitive.Regex{
			Pattern: movieName,
			Options: "i",
		},
	}
	cur, err := collection.Find(context.TODO(), query)
	if err != nil {
		return response
	}

	for cur.Next(context.TODO()) {
		var movie Movie
		err = cur.Decode(&movie)
		if err != nil {
			return response
		}
		movies = append(movies, &movie)
	}

	response = ResponseMovies{
		Status: http.StatusOK,
		Body:   movies,
	}

	return response
}

// AddRating adds new rating and updates the previous rating accordingly
func AddRating(req RatingRequest) ResponseRating {
	userDetails := GetUser(req.UserName)

	if userDetails.Status != http.StatusOK {
		return ResponseRating{
			Status: http.StatusNotFound,
			Body:   "User not found",
		}
	}

	for _, value := range userDetails.Body.Rated {
		if value.Movie == req.Movie {
			return ResponseRating{
				Status: http.StatusOK,
				Body:   "Movie is already rated by user",
			}
		}
	}

	rating := Rated{
		Movie:  req.Movie,
		Rating: req.Rating,
	}
	userDetails.Body.Rated = append(userDetails.Body.Rated, rating)

	collection := client.Database("Movies").Collection("users")
	_, err := collection.UpdateMany(
		context.TODO(),
		primitive.M{
			"username": req.UserName,
		},
		primitive.D{
			{"$set", primitive.D{{"rated", userDetails.Body.Rated}}},
		},
	)

	if err != nil {
		return ResponseRating{
			Status: http.StatusNotFound,
			Body:   "User not found",
		}
	}

	getMovie := GetMovies(req.Movie)
	movieDetails := getMovie.Body[0]

	newRating := ((movieDetails.Rating * movieDetails.RatedBy) + req.Rating) / (movieDetails.RatedBy + 1)

	newRating = math.Round(newRating*100) / 100

	collection = client.Database("Movies").Collection("movies")
	_, err = collection.UpdateOne(
		context.TODO(),
		primitive.M{
			"name": req.Movie,
		},
		primitive.D{
			{"$set", primitive.D{{"rating", newRating}}},
			{"$set", primitive.D{{"ratedBy", movieDetails.RatedBy + 1}}},
		},
	)

	if err != nil {
		return ResponseRating{
			Status: http.StatusInternalServerError,
			Body:   "Failed to get movie",
		}
	}

	return ResponseRating{
		Status: http.StatusOK,
		Body:   "Rating updated successfully",
	}
}

// AddComments adds new comment made by the user
func AddComments(req RequestComment) ResponseComment {
	getMovie := GetMovies(req.Movie)
	movieDetails := getMovie.Body[0]

	movieDetails.Comments = append(movieDetails.Comments, req.Comment)

	collection := client.Database("Movies").Collection("movies")
	_, err := collection.UpdateOne(
		context.TODO(),
		primitive.M{
			"name": req.Movie,
		},
		primitive.D{
			{"$set", primitive.D{{"comments", movieDetails.Comments}}},
		},
	)

	if err != nil {
		return ResponseComment{
			Status: http.StatusInternalServerError,
			Body:   "Failed to add comment",
		}
	}

	return ResponseComment{
		Status: http.StatusOK,
		Body:   "Successfully added comment",
	}
}

// GetUser returns the user details with the movies rated by user
func GetUser(username string) ResponseUser {
	var user User

	collection := client.Database("Movies").Collection("users")

	query := primitive.M{
		"username": username,
	}
	cur := collection.FindOne(context.TODO(), query)
	err := cur.Decode(&user)

	if err != nil {
		return ResponseUser{
			Status: http.StatusInternalServerError,
			Body:   user,
		}
	}

	return ResponseUser{
		Status: http.StatusOK,
		Body:   user,
	}
}
