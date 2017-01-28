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
	"log"
	"testing"

	"github.com/googleapis/openapi-compiler/plugins/go/examples/bookstore/bookstore"
)

const service = "http://localhost:8080"

//const service = "http://generated-bookstore.appspot.com"

func TestBookstore(t *testing.T) {
	{
		err := bookstore.DeleteShelves()
		if err != nil {
			panic(err)
		}
	}
	{
		response, err := bookstore.ListShelves()
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 0 {
			t.Fail()
		}
	}
	{
		var shelf bookstore.Shelf
		shelf.Theme = "mysteries"
		response, err := bookstore.CreateShelf(shelf)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
	}
	{
		var shelf bookstore.Shelf
		shelf.Theme = "comedies"
		response, err := bookstore.CreateShelf(shelf)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
	}
	{
		response, err := bookstore.GetShelf(1)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
	}
	{
		response, err := bookstore.ListShelves()
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 2 {
			t.Fail()
		}
	}

	{
		err := bookstore.DeleteShelf(2)
		if err != nil {
			panic(err)
		}
	}

	{
		response, err := bookstore.ListShelves()
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 1 {
			t.Fail()
		}
	}

	{
		response, err := bookstore.ListBooks(1)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Books) != 0 {
			t.Fail()
		}
	}

	{
		var book bookstore.Book
		book.Author = "Agatha Christie"
		book.Title = "And Then There Were None"
		response, err := bookstore.CreateBook(1, book)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
	}

	{
		var book bookstore.Book
		book.Author = "Agatha Christie"
		book.Title = "Murder on the Orient Express"
		response, err := bookstore.CreateBook(1, book)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
	}

	{
		book, err := bookstore.GetBook(1, 1)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", book)
	}

	{
		response, err := bookstore.ListBooks(1)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Books) != 2 {
			t.Fail()
		}
	}

	{
		err := bookstore.DeleteBook(1, 2)
		if err != nil {
			panic(err)
		}
	}

	{
		response, err := bookstore.ListBooks(1)
		if err != nil {
			panic(err)
		}
		log.Printf("%+v", response)
		if len(response.Books) != 1 {
			t.Fail()
		}
	}
}
