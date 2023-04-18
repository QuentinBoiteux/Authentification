package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

//
// STRUCTURES
//

var fileIndex = 0

type UserAccount struct {
	Id        int
	Name      string // Name of the user
	Image     string // Path src
	Email     string
	UUID      string
	Password  string
	Salt      string
	Post      []Post
	Comment   []Comment
	Admin     bool // true: the user is the Admin
	Connected bool
}

type Post struct {
	Id        int
	UserId    int
	Username  string
	Title     string
	Body      string
	Likes     int
	Dislikes  int
	Comments  []Comment
	PostTime  string
	Category  string
	ImagePath string
	Connected bool
}

type Like struct {
	Id     int
	UserId int
	PostId int
}

type Dislike struct {
	Id     int
	UserId int
	PostId int
}

type Comment struct {
	Id        int
	UserId    int
	Username  string
	PostId    int
	Body      string
	Likes     int
	Dislikes  int
	Connected bool
}

type Likecomment struct {
	Id     int
	UserId int
	PostId int
}

type Dislikecomment struct {
	Id     int
	UserId int
	PostId int
}

type AuthGoogle struct {
	Access_Token  string `json:"access_token"`
	Expires_In    int    `json:"expires_in"`
	Refresh_Token string `json:"refresh_token"`
	Id_Token      string `json:"id_token"`
	Scope         string `json:"scope"`
	Token_Type    string `json:"token_type"`
}

type GoogleUser struct {
	Name           string `json:"name"`
	Picture        string `json:"picture"`
	Email          string `json:"email"`
	Email_Verified string `json:"email_verified"`
}

type AuthGitHub struct {
	Access_Token string `json:"access_token"`
	Scope        string `json:"scope"`
	Token_Type   string `json:"token_type"`
}

type GithubUser struct {
	Avatar_Url string `json:"avatar_url"`
	Name       string `json:"login"`
	Email      string `json:"email"`
}

//
// DATABASE MANAGEMENT
//

var (
	database *sql.DB
	posts    Post
	comments Comment
	user     UserAccount
	uAccount []UserAccount
)

func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/post-list.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to parse template.", http.StatusInternalServerError)
		return
	}

	rows, err := database.Query("SELECT * FROM posts")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to query data from the database.", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	for rows.Next() {
		posts := []Post{}
		post := Post{}
		err := rows.Scan(&post.Id, &post.Title, &post.Body, &post.Likes, &post.Dislikes, &post.PostTime, &post.Category, &post.ImagePath)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}

	err = t.Execute(w, posts)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to execute template.", http.StatusInternalServerError)
		return
	}
}

func GetUidFromPostID(pid int) int {
	var userId int
	row := database.QueryRow("SELECT user_id FROM posts WHERE id = ?", pid)
	err := row.Scan(&userId)
	if err != nil {
		panic(err)
	}
	return userId
}

func LikeCommentExists(uid int, cid int) bool {
	count := 0
	row := database.QueryRow("SELECT COUNT(*) FROM likescomment WHERE user_id = ? AND comment_id = ?", uid, cid)
	err := row.Scan(&count)
	if err != nil {
		panic(err)
	}
	return count != 0
}

func DislikeCommentExists(uid int, cid int) bool {
	count := 0
	row := database.QueryRow("SELECT COUNT(*) FROM dislikescomment WHERE user_id = ? AND comment_id = ?", uid, cid)
	err := row.Scan(&count)
	if err != nil {
		panic(err)
	}
	return count != 0
}

func LikeExists(uid int, pid int) bool {
	count := 0
	row := database.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND posts_id = ?", uid, pid)
	err := row.Scan(&count)
	if err != nil {
		panic(err)
	}
	return count != 0
}

func DislikeExists(uid int, pid int) bool {
	count := 0
	row := database.QueryRow("SELECT COUNT(*) FROM dislikes WHERE user_id = ? AND posts_id = ?", uid, pid)
	err := row.Scan(&count)
	if err != nil {
		panic(err)
	}
	return count != 0
}

func GetUsernameFromUserId(uid int) string {
	database, _ := sql.Open("sqlite3", "./database.db")
	var username string
	fmt.Println(uid)
	row := database.QueryRow("SELECT username FROM users WHERE id = ?", uid)
	err := row.Scan(&username)
	if err != nil {
		panic(err)
	}
	return username
}

