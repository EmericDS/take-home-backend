package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Document represents a file uploaded to the service
// with its metadata stored in the database
type Document struct {
	ID         string    `json:"id"`         // Unique identifier for the document
	Name       string    `json:"name"`       // Original filename
	URL        string    `json:"url"`        // URL to download the document
	UploadedAt time.Time `json:"uploaded_at"` // Timestamp of the upload
}

// Directory where uploaded files are stored
const uploadDir = "/app/uploads"

// Global database connection
var db *sql.DB

// HTML template for the homepage with upload form - do not modify
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Test Upload</title>
</head>
<body>
    <h2>Upload de fichier</h2>
    <form action="http://localhost:8080/upload" method="post" enctype="multipart/form-data">
        <input type="file" name="file">
        <input type="submit" value="send">
    </form>
</body>
</html>`

// init initializes the application by setting up the database connection
// and creating required directory structures
func init() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Wait for PostgreSQL to start
	log.Println("Waiting for PostgreSQL to start...")
	time.Sleep(5 * time.Second)

	// Get database connection string from environment variable or use default
	connStr := os.Getenv("POSTGRES_DSN")
	if connStr == "" {
		connStr = "postgres://upload-service:password@postgres:5432/main?sslmode=disable"
	}
	
	log.Printf("Attempting to connect to PostgreSQL with: %s", connStr)
	
	// Multiple connection attempts for better reliability
	maxRetries := 5
	var err error
	for i := 0; i < maxRetries; i++ {
		// Open database connection
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Error opening connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}
		
		// Verify connection is working
		err = db.Ping()
		if err == nil {
			log.Printf("Successfully connected to PostgreSQL")
			break
		}
		
		log.Printf("Error pinging database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	
	// If we couldn't connect after all retries, exit
	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// Create documents table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// uploadHandler processes file upload requests
// It saves the file to disk and stores metadata in the database
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique ID for the file
	id := uuid.New().String()
	fileName := header.Filename
	filePath := filepath.Join(uploadDir, id)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file content to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Store file metadata in the database
	_, err = db.Exec("INSERT INTO documents (id, name, uploaded_at) VALUES ($1, $2, $3)",
		id, fileName, time.Now())
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// Return success status
	w.WriteHeader(http.StatusCreated)
}

// listHandler returns a JSON list of all uploaded documents
func listHandler(w http.ResponseWriter, r *http.Request) {
	// Query all documents from the database
	rows, err := db.Query("SELECT id, name, uploaded_at FROM documents")
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate through rows and build document list
	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Name, &doc.UploadedAt)
		if err != nil {
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		// Build download URL for each document
		doc.URL = fmt.Sprintf("http://localhost:8080/dl/%s", doc.ID)
		documents = append(documents, doc)
	}

	// Return document list as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// downloadHandler serves file downloads by ID
// It retrieves file metadata from the database and serves the file
// with appropriate headers to force download rather than in-browser display
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Extract document ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Retrieve the filename from the database
	var fileName string
	err := db.QueryRow("SELECT name FROM documents WHERE id = $1", id).Scan(&fileName)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Construct file path
	filePath := filepath.Join(uploadDir, id)
	
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	
	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	
	// Set headers to force download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	
	// Send the file content to the client
	io.Copy(w, file)
}

// indexHandler serves the homepage with the upload form
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlTemplate))
}

// main sets up the HTTP routes and starts the server
func main() {
	// Create a new router using gorilla/mux
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/", indexHandler).Methods("GET")            // Homepage with upload form
	router.HandleFunc("/upload", uploadHandler).Methods("POST")    // File upload endpoint
	router.HandleFunc("/documents", listHandler).Methods("GET")    // List all uploaded documents
	router.HandleFunc("/dl/{id}", downloadHandler).Methods("GET")  // Download a document by ID

	// Start the HTTP server
	log.Printf("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
