package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const ()

var (
	db *sql.DB
)

// MODELS --------------------------------------------------------------------------

type Book struct {
	Id     int
	Name   string
	Author string
}

type BookRequest struct {
	Name   string
	Author string
}

type BookResponse struct {
	Id int
}

type Client struct {
	Id   int
	Name string
}

type ClientRequest struct {
	Name string
}

type ClientResponse struct {
	Id int
}

type Library struct {
	Id     int
	Date   string
	Active bool
}

type LibraryJoin struct {
	Library Library
	Book    Book
	Client  Client
}

type LibraryRequest struct {
	Date   string
	Active bool
}

type LibraryRequestJoin struct {
	Library LibraryRequest
	Book    Book
	Client  Client
}

type LibraryResponse struct {
	Id int
}

// FUNC -----------------------------------------------------------------------------

func getConfig() {
	var fileLines []string
	readFile, _ := os.Open("library.config")
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	readFile.Close()

	connectionString := fileLines[0] + ":" + fileLines[1] + "@" + fileLines[2] + "/" + fileLines[3] + "?parseTime=true"

	// Create the database handle, confirm driver is present
	db, _ = sql.Open("mysql", connectionString)
}

func log2File() {
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
}

func handleRequests() { // router
	router := mux.NewRouter()

	router.HandleFunc("/api/books/{id}", getBook).Methods("GET")       // returns book by id
	router.HandleFunc("/api/books", getBooks).Methods("GET")           // returns all books
	router.HandleFunc("/api/books", postBook).Methods("POST")          // creates book, returns id of created book
	router.HandleFunc("/api/books/{id}", putBook).Methods("PUT")       // updates book by id
	router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE") // deletes book by id

	router.HandleFunc("/api/clients/{id}", getClient).Methods("GET")       // returns client by id
	router.HandleFunc("/api/clients", getClients).Methods("GET")           // returns all clients
	router.HandleFunc("/api/clients", postClient).Methods("POST")          // creates client, returns id of created client
	router.HandleFunc("/api/clients/{id}", putClient).Methods("PUT")       // updates client by id
	router.HandleFunc("/api/clients/{id}", deleteClient).Methods("DELETE") // deletes client by id

	router.HandleFunc("/api/libraries/{id}", getLibrary).Methods("GET")       // returns borrow by id
	router.HandleFunc("/api/libraries", getLibraries).Methods("GET")          // returns all borrowed books
	router.HandleFunc("/api/libraries", postLibrary).Methods("POST")          // creates borrow, returns id of created borrow
	router.HandleFunc("/api/libraries/{id}", putLibrary).Methods("PUT")       // updates borrow by id
	router.HandleFunc("/api/libraries/{id}", deleteLibrary).Methods("DELETE") // deletes borrow by id

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowCredentials: true,
	})
	handler := cors.Handler(router)
	log.Fatal(http.ListenAndServe(":10000", handler))
}

func main() {
	getConfig()
	log2File()

	// Connect and check the server version
	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	log.Println("Connected to:", version)
	fmt.Println("Connected to:", version)

	handleRequests()

	defer db.Close()
}

// ENDPOINTS -------------------------------------------------------------------------

// Books

// GET /api/books/1
func getBook(w http.ResponseWriter, r *http.Request) {
	var name, author string

	vars := mux.Vars(r)
	id := vars["id"]

	// validate if id == int
	int_id, errAtoi := strconv.Atoi(id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("GET /api/books/" + id + " " + errAtoi.Error())
		return
	}

	// repository
	errScan := db.QueryRow("SELECT name, author FROM book WHERE id = ?", int_id).Scan(&name, &author)
	if errScan != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/books/" + id + " " + errScan.Error())
		return
	}
	// number too low or too high -> empty fields // NOT USED
	if name == "" || author == "" {
		w.WriteHeader(http.StatusNoContent)
		log.Println("GET /api/books/" + id + " empty fields")
		return
	}
	book := BookRequest{Name: name, Author: author}

	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(book)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/books/" + id + " " + errEncode.Error())
		return
	}
}

// GET /api/books
func getBooks(w http.ResponseWriter, r *http.Request) {
	var id int
	var name, author string
	var books []Book

	// repository
	rows, errQuery := db.Query("SELECT id, name, author FROM book")
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/books " + errQuery.Error())
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &name, &author)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("GET /api/books " + errScan.Error())
			return
		}
		books = append(books, Book{Id: id, Name: name, Author: author})
	}

	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(books)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/books " + errEncode.Error())
		return
	}
}

// POST /api/books BookRequest{}
func postBook(w http.ResponseWriter, r *http.Request) {
	var payload BookRequest
	var response BookResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/books " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/books " + errUnmarshal.Error())
		return
	}
	// wrong JSON
	if payload.Name == "" || payload.Author == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("POST /api/books empty fields in JSON")
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO book (Name, Author) VALUES (?, ?)", payload.Name, payload.Author)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/books " + errQuery.Error())
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/books" + errLII.Error())
		return
	}
	response = BookResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	errEncode := json.NewEncoder(w).Encode(response)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/books " + errEncode.Error())
		return
	}
}

// PUT /api/books/1 BookRequest{}
func putBook(w http.ResponseWriter, r *http.Request) {
	var payload Book

	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/books/" + vars_id + " " + errAtoi.Error())
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/books/" + vars_id + " " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/books/" + vars_id + " " + errUnmarshal.Error())
		return
	}
	// wrong JSON or /{id}
	if payload.Name == "" || payload.Author == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/books/" + vars_id + " wrong JSON or id")
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE book SET Name = ?, Author = ? WHERE Id = ?", payload.Name, payload.Author, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/books/" + vars_id + " " + errQuery.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/books/1
func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]

	// validate if id == int, id !< 1
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("DELETE /api/books/" + vars_id + " " + errAtoi.Error())
		return
	}
	if int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("DELETE /api/books/" + vars_id + "  id < 1")
		return
	}

	// repository
	_, errQuery := db.Exec("DELETE FROM book WHERE id = ?", int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("DELETE /api/books/" + vars_id + " " + errQuery.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Clients

// GET /api/clients/1
func getClient(w http.ResponseWriter, r *http.Request) {
	var name string

	vars := mux.Vars(r)
	id := vars["id"]

	// validate if id == int
	int_id, errAtoi := strconv.Atoi(id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("GET /api/clients/" + id + " " + errAtoi.Error())
		return
	}

	// repository
	errScan := db.QueryRow("SELECT name FROM client WHERE id = ?", int_id).Scan(&name)
	if errScan != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/clients/" + id + " " + errScan.Error())
		return
	}
	// number too low or too high -> empty field
	if name == "" {
		w.WriteHeader(http.StatusNoContent)
		log.Println("GET /api/clients/" + id + " empty fields")

		return
	}
	client := ClientRequest{Name: name}
	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(client)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/clients/" + id + " " + errEncode.Error())
		return
	}
}

// GET /api/clients
func getClients(w http.ResponseWriter, r *http.Request) {
	var id int
	var name string
	var clients []Client

	// repository
	rows, errQuery := db.Query("SELECT id, name FROM client")
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/clients/ " + errQuery.Error())
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &name)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("GET /api/clients/ " + errScan.Error())
			return
		}
		clients = append(clients, Client{Id: id, Name: name})
	}

	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(clients)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/clients/ " + errEncode.Error())
		return
	}
}

// POST /api/clients ClientRequest{}
func postClient(w http.ResponseWriter, r *http.Request) {
	var payload ClientRequest
	var response ClientResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/clients/ " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/clients/ " + errUnmarshal.Error())
		return
	}
	// wrong JSON
	if payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("POST /api/clients/  empty fields in JSON")
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO client (Name) VALUES (?)", payload.Name)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/clients/ " + errQuery.Error())
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/clients/ " + errLII.Error())
		return
	}
	response = ClientResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	errEncode := json.NewEncoder(w).Encode(response)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/clients " + errEncode.Error())
		return
	}
}

// PUT /api/clients/1 ClientRequest{}
func putClient(w http.ResponseWriter, r *http.Request) {
	var payload Client

	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/clients/" + vars_id + " " + errAtoi.Error())
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/clients/" + vars_id + " " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/clients/" + vars_id + " " + errUnmarshal.Error())
		return
	}
	// wrong JSON or /{id}
	if payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/clients/" + vars_id + "  wrong JSON or id")
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE client SET Name = ? WHERE Id = ?", payload.Name, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/clients/" + vars_id + " " + errQuery.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/clients/1
func deleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]

	// validate if id == int, id ! < 1
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("DELETE /api/clients/" + vars_id + " " + errAtoi.Error())
		return
	}
	if int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("DELETE /api/clients/" + vars_id + "  id < 1")
		return
	}

	// repository
	_, errQuery := db.Exec("DELETE FROM client WHERE id = ?", vars_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("DELETE /api/clients/" + vars_id + " " + errQuery.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Libraries

// GET /api/libraries/1
func getLibrary(w http.ResponseWriter, r *http.Request) {
	var idBook, idClient int
	var bookName, bookAuthor, clientName, date string
	var active bool

	vars := mux.Vars(r)
	id := vars["id"]

	// validate if id == int
	int_id, errAtoi := strconv.Atoi(id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("GET /api/libraries/" + id + " " + errAtoi.Error())
		return
	}

	// repository
	errScan := db.QueryRow("SELECT id_book, book.name, book.author, id_client, client.name, date, active FROM library INNER JOIN book ON library.id_book = book.id INNER JOIN client ON library.id_client = client.id WHERE library.id = ?", int_id).Scan(&idBook, &bookName, &bookAuthor, &idClient, &clientName, &date, &active)
	if errScan != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/libraries/" + id + " " + errScan.Error())
		return
	}
	// number too low or too high -> empty field
	if idBook == 0 || idClient == 0 || date == "" {
		w.WriteHeader(http.StatusNoContent)
		log.Println("GET /api/libraries/" + id + "  wrong JSON or ID")
		return
	}
	library := LibraryRequestJoin{
		LibraryRequest{Date: date, Active: active},
		Book{Id: idBook, Name: bookName, Author: bookAuthor},
		Client{Id: idClient, Name: clientName},
	}
	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(library)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/libraries/" + id + " " + errEncode.Error())
		return
	}
}

// GET /api/libraries
func getLibraries(w http.ResponseWriter, r *http.Request) {
	var id, id_book, id_client int
	var bookName, bookAuthor, clientName, date string
	var active bool
	var libraries []LibraryJoin

	// repository
	rows, errQuery := db.Query("SELECT library.id, id_book, book.name, book.author, id_client, client.name, date, active FROM library INNER JOIN book ON library.id_book = book.id INNER JOIN client ON library.id_client = client.id ")
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/libraries" + errQuery.Error())
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &id_book, &bookName, &bookAuthor, &id_client, &clientName, &date, &active)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("GET /api/libraries" + errScan.Error())
			return
		}
		libraries = append(libraries, LibraryJoin{
			Library{Id: id, Date: date, Active: active},
			Book{Id: id_book, Name: bookName, Author: bookAuthor},
			Client{Id: id_client, Name: clientName},
		})
	}

	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(libraries)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GET /api/libraries " + errEncode.Error())
		return
	}
}

// POST /api/libraries LibraryRequestJoin{}
func postLibrary(w http.ResponseWriter, r *http.Request) {
	var payload LibraryRequestJoin
	var response LibraryResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/libraries " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/libraries " + errUnmarshal.Error())
		return
	}
	// wrong JSON
	if payload.Book.Id == 0 || payload.Client.Id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("POST /api/libraries wrong JSON or ID")
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO library (id_book, id_client, active) VALUES (?, ?, ?)", payload.Book.Id, payload.Client.Id, payload.Library.Active)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/libraries " + errQuery.Error())
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/libraries " + errLII.Error())
		return
	}
	response = LibraryResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	errEncode := json.NewEncoder(w).Encode(response)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("POST /api/libraries " + errEncode.Error())
		return
	}
}

// PUT /api/libraries/1 LibraryRequestJoin{}
func putLibrary(w http.ResponseWriter, r *http.Request) {
	var payload LibraryRequestJoin

	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/libraries/" + vars_id + " " + errAtoi.Error())
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/libraries/" + vars_id + " " + errIO.Error())
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/libraries/" + vars_id + " " + errUnmarshal.Error())
		return
	}
	// wrong JSON or /{id}
	if payload.Book.Id == 0 || payload.Client.Id == 0 || payload.Library.Date == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("PUT /api/libraries/" + vars_id + " wrong JSON or ID")
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE library SET Id_book = ?, Id_client = ?, Date = ?, Active = ? WHERE Id = ?", payload.Book.Id, payload.Client.Id, payload.Library.Date, payload.Library.Active, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("PUT /api/libraries/" + vars_id + " " + errQuery.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/libraries/1
func deleteLibrary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int, id !< 1
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("DELETE /api/libraries/" + vars_id + " " + errAtoi.Error())
		return
	}
	if int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("DELETE /api/libraries/" + vars_id + "  id < 1")
		return
	}

	// repository
	_, errQuery := db.Exec("DELETE FROM library WHERE id = ?", int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
