package main

import (
	"context"
	"math"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResponseMovies struct {
	Status int      `json:"status"`
	Body   []*Movie `json:"body"`
}

type Movie struct {
	Name     string    `json:"name"`
	Rating   float64   `json:"rating"`
	RatedBy  float64   `json:"ratedBy"`
	Comments []Comment `json:"comments"`
}

type User struct {
	UserName string  `json:"userName"`
	Rated    []Rated `json:"rated,omitempty"`
}

type Rated struct {
	Movie  string  `json:"movie"`
	Rating float64 `json:"rating"`
}

type Comment struct {
	UserName string `json:"username"`
	Comment  string `json:"comment"`
}

type RatingRequest struct {
	UserName string  `json:"userName"`
	Movie    string  `json:"movie"`
	Rating   float64 `json:"rating"`
}

type ResponseRating struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

type RequestComment struct {
	Movie   string  `json:"movieName"`
	Comment Comment `json:"comment"`
}

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

func AddRating(req RatingRequest) ResponseRating {
	var user User

	collection := client.Database("Movies").Collection("users")

	query := primitive.M{
		"username": req.UserName,
	}
	cur := collection.FindOne(context.TODO(), query)
	err := cur.Decode(&user)
	for _, value := range user.Rated {
		if value.Movie == req.Movie {
			return ResponseRating{
				Status: http.StatusOK,
				Body:   "Movie is already rated by user",
			}
		}
	}
	if err != nil {
		return ResponseRating{
			Status: http.StatusNotFound,
			Body:   "User not found",
		}
	}

	rating := Rated{
		Movie:  req.Movie,
		Rating: req.Rating,
	}
	user.Rated = append(user.Rated, rating)

	_, err = collection.UpdateOne(
		context.TODO(),
		primitive.M{
			"username": req.UserName,
		},
		primitive.D{
			{"$set", primitive.D{{"rated", user.Rated}}},
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

	newRating := ((movieDetails.Rating*movieDetails.RatedBy)+req.Rating)/movieDetails.RatedBy + 1

	newRating = math.Round(newRating*100) / 100

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
