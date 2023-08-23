package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
	// "os"
)

var db *sql.DB
//var err error
var filepath = path.Join("views", "register.html")
var tmpl, err = template.ParseFiles(filepath)

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
	http.HandleFunc("/registerr", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/home", home)
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

				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
		}
	} else {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
	}
}


func login(w http.ResponseWriter, r *http.Request){ 
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "login.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	users := QueryUser(username)

	if (user{}) != users {
		err := bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))
		if err != nil {
			session := sessions.Start(w, r)
			session.Set("username", users.Username)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	}
	fmt.Println("Login Success")

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func home(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/home", http.StatusMovedPermanently)
	}

	var data = map[string]interface{}{
		"username": session.GetString("username"),
		"message":  "Welcome to the Go !",
	}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	// var t, err = template.ParseFiles("home.html")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// t.Execute(w, data)
	// //return



}

// func home(w http.ResponseWriter, r *http.Request) {
// 	// Check if the user is authenticated
// 	session := sessions.Start(w, r)
// 	username := session.GetString("username")
// 	if len(username) == 0 {
// 		http.Redirect(w, r, "/home", http.StatusSeeOther)
// 		return
// 	}

// 	// Get user information from the database
// 	users := QueryUser(username)
// 	if (user{}) == users {
// 		http.Error(w, "User not found", http.StatusInternalServerError)
// 		return
// 	}

// 	// Define the data to be passed to the template
// 	data := struct {
// 		Username     string
// 		Nim          string
// 		Nama         string
// 		AsalInstansi string
// 		MulaiPkl     time.Time
// 		SelesaiPkl   time.Time
// 		UploadFile   string
// 		Role         int
// 		Status       int
// 	}{
// 		Username:     users.Username,
// 		Nim:          users.Nim,
// 		Nama:         users.Nama,
// 		AsalInstansi: users.AsalInstansi,
// 		MulaiPkl:     users.MulaiPkl,
// 		SelesaiPkl:   users.SelesaiPkl,
// 		UploadFile:   users.UploadFile,
// 		Role:         users.Role,
// 		Status:       users.Status,
// 	}

// 	// Load and parse the template
// 	tmpl, err := template.ParseFiles("home.html")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Execute the template with the data
// 	err = tmpl.Execute(w, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }


// func logout(w http.ResponseWriter, r *http.Request) {
// 	session := sessions.Start(w, r)
// 	session.Clear()
// 	sessions.Destroy(w, r)
// 	http.Redirect(w, r, "/home", http.StatusFound)
// }

func main() {
	//connectServer()
	connect_db()
	routes()

	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		


    })

	http.Handle("/static/", 
        http.StripPrefix("/static/", 
            http.FileServer(http.Dir("assets"))))

	fmt.Println("Server running on port :8000")
	http.ListenAndServe(":8000", nil)
}