package main

import (
	"fmt"
	"net/http"
	"os"

	"forum/database"
	"forum/handlers"
)

func main() {
	db, err := database.InitializeDB("./my_database.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	Db := &handlers.Handeldb{DB: db}

	_ = []string{"posts", "likes_dislikes", "sessions", "comments", "comment_likes"}

	http.HandleFunc("/", Db.HomePage)
	http.HandleFunc("/logup", Db.RegisterPage)
	http.HandleFunc("/login", Db.LoginPage)
	http.HandleFunc("/create-post", Db.CreatePostPage)
	http.HandleFunc("/like-post", Db.LikePost)
	http.HandleFunc("/logout", Db.Logout)
	http.HandleFunc("/profile", Db.Profile)
	http.HandleFunc("/posts", Db.AddPosts)
	http.HandleFunc("/fetch-posts", Db.FetchPosts)
	http.HandleFunc("/comments", Db.Addcomment)

	css := http.StripPrefix("/css/", http.FileServer(http.Dir("./css")))
	http.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		_, err := os.ReadFile("." + r.URL.Path)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		css.ServeHTTP(w, r)
	})
	port := ":8080"
	fmt.Printf("http://localhost%s\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

/* package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type handeldb struct {
	DB *sql.DB
}

func main() {
	db, err := sql.Open("sqlite3", "./my_database.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	Db := &handeldb{DB: db}
	Db.createTable()

	http.HandleFunc("/", Db.homePage)
	http.HandleFunc("/logup", Db.registerPage)
	http.HandleFunc("/login", Db.loginPage)
	http.HandleFunc("/create-post", Db.createPostPage)
	http.HandleFunc("/like-post", Db.likePost)
	http.HandleFunc("/logout", Db.logout)
	http.HandleFunc("/profile", Db.profile)
	http.HandleFunc("/posts", Db.addposts)
	http.HandleFunc("/fetch-posts", Db.fetchPosts)
	css := http.StripPrefix("/css/", http.FileServer(http.Dir("./css")))
	http.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		_, err := os.ReadFile("." + r.URL.Path)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		css.ServeHTTP(w, r)
	})

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func (db *handeldb) fetchPosts(w http.ResponseWriter, r *http.Request) {
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

	// Query database for posts with pagination
	query := "SELECT id, title, content FROM posts ORDER BY id DESC LIMIT ? OFFSET ?"
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (db *handeldb) profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
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
	var posts []Post
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
	for _, post := range posts {
		fmt.Fprintf(w, "post id: %d\n\n titel: %s\n\n content: %s\n", post.ID, post.Title, post.Content)
	}
}

func (db *handeldb) addposts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Session Token Not Found", http.StatusUnauthorized)
		return
	}
	usrId := ""
	getUserId := "SELECT user_id FROM sessions WHERE token = ?"
	err = db.DB.QueryRow(getUserId, cookie.Value).Scan(&usrId)
	if err != nil {
		fmt.Println(err)
	}
	query := "INSERT INTO posts (user_id, title, content, category_id) VALUES (?, ?, ?, ?)"

	r.ParseForm()
	title := r.FormValue("title")
	if _, err := db.DB.Exec(query, usrId, title, "-", 1); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (db *handeldb) createTable() {
	sqlTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		category_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS likes_dislikes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER,
		is_like BOOLEAN NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(post_id) REFERENCES posts(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`
	_, err := db.DB.Exec(sqlTable)
	if err != nil {
		fmt.Println(err)
	}
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (db *handeldb) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		db.DB.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (db *handeldb) registerPage(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "invalid email", http.StatusBadRequest)
			return
		}

		if len(password) < 10 {
			http.Error(w, "Password is too short", http.StatusBadRequest)
			return
		}

		exists := db.UseserExist(email)
		if exists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		if strings.Contains(username, " ") {
			http.Error(w, "invalid name", http.StatusForbidden)
			return
		}
		_, err := db.DB.Exec(username)
		if err == nil {
			fmt.Println(err)
			http.Error(w, "sql injection detected", http.StatusForbidden)
			return
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err = db.DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`, username, email, hashedPassword)
		if err != nil {
			http.Error(w, "Unable to register user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	tmp, _ := template.ParseFiles("./templates/logup.html")
	tmp.Execute(w, nil)
}

func (db *handeldb) loginPage(w http.ResponseWriter, r *http.Request) {
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

type Data struct {
	Logdin bool
	Name   string
	Posts  []Post
}

type Post struct {
	ID      int
	Title   string
	Content string
}

func (db *handeldb) homePage(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("./templates/home.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", 404)
		return
	}

	cookie, err := r.Cookie("session_token")
	loggedIn := true
	name := ""
	if err != nil {
		loggedIn = false
	} else {
		var userID int
		err := db.DB.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&userID)
		if err != nil {
			loggedIn = false
		} else {
			query := "SELECT username FROM users WHERE id = ?"
			err = db.DB.QueryRow(query, userID).Scan(&name)
		}
	}

	query := "SELECT id, title, content FROM posts ORDER BY id DESC LIMIT 20"
	row, err := db.DB.Query(query)
	if err != nil {
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}

	defer row.Close()
	var posts []Post
	for row.Next() {
		var id int
		var title, content string
		if err := row.Scan(&id, &title, &content); err != nil {
			return
		}
		posts = append(posts, Post{ID: id, Title: title, Content: content})
	}

	data := Data{
		Logdin: loggedIn,
		Name:   name,
		Posts:  posts,
	}
	tmp.Execute(w, data)
}

func (db *handeldb) createPostPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")

		var userID int
		err := db.DB.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&userID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_, err = db.DB.Exec(`INSERT INTO posts (user_id, title, content, category_id) VALUES (?, ?, ?, ?)`, userID, title, content, 1)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	tmp, _ := template.ParseFiles("./templates/create_post.html")
	tmp.Execute(w, nil)
}

/* like posts handler */
/*
func (db *handeldb) likePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID := r.FormValue("post_id")

	var userID int
	err = db.DB.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	query := "SELECT id FROM likes_dislikes WHERE user_id = ? AND post_id = ?"
	var likeID int
	err = db.DB.QueryRow(query, userID, postID).Scan(&likeID)

	if err != nil {
		if err == sql.ErrNoRows {
			// No like exists, add the like
			insertQuery := "INSERT INTO likes_dislikes (user_id, post_id, is_like) VALUES (?, ?, true)"
			_, insertErr := db.DB.Exec(insertQuery, userID, postID)
			if insertErr != nil {
				fmt.Println("failed to add like:", insertErr)
			}
		} else {
			fmt.Println("error checking like:", err)
		}
	} else {
		// Like exists, remove it
		fmt.Println(likeID)
		deleteQuery := "DELETE FROM likes_dislikes WHERE id = ?"
		_, deleteErr := db.DB.Exec(deleteQuery, likeID)
		if deleteErr != nil {
			fmt.Println("failed to remove like:", deleteErr)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Post liked successfully"}`))
}

func (db *handeldb) UseserExist(email string) bool {
	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", email).Scan(&exists); err != nil {
		fmt.Println(err)
	}
	return exists
}
*/
