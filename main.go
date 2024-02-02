package main

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
)

var (
	password = "123qweASD"
	port     = 1433
	server   = "localhost\\SQLEXPRESS"
	user     = "sa"
	database = "GoBlog"
)

var DbInstance *sql.DB

type User struct {
	Name                  string
	Age                   uint16
	Money                 int16
	avg_grades, happiness float64
	Hobbies               []string
}

type Article struct {
	Id int
	Title, Anons, FullText string
}

type UserItem struct {
	Name string `json:"name"`
	Age uint16 `json:"age"`
}

func (u User) getAllInfo() string {
	return fmt.Sprintf("User name is: %s. He is %d and he has money "+
		"equal: %d", u.Name, u.Age, u.Money)
}

func (u *User) setNewName(newName string) {
	u.Name = newName
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	res, err := DbInstance.Query(`SELECT [Id], [Title], [Anons], [FullText] FROM [GoBlog].[dbo].[Articles]`)
	if err != nil {
		log.Fatal("Select rows failed:", err.Error())
	}

	var posts = []Article{}

	for res.Next(){
		var article Article
		err = res.Scan(&article.Id, &article.Title, &article.Anons, &article.FullText)
		if err != nil {
			log.Fatal("Select article row failed:", err.Error())
		}
		posts = append(posts, article)
	}

	defer res.Close()

	tmpl.ExecuteTemplate(w, "index", posts)
}

func create(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(writer, err.Error())
	}
	tmpl.ExecuteTemplate(writer, "create", nil)
}

func contacts_page(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Contacts page")
}

func save_article(writer http.ResponseWriter, request *http.Request) {
	title := request.FormValue("title")
	anons := request.FormValue("anons")
	full_text := request.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		fmt.Fprintf(writer, "Не все данные заполнены")
	}else {
		insert, err := DbInstance.Query(`INSERT INTO [dbo].[Articles] ([Title], [Anons], [FullText]) 
		VALUES (?, ?, ?)`, title, anons, full_text)
		if err != nil {
			log.Fatal("Insert row failed:", err.Error())
		}
		defer insert.Close()

		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}
}

func show_post(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	post := Article{}

	row := DbInstance.QueryRow(`SELECT [Id], [Title], [Anons], [FullText] 
		FROM [GoBlog].[dbo].[Articles] WHERE [Id]=?`, vars["id"])

	err := row.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)

	if row == nil || err != nil || post.Id == 0{
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, "Not found")
	}else {
		tmpl, err := template.ParseFiles("templates/show_post.html",
			"templates/header.html", "templates/footer.html")
		if err != nil {
			fmt.Fprintf(writer, err.Error())
		}
		tmpl.ExecuteTemplate(writer, "show_post", post)
	}
}


func handleRequest() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router := mux.NewRouter()
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/create", create).Methods("GET")
	router.HandleFunc("/save_article", save_article).Methods("POST")
	router.HandleFunc("/contacts/", contacts_page).Methods("GET")
	router.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET")

	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}



func main() {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s",
		server, user, password, port, database)

	db, err := sql.Open("mssql", connString)

	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}

	DbInstance = db
	defer db.Close()

	//insert, err := db.Query(`INSERT INTO [dbo].[Users] ([Name], [Age]) VALUES ('Bob', 35)`)
	//if err != nil {
	//	log.Fatal("Insert row failed:", err.Error())
	//}
	//defer insert.Close()

	//res, err := db.Query(`SELECT [Name], [Age] FROM [GoBlog].[dbo].[Users]`)
	//if err != nil {
	//	log.Fatal("Select rows failed:", err.Error())
	//}
	//
	//for res.Next(){
	//	var user UserItem
	//	err = res.Scan(&user.Name, &user.Age)
	//	if err != nil {
	//		log.Fatal("Select user row failed:", err.Error())
	//	}
	//	log.Println(user)
	//}
	//
	//defer res.Close()

	handleRequest()
}
