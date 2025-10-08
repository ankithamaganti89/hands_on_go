package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// User represents the structure of a user in the database
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

var db *sql.DB

func main() {
	// Database connection string
	// Format: "user:password@tcp(host:port)/dbname"
	// For Dockerized MySQL, host is typically "127.0.0.1" or "localhost" if running on the same machine
	// or the service name if running within a Docker Compose network.
	// Assuming a local Dockerized MySQL with default user/pass and database.
	dsn := "dockeruser:dockerpass@tcp(127.0.0.1:3306)/hands_on_go"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to MySQL database!")

	// Register the HTTP handler
	http.HandleFunc("/get-user", getUserHandler)

	// Start the HTTP server
	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// getUserHandler handles requests to the /get-user endpoint
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 500)
		return
	}

	// Get the 'id' parameter from the URL query
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing 'id' parameter", 500)
		return
	}

	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid 'id' parameter", 500)
		return
	}

	// Query the database for the user
	user := User{}
	err = db.QueryRow("SELECT id, first_name, last_name FROM users WHERE id = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", 500)
		} else {
			log.Printf("Error querying user: %v", err)
			http.Error(w, "Internal server error", 500)
		}
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode the user struct to JSON and write to the response body
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", 500)
	}
}
