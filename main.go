package main

import (
	"encoding/json" //This package is for encoding and decoding the JSON data
	"fmt"           //This package is used for formatting input/output function like printf/scanf
	"net/http"      //This provides HTTP client and server implementations
	"sync"          //This package has basic synchronization primitives to handle concurrent prgramming
)

// Below is the User struct description of  the User model
type User struct {
	ID       int    `json:"id"` //User Id which is identical for each user
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"` // Plain  password in text
}

// In-memory  store (slice)
var users []User
var mu sync.Mutex //  mutual exclusion lock to allow only one goroutine can access a critical section at a time.
var nextID = 1    // For generating the user ID

// Register a new user (Create)
func registerUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var newUser User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the new user to the slice
	mu.Lock() // used to lock a mutex (mutual exclusion lock)and allows single goroutine can access a shared resource at a time.
	newUser.ID = nextID
	users = append(users, newUser)
	nextID++
	mu.Unlock()

	// Respond with success after registering the user
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// Login user (Read)
func loginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var loginDetails struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginDetails); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the user in the slice
	mu.Lock()
	var foundUser *User
	for i := range users {
		if users[i].Username == loginDetails.Username {
			foundUser = &users[i]
			break
		}
	}
	mu.Unlock()

	// If user is not found
	if foundUser == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Compare the stored password with the provided one
	if foundUser.Password != loginDetails.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Successful login msg
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(foundUser)
}

// Update user information (Update)
func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var updatedUser User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find and update the user in the slice
	mu.Lock()
	var foundUser *User
	for i := range users {
		if users[i].ID == updatedUser.ID {
			foundUser = &users[i]
			break
		}
	}
	mu.Unlock()

	// If user is not found
	if foundUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update the user info
	mu.Lock()
	if updatedUser.Email != "" {
		foundUser.Email = updatedUser.Email
	}
	if updatedUser.Password != "" {
		foundUser.Password = updatedUser.Password // Store the updated passwd
	}
	mu.Unlock()

	// Respond with updated user info
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(foundUser)
}

// Delete user by ID (Delete)
func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var idToDelete struct {
		ID int `json:"id"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&idToDelete); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find and delete the user from the slice
	mu.Lock()
	var deleted bool
	for i := range users {
		if users[i].ID == idToDelete.ID {
			// Remove user from the slice
			users = append(users[:i], users[i+1:]...)
			deleted = true
			break
		}
	}
	mu.Unlock()

	// Respond with success msg or failure msg
	if deleted {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "User is not found", http.StatusNotFound)
	}
}

// Setup routes for each operation
func setupRoutes() {
	http.HandleFunc("/register", registerUser)
	http.HandleFunc("/login", loginUser)
	http.HandleFunc("/update", updateUser)
	http.HandleFunc("/delete", deleteUser)
}

func main() {
	// Set up routes
	setupRoutes()

	// Start the server
	fmt.Println("Starting the server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
