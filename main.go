package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Task struct {
	ID   int
	Text string
}

var tasks []Task
var idCounter int
var mu sync.Mutex

func main() {
	http.HandleFunc("/add", addTask)
	http.HandleFunc("/remove", removeTask)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	taskText := r.URL.Query().Get("task")
	if taskText == "" {
		http.Error(w, "Task text is required", http.StatusBadRequest)
		return
	}

	idCounter++
	task := Task{
		ID:   idCounter,
		Text: taskText,
	}

	tasks = append(tasks, task)
	fmt.Fprintf(w, "<li data-task-id='%d'>%s <a href='#' hx-post='/remove?id=%d' hx-target='closest li' hx-swap='outerHTML'>[Remove]</a></li>", task.ID, task.Text, task.ID)
}

func removeTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	taskID := r.URL.Query().Get("id")
	for index, task := range tasks {
		if fmt.Sprintf("%d", task.ID) == taskID {
			tasks = append(tasks[:index], tasks[index+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusOK)
}
