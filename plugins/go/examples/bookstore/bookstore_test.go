/*
 Copyright 2016 Google Inc. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
)

const service = "http://localhost:8080"
//const service = "http://generated-bookstore.appspot.com"

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

func ListShelves() (response *ListShelvesResponse) {
	path := fmt.Sprintf("%s/shelves", service)
	resp, err := http.Get(path)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &ListShelvesResponse{}
	decoder.Decode(response)
	return response
}

func DeleteShelves() (response *Value) {
	path := fmt.Sprintf("%s/shelves", service)
	req, err := http.NewRequest("DELETE", path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Value{}
	decoder.Decode(response)
	return response
}

func CreateShelf(request *CreateShelfRequest) (response *Shelf) {
	path := fmt.Sprintf("%s/shelves", service)
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(request)
	resp, err := http.Post(path, "application/json", body)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Shelf{}
	decoder.Decode(response)
	return response
}

func GetShelf(request *GetShelfRequest) (response *Shelf) {
	path := fmt.Sprintf("%s/shelves/%d", service, request.Shelf)
	resp, err := http.Get(path)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Shelf{}
	decoder.Decode(response)
	return response
}

func DeleteShelf(request *DeleteShelfRequest) (response *Value) {
	path := fmt.Sprintf("%s/shelves/%d", service, request.Shelf)
	req, err := http.NewRequest("DELETE", path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Value{}
	decoder.Decode(response)
	return response
}

func ListBooks(request *ListBooksRequest) (response *ListBooksResponse) {
	path := fmt.Sprintf("%s/shelves/%d/books", service, request.Shelf)
	resp, err := http.Get(path)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &ListBooksResponse{}
	decoder.Decode(response)
	return response
}

func CreateBook(request *CreateBookRequest) (response *Book) {
	path := fmt.Sprintf("%s/shelves/%d/books", service, request.Shelf)
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(request)
	resp, err := http.Post(path, "application/json", body)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Book{}
	decoder.Decode(response)
	return response
}

func GetBook(request *GetBookRequest) (response *Book) {
	path := fmt.Sprintf("%s/shelves/%d/books/%d", service, request.Shelf, request.Book)
	resp, err := http.Get(path)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Book{}
	decoder.Decode(response)
	return response
}

func DeleteBook(request *DeleteBookRequest) (response *Value) {
	path := fmt.Sprintf("%s/shelves/%d/books/%d", service, request.Shelf, request.Book)
	req, err := http.NewRequest("DELETE", path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	response = &Value{}
	decoder.Decode(response)
	return response
}

func TestBookstore(t *testing.T) {
	{
		response := DeleteShelves()
		log.Printf("%+v", response)
	}

	{
		response := ListShelves()
		log.Printf("%+v", response)
		if len(response.Shelves) != 0 {
			t.Fail()
		}
	}

	{
		var request CreateShelfRequest
		request.Shelf.Theme = "mysteries"
		response := CreateShelf(&request)
		log.Printf("%+v", response)
	}

	{
		var request CreateShelfRequest
		request.Shelf.Theme = "comedies"
		response := CreateShelf(&request)
		log.Printf("%+v", response)
	}

	{
		var request GetShelfRequest
		request.Shelf = 1
		response := GetShelf(&request)
		log.Printf("%+v", response)
	}

	{
		response := ListShelves()
		log.Printf("%+v", response)
		if len(response.Shelves) != 2 {
			t.Fail()
		}
	}

	{
		var request DeleteShelfRequest
		request.Shelf = 2
		response := DeleteShelf(&request)
		log.Printf("%+v", response)
	}

	{
		response := ListShelves()
		log.Printf("%+v", response)
		if len(response.Shelves) != 1 {
			t.Fail()
		}
	}

	{
		var request ListBooksRequest
		request.Shelf = 1
		response := ListBooks(&request)
		log.Printf("%+v", response)
		if len(response.Books) != 0 {
			t.Fail()
		}
	}

	{
		var request CreateBookRequest
		request.Shelf = 1
		request.Book.Author = "Agatha Christie"
		request.Book.Title = "And Then There Were None"
		response := CreateBook(&request)
		log.Printf("%+v", response)
	}

	{
		var request CreateBookRequest
		request.Shelf = 1
		request.Book.Author = "Agatha Christie"
		request.Book.Title = "Murder on the Orient Express"
		response := CreateBook(&request)
		log.Printf("%+v", response)
	}

	{
		var request GetBookRequest
		request.Shelf = 1
		request.Book = 1
		response := GetBook(&request)
		log.Printf("%+v", response)
	}

	{
		var request ListBooksRequest
		request.Shelf = 1
		response := ListBooks(&request)
		log.Printf("%+v", response)
		if len(response.Books) != 2 {
			t.Fail()
		}
	}

	{
		var request DeleteBookRequest
		request.Shelf = 1
		request.Book = 2
		response := DeleteBook(&request)
		log.Printf("%+v", response)
	}

	{
		var request ListBooksRequest
		request.Shelf = 1
		response := ListBooks(&request)
		log.Printf("%+v", response)
		if len(response.Books) != 1 {
			t.Fail()
		}
	}

}
