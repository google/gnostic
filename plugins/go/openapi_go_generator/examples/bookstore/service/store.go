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

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator/examples/bookstore/bookstore"
)

func shelf_id(s *bookstore.Shelf) int64 {
	parts := strings.Split(s.Name, "/")
	id, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return id
}

func book_id(b *bookstore.Book) int64 {
	parts := strings.Split(b.Name, "/")
	id, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return id
}

type ShelfList struct {
	Shelves []bookstore.Shelf
}

type BookList struct {
	Books []bookstore.Book
}

// Store holds the contents of a bookstore.
type Store struct {
	Shelves     map[int64]*bookstore.Shelf
	BookMaps    map[int64]map[int64]*bookstore.Book
	LastShelfID int64
	LastBookID  int64
	Mutex       sync.Mutex
}

func (s *Store) checkShelvesLocked() (status int, err error) {
	if s.Shelves == nil {
		s.Shelves = make(map[int64]*bookstore.Shelf)
		s.BookMaps = make(map[int64]map[int64]*bookstore.Book)
	}
	return 200, nil
}

func (s *Store) getShelfLocked(sid int64) (shelf *bookstore.Shelf, status int, err error) {
	s.checkShelvesLocked()
	shelf, ok := s.Shelves[sid]
	if !ok {
		return nil, status, errors.New(fmt.Sprintf("Couldn't find shelf %q", sid))
	}
	return shelf, http.StatusOK, nil
}

func (s *Store) checkBooksLocked(shelf *bookstore.Shelf) (status int, err error) {
	if s.BookMaps[shelf_id(shelf)] == nil {
		s.BookMaps[shelf_id(shelf)] = make(map[int64]*bookstore.Book)
	}
	return http.StatusOK, nil
}

func getBookLocked(s *Store, sid int64, bid int64) (shelf *bookstore.Shelf, book *bookstore.Book, status int, err error) {
	shelf, status, err = s.getShelfLocked(sid)
	if err != nil {
		return nil, nil, status, err
	}
	s.checkBooksLocked(shelf)
	book, ok := s.BookMaps[sid][bid]
	if !ok {
		return nil, nil, http.StatusNotFound, errors.New(fmt.Sprintf("Couldn't find book %q on shelf %q", bid, sid))
	}
	return shelf, book, status, nil
}

// Lists the shelves available at the store.
func (s *Store) ListShelves() (shelflist ShelfList, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	sl := ShelfList{Shelves: make([]bookstore.Shelf, 0, len(s.Shelves))}
	for _, shelf := range s.Shelves {
		sl.Shelves = append(sl.Shelves, *shelf)
	}

	return sl, http.StatusOK, nil
}

// Creates a new bookstore shelf.
func (s *Store) CreateShelf(shelf bookstore.Shelf) (newShelf *bookstore.Shelf, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.LastShelfID++
	sid := s.LastShelfID
	shelf.Name = fmt.Sprintf("shelves/%d", sid)
	s.checkShelvesLocked()
	s.Shelves[sid] = &shelf
	return &shelf, http.StatusOK, nil
}

// Returns an existing bookstore shelf.
func (s *Store) GetShelf(sid int64) (shelf *bookstore.Shelf, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, status, err = s.getShelfLocked(sid)
	if err != nil {
		return &bookstore.Shelf{}, status, err
	}
	return shelf, status, nil
}

// Deletes a bookstore shelf.
func (s *Store) DeleteShelf(sid int64) (status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, status, err := s.getShelfLocked(sid); err != nil {
		return status, err
	}
	delete(s.Shelves, sid)
	return http.StatusOK, nil
}

// Lists the books on a bookstore shelf.
func (s *Store) ListBooks(sid int64) (bookList BookList, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, status, err := s.getShelfLocked(sid)
	if err != nil {
		return BookList{}, status, err
	}

	shelfBooks := s.BookMaps[shelf_id(shelf)]

	bl := BookList{Books: make([]bookstore.Book, 0, len(shelfBooks))}
	for _, book := range shelfBooks {
		bl.Books = append(bl.Books, *book)
	}

	return bl, status, nil
}

// Creates a new book on a bookstore shelf.
func (s *Store) CreateBook(sid int64, book bookstore.Book) (newBook bookstore.Book, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, status, err := s.getShelfLocked(sid)
	if err != nil {
		return bookstore.Book{}, status, err
	}

	s.LastBookID++
	bid := s.LastBookID
	book.Name = fmt.Sprintf("%s/books/%d", shelf.Name, bid)
	s.checkBooksLocked(shelf)
	s.BookMaps[sid][bid] = &book

	log.Printf("CREATED AND SAVED BOOK %+v in shelf %+v", book, s.BookMaps[shelf_id(shelf)])
	return book, status, nil
}

// Returns an existing book from a bookstore shelf.
func (s *Store) GetBook(sid int64, bid int64) (book *bookstore.Book, status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	_, book, status, err = getBookLocked(s, sid, bid)
	if err != nil {
		return &bookstore.Book{}, status, err
	}

	return book, status, nil
}

// Deletes a book from a bookstore shelf.
func (s *Store) DeleteBook(sid int64, bid int64) (status int, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	delete(s.BookMaps[sid], bid)
	return http.StatusOK, nil
}
