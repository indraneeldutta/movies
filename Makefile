deps:
	go get -u go.mongodb.org/mongo-driver/mongo
	go get -u github.com/gorilla/mux

build:
	go build -o bin/main

run:
	./bin/main