package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


var db *pgxpool.Pool

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Todo struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	var err error
	
	// Check environment variables are set
	if os.Getenv("SUPABASE_API_KEY") == "" || os.Getenv("SUPABASE_PROJECT_REF") == "" {
		log.Fatal("SUPABASE_API_KEY and SUPABASE_PROJECT_REF environment variables must be set")
	}
	
	// Initialize PostgreSQL connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err = pgxpool.New(context.Background(), dbURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to database directly: %v", err)
		} else {
			defer db.Close()
			log.Println("PostgreSQL connection established")
		}
	}

	// Setup routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/todos", todosHandler)
	
	// Enable CORS for frontend
	http.HandleFunc("/", corsMiddleware(http.NotFoundHandler()))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if user exists and verify password
	var user User
	err := db.QueryRow(context.Background(), "SELECT id, email, password_hash FROM todousers WHERE email = $1", loginReq.Email).
		Scan(&user.ID, &user.Email, &user.PasswordHash)
	
	if err != nil {
		log.Printf("Database query error: %v", err)
		if err == pgx.ErrNoRows {
			// Create new user if doesn't exist
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(loginReq.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}

			err = db.QueryRow(context.Background(), "INSERT INTO todousers (email, password_hash, created_at) VALUES ($1, $2, $3) RETURNING id",
				loginReq.Email, string(hashedPassword), time.Now()).Scan(&user.ID)
			if err != nil {
				log.Printf("Failed to create user: %v", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
			user.Email = loginReq.Email
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// Verify password for existing user
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginReq.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	// Create session
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)
	
	_, err = db.Exec(context.Background(), "INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)",
		sessionID, user.ID, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := struct {
		SessionID string `json:"session_id"`
		UserID    int    `json:"user_id"`
		Email     string `json:"email"`
	}{
		SessionID: sessionID,
		UserID:    user.ID,
		Email:     user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate request
	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	var userID int
	err := db.QueryRow(context.Background(), "SELECT user_id FROM sessions WHERE id = $1 AND expires_at > $2",
		sessionID, time.Now()).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		getTodos(w, r, userID)
	case "POST":
		createTodo(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTodos(w http.ResponseWriter, _ *http.Request, userID int) {
	rows, err := db.Query(context.Background(), "SELECT id, title, completed, created_at FROM todos WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		todo.UserID = userID
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request, userID int) {
	var todoReq struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&todoReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if todoReq.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	var todo Todo
	err := db.QueryRow(context.Background(), "INSERT INTO todos (user_id, title, completed, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, todoReq.Title, false, time.Now()).Scan(&todo.ID)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	todo.UserID = userID
	todo.Title = todoReq.Title
	todo.Completed = false
	todo.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}