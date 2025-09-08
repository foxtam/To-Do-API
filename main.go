package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Task struct {
	ID    int     `json:"id"`
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

var tasks = make(map[int]*Task)
var currentID int

func main() {
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/tasks/", handleTasksByID)

	fmt.Println("Сервер запущен...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handleTasksByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idStr == "" {
		http.Error(w, "Empty id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Wrong id", http.StatusBadRequest)
		return
	}

	task, ok := tasks[id]
	if !ok {
		http.Error(w, "No such task id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var updatedTask Task
		err := json.NewDecoder(r.Body).Decode(&updatedTask)
		if err != nil {
			http.Error(w, "Bad json", http.StatusBadRequest)
			return
		}

		if updatedTask.Title != nil {
			if *updatedTask.Title == "" {
				http.Error(w, "Title is required", http.StatusBadRequest)
				return
			}
			task.Title = updatedTask.Title
		}
		if updatedTask.Done != nil {
			task.Done = updatedTask.Done
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(task)

	case http.MethodDelete:
		delete(tasks, id)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tasks)

	case http.MethodPost:
		var newTask Task
		err := json.NewDecoder(r.Body).Decode(&newTask)
		if err != nil {
			http.Error(w, "Bad json", http.StatusBadRequest)
			return
		}

		currentID++
		newTask.ID = currentID
		if newTask.Title == nil || *newTask.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}
		if newTask.Done == nil {
			newTask.Done = new(bool)
		}
		tasks[currentID] = &newTask

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newTask)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
