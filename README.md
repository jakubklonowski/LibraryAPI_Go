# RESTful API for Library
## Author
The author of this app is Jakub Klonowski (jakubpklonowski@gmail.com).

## Requirements
Technologies used include:
- go1.18
- go libraries:
    - github.com/go-sql-driver/mysql v1.7.0
	- github.com/gorilla/mux v1.8.0
    - github.com/rs/cors v1.8.3

## Manual
If you want to host API update config file and run file generated for your OS. Otherwise you can open project files and run `go run .` command. To consume api connect to [localhost:10000/api](localhost:10000/api) and choose subsequent endpoint.

This api was designed according to REST standard (names convention, return statuses etc).

### Endpoints & objects structs
#### /api/books - GET
    request: {

    }

    response: {
        [
            {
                Id: 0,
                Name: "",
                Author: ""
            }
        ]
    }

#### /api/books - POST
    request: {
        "Name": "",
        "Author": ""
    }

    response: {
        "Id": 0
    }

#### /api/books/{id} - GET
    request: {

    }

    response: {
        "Name": "",
        "Author": ""
    }

#### /api/books/{id} - PUT
    request: {
        "Name": "",
        "Author": ""
    }

    response: {

    }

#### /api/books/{id} - DELETE
    request: {

    }

    response: {

    }

#### /api/clients - GET
    request: {

    }

    response: {
        [
            {
                "Id": 0,
                "Name": ""
            }
        ]
    }

#### /api/clients - POST
    request: {
        "Name": ""
    }

    response: {
        "Id": 0
    }

#### /api/clients/{id} - GET
    request: {

    }

    response: {
        "Name": ""
    }

#### /api/clients/{id} - PUT
    request: {
        "Name": ""
    }

    response: {

    }

#### /api/clients/{id} - DELETE
    request: {

    }

    response: {

    }

    /api/libraries - GET, POST
    /api/libraries/{id} - GET, PUT, DELETE