func APINewPostHandler(w http.ResponseWriter, r *http.Request) {
	database, _ := sql.Open("sqlite3", "./database.db")
	if r.Method == http.MethodPost {
		dt := time.Now()
		userId := AuthentifiedUser(r)
		if userId == -1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		filepath := ""

		title := r.FormValue("title")
		body := r.FormValue("body")
		categ := r.FormValue("category")
		cat := strings.ReplaceAll(categ, " ", " #")
		file, _, err := r.FormFile("file")
		if err == nil {
			filepath = fmt.Sprintf("static/uploads/%d.png", fileIndex)
			dst, err := os.Create(filepath)
			if err != nil {
				log.Fatalln(err)
			}
			io.Copy(dst, file)
			fileIndex++
		}

		_, err = database.Exec("INSERT INTO posts (user_id, username, title, body, likes, dislikes, postime, category, image) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
			userId, GetUsernameFromUserId(userId), title, body, 0, 0, dt.Format("Posted the 01-02-2006 at 15:04:05"), strings.ToUpper(cat), filepath)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to insert data into the database.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func APIDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userId := AuthentifiedUser(r)
		if userId == -1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		id := r.FormValue("id")
		newId, _ := strconv.Atoi(id)
		postNewId := GetUidFromPostID(newId)
		if postNewId != userId {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		_, err := database.Exec("DELETE FROM posts WHERE id = ?", id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to delete post from the database.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)

	}
}

var count = make(map[int]int)

func APINewCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		uid := AuthentifiedUser(r)
		if uid == -1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		postId := r.FormValue("id")
		body := r.FormValue("commentBox")

		postIdInt, err := strconv.Atoi(postId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid post id.", http.StatusBadRequest)
			return
		}

		username := GetUsernameFromUserId(uid)

		_, err = database.Exec("INSERT INTO comments(posts_id, user_id, username, body, likes, dislikes) VALUES(?,?,?,?,?,?)", postIdInt, uid, username, body, 0, 0)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to insert comment into the database.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func APILikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	uid := AuthentifiedUser(r)
	commentIdString := r.URL.Query().Get("comment_id")
	commentId, _ := strconv.Atoi(commentIdString)
	if uid == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !LikeCommentExists(uid, commentId) {
		_, err := database.Exec("INSERT INTO likescomment (user_id, comment_id) VALUES(?, ?)", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
		_, err = database.Exec("DELETE FROM dislikescomment WHERE user_id = ? AND comment_id = ?", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := database.Exec("DELETE FROM likescomment WHERE user_id = ? AND comment_id = ?", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := database.Query("SELECT comment_id, COUNT(*) as likescomment_count FROM likescomment GROUP BY comment_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	likes := make(map[int]int)

	for rows.Next() {
		var commentId int
		var likesCCount int
		err := rows.Scan(&commentId, &likesCCount)
		if err != nil {
			log.Fatal(err)
		}
		likes[commentId] = likesCCount
		fmt.Println("likeComment", likes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE comments SET likes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for cId, likesCount := range likes {
		_, err := database.Exec("UPDATE comments SET likes = ? WHERE id = ?", likesCount, cId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err = database.Query("SELECT comment_id, COUNT(*) as dislikescomment_count FROM dislikescomment GROUP BY comment_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dislikes := make(map[int]int)

	for rows.Next() {
		var commentId int
		var dislikesCCount int
		err := rows.Scan(&commentId, &dislikesCCount)
		if err != nil {
			log.Fatal(err)
		}
		dislikes[commentId] = dislikesCCount
		fmt.Println("dislikeComment", dislikes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE comments SET dislikes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for cId, dislikesCount := range dislikes {
		_, err := database.Exec("UPDATE comments SET dislikes = ? WHERE id = ?", dislikesCount, cId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func APIDislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	uid := AuthentifiedUser(r)
	commentIdString := r.URL.Query().Get("comment_id")
	commentId, _ := strconv.Atoi(commentIdString)
	if uid == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !DislikeCommentExists(uid, commentId) {
		_, err := database.Exec("INSERT INTO dislikescomment (user_id, comment_id) VALUES(?, ?)", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
		_, err = database.Exec("DELETE FROM likescomment WHERE user_id = ? AND comment_id = ?", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := database.Exec("DELETE FROM dislikescomment WHERE user_id = ? AND comment_id = ?", uid, commentId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := database.Query("SELECT comment_id, COUNT(*) as dislikescomment_count FROM dislikescomment GROUP BY comment_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dislikes := make(map[int]int)

	for rows.Next() {
		var commentId int
		var dislikesCCount int
		err := rows.Scan(&commentId, &dislikesCCount)
		if err != nil {
			log.Fatal(err)
		}
		dislikes[commentId] = dislikesCCount
		fmt.Println("dislikeComment", dislikes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE comments SET dislikes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for cId, dislikesCCount := range dislikes {
		_, err := database.Exec("UPDATE comments SET dislikes = ? WHERE id = ?", dislikesCCount, cId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err = database.Query("SELECT comment_id, COUNT(*) as likescomment_count FROM likescomment GROUP BY comment_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	likes := make(map[int]int)

	for rows.Next() {
		var commentId int
		var likesCCount int
		err := rows.Scan(&commentId, &likesCCount)
		if err != nil {
			log.Fatal(err)
		}
		likes[commentId] = likesCCount
		fmt.Println("likeComment", likes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE comments SET likes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for cId, likesCCount := range likes {
		_, err := database.Exec("UPDATE comments SET likes = ? WHERE id = ?", likesCCount, cId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func APILikeHandler(w http.ResponseWriter, r *http.Request) {
	uid := AuthentifiedUser(r)
	postIdString := r.URL.Query().Get("post_id")
	postId, _ := strconv.Atoi(postIdString)
	if uid == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !LikeExists(uid, postId) {
		_, err := database.Exec("INSERT INTO likes (user_id, posts_id) VALUES(?, ?)", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
		_, err = database.Exec("DELETE FROM dislikes WHERE user_id = ? AND posts_id = ?", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := database.Exec("DELETE FROM likes WHERE user_id = ? AND posts_id = ?", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := database.Query("SELECT posts_id, COUNT(*) as likes_count FROM likes GROUP BY posts_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	likes := make(map[int]int)

	for rows.Next() {
		var postId int
		var likesCount int
		err := rows.Scan(&postId, &likesCount)
		if err != nil {
			log.Fatal(err)
		}
		likes[postId] = likesCount
		fmt.Println("like", likes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE posts SET likes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for pId, likesCount := range likes {
		_, err := database.Exec("UPDATE posts SET likes = ? WHERE id = ?", likesCount, pId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err = database.Query("SELECT posts_id, COUNT(*) as dislikes_count FROM dislikes GROUP BY posts_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dislikes := make(map[int]int)

	for rows.Next() {
		var postId int
		var dislikesCount int
		err := rows.Scan(&postId, &dislikesCount)
		if err != nil {
			log.Fatal(err)
		}
		dislikes[postId] = dislikesCount
		fmt.Println("dislike", dislikes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE posts SET dislikes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for pId, dislikesCount := range dislikes {
		_, err := database.Exec("UPDATE posts SET dislikes = ? WHERE id = ?", dislikesCount, pId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func APIDislikeHandler(w http.ResponseWriter, r *http.Request) {
	uid := AuthentifiedUser(r)
	postIdString := r.URL.Query().Get("post_id")
	postId, _ := strconv.Atoi(postIdString)
	if uid == -1 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !DislikeExists(uid, postId) {
		_, err := database.Exec("INSERT INTO dislikes (user_id, posts_id) VALUES(?, ?)", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
		_, err = database.Exec("DELETE FROM likes WHERE user_id = ? AND posts_id = ?", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := database.Exec("DELETE FROM dislikes WHERE user_id = ? AND posts_id = ?", uid, postId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := database.Query("SELECT posts_id, COUNT(*) as dislikes_count FROM dislikes GROUP BY posts_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dislikes := make(map[int]int)

	for rows.Next() {
		var postId int
		var dislikesCount int
		err := rows.Scan(&postId, &dislikesCount)
		if err != nil {
			log.Fatal(err)
		}
		dislikes[postId] = dislikesCount
		fmt.Println("dislike", dislikes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE posts SET dislikes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for pId, dislikesCount := range dislikes {
		_, err := database.Exec("UPDATE posts SET dislikes = ? WHERE id = ?", dislikesCount, pId)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err = database.Query("SELECT posts_id, COUNT(*) as likes_count FROM likes GROUP BY posts_id")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	likes := make(map[int]int)

	for rows.Next() {
		var postId int
		var likesCount int
		err := rows.Scan(&postId, &likesCount)
		if err != nil {
			log.Fatal(err)
		}
		likes[postId] = likesCount
		fmt.Println("like", likes)
	}
	rows.Close()
	_, err = database.Exec("UPDATE posts SET likes = 0")
	if err != nil {
		log.Fatal(err)
	}

	for pId, likesCount := range likes {
		_, err := database.Exec("UPDATE posts SET likes = ? WHERE id = ?", likesCount, pId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func APIfilter(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		filterBox := r.FormValue("filter")
		http.Redirect(w, r, "/?filter="+url.QueryEscape(filterBox), http.StatusFound)
	}
}

func getCommentsFromPostId(postId int) []Comment {
	rows, err := database.Query("SELECT * FROM comments WHERE posts_id = ?", postId)
	if err != nil {
		log.Println(err)
		return []Comment{}
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var id int
		var userId int
		var username string
		var postId int
		var body string
		var likes int
		var dislikes int

		err = rows.Scan(&id, &postId, &userId, &username, &body, &likes, &dislikes)
		if err != nil {
			panic(err)
		}

		comments = append(comments, Comment{Id: id, UserId: userId, Username: username, PostId: postId, Body: body, Likes: likes, Dislikes: dislikes})
	}

	return comments
}

func GoogleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.FormValue("code") != "" {
			code := r.FormValue("code")

			checkGoogleUserLogged, _, uEmail, _ := GoogleAuthLog(code)
			fmt.Print(uEmail)
			if checkGoogleUserLogged {
				uuidGenerated := uuid.NewV4()
				uuidUser := uuidGenerated.String()
				cookie := http.Cookie{
					Value:  uuidUser,
					Name:   "session",
					MaxAge: 7200,
				}
				http.SetCookie(w, &cookie)
				AddSession(GetUserIdFromEmail(uEmail), uuidUser)
				database, _ := sql.Open("sqlite3", "./database.db")
				uName := strings.Split(uEmail, "@")[0]
				// check if user exists by email
				var id int
				_ = database.QueryRow("SELECT id FROM users WHERE mail = ?", uEmail).Scan(&id)

				_, err := database.Exec("INSERT INTO users (username, mail, password, salt) VALUES (?,?,?,?)", uName, uEmail, "", "")
				fmt.Println(err)

				/*dbId := GetUserIdFromEmail(uEmail)
				uuid := uuid.NewV4()
				if _, err := database.Exec("INSERT INTO sessions (user_id, session_uuid) VALUES (?,?)", dbId, uuid.String()); err != nil {
					panic(err)
				}*/

				user.Connected = true
				posts.Connected = true
				comments.Connected = true
				/*SetGoogleUserUUID(uEmail)*/
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/auth", http.StatusFound)
		}
	}
}

func GitHubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.FormValue("code") != "" {
			code := r.FormValue("code")
			checkGitHubUserLoogged, gitHubUserName, GitHub_Email, _ := GitHubLog(code)

			if checkGitHubUserLoogged {
				uuidGenerated := uuid.NewV4()
				uuidGithubUser := uuidGenerated.String()
				cookie := http.Cookie{
					Value:  uuidGithubUser,
					Name:   "session",
					MaxAge: 7200,
				}
				AddSession(GetUserIdFromEmail(GitHub_Email), uuidGithubUser)
				http.SetCookie(w, &cookie)

				database, _ := sql.Open("sqlite3", "./database.db")

				// check if user exists by email

				fmt.Println(GitHub_Email)
				_, err := database.Exec("INSERT INTO users (username, mail, password, salt) VALUES (?,?,?,?)", gitHubUserName, GitHub_Email, "", "")
				fmt.Println(err)

				user.Connected = true
				posts.Connected = true
				comments.Connected = true

				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/auth", http.StatusFound)
		}
	}
}

func CheckGoogleUserLogin(email string, email_verified string, uuid string) bool {
	if email_verified == "false" {
		return false
	} else {
		_, err := database.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, email)
		if err != nil {
			fmt.Println(err)
			return false
		} else {
			return true
		}
	}
}

func SetGoogleUserUUID(userEmail string) string {
	uuidGenerated := uuid.NewV4()
	uuid := uuidGenerated.String()
	_, err := database.Exec("UPDATE users SET UUID = ? WHERE email = ?", uuid, userEmail)
	if err != nil {
		fmt.Println(err)
	}
	return uuid
}

func SetGitHubUUID(userName string) string {
	uuidGenerated := uuid.NewV4()
	uuid := uuidGenerated.String()
	_, err := database.Exec("UPDATE sessions SET UUID = ? WHERE, user.Post = data.GetUserProfil()E name = ?", uuid, userName)
	if err != nil {
		fmt.Println("Error function SetGitUUID dataBase:")
		fmt.Println(err)
	}
	return uuid
}

func GoogleAuthLog(code string) (bool, string, string, string) {
	data := url.Values{}
	data.Set("client_id", "391776863434-om8h81q1ssu6gq4qtufuoh661mb1uoeh.apps.googleusercontent.com")
	data.Set("client_secret", "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")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:443/google")
	data.Set("grant_type", "authorization_code")
	responseGoogle, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	defer responseGoogle.Body.Close()
	var googleTokenJSON AuthGoogle
	err = json.NewDecoder(responseGoogle.Body).Decode(&googleTokenJSON)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(googleTokenJSON)

	googleAuthResponse, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + googleTokenJSON.Id_Token)
	if err != nil {
		log.Fatal(err)
	}
	defer googleAuthResponse.Body.Close()
	var googleUser GoogleUser
	err = json.NewDecoder(googleAuthResponse.Body).Decode(&googleUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(googleUser.Email)

	return true, googleUser.Name, googleUser.Email, googleUser.Email_Verified
}

func GoogleAuthRegister(code string, hashPassword string) (bool, string, string) {
	data := url.Values{}
	data.Set("client_id", "391776863434-om8h81q1ssu6gq4qtufuoh661mb1uoeh.apps.googleusercontent.com")
	data.Set("client_secret", "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")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:443/google")
	data.Set("grant_type", "authorization_code")
	responseGoogle, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	defer responseGoogle.Body.Close()
	var googleTokenJSON AuthGoogle
	err = json.NewDecoder(responseGoogle.Body).Decode(&googleTokenJSON)
	if err != nil {
		log.Fatal(err)
	}
	// Rfresh_Token := googleTokenJSON.Refresh_Token
	// refresh_token := "1//03141UoOFJOiJCgYIARAAGAMSNwF-L9Irjnoum5-ga4HAMEgCNKgxA4GUcxt90qDVCa23nw0ZLZfHUDB7FJ7_JV08LIUCQSBc4r4"
	googleAuthResponse, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + googleTokenJSON.Id_Token)
	if err != nil {
		log.Fatal(err)
	}
	defer googleAuthResponse.Body.Close()
	var googleUser GoogleUser
	err = json.NewDecoder(googleAuthResponse.Body).Decode(&googleUser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(googleUser)

	count := 0
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE mail = ?", googleUser.Email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count > 0 {
		// fmt.Println("google user already registered")
		return false, googleUser.Email, ""
	} else {
		if googleUser.Email != "" {
			_, err = database.Exec("INSERT INTO users (username, mail, password, google, github) VALUES (?, ?, ?, ?,?)", googleUser.Name, googleUser.Email, hashPassword, 1, 0)
			if err != nil {
				log.Fatal(err)
			}
			return true, googleUser.Email, googleUser.Name
		} else {
			return false, "", ""
		}
	}
}

func GitHubRegister(code string) (bool, string, string) {
	data := url.Values{}
	data.Set("client_id", "18685620b53c0b8595cd")
	data.Set("client_secret", "246319a5da6fb046790f2663ac70c27ec8d1be61")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:443/github")
	responseGitHub, err := http.PostForm("https://github.com/login/oauth/access_token", data)
	if err != nil {
		log.Fatal(err)
	}
	if responseGitHub.StatusCode != http.StatusOK {
		log.Fatalf("Error: %v", responseGitHub.Status)
	}
	// read the response
	body, err := ioutil.ReadAll(responseGitHub.Body)
	if err != nil {
		log.Fatal(err)
	}
	// close the response
	responseGitHub.Body.Close()
	// parse the response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatal(err)
	}
	// get the token
	token := values.Get("access_token")
	client := &http.Client{}
	reqGitHubUser, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}
	reqGitHubUser.Header.Set("Authorization", "Bearer "+token)
	reqGitHubUser.Header.Set("Accept", "application/vnd.github+json")
	reqGitHubUser.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	// fmt.Println(reqGitHubUser.Header)
	responseGitHubUser, err := client.Do(reqGitHubUser)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(responseGitHubUser.Status)
	var githubUserJSONToken GithubUser
	json.NewDecoder(responseGitHubUser.Body).Decode(&githubUserJSONToken)
	defer responseGitHubUser.Body.Close()
	// fmt.Println(githubUserJSONToken)
	count := 0
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", githubUserJSONToken.Email).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count > 0 {
		// fmt.Println("Github user already register !")
		return false, "", githubUserJSONToken.Name
	} else {
		// hashPassword := script.GenerateHash(script.GenerateRandomString())
		if githubUserJSONToken.Name != "" {
			hashPassword, _, _ := hashPassword(randomString(12))
			_, err = database.Exec("INSERT INTO users (username, mail,password, google, github) VALUES (?, ?, ?,?,?)", githubUserJSONToken.Name, githubUserJSONToken.Email, hashPassword, 0, 1)
			if err != nil {
				fmt.Println("Erreur EXEC INSERT fonction GitHubRegister")
				log.Fatal(err)
			}
			return true, githubUserJSONToken.Email, githubUserJSONToken.Name
		} else {
			return false, "", ""
		}
	}
}

func GitHubLog(code string) (bool, string, string, string) {
	data := url.Values{}
	data.Set("client_id", "18685620b53c0b8595cd")
	data.Set("client_secret", "246319a5da6fb046790f2663ac70c27ec8d1be61")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:443/github")
	responseGitHub, err := http.PostForm("https://github.com/login/oauth/access_token", data)
	if err != nil {
		log.Fatal(err)
	}
	if responseGitHub.StatusCode != http.StatusOK {
		log.Fatalf("Error: %v", responseGitHub.Status)
	}
	// read the response
	body, err := ioutil.ReadAll(responseGitHub.Body)
	if err != nil {
		log.Fatal(err)
	}
	// close the response
	responseGitHub.Body.Close()
	// parse the response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("string(body): %v\n", string(body))
	// get the token
	token := values.Get("access_token")
	fmt.Println("Token:", token)
	client := &http.Client{}
	reqGitHubUser, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Fatal(err)
	}
	reqGitHubUser.Header.Set("Authorization", "Bearer "+token)
	reqGitHubUser.Header.Set("Accept", "application/vnd.github+json")
	reqGitHubUser.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	// fmt.Println(reqGitHubUser.Header)
	responseGitHubUser, err := client.Do(reqGitHubUser)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(responseGitHubUser.Status)
	var githubUserJSONToken GithubUser
	json.NewDecoder(responseGitHubUser.Body).Decode(&githubUserJSONToken)
	fmt.Println(githubUserJSONToken)
	defer responseGitHubUser.Body.Close()
	var mail, userName, userAvatar string
	userName = githubUserJSONToken.Name
	userAvatar = githubUserJSONToken.Avatar_Url
	mail = githubUserJSONToken.Name + "@github.forum"
	// fmt.Printf("githubUserJSONToken.Email: %v\n", githubUserJSONToken.Email)
	/*count := 0
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", githubUserJSONToken.Name).Scan(&count)
	if err != nil {
		fmt.Println("error reading database to found email !!")
	}
	if count == 1 {
		_, err = database.Exec("UPDATE users SET IMAGE = ? WHERE name = ?", githubUserJSONToken.Avatar_Url, githubUserJSONToken.Name)
		if err != nil {
			fmt.Println("Error in the GitHubAuthLog function, sql Exec setting name, image with email:")
			fmt.Println(err)
			return false, "", "", ""
		}*/
	fmt.Println("c'est un print", userName, mail, userAvatar)
	return true, userName, mail, userAvatar
	/*} else {
		return false, "", "", ""
	}*/
}

func AddSession(id string, uuid string) {
	fmt.Println(id, uuid)

	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println("erreur ouverture de la base de donnée")
		log.Fatal(err)
	}

	fmt.Printf("uuidUser: %v\n", uuid)
	fmt.Printf("userName: %v\n", id)

	if id != "" || uuid != "" {
		_, err := database.Exec("INSERT INTO sessions (user_id, session_uuid) VALUES (?, ?)", id, uuid)
		if err != nil {
			fmt.Println("Erreur à l'insertion de donnée dans session, func AddSession:")
			log.Fatal(err)
		}
	} else {
		fmt.Println("name uuid cookie vide !")
	}
}

func GetUserIdFromEmail(mail string) string {
	fmt.Println("mail: ", mail)
	var userId int
	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		panic(err)
	}
	row := database.QueryRow("SELECT id FROM users WHERE username = ? OR mail = ?", mail, mail)
	err = row.Scan(&userId)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return ""
		} else {
			panic(err)
		}
	}
	return strconv.Itoa(userId)
}
