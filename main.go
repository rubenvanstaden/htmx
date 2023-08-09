package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Message struct {
	ID      int
	Name    string
	Email   string
	Content string
}

var messages []Message
var idCounter int
var mu sync.Mutex

func main() {
	http.HandleFunc("/contact", contactHandler)
    http.HandleFunc("/latest-messages", latestMessagesHandler) // new endpoint
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    name := r.PostFormValue("name")
    email := r.PostFormValue("email")
    content := r.PostFormValue("content")

    if name == "" || email == "" || content == "" {
        http.Error(w, "All fields are required", http.StatusBadRequest)
        return
    }

	idCounter++
	message := Message{
		ID:      idCounter,
		Name:    name,
		Email:   email,
		Content: content,
	}

	messages = append(messages, message)

	fmt.Fprintf(w, "<div><strong>%s (%s)</strong>: %s</div>", message.Name, message.Email, message.Content)
}

func latestMessagesHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	for _, message := range messages {
		fmt.Fprintf(w, "<div><strong>%s (%s)</strong>: %s</div>", message.Name, message.Email, message.Content)
	}
}
