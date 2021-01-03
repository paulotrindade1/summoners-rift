package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/segmentio/ksuid"
)

// Post ...
type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// PostResponse ...
type PostResponse struct {
	Type string `json:"type"`
	Data Post   `json:"data"`
}

var listPosts []Post

func posts(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		var post Post
		if err := json.Unmarshal(body, &post); err != nil {
			http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
			return
		}
		post.ID = generateID()
		listPosts = append(listPosts, post)

		postRequest, _ := json.Marshal(&PostResponse{
			Type: "PostCreated",
			Data: post,
		})
		sendPostToEventBus(postRequest)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Post Created!")

	} else if req.Method == "GET" {
		posts, err := json.Marshal(listPosts)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(posts))
	} else {
		fmt.Fprintln(w, "Method not implemented to /posts")
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func events(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		var postResponse PostResponse
		if err := json.Unmarshal(body, &postResponse); err != nil {
			http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
			return
		}
		fmt.Printf("Received event - %v\n", postResponse.Type)
	} else {
		fmt.Println("Method not implemented to /events")
		w.WriteHeader(http.StatusNotImplemented)
	}

}
func generateID() string {
	id := ksuid.New()
	return id.String()
}

func sendPostToEventBus(postRequest []byte) {
	url := "http://localhost:4005/events"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postRequest))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("Error sending to event bus")
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
}

func main() {
	http.HandleFunc("/posts", posts)
	http.HandleFunc("/events", events)
	http.ListenAndServe(":4000", nil)
}
