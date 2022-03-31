package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"paaseasy/functions"
	"strings"
	"unicode"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"

	"gopkg.in/ini.v1"
)

type pathpass struct {
	path string
	pass string
}

var tpl *template.Template
var db *sql.DB
var store *sessions.CookieStore
var cfg *ini.File

//var store = sessions.NewCookieStore([]byte("super-secret"))

func main() {
	var err error
	cfg, err = ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	store = sessions.NewCookieStore([]byte(cfg.Section("").Key("supersecret").String()))
	tpl, _ = template.ParseGlob("templates/*.html")
	db, err = sql.Open("mysql", cfg.Section("").Key("database_user").String()+":"+cfg.Section("").Key("database_password").String()+"@tcp(localhost:3306)/"+cfg.Section("").Key("database_name").String())
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/loginauth", loginAuthHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerauth", registerAuthHandler)
	http.HandleFunc("/about", Auth(aboutHandler))
	http.HandleFunc("/createpaas", Auth(createnewpaas))
	http.HandleFunc("/createnewauth", Auth(createnewauth))
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func Auth(HandlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["userID"]
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return
		}

		HandlerFunc.ServeHTTP(w, r)
	}
}

// loginHandler serves form for users to login with
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****loginHandler running*****")
	tpl.ExecuteTemplate(w, "login.html", nil)
}

// loginAuthHandler authenticates user login
func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****loginAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println("username:", username, "password:", password)
	// retrieve password from db to compare (hash) with user supplied password's hash
	var userID, hash string
	stmt := "SELECT UserID, Hash FROM bcrypt WHERE Username = ?"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&userID, &hash)
	fmt.Println("hash from db:", hash)
	if err != nil {
		fmt.Println("error selecting Hash in db by Username")
		tpl.ExecuteTemplate(w, "login.html", "check username and password")
		return
	}
	// func CompareHashAndPassword(hashedPassword, password []byte) error
	// CompareHashAndPassword() returns err with a value of nil for a match
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		// Get always returns a session, even if empty
		// returns error if exists and could not be decoded
		// Get(r *http.Request, name string) (*Session, error)
		session, _ := store.Get(r, "session")
		// session struct has field Values map[interface{}]interface{}
		session.Values["userID"] = userID
		// save before writing to response/return from handler
		session.Save(r, w)
		tpl.ExecuteTemplate(w, "userfirstpage.html", username)
		return
	}
	fmt.Println("incorrect password")
	tpl.ExecuteTemplate(w, "login.html", "check username and password")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****indexHandler running*****")
	session, _ := store.Get(r, "session")
	_, ok := session.Values["userID"]
	fmt.Println("ok:", ok)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound) // http.StatusFound is 302
		return
	}
	tpl.ExecuteTemplate(w, "index.html", "Logged In")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "about.html", "Logged In")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****logoutHandler running*****")
	session, _ := store.Get(r, "session")
	// The delete built-in function deletes the element with the specified key (m[key]) from the map.
	// If m is nil or there is no such element, delete is a no-op.
	delete(session.Values, "userID")
	session.Save(r, w)
	tpl.ExecuteTemplate(w, "login.html", "Logged Out")
}

// registerHandler serves form for registring new users
func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****registerHandler running*****")
	tpl.ExecuteTemplate(w, "register.html", nil)
}

