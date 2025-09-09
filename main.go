package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID    int     `json:"id"`
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

var mu sync.RWMutex
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

	mu.RLock()
	taskPtr, ok := tasks[id]
	mu.RUnlock()
	if !ok {
		http.Error(w, "No such taskPtr id", http.StatusBadRequest)
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

		mu.Lock()
		if updatedTask.Title != nil {
			if *updatedTask.Title == "" {
				http.Error(w, "Title is required", http.StatusBadRequest)
				return
			}
			taskPtr.Title = updatedTask.Title
		}
		if updatedTask.Done != nil {
			taskPtr.Done = updatedTask.Done
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(taskPtr)

	case http.MethodDelete:
		mu.Lock()
		delete(tasks, id)
		mu.Unlock()

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		mu.RLock()
		_ = json.NewEncoder(w).Encode(tasks)
		mu.RUnlock()

	case http.MethodPost:
		var newTask Task
		err := json.NewDecoder(r.Body).Decode(&newTask)
		if err != nil {
			http.Error(w, "Bad json", http.StatusBadRequest)
			return
		}

		if newTask.Title == nil || *newTask.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}
		if newTask.Done == nil {
			newTask.Done = new(bool)
		}

		mu.Lock()
		currentID++
		newTask.ID = currentID
		tasks[currentID] = &newTask
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newTask)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
