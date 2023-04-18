package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	sqlite3 "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "POST request successful")
	name := r.FormValue("name")
	address := r.FormValue("address")
	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Address = %s\n", address)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.FormValue("code") != "" {
			checkGitHub_User_Registered, _, userGitHubName := GitHubRegister(r.FormValue("code"))
			randomPass, _, _ := hashPassword(randomString(12))

			checkGoogleUserRegistered, googleUserEmail, userGoogleName := GoogleAuthRegister(r.FormValue("code"), randomPass)
			if checkGitHub_User_Registered || userGitHubName != "" {
				uuidGitHubUser := SetGitHubUUID(userGitHubName)
				cookie := http.Cookie{
					Value:  uuidGitHubUser,
					Name:   "session",
					MaxAge: 120,
				}
				http.SetCookie(w, &cookie)
				AddSession(userGitHubName, uuidGitHubUser)
				user.Connected = true
				posts.Connected = true
				comments.Connected = true
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			if checkGoogleUserRegistered && googleUserEmail != "" {
				uuidGoogleUser := SetGoogleUserUUID(googleUserEmail)
				cookie := http.Cookie{
					Value:  uuidGoogleUser,
					Name:   "session",
					MaxAge: 120,
				}
				http.SetCookie(w, &cookie)
				AddSession(userGoogleName, uuidGoogleUser)
				user.Connected = true
				posts.Connected = true
				comments.Connected = true
				http.Redirect(w, r, "/", http.StatusFound)
				return
			} else if googleUserEmail != "" {
				http.Redirect(w, r, "/auth", http.StatusFound)
				return
			} else {
				fmt.Println("Error register Google User !")
				return
			}
		}
		username := r.FormValue("username")
		mail := r.FormValue("mail")
		password := r.FormValue("password")
		hashPassword, salt, _ := hashPassword(password)
		entropy := calculateEntropy(password)
		fmt.Println(username, mail, password)
		if entropy < 1.0 {
			fmt.Fprintf(w, "Le mot de passe n'est pas assez fort")
		} else {
			database, _ := sql.Open("sqlite3", "./database.db")

			// insert the new user into the database

			_, err := database.Exec("INSERT INTO users (username, mail, password, salt) VALUES (?, ?, ?, ?)", username, mail, hashPassword, salt)

			if sqliteErr, ok := err.(sqlite3.Error); ok {
				if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					fmt.Fprintf(w, "L'adresse mail ou le Pseudo existe déjà.")
					return
				} else {
					panic(err)
				}
			}

			tmpl := template.Must(template.ParseFiles("templates/thank-you.html"))
			if err = tmpl.Execute(w, nil); err != nil {
				panic(err)
			}
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mail := r.FormValue("mail")
		password := r.FormValue("password")

		var dbId int
		var dbMail string
		var dbPassword string
		var dbSalt string

		err := database.QueryRow("SELECT id, mail, password, salt FROM users WHERE mail = ?", mail).Scan(&dbId, &dbMail, &dbPassword, &dbSalt)
		if err != nil {
			fmt.Fprintf(w, "%s: %s doesn't exist", err.Error(), mail)
			return
		}

		passwordcheck := checkPasswordHash(password, dbPassword, dbSalt)

		if dbMail == mail && passwordcheck {
			uuid := uuid.NewV4()

			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   uuid.String(),
				Expires: time.Now().Add(time.Hour),
				Path:    "/",
			})

			if _, err := database.Exec("INSERT INTO sessions (user_id, session_uuid) VALUES (?,?)", dbId, uuid.String()); err != nil {
				panic(err)
			}

			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			fmt.Fprintf(w, "Wrong password")
			return
		}
	}
}

/*func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}*/

func hashPassword(password string) (string, string, error) {
	b := make([]byte, 10)
	rand.Read(b)
	salt := fmt.Sprintf("%x", b)

	saltedPassword := []byte(password + salt)
	bytes, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.DefaultCost)
	return string(bytes), salt, err
}

/*func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}*/

func checkPasswordHash(password, hash, salt string) bool {
	saltedPassword := []byte(password + salt)
	fmt.Println(hash, string(saltedPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hash), saltedPassword)
	fmt.Print(err)
	return err == nil
}

func AuthentifiedUser(r *http.Request) int {
	cookie, err := r.Cookie("session")
	if err != nil {
		return -1
	}
	var user_id int
	var session_uuid string
	err = database.QueryRow("SELECT user_id, session_uuid FROM sessions WHERE session_uuid = ?", cookie.Value).Scan(&user_id, &session_uuid)
	if err != nil {
		return -1
	}
	fmt.Println(user_id)
	return user_id
}

func calculateEntropy(password string) float64 {
	// Initialisez une map pour compter le nombre d'occurrences de chaque caractère dans le mot de passe
	charCounts := make(map[rune]int)
	for _, c := range password {
		charCounts[c]++
	}

	// Calculez l'entropie en utilisant la formule suivante :
	// entropy = -1 * sum(p(c) * log2(p(c)))
	// où p(c) est la probabilité d'un caractère c dans le mot de passe
	entropy := 0.0
	for _, count := range charCounts {
		p := float64(count) / float64(len(password))
		entropy -= p * math.Log2(p)

	}

	return entropy
}

func getSalt(length int) string {
	salt := ""
	if _, err := os.Stat("salt"); err != nil {
		// File doesn't exists
		b := make([]byte, length)
		rand.Read(b)
		salt = fmt.Sprintf("%x", b)

		file, _ := os.Create("salt")
		defer file.Close()
		file.Write([]byte(salt))
	} else {
		// File exists
		rawSalt, _ := ioutil.ReadFile("salt")
		salt = string(rawSalt)
	}
	return salt
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/auth.html")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func renderPosts(w http.ResponseWriter, r *http.Request) {
	userId := AuthentifiedUser(r)
	filterBox := strings.ToUpper(r.FormValue("filter"))
	filterLikes := r.URL.Query().Get("likes") == "true"
	filterUserLikes := r.URL.Query().Get("user_likes") == "true"
	filterUserPosts := r.URL.Query().Get("user_posts") == "true"
	var rows *sql.Rows
	var err error
	if filterBox != "" {
		rows, err = database.Query("SELECT * FROM posts WHERE category LIKE ?", "%"+filterBox+"%")
	} else if filterLikes {
		rows, err = database.Query("SELECT * FROM posts ORDER BY likes DESC")
	} else if filterUserLikes {
		rows, err = database.Query("SELECT * FROM posts WHERE id IN (SELECT posts_id FROM likes WHERE user_id = ?) ORDER BY likes DESC", userId)
	} else if filterUserPosts {
		rows, err = database.Query("SELECT * FROM posts WHERE user_id = ? ORDER BY postime DESC", userId)
	} else {
		rows, err = database.Query("SELECT * FROM posts")
	}
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var id int
		var userId int
		var username string
		var title string
		var body string
		var likes int
		var dislikes int
		var postime string
		var category string
		var imagePath string

		err = rows.Scan(&id, &userId, &username, &title, &body, &likes, &dislikes, &postime, &category, &imagePath)
		if err != nil {
			panic(err)
		}

		posts = append(posts, Post{Id: id, UserId: userId, Username: username, Title: title, Body: body, Likes: likes, Dislikes: dislikes, PostTime: postime, Category: category, Comments: getCommentsFromPostId(id), ImagePath: imagePath})
	}
	if !filterLikes {
		for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
			posts[i], posts[j] = posts[j], posts[i]
		}
	}

	tmpl := template.Must(template.ParseFiles("./templates/posts.html"))
	err = tmpl.Execute(w, posts)
	if err != nil {
		panic(err)
	}
}

func main() {
	database, _ = sql.Open("sqlite3", "./database.db")
	defer database.Close()

	database.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT NOT NULL UNIQUE, mail TEXT NOT NULL UNIQUE, password TEXT NOT NULL, salt TEXT NOT NULL)")
	database.Exec("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, username TEXT, title TEXT, body TEXT, likes INTEGER, dislikes INTEGER NOT NULL, postime TEXT, category TEXT, image TEXT)")
	database.Exec("CREATE TABLE IF NOT EXISTS likes (id INTEGER PRIMARY KEY AUTOINCREMENT,user_id INTEGER, posts_id INTEGER)")
	database.Exec("CREATE TABLE IF NOT EXISTS dislikes (id INTEGER PRIMARY KEY AUTOINCREMENT,user_id INTEGER, posts_id INTEGER)")
	database.Exec("CREATE TABLE IF NOT EXISTS likescomment (id INTEGER PRIMARY KEY AUTOINCREMENT,user_id INTEGER, comment_id INTEGER)")
	database.Exec("CREATE TABLE IF NOT EXISTS dislikescomment (id INTEGER PRIMARY KEY AUTOINCREMENT,user_id INTEGER, comment_id INTEGER)")
	database.Exec("CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, posts_id INTEGER, user_id INTEGER, username TEXT, body TEXT, likes INTEGER, dislikes INTEGER)")
	database.Exec("DROP TABLE sessions")
	database.Exec("CREATE TABLE sessions (user_id INTEGER, session_uuid TEXT)")

	lmt := tollbooth.NewLimiter(5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Print("lol")
			http.NotFound(w, r)
			return
		}

		if AuthentifiedUser(r) == -1 {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		renderPosts(w, r)
	})

	http.Handle("/api/post/new", tollbooth.LimitFuncHandler(lmt, APINewPostHandler))
	http.Handle("/api/filter", tollbooth.LimitFuncHandler(lmt, APIfilter))
	http.Handle("/api/post/delete", tollbooth.LimitFuncHandler(lmt, APIDeletePostHandler))

	http.Handle("/api/comment/new", tollbooth.LimitFuncHandler(lmt, APINewCommentHandler))
	http.Handle("/api/like", tollbooth.LimitFuncHandler(lmt, APILikeHandler))
	http.Handle("/api/dislike", tollbooth.LimitFuncHandler(lmt, APIDislikeHandler))
	http.Handle("/api/likecomment", tollbooth.LimitFuncHandler(lmt, APILikeCommentHandler))
	http.Handle("/api/dislikecomment", tollbooth.LimitFuncHandler(lmt, APIDislikeCommentHandler))
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", tollbooth.LimitHandler(lmt, fs)))
	http.Handle("/google", tollbooth.LimitFuncHandler(lmt, GoogleHandler))
	http.Handle("/github", tollbooth.LimitFuncHandler(lmt, GitHubHandler))
	http.Handle("/auth", tollbooth.LimitFuncHandler(lmt, authHandler))
	http.Handle("/login", tollbooth.LimitFuncHandler(lmt, loginHandler))
	http.Handle("/register", tollbooth.LimitFuncHandler(lmt, registerHandler))
	http.Handle("/form", tollbooth.LimitFuncHandler(lmt, formHandler))

	fmt.Printf("SERVER AWAITS: https://localhost/\n")
	startHTTPSServer()
}

func startHTTPSServer() {
	// Redirect all incoming HTTP requests to HTTPS
	go func() {
		log.Fatal(http.ListenAndServe(":40", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "https://"+req.Host+req.RequestURI, http.StatusMovedPermanently)
		})))
	}()

	srv := &http.Server{
		Addr: ":443",
	}
	log.Fatal(srv.ListenAndServeTLS("./localhost.crt", "./localhost.key"))
}

func randomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)[:length]
}
