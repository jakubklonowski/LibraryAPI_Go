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

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
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
	Id       int
	IdBook   int
	IdClient int
	Date     string
	Active   bool
}

type LibraryRequest struct {
	IdBook   int
	IdClient int
	Date     string
	Active   bool
}

type LibraryResponse struct {
	Id int
}

// FUNC -----------------------------------------------------------------------------

func getConfig() {
	var fileLines []string
	readFile, _ := os.Open("library.config") // _ to err
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	readFile.Close()

	connectionString := fileLines[0] + ":" + fileLines[1] + "@" + fileLines[2] + fileLines[3] + "/" + fileLines[4] + "?parseTime=true"

	// Create the database handle, confirm driver is present
	db, _ = sql.Open("mysql", connectionString)
}

func handleRequests() {
	router := mux.NewRouter()

	router.HandleFunc("/api/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/api/books", getBooks).Methods("GET")
	router.HandleFunc("/api/books", postBook).Methods("POST")
	router.HandleFunc("/api/books/{id}", putBook).Methods("PUT")
	router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	router.HandleFunc("/api/clients/{id}", getClient).Methods("GET")
	router.HandleFunc("/api/clients", getClients).Methods("GET")
	router.HandleFunc("/api/clients", postClient).Methods("POST")
	router.HandleFunc("/api/clients/{id}", putClient).Methods("PUT")
	router.HandleFunc("/api/clients/{id}", deleteClient).Methods("DELETE")

	router.HandleFunc("/api/libraries/{id}", getLibrary).Methods("GET")
	router.HandleFunc("/api/libraries", getLibraries).Methods("GET")
	router.HandleFunc("/api/libraries", postLibrary).Methods("POST")
	router.HandleFunc("/api/libraries/{id}", putLibrary).Methods("PUT")
	router.HandleFunc("/api/libraries/{id}", deleteLibrary).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":10000", router))
}

func main() {
	getConfig()

	// Connect and check the server version
	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
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

	db.QueryRow("SELECT name, author FROM book WHERE id = ?", id).Scan(&name, &author)
	book := BookRequest{Name: name, Author: author}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

// GET /api/books
func getBooks(w http.ResponseWriter, r *http.Request) {
	var id int
	var name, author string
	var books []Book

	rows, _ := db.Query("SELECT id, name, author FROM book")
	for rows.Next() {
		rows.Scan(&id, &name, &author)
		books = append(books, Book{Id: id, Name: name, Author: author})
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

// POST /api/books BookRequest{}
func postBook(w http.ResponseWriter, r *http.Request) {
	var payload BookRequest
	var response BookResponse

	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &payload)
	result, _ := db.Exec("INSERT INTO book (Name, Author) VALUES (?, ?)", payload.Name, payload.Author)
	id, _ := result.LastInsertId()
	response = BookResponse{Id: int(id)}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// PUT /api/books/1 BookRequest{}
func putBook(w http.ResponseWriter, r *http.Request) {
	var payload Book

	vars := mux.Vars(r)
	vars_id := vars["id"]
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &payload)

	db.Exec("UPDATE book SET Name = ?, Author = ? WHERE Id = ?", payload.Name, payload.Author, vars_id)
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/books/1
func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]

	db.Exec("DELETE FROM book WHERE id = ?", vars_id)
	w.WriteHeader(http.StatusNoContent)
}

// Clients

// GET /api/clients/1
func getClient(w http.ResponseWriter, r *http.Request) {
	var name string

	vars := mux.Vars(r)
	id := vars["id"]

	db.QueryRow("SELECT name FROM client WHERE id = ?", id).Scan(&name)
	client := ClientRequest{Name: name}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(client)
}

// GET /api/clients
func getClients(w http.ResponseWriter, r *http.Request) {
	var id int
	var name string
	var clients []Client

	rows, _ := db.Query("SELECT id, name FROM client")
	for rows.Next() {
		rows.Scan(&id, &name)
		clients = append(clients, Client{Id: id, Name: name})
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clients)
}

// POST /api/clients ClientRequest{}
func postClient(w http.ResponseWriter, r *http.Request) {
	var payload ClientRequest
	var response ClientResponse

	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &payload)
	result, _ := db.Exec("INSERT INTO client (Name) VALUES (?)", payload.Name)
	id, _ := result.LastInsertId()
	response = ClientResponse{Id: int(id)}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// PUT /api/clients/1 ClientRequest{}
func putClient(w http.ResponseWriter, r *http.Request) {
	var payload Client

	vars := mux.Vars(r)
	vars_id := vars["id"]
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &payload)

	db.Exec("UPDATE client SET Name = ? WHERE Id = ?", payload.Name, vars_id)
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/clients/1
func deleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]

	db.Exec("DELETE FROM client WHERE id = ?", vars_id)
	w.WriteHeader(http.StatusNoContent)
}

// Libraries

// GET /api/libraries/1
func getLibrary(w http.ResponseWriter, r *http.Request) {
	var idBook, idClient int
	var date string
	var active bool

	vars := mux.Vars(r)
	id := vars["id"]

	db.QueryRow("SELECT id_book, id_client, date, active FROM library WHERE id = ?", id).Scan(&idBook, &idClient, &date, &active)
	library := LibraryRequest{IdBook: idBook, IdClient: idClient, Date: date, Active: active}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(library)
}

// GET /api/libraries
func getLibraries(w http.ResponseWriter, r *http.Request) {
	var id, id_book, id_client int
	var date string
	var active bool
	var libraries []Library

	rows, _ := db.Query("SELECT id, id_book, id_client, date, active FROM library")
	for rows.Next() {
		rows.Scan(&id, &id_book, &id_client, &date, &active)
		libraries = append(libraries, Library{Id: id, IdBook: id_book, IdClient: id_client, Date: date, Active: active})
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(libraries)
}

// POST /api/libraries LibraryRequest{}
func postLibrary(w http.ResponseWriter, r *http.Request) {
	var payload LibraryRequest
	var response LibraryResponse

	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &payload)
	result, _ := db.Exec("INSERT INTO library (id_book, id_client, date, active) VALUES (?, ?, ?, ?)", payload.IdBook, payload.IdClient, payload.Date, payload.Active)
	id, _ := result.LastInsertId()
	response = LibraryResponse{Id: int(id)}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// PUT /api/libraries/1 LibraryRequest{}
func putLibrary(w http.ResponseWriter, r *http.Request) {
	var payload LibraryRequest

	vars := mux.Vars(r)
	id := vars["id"]

	db.Exec("UPDATE library SET id_book = ?, id_client = ?, date = ?, active = ? WHERE id = ?", payload.IdBook, payload.IdClient, payload.Date, payload.Active, id)
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/libraries/1
func deleteLibrary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db.Exec("DELETE FROM library WHERE id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}
