package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
	// "os"
)

var db *sql.DB
var err error

type user struct {
	ID           int
	Username     string
	Password     string
	Nim          string
	Nama         string
	AsalInstansi string
	MulaiPkl     time.Time
	SelesaiPkl   time.Time
	UploadFile   string
	Role         int
	Status       int
}

func connect_db() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1)/web_pkl")

	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}

func routes() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/home", home)
	http.HandleFunc("admin/homeAdmin", homeAdmin)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/about", about)
}

// func connectServer(){
// 	directory := http.Dir("./resources")
// 	fileServer := http.FileServer(directory)

// 	mux := http.NewServeMux()
// 	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
// }

func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(`
		SELECT id, 
		username, 
		password,
		nim,
		nama,
		asal_instansi,
		mulai_pkl,
		selesai_pkl,
		upload_file,
		role,
		status
		FROM users WHERE username=?
		`, username).
		Scan(
			&users.ID,
			&users.Username,
			&users.Password,
			&users.Nim,
			&users.Nama,
			&users.AsalInstansi,
			&users.MulaiPkl,
			&users.SelesaiPkl,
			&users.UploadFile,
			&users.Role,
			&users.Status,
		)
	return users
}

func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	return true
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "register.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	nim := r.FormValue("nim")
	nama := r.FormValue("nama")
	asal_instansi := r.FormValue("asal_instansi")
	mulai_pkl := r.FormValue("mulai_pkl")
	selesai_pkl := r.FormValue("selesai_pkl")
	upload_file := r.FormValue("upload_file")
	role := r.FormValue("role")
	status := r.FormValue("status")

	users := QueryUser(username)

	if (user{}) == users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if len(hashedPassword) != 0 && checkErr(w, r, err) {
			stmt, err := db.Prepare("INSERT INTO users SET username=?, password=?, nim=?, nama=?, asal_instansi=?, mulai_pkl=?, selesai_pkl=?, upload_file=?, role=?, status=?")
			if err == nil {
				_, err := stmt.Exec(&username, &hashedPassword, &nim, &nama, &asal_instansi, &mulai_pkl, &selesai_pkl, &upload_file, &role, &status)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}
	} else {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
	}
}

func about(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "about.html")
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "login.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println(username)
	fmt.Println(password)

	users := QueryUser(username)

	//deskripsi dan compare password
	var password_tes = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))

	if password_tes == nil {
		//login success
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		session.Set("password", users.Password)
		http.Redirect(w, r, "admin/homeAdmin", http.StatusFound)
		fmt.Println("Sukses")
	} else {
		//login failed
		fmt.Println("Gagal")
		fmt.Fprint(w, "Gagal")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func homeAdmin(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "admin/homeAdmin", http.StatusMovedPermanently)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
		"message":  "Welcome to the Go !",
	}
	var t, err = template.ParseFiles("home_admin.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Execute(w, data)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "home.html")
		return
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/home", http.StatusFound)
}

func main() {
	//connectServer()
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.Handle("/video/", http.StripPrefix("/video/", http.FileServer(http.Dir("video"))))
	connect_db()
	routes()

	defer db.Close()

	fmt.Println("Server running on port :2003")
	http.ListenAndServe(":2003", nil)
}