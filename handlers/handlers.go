package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

func (db *Handeldb) AddPosts(w http.ResponseWriter, r *http.Request) {
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
	// file, handler, err := r.FormFile("filename")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// dest, err := os.Create("./comming_data/"+ handler.Filename)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("./comming_data/"+ handler.Filename)
	// defer dest.Close()
	// _, err = io.Copy(dest, file)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	title := r.FormValue("title")
	content := r.FormValue("content")
	if _, err := db.DB.Exec(query, usrId, title, strings.ReplaceAll(strings.TrimSpace(content), "\r\n", "<br>"), 1); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (db *Handeldb) Logout(w http.ResponseWriter, r *http.Request) {
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

type Data struct {
	Logdin bool
	Name   string
	Posts  []Post
}

type Post struct {
	ID      int
	UserName string
	Title   string
	Content string
}

func (db *Handeldb) HomePage(w http.ResponseWriter, r *http.Request) {
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
			if err != nil {
				fmt.Println("could not get user name", err)
			}
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

func (db *Handeldb) CreatePostPage(w http.ResponseWriter, r *http.Request) {
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

func (db *Handeldb) LikePost(w http.ResponseWriter, r *http.Request) {
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
	var like bool
	if r.FormValue("like") == "true" {
		like = true
	} else if r.FormValue("like") == "false" {
		like = false
	} else {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

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
			insertQuery := "INSERT INTO likes_dislikes (user_id, post_id, is_like) VALUES (?, ?, ?)"
			_, insertErr := db.DB.Exec(insertQuery, userID, postID, like)
			if insertErr != nil {
				fmt.Println("failed to add like:", insertErr)
			}
		} else {
			fmt.Println("error checking like:", err)
		}
	} else {
		var likeval bool
		err := db.DB.QueryRow("SELECT is_like FROM likes_dislikes WHERE id = ?", likeID).Scan(&likeval)
		if err != nil {
			fmt.Println(err)
		}
		if likeval != like {
			updatelike := `UPDATE likes_dislikes
			SET is_like = ?
			WHERE id = ?`
			_, err := db.DB.Exec(updatelike, like, likeID)
			if err != nil {
				fmt.Println("error updating like:", err)
				http.Error(w, "error updating like", http.StatusInternalServerError)
				return
			}
		} else {

			deleteQuery := "DELETE FROM likes_dislikes WHERE id = ?"
			_, deleteErr := db.DB.Exec(deleteQuery, likeID)
			if deleteErr != nil {
				fmt.Println("failed to remove like:", deleteErr)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Post liked successfully"}`))
}

func (db *Handeldb) UseserExist(email string) bool {
	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", email).Scan(&exists); err != nil {
		fmt.Println(err)
	}
	return exists
}

func (db *Handeldb) Addcomment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userId int
	getuserId := "SELECT user_id FROM sessions WHERE token = ?"

	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}

	content := r.FormValue("comment")
	fmt.Println(content)
	err = db.DB.QueryRow(getuserId, cookie.Value).Scan(&userId)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	insetComment := "INSERT INTO comments (user_id, post_id, content) VALUES (?,?,?)"
	_, err = db.DB.Exec(insetComment, userId, 5, content)
	if err != nil {
		fmt.Println(err.Error())
	}

}