// registerAuthHandler creates new user in database
func registerAuthHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("*****registerAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")
	// check username for only alphaNumeric characters
	var nameAlphaNumeric = true
	for _, char := range username {
		// func IsLetter(r rune) bool, func IsNumber(r rune) bool
		// if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
		if unicode.IsLetter(char) == false && unicode.IsNumber(char) == false {
			nameAlphaNumeric = false
		}
	}
	// check username pswdLength
	var nameLength bool
	if 5 <= len(username) && len(username) <= 50 {
		nameLength = true
	}
	// check password criteria
	password := r.FormValue("password")
	fmt.Println("password:", password, "\npswdLength:", len(password))
	// variables that must pass for password creation criteria
	var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpaces bool
	pswdNoSpaces = true
	for _, char := range password {
		switch {
		// func IsLower(r rune) bool
		case unicode.IsLower(char):
			pswdLowercase = true
		// func IsUpper(r rune) bool
		case unicode.IsUpper(char):
			pswdUppercase = true
		// func IsNumber(r rune) bool
		case unicode.IsNumber(char):
			pswdNumber = true
		// func IsPunct(r rune) bool, func IsSymbol(r rune) bool
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			pswdSpecial = true
		// func IsSpace(r rune) bool, type rune = int32
		case unicode.IsSpace(int32(char)):
			pswdNoSpaces = false
		}
	}
	if 11 < len(password) && len(password) < 60 {
		pswdLength = true
	}
	fmt.Println("pswdLowercase:", pswdLowercase, "\npswdUppercase:", pswdUppercase, "\npswdNumber:", pswdNumber, "\npswdSpecial:", pswdSpecial, "\npswdLength:", pswdLength, "\npswdNoSpaces:", pswdNoSpaces, "\nnameAlphaNumeric:", nameAlphaNumeric, "\nnameLength:", nameLength)
	if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpaces || !nameAlphaNumeric || !nameLength {
		tpl.ExecuteTemplate(w, "register.html", "please check username and password criteria")
		return
	}
	// check if username already exists for availability
	stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		fmt.Println("username already exists, err:", err)
		tpl.ExecuteTemplate(w, "register.html", "username already taken")
		return
	}
	// create hash from password
	var hash []byte
	// func GenerateFromPassword(password []byte, cost int) ([]byte, error)
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Println("hash:", hash)
	fmt.Println("string(hash):", string(hash))
	// func (db *DB) Prepare(query string) (*Stmt, error)
	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO bcrypt (Username, Hash) VALUES (?, ?);")
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	defer insertStmt.Close()
	var result sql.Result
	//  func (s *Stmt) Exec(args ...interface{}) (Result, error)
	result, err = insertStmt.Exec(username, hash)
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	fmt.Println("rowsAff:", rowsAff)
	fmt.Println("lastIns:", lastIns)
	fmt.Println("err:", err)
	if err != nil {
		fmt.Println("error inserting new user")
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Fprint(w, "congrats, your account has been successfully created")
}
func createnewpaas(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "createnew.html", nil)
}

func createnewauth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("****Creating new Paas****")
	r.ParseForm()
	var data pathpass
	repourl := r.FormValue("repourl")
	branch := r.FormValue("branch")
	port := r.FormValue("port")
	//TODO : show guide page and path and password
	data.pass = functions.RandomString(12)
	path := functions.RandomString(7)
	data.path = strings.ToLower(path)
	tpl.ExecuteTemplate(w, "guidpage.html", data)
	fmt.Println(data.pass, data.path)
	//TODO : add url and password to HOOKS.json file
	work_directory := "/tmp/" + data.path
	execute_commands := work_directory + "/commands.sh"
	err := os.Mkdir(work_directory, 0755)
	if err != nil {
		log.Fatal(err)
	}

	//export last version from DB
	/*var last_version string
	stmt := "SELECT lastversion FROM lastversion WHERE path = ?"
	row := db.QueryRow(stmt, data.path)
	err = row.Scan(&last_version)
	if err != sql.ErrNoRows {
		fmt.Println("err:", err)
		return
	}
	last_int, _ := strconv.Atoi(last_version)*/
	dbname := cfg.Section("").Key("database_name").String()
	dbuser := cfg.Section("").Key("database_user").String()
	dbpassword := cfg.Section("").Key("database_password").String()
	gitrepo := cfg.Section("").Key("git_repo_secret").String()
	functions.Versionhandler(true, data.path, "0", dbuser, dbpassword, dbname)
	functions.CommandCreator(work_directory, data.path, dbuser, dbpassword, dbname, gitrepo)
	functions.Updater(string(data.path), string(data.pass), execute_commands, work_directory)
	fmt.Println("repo URL", repourl)
	fmt.Println("Branch", branch)
	fmt.Println("Port", port)
}
