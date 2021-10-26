package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ToDo struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty"`
	Completion   *bool              `json:"completion,omitempty" bson:"completion,omitempty"`
	CreationTime time.Time          `json:"creationTime,omitempty" bson:"creationTime,omitempty"`
}

var client *mongo.Client

func createTaskEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var todo ToDo
	_ = json.NewDecoder(r.Body).Decode(&todo)
	collection := client.Database("testgo").Collection("gotest")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, todo)
	json.NewEncoder(w).Encode(result)
}

func getTaskEndPoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var all_todo []ToDo
	collection := client.Database("testgo").Collection("gotest")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var temp_todo ToDo
		cursor.Decode(&temp_todo)
		all_todo = append(all_todo, temp_todo)
	}
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(all_todo)
}

func getOneTaskeEndPoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var todo ToDo
	collection := client.Database("testgo").Collection("gotest")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, ToDo{ID: id}).Decode(&todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(todo)
}

func deleteTaskEndPoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("testgo").Collection("gotest")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.DeleteOne(ctx, ToDo{ID: id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	fmt.Println(result.DeletedCount)
}

func updateTaskndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("testgo").Collection("gotest")
	update := bson.M{"$set": bson.M{"description": "s3il.com"}}

	result := collection.FindOneAndUpdate(ctx, ToDo{ID: id}, update)
	if result.Err() != nil {
		log.Printf("task update failed: %v\n", result.Err())
	}
	json.NewEncoder(w).Encode("Updated Successfully")
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	route := mux.NewRouter()
	route.HandleFunc("/task", createTaskEndpoint).Methods("POST")
	route.HandleFunc("/all_task", getTaskEndPoint).Methods("GET")
	route.HandleFunc("/task/{id}", getOneTaskeEndPoint).Methods("GET")
	route.HandleFunc("/task/{id}", deleteTaskEndPoint).Methods("DELETE")
	route.HandleFunc("/task/{id}", updateTaskndpoint).Methods("PUT")
	http.ListenAndServe(":12345", route)
}
