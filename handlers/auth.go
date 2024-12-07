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

// User registration
func (db *Handeldb) RegisterPage(w http.ResponseWriter, r *http.Request) {
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

	tmp, _ := template.ParseFiles("./templates/logup.html")
	tmp.Execute(w, nil)
}

// User login
func (db *Handeldb) LoginPage(w http.ResponseWriter, r *http.Request) {
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
			fmt.Println(token)
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

	tmp, _ := template.ParseFiles("./templates/login.html")
	tmp.Execute(w, nil)
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
	name := r.URL.Query().Get("name")
	limit := 20

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	var posts []Post

	if name == "home" || name == "" {
		query := "SELECT id, user_id, title, content FROM posts ORDER BY id DESC LIMIT ? OFFSET ?"
		rows, err := db.DB.Query(query, limit, offset)
		if err != nil {
			http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var post Post
			var user_id int
			if err := rows.Scan(&post.ID, &user_id, &post.Title, &post.Content); err != nil {
				http.Error(w, "Error scanning posts", http.StatusInternalServerError)
				return
			}
			getuserName := "SELECT username FROM users WHERE id = ?"
			db.DB.QueryRow(getuserName, user_id).Scan(&post.UserName)
			posts = append(posts, post)
		}

	} else if name == "profile" {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

		var id int
		err = db.DB.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&id)
		if err != nil {
			fmt.Println(err.Error())
		}

		query := "SELECT id, title, content FROM posts WHERE user_id = ?"
		row, err := db.DB.Query(query, id)
		if err != nil {
			http.Error(w, "couldn't retreve data from database", http.StatusInternalServerError)
			return
		}

		defer row.Close()
		for row.Next() {
			var post_id int
			var title, content string
			if err := row.Scan(&post_id, &title, &content); err != nil {
				fmt.Println(err)
				http.Error(w, "Error during iteration over rows", http.StatusInternalServerError)
				return
			}
			posts = append(posts, Post{ID: post_id, Title: title, Content: content})
		}

		if err = row.Err(); err != nil {
			http.Error(w, "Error during iteration over rows", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (db *Handeldb) Profile(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	tmp, err := template.ParseGlob("templates/*.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Println("profile")

	err = tmp.ExecuteTemplate(w, "profile.html", nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
