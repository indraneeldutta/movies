# Movies

Below are the REST APIs created for Movies
1. `/movies` - shows all movies added and its details
2. `/movies/{moviename}` - shows movie which matches the name and its details
3. `/addrating` - Adds rating to a movie
4. `/addcomment` - Adds comment to movie
5. `/userdetails` - shows the movies rated by the user along with details

## Setup
Clone this repo to your GOPATH

The program uses MongoDB as its database. The DB Dumps can be found in `dbDump` folder in JSON format. Consists of two collections `movies` and `users` with prepopulated data. Feel free to edit the data before import.

1. Change your connection string in `main.go` line 16 to your mongoDB connection. (Assumption #7)

2. Make sure to install dependencies to run the program.
- Run `make deps`

OR
- `go get -u go.mongodb.org/mongo-driver/mongo` and `go get -u github.com/gorilla/mux`  
3. run `make build` OR `go build -o bin/main`
4. run `make run` OR `./bin/main`

## Usage
The program serves default to port 8080. This can be changed on `main.go` line 27

Import `PostmanCollection.json` to your Postman

Execute the desired API by changing the values in Body for `POST` requests

## Assumptions

1. No ACL. Therefore the user data has to be added manually in DB. 
2. No API to add movies. New Movies can be added manually if required. 
3. For now username is sent manually from the postman, by right should be auto populated through session.
4. The user schema just has username for user details. but can accommodate other values related to user.
5. The program is using username as identifier. But can be improved to having unique ID to every user which is then supplied by session for details retrival.
6. One user can rate a movie once. Can be improved to update the rating previously rated.
7. The connection string can be moved to secure location. For now its left in `main.go` file for easy change and test.