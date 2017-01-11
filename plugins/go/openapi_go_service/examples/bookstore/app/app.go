package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func intValue(s string) (v int64) {
	v, _ = strconv.ParseInt(s, 10, 64)
	return v
}

type Book struct {
	Author string `json:"author"`
	Name   string `json:"name"`
	Title  string `json:"title"`
}

type ListBooksResponse struct {
	Books []Book `json:"books"`
}

type Shelf struct {
	Name  string `json:"name"`
	Theme string `json:"theme"`
}

type ListShelvesResponse struct {
	Shelves []Shelf `json:"shelves"`
}

type CreateShelfRequest struct {
	Shelf Shelf `json:"shelf"`
}

type GetShelfRequest struct {
	Shelf int64 `json:"shelf"`
}

type DeleteShelfRequest struct {
	Shelf int64 `json:"shelf"`
}

type ListBooksRequest struct {
	Shelf int64 `json:"shelf"`
}

type CreateBookRequest struct {
	Shelf int64 `json:"shelf"`
	Book  Book  `json:"book"`
}

type GetBookRequest struct {
	Shelf int64 `json:"shelf"`
	Book  int64 `json:"book"`
}

type DeleteBookRequest struct {
	Shelf int64 `json:"shelf"`
	Book  int64 `json:"book"`
}

type ListValue struct {
	Values []Value `json:"values"`
}

type Struct struct {
	Fields []FieldsEntry `json:"fields"`
}

type FieldsEntry struct {
	Key   string `json:"key"`
	Value Value  `json:"value"`
}

type Empty struct {
}

type Value struct {
	Null_value   int       `json:"nullValue"`
	Number_value float64   `json:"numberValue"`
	String_value string    `json:"stringValue"`
	Bool_value   bool      `json:"boolValue"`
	Struct_value Struct    `json:"structValue"`
	List_value   ListValue `json:"listValue"`
}

//
func handleListShelves(w http.ResponseWriter, r *http.Request) {
	log.Printf("ListShelves")
	var err error

	// instantiate the request type
	var request Empty

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	// instantiate the response type
	var response ListShelvesResponse

	// call the processor
	err = processListShelves(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleDeleteShelves(w http.ResponseWriter, r *http.Request) {
	log.Printf("DeleteShelves")
	var err error

	// instantiate the request type
	var request Empty

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	// instantiate the response type
	var response Value

	// call the processor
	err = processDeleteShelves(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleCreateShelf(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateShelf")
	var err error

	// instantiate the request type
	var request CreateShelfRequest

	// deserialize request from post data
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}
	log.Printf("REQUEST %+v", request)

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	// instantiate the response type
	var response Shelf

	// call the processor
	err = processCreateShelf(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleGetShelf(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetShelf")
	var err error

	// instantiate the request type
	var request GetShelfRequest

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	// instantiate the response type
	var response Shelf

	// call the processor
	err = processGetShelf(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleDeleteShelf(w http.ResponseWriter, r *http.Request) {
	log.Printf("DeleteShelf")
	var err error

	// instantiate the request type
	var request DeleteShelfRequest

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	// instantiate the response type
	var response Value

	// call the processor
	err = processDeleteShelf(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleListBooks(w http.ResponseWriter, r *http.Request) {
	log.Printf("ListBooks")
	var err error

	// instantiate the request type
	var request ListBooksRequest

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	// instantiate the response type
	var response ListBooksResponse

	// call the processor
	err = processListBooks(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleCreateBook(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateBook")
	var err error

	// instantiate the request type
	var request CreateBookRequest

	// deserialize request from post data
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}
	log.Printf("REQUEST %+v", request)

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	// instantiate the response type
	var response Book

	// call the processor
	err = processCreateBook(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleGetBook(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetBook")
	var err error

	// instantiate the request type
	var request GetBookRequest

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	if value, ok := vars["book"]; ok {
		request.Book = intValue(value)
	} else if len(r.Form["book"]) > 0 {
		request.Book = intValue(r.Form["book"][0])
	}

	// instantiate the response type
	var response Book

	// call the processor
	err = processGetBook(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

//
func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	log.Printf("DeleteBook")
	var err error

	// instantiate the request type
	var request DeleteBookRequest

	// get request fields in path and query parameters
	vars := mux.Vars(r)
	r.ParseForm()
	log.Printf("MUX VARS %+v", vars)
	log.Printf("QUERY PARMS %+v", r.Form)

	if value, ok := vars["shelf"]; ok {
		request.Shelf = intValue(value)
	} else if len(r.Form["shelf"]) > 0 {
		request.Shelf = intValue(r.Form["shelf"][0])
	}

	if value, ok := vars["book"]; ok {
		request.Book = intValue(value)
	} else if len(r.Form["book"]) > 0 {
		request.Book = intValue(r.Form["book"][0])
	}

	// instantiate the response type
	var response Value

	// call the processor
	err = processDeleteBook(&request, &response)

	if err == nil {
		// serialize the response
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	} else {
		fmt.Fprintf(w, "ERROR: %v", err)
	}
}

func init() {
	var router = mux.NewRouter()

	router.HandleFunc("/shelves", handleListShelves).Methods("GET")
	router.HandleFunc("/shelves", handleDeleteShelves).Methods("DELETE")
	router.HandleFunc("/shelves", handleCreateShelf).Methods("POST")
	router.HandleFunc("/shelves/{shelf}", handleGetShelf).Methods("GET")
	router.HandleFunc("/shelves/{shelf}", handleDeleteShelf).Methods("DELETE")
	router.HandleFunc("/shelves/{shelf}/books", handleListBooks).Methods("GET")
	router.HandleFunc("/shelves/{shelf}/books", handleCreateBook).Methods("POST")
	router.HandleFunc("/shelves/{shelf}/books/{book}", handleGetBook).Methods("GET")
	router.HandleFunc("/shelves/{shelf}/books/{book}", handleDeleteBook).Methods("DELETE")
	http.Handle("/", router)

	initialize_service()
}

func main() {
	http.ListenAndServe(":8080", nil)
}
