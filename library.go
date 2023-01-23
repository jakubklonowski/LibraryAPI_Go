package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
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

// STRUCT --------------------------------------------------------------------------

type Book struct {
	Id     int
	Name   string
	Author string
}

type Client struct {
	Id   int
	Name string
}

type Library struct {
	Id       int
	IdBook   int
	IdClient int
	Date     string
	Active   bool
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
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", putBook).Methods("PUT")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELTE")
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

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func getBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println(id)
	json.NewEncoder(w).Encode(Book{Id: 1, Name: "Atlas Shrugged", Author: "Ayn Rand"})
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	Books := []Book{
		{Id: 1, Name: "Atlas Shrugged", Author: "Ayn Rand"},
		{Id: 2, Name: "Don Quixote", Author: "Miguel de Cervantes"},
	}
	json.NewEncoder(w).Encode(Books)
}

func putBook(w http.ResponseWriter, r *http.Request) {

}

func deleteBook(w http.ResponseWriter, r *http.Request) {

}
