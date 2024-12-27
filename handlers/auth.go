package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Handeldb struct {
	DB *sql.DB
}


func (db *Handeldb) RegisterPage(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_token")
	if err == nil {
		fmt.Println("cookie does exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		if email == "" || username == "" || password == "" {
			http.Error(w, "Please provide all details", http.StatusBadRequest)
			return
		}

		pattern := `[a-zA-Z0-9]{5,}@`
		reg := regexp.MustCompile(pattern)

		if !reg.MatchString(email) {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}

		if len(password) < 10 {
			http.Error(w, "Password is too short", http.StatusBadRequest)
			return
		}

		exists := db.UserExists(email)
		if exists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		if strings.Contains(username, " ") {
			http.Error(w, "Invalid name", http.StatusForbidden)
			return
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`, username, email, hashedPassword)
		if err != nil {
			http.Error(w, "Unable to register user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmp, err := template.ParseGlob("./templates/*.html")
	if err != nil {
		fmt.Println(err)
	}
	err = tmp.ExecuteTemplate(w, "logup.html", nil)
	if err != nil {
		fmt.Println(err)
	}
}

// User login
func (db *Handeldb) LoginPage(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_token")
	if err == nil {
		fmt.Println("cookie does exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var storedHashedPassword string
		var userId int

		err := db.DB.QueryRow(`SELECT id, password FROM users WHERE email = ?`, email).Scan(&userId, &storedHashedPassword)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		var token string
		err = db.DB.QueryRow("SELECT token FROM sessions WHERE user_id = ?", userId).Scan(&token)
		if err == nil {
			sessionCookie := &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, sessionCookie)

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		fmt.Println("rTCGGkNLQsmj6r3be9UNBaeKiEI-pQM6OcIT7d0zzws= ")
		if err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		sessionToken := generateSessionToken()
		_, err = db.DB.Exec(`INSERT INTO sessions (token, user_id) VALUES (?, ?)`, sessionToken, userId)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		fmt.Println(sessionToken)
		sessionCookie := &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, sessionCookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmp, _ := template.ParseGlob("./templates/*.html")
	tmp.ExecuteTemplate(w, "login.html", nil)
}

// Generate session token
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Check if user exists
func (db *Handeldb) UserExists(email string) bool {
	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", email).Scan(&exists); err != nil {
		fmt.Println(err)
	}
	return exists
}

func (db *Handeldb) FetchPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	limit := 20

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	var posts []Post

	query := "SELECT id, (SELECT username FROM users WHERE id = user_id), title, content FROM posts ORDER BY id DESC LIMIT ? OFFSET ?"
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		// var user_id int
		if err := rows.Scan(&post.ID, &post.UserName, &post.Title, &post.Content); err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (db *Handeldb) ProfileData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var id int
	err = db.DB.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&id)
	if err != nil {
		fmt.Println(err.Error())
	}

	query := "SELECT id,(SELECT username FROM users WHERE id = user_id), title, content FROM posts WHERE user_id = ?"
	row, err := db.DB.Query(query, id)
	if err != nil {
		http.Error(w, "couldn't retreve data from database", http.StatusInternalServerError)
		return
	}
	var posts []Post
	defer row.Close()
	for row.Next() {
		var post Post
		if err := row.Scan(&post.ID, &post.UserName, &post.Title, &post.Content); err != nil {
			fmt.Println(err)
			http.Error(w, "Error during iteration over rows", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	if err = row.Err(); err != nil {
		http.Error(w, "Error during iteration over rows", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
