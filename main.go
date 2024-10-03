package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Task struct represents a single task
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending" or "completed"
}

var (
	tasks  = []Task{}
	nextID = 1
	mu     sync.Mutex
)

// Create a new task (POST /tasks)
func createTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var newTask Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newTask.ID = nextID
	nextID++

	tasks = append(tasks, newTask)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// Get all tasks (GET /tasks)
func getAllTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Get task by ID (GET /tasks/{id})
func getTaskByID(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	idStr := r.URL.Path[len("/tasks/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

// Update task (PUT /tasks/{id})
func updateTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	idStr := r.URL.Path[len("/tasks/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedTask Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			// Update task fields
			tasks[i].Title = updatedTask.Title
			tasks[i].Description = updatedTask.Description
			tasks[i].Status = updatedTask.Status

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

// Delete task (DELETE /tasks/{id})
func deleteTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	idStr := r.URL.Path[len("/tasks/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func main() {
	http.HandleFunc("/tasks", createTask)        // Create new task
	http.HandleFunc("/tasks", getAllTasks)       // Get all tasks
	http.HandleFunc("/tasks/", getTaskByID)      // Get task by ID
	http.HandleFunc("/tasks/", updateTask)       // Update task by ID
	http.HandleFunc("/tasks/", deleteTask)       // Delete task by ID

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
