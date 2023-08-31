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
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1)/web_pkl?parseTime=true")

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
	http.HandleFunc("/home_user", home_user)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/profile", profile)
	//http.HandleFunc("/login_admin", loginAdmin)
	http.HandleFunc("/home_admin", home_admin)
	http.HandleFunc("/edit_profile", editProfile)
	http.HandleFunc("/save_profile", saveProfile)
}

// func QueryAdmin(username string) user {
// 	var admin = user{}
// 	err = db.QueryRow(`
// 		SELECT id, 
// 		username, 
// 		password,
// 		role,
// 		status
// 		FROM admins WHERE username=?
// 		`, username).
// 		Scan(
// 			&admin.ID,
// 			&admin.Username,
// 			&admin.Password,
// 			&admin.Role,
// 			&admin.Status,
// 		)
// 	return admin
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

// func loginAdmin(w http.ResponseWriter, r *http.Request) {
// 	session := sessions.Start(w, r)
// 	if len(session.GetString("username")) != 0 {
// 		http.Redirect(w, r, "/", http.StatusFound)
// 		return
// 	}

// 	if r.Method != "POST" {
// 		http.ServeFile(w, r, "login_admin.html")
// 		return
// 	}

// 	username := r.FormValue("username")
// 	password := r.FormValue("password")

// 	admin := QueryAdmin(username)

// 	var password_tes = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))

// 	if password_tes == nil {
// 		session := sessions.Start(w, r)
// 		session.Set("username", admin.Username)
// 		session.Set("password", admin.Password)
// 		http.Redirect(w, r, "/home_admin", http.StatusFound)
// 		fmt.Println("Admin login success")
// 	} else {
// 		fmt.Println("Admin login failed")
// 		fmt.Fprint(w, "Admin login failed")
// 		http.Redirect(w, r, "/login_admin", http.StatusFound)
// 	}
// }


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
		http.Redirect(w, r, "/home_user", http.StatusFound)
		fmt.Println("Sukses")
	} else {
		//login failed
		fmt.Println("Gagal")
		fmt.Fprint(w, "Gagal")
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func home_user(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/home_user", http.StatusMovedPermanently)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
	}
	var t, err = template.ParseFiles("home_user.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Execute(w, data)
}

func home_admin(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login_admin", http.StatusMovedPermanently)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
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

func GetUserByUsername(username string) (user, error) {
	var u user
	err := db.QueryRow("SELECT * FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Nim, &u.Nama, &u.AsalInstansi,
			&u.MulaiPkl, &u.SelesaiPkl, &u.UploadFile, &u.Role, &u.Status)
	if err != nil {
		return u, err
	}
	return u, nil
}


func profile(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	username := session.GetString("username")

	if len(username) == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	u, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{
		"username":     u.Username,
		//"password":     u.Password,
		"nim":          u.Nim,
		"nama":         u.Nama,
		"asal_instansi": u.AsalInstansi,
		"mulai_pkl":    u.MulaiPkl,
		"selesai_pkl":  u.SelesaiPkl,
		"upload_file":  u.UploadFile,
		"role":         u.Role,
		"status":       u.Status,
	}

	var t *template.Template
	t, err = template.ParseFiles("profile.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

func editProfile(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	username := session.GetString("username")

	if len(username) == 0 {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	u, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = map[string]interface{}{
		"username":     u.Username,
		"nim":          u.Nim,
		"nama":         u.Nama,
		"asal_instansi": u.AsalInstansi,
		"mulai_pkl":    u.MulaiPkl,
		"selesai_pkl":  u.SelesaiPkl,
		"upload_file":  u.UploadFile,
		"role":         u.Role,
		"status":       u.Status,
	}

	var t *template.Template
	t, err = template.ParseFiles("edit_profile.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

func saveProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/edit_profile", http.StatusFound)
		return
	}

	session := sessions.Start(w, r)
	username := session.GetString("username")

	if len(username) == 0 {
		http.Redirect(w, r, "/home_user", http.StatusFound)
		return
	}
	Nim := r.FormValue("nim")
	Nama := r.FormValue("nama")
	AsalInstansi := r.FormValue("asal_instansi")
	MulaiPkl := r.FormValue("mulai_pkl")
	SelesaiPkl := r.FormValue("selesai_pkl")
	UploadFile := r.FormValue("upload_file")
	Role := r.FormValue("role")
	Status := r.FormValue("status")

	
	layout := "2006-01-02"
	mulaiPkl, err := time.Parse(layout, MulaiPkl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	selesaiPkl, err := time.Parse(layout, SelesaiPkl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

		result, err := db.Exec(`
		UPDATE users
		SET nim=?, 
		nama=?, 
		asal_instansi=?, 
		mulai_pkl=?, 
		selesai_pkl=?, 
		upload_file=?, 
		role=?, 
		status=?
		WHERE username=?
	`,
	Nim, 
	Nama, 
	AsalInstansi, 
	mulaiPkl, 
	selesaiPkl, 
	UploadFile, 
	Role, 
	Status, 
	username,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

affectedRows, _ := result.RowsAffected()
if affectedRows == 0 {
	http.Error(w, "Tidak ada perubahan", http.StatusInternalServerError)
	return
}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
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

	fmt.Println("Server running on port :2004")
	http.ListenAndServe(":2004", nil)
}