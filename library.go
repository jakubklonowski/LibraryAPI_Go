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

	// validate if id == int
	int_id, errAtoi := strconv.Atoi(id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// repository
	errScan := db.QueryRow("SELECT name, author FROM book WHERE id = ?", int_id).Scan(&name, &author)
	if errScan != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// number too low or too high -> empty fields
	if name == "" || author == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	book := BookRequest{Name: name, Author: author}
	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(book)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &name, &author)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		books = append(books, Book{Id: id, Name: name, Author: author})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

// POST /api/books BookRequest{}
func postBook(w http.ResponseWriter, r *http.Request) {
	var payload BookRequest
	var response BookResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON
	if payload.Name == "" || payload.Author == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO book (Name, Author) VALUES (?, ?)", payload.Name, payload.Author)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response = BookResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
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
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON or /{id}
	if payload.Name == "" || payload.Author == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE book SET Name = ?, Author = ? WHERE Id = ?", payload.Name, payload.Author, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	if errAtoi != nil || int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	_, errQuery := db.Exec("DELETE FROM book WHERE id = ?", int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		return
	}

	// repository
	errScan := db.QueryRow("SELECT name FROM client WHERE id = ?", int_id).Scan(&name)
	if errScan != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// number too low or too high -> empty field
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	client := ClientRequest{Name: name}
	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(client)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &name)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		clients = append(clients, Client{Id: id, Name: name})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clients)
}

// POST /api/clients ClientRequest{}
func postClient(w http.ResponseWriter, r *http.Request) {
	var payload ClientRequest
	var response ClientResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON
	if payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO client (Name) VALUES (?)", payload.Name)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response = ClientResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
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
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON or /{id}
	if payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE client SET Name = ? WHERE Id = ?", payload.Name, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/clients/1
func deleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int, id !< 1
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil || int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	_, errQuery := db.Exec("DELETE FROM client WHERE id = ?", vars_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

	// validate if id == int
	int_id, errAtoi := strconv.Atoi(id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	errScan := db.QueryRow("SELECT id_book, id_client, date, active FROM library WHERE id = ?", int_id).Scan(&idBook, &idClient, &date, &active)
	if errScan != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// number too low or too high -> empty field
	if idBook == 0 || idClient == 0 || date == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	library := LibraryRequest{IdBook: idBook, IdClient: idClient, Date: date, Active: active}
	w.WriteHeader(http.StatusOK)
	errEncode := json.NewEncoder(w).Encode(library)
	if errEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GET /api/libraries
func getLibraries(w http.ResponseWriter, r *http.Request) {
	var id, id_book, id_client int
	var date string
	var active bool
	var libraries []Library

	// repository
	rows, errQuery := db.Query("SELECT id, id_book, id_client, date, active FROM library")
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		errScan := rows.Scan(&id, &id_book, &id_client, &date, &active)
		if errScan != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		libraries = append(libraries, Library{Id: id, IdBook: id_book, IdClient: id_client, Date: date, Active: active})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(libraries)
}

// POST /api/libraries LibraryRequest{}
func postLibrary(w http.ResponseWriter, r *http.Request) {
	var payload LibraryRequest
	var response LibraryResponse

	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON
	if payload.IdBook == 0 || payload.IdClient == 0 || payload.Date == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	result, errQuery := db.Exec("INSERT INTO library (id_book, id_client, date, active) VALUES (?, ?, ?, ?)", payload.IdBook, payload.IdClient, payload.Date, payload.Active)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, errLII := result.LastInsertId()
	if errLII != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response = LibraryResponse{Id: int(id)}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// PUT /api/libraries/1 LibraryRequest{}
func putLibrary(w http.ResponseWriter, r *http.Request) {
	var payload Library

	vars := mux.Vars(r)
	vars_id := vars["id"]
	// validate if id == int
	int_id, errAtoi := strconv.Atoi(vars_id)
	if errAtoi != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestBody, errIO := ioutil.ReadAll(r.Body)
	if errIO != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errUnmarshal := json.Unmarshal(requestBody, &payload)
	if errUnmarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// wrong JSON or /{id}
	if payload.IdBook == 0 || payload.IdClient == 0 || payload.Date == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// repository
	_, errQuery := db.Exec("UPDATE library SET Id_book = ?, Id_client = ?, Date = ?, Active = ? WHERE Id = ?", payload.IdBook, payload.IdClient, payload.Date, payload.Active, int_id)
	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	if errAtoi != nil || int_id < 1 {
		w.WriteHeader(http.StatusBadRequest)
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
