/*
 Copyright 2017 Google Inc. All Rights Reserved.

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
	// create a client
	b := bookstore.NewClient(service)
	// reset the service by deleting all shelves
	{
		err := b.DeleteShelves()
		if err != nil {
			t.Fail()
		}
	}
	// verify that the service has no shelves
	{
		response, err := b.ListShelves()
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 0 {
			t.Fail()
		}
	}
	// add a shelf
	{
		var shelf bookstore.Shelf
		shelf.Theme = "mysteries"
		response, err := b.CreateShelf(shelf)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
	}
	// add another shelf
	{
		var shelf bookstore.Shelf
		shelf.Theme = "comedies"
		response, err := b.CreateShelf(shelf)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
	}
	// get the first shelf that was added
	{
		response, err := b.GetShelf(1)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
	}
	// list shelves and verify that there are 2
	{
		response, err := b.ListShelves()
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 2 {
			t.Fail()
		}
	}
	// delete a shelf
	{
		err := b.DeleteShelf(2)
		if err != nil {
			t.Fail()
		}
	}
	// list shelves and verify that there is only 1
	{
		response, err := b.ListShelves()
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Shelves) != 1 {
			t.Fail()
		}
	}
	// list books on a shelf, verify that there are none
	{
		response, err := b.ListBooks(1)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Books) != 0 {
			t.Fail()
		}
	}
	// create a book
	{
		var book bookstore.Book
		book.Author = "Agatha Christie"
		book.Title = "And Then There Were None"
		response, err := b.CreateBook(1, book)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
	}
	// create another book
	{
		var book bookstore.Book
		book.Author = "Agatha Christie"
		book.Title = "Murder on the Orient Express"
		response, err := b.CreateBook(1, book)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
	}
	// get the first book that was added
	{
		book, err := b.GetBook(1, 1)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", book)
	}
	// list the books on a shelf and verify that there are 2
	{
		response, err := b.ListBooks(1)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Books) != 2 {
			t.Fail()
		}
	}
	// delete a book
	{
		err := b.DeleteBook(1, 2)
		if err != nil {
			t.Fail()
		}
	}
	// list the books on a shelf and verify that is only 1
	{
		response, err := b.ListBooks(1)
		if err != nil {
			t.Fail()
		}
		log.Printf("%+v", response)
		if len(response.Books) != 1 {
			t.Fail()
		}
	}
}
