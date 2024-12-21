// main.go
package main

import (
    "database/sql"
    "encoding/json"
   "fmt"
    "log"
    "net/http"
//    "os"
    "time"
    "strconv"
    "analyze_study/config"
    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/mux"
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
)

// Structs for our models
type User struct {
    ID         int     `json:"id"`
    Fullname   string  `json:"fullname"`
    Email      string  `json:"email"`
    Category   string  `json:"category"`
    Commission float64 `json:"commission"`
    Password   string  `json:"password,omitempty"`
}

type Department struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Shortname string `json:"shortname"`
    Dean      string `json:"dean"`
    Secretary string `json:"secretary"`
}

type Group struct {
    ID            int    `json:"id"`
    Name          string `json:"name"`
    Year          int    `json:"year"`
    CountStudents int    `json:"count_students"`
}

type Analysis struct {
    ID           int     `json:"id"`
    Year         int     `json:"year"`
    Semester     int     `json:"semester"`
    Subject      string  `json:"subject"`
    GroupID      int     `json:"id_group"`
    DepartmentID int     `json:"id_department"`
    CountStud    int     `json:"count_stud"`
    Count5       int     `json:"count5"`
    Count4       int     `json:"count4"`
    Count3       int     `json:"count3"`
    Count2       int     `json:"count2"`
    UserID       int     `json:"id_user"`
    Overall      float64 `json:"overall"`
    Average      float64 `json:"average"`
}
type ErrorResponse struct {
    Error string `json:"error"`
}

type SuccessResponse struct {
    Message string `json:"message"`
}

var db *sql.DB
var jwtKey = []byte("your_secret_key")

// JWT claims struct
type Claims struct {
    UserID int `json:"user_id"`
    jwt.StandardClaims
}



// Authentication middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Verify user credentials from database
    var dbUser User
    var hashedPassword string
    err := db.QueryRow("SELECT id, password FROM users WHERE fullname = ?", user.Fullname).
        Scan(&dbUser.ID, &hashedPassword)

    if err != nil || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)) != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Create token
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: dbUser.ID,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "token": tokenString,
    })
}

// Example of a protected route handler (Users)
func getUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, fullname, category, commission FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Fullname, &user.Category, &user.Commission); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    result, err := db.Exec("INSERT INTO users (fullname, category, commission, password) VALUES (?, ?, ?, ?)",
        user.Fullname, user.Category, user.Commission, hashedPassword)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    user.ID = int(id)
    user.Password = "" // Don't return the password
    json.NewEncoder(w).Encode(user)
}

// getUser handles fetching a single user by ID
func getUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    var user User
    err = db.QueryRow(`
        SELECT id, fullname, category, commission 
        FROM users 
        WHERE id = ?`, id).
        Scan(&user.ID, &user.Fullname, &user.Category, &user.Commission)

    if err == sql.ErrNoRows {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
        return
    }

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// updateUser handles updating an existing user
func updateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Check if user exists
    var exists bool
    err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
        return
    }

    if !exists {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
        return
    }

    // Decode request body
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
        return
    }

    // Start a transaction
    tx, err := db.Begin()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
        return
    }
    defer tx.Rollback()

    // Update user information
    var updateQuery string
    var args []interface{}

    if user.Password != "" {
        // If password is provided, hash it and update everything
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to hash password"})
            return
        }

        updateQuery = `
            UPDATE users 
            SET fullname = ?, category = ?, commission = ?, password = ?
            WHERE id = ?`
        args = []interface{}{user.Fullname, user.Category, user.Commission, hashedPassword, id}
    } else {
        // If no password provided, update only other fields
        updateQuery = `
            UPDATE users 
            SET fullname = ?, category = ?, commission = ?
            WHERE id = ?`
        args = []interface{}{user.Fullname, user.Category, user.Commission, id}
    }

    result, err := tx.Exec(updateQuery, args...)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update user"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil || rowsAffected == 0 {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update user"})
        return
    }

    // Commit the transaction
    if err = tx.Commit(); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to commit transaction"})
        return
    }

    // Return updated user (without password)
    user.ID = id
    user.Password = ""
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// deleteUser handles deleting a user
func deleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Start a transaction
    tx, err := db.Begin()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
        return
    }
    defer tx.Rollback()

    // Check if user has any related records in analyses table
    var hasAnalyses bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM analyses WHERE id_user = ?)", id).Scan(&hasAnalyses)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
        return
    }

    if hasAnalyses {
        // Update analyses to set id_user to NULL where this user is referenced
        _, err = tx.Exec("UPDATE analyses SET id_user = NULL WHERE id_user = ?", id)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update related analyses"})
            return
        }
    }

    // Delete the user
    result, err := tx.Exec("DELETE FROM users WHERE id = ?", id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete user"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete user"})
        return
    }

    if rowsAffected == 0 {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
        return
    }

    // Commit the transaction
    if err = tx.Commit(); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to commit transaction"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(SuccessResponse{Message: "User deleted successfully"})
}

func main() {

    config, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    // Database connection
    
    db, err := sql.Open("mysql", config.Database.DSN())
    if err != nil {
        log.Fatal(err)
    }
    db.SetMaxOpenConns(config.Database.MaxOpenConns)
    db.SetMaxIdleConns(config.Database.MaxIdleConns)
    db.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
    
    defer db.Close()
// Use JWT secret from config
    jwtKey = []byte(config.JWT.SecretKey)

    // Router setup
    r := mux.NewRouter()

    // Auth endpoints
    r.HandleFunc("/api/login", loginHandler).Methods("POST")

    // Protected routes
    api := r.PathPrefix("/api").Subrouter()
    api.Use(authMiddleware)

    // User routes
    api.HandleFunc("/users", getUsers).Methods("GET")
    api.HandleFunc("/users", createUser).Methods("POST")
    api.HandleFunc("/users/{id}", getUser).Methods("GET")
    api.HandleFunc("/users/{id}", updateUser).Methods("PUT")
    api.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

    // Department routes
//    api.HandleFunc("/departments", getDepartments).Methods("GET")
//   api.HandleFunc("/departments", createDepartment).Methods("POST")
    //api.HandleFunc("/departments/{id}", getDepartment).Methods("GET")
//    api.HandleFunc("/departments/{id}", updateDepartment).Methods("PUT")
//    api.HandleFunc("/departments/{id}", deleteDepartment).Methods("DELETE")

    // Group routes
//    api.HandleFunc("/groups", getGroups).Methods("GET")
//    api.HandleFunc("/groups", createGroup).Methods("POST")
//    api.HandleFunc("/groups/{id}", getGroup).Methods("GET")
//    api.HandleFunc("/groups/{id}", updateGroup).Methods("PUT")
//    api.HandleFunc("/groups/{id}", deleteGroup).Methods("DELETE")

    // Analysis routes
//    api.HandleFunc("/analyses", getAnalyses).Methods("GET")
//    api.HandleFunc("/analyses", createAnalysis).Methods("POST")
//    api.HandleFunc("/analyses/{id}", getAnalysis).Methods("GET")
//    api.HandleFunc("/analyses/{id}", updateAnalysis).Methods("PUT")
//    api.HandleFunc("/analyses/{id}", deleteAnalysis).Methods("DELETE")

//fmt.Println("Server is listening on port 8081")
    //http.ListenAndServe(":8080", api)
  //  log.Fatal(http.ListenAndServe(":8081", r))
  serverAddr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
    log.Printf("Server starting on %s", serverAddr)
    log.Fatal(http.ListenAndServe(serverAddr, r))
}