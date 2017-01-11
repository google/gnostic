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

package main

var s *Store

func initialize_service() {
	s = &Store{}
}

func processListShelves(request *Empty, response *ListShelvesResponse) (err error) {
	sl := s.ListShelves()
	(*response).Shelves = sl.Shelves
	return nil
}

func processDeleteShelves(request *Empty, response *Value) (err error) {
	initialize_service()
	return nil
}

func processCreateShelf(request *CreateShelfRequest, response *Shelf) (err error) {
	*response = s.CreateShelf(request.Shelf)
	return nil
}

func processGetShelf(request *GetShelfRequest, response *Shelf) (err error) {
	*response, err = s.GetShelf(request.Shelf)
	return err
}

func processDeleteShelf(request *DeleteShelfRequest, response *Value) (err error) {
	return s.DeleteShelf(request.Shelf)
}

func processListBooks(request *ListBooksRequest, response *ListBooksResponse) (err error) {
	bl, err := s.ListBooks(request.Shelf)
	(*response).Books = bl.Books
	return err
}

func processCreateBook(request *CreateBookRequest, response *Book) (err error) {
	*response, err = s.CreateBook(request.Shelf, request.Book)
	return err
}

func processGetBook(request *GetBookRequest, response *Book) (err error) {
	*response, err = s.GetBook(request.Shelf, request.Book)
	return err
}

func processDeleteBook(request *DeleteBookRequest, response *Value) (err error) {
	return s.DeleteBook(request.Shelf, request.Book)
}
