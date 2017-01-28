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

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func (s Shelf) Id() int64 {
	parts := strings.Split(s.Name, "/")
	id, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return id
}

func (b Book) Id() int64 {
	parts := strings.Split(b.Name, "/")
	id, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return id
}

type ShelfList struct {
	Shelves []Shelf
}

type BookList struct {
	Books []Book
}

// Store holds the contents of a bookstore.
type Store struct {
	Shelves     map[int64]*Shelf
	BookMaps    map[int64]map[int64]*Book
	LastShelfID int64
	LastBookID  int64
	Mutex       sync.Mutex
}

func (s *Store) checkShelvesLocked() {
	if s.Shelves == nil {
		s.Shelves = make(map[int64]*Shelf)
		s.BookMaps = make(map[int64]map[int64]*Book)
	}
}

func (s *Store) getShelfLocked(sid int64) (*Shelf, error) {
	s.checkShelvesLocked()
	shelf, ok := s.Shelves[sid]
	if !ok {
		return nil, httpErrorf(http.StatusNotFound, "Couldn't find shelf %q", sid)
	}
	return shelf, nil
}

func (s *Store) checkBooksLocked(shelf *Shelf) {
	if s.BookMaps[shelf.Id()] == nil {
		s.BookMaps[shelf.Id()] = make(map[int64]*Book)
	}
}

func getBookLocked(s *Store, sid int64, bid int64) (*Shelf, *Book, error) {
	shelf, err := s.getShelfLocked(sid)
	if err != nil {
		return nil, nil, err
	}
	s.checkBooksLocked(shelf)
	book, ok := s.BookMaps[sid][bid]
	if !ok {
		return nil, nil, httpErrorf(http.StatusNotFound, "Couldn't find book %q on shelf %q", bid, sid)
	}
	return shelf, book, nil
}

// Lists the shelves available at the store.
func (s *Store) ListShelves() ShelfList {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	sl := ShelfList{Shelves: make([]Shelf, 0, len(s.Shelves))}
	for _, shelf := range s.Shelves {
		sl.Shelves = append(sl.Shelves, *shelf)
	}

	return sl
}

// Creates a new bookstore shelf.
func (s *Store) CreateShelf(shelf Shelf) Shelf {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.LastShelfID++
	sid := s.LastShelfID
	shelf.Name = fmt.Sprintf("shelves/%d", sid)
	s.checkShelvesLocked()
	s.Shelves[sid] = &shelf
	return shelf
}

// Returns an existing bookstore shelf.
func (s *Store) GetShelf(sid int64) (Shelf, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, err := s.getShelfLocked(sid)
	if err != nil {
		return Shelf{}, err
	}
	return *shelf, nil
}

// Deletes a bookstore shelf.
func (s *Store) DeleteShelf(sid int64) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, err := s.getShelfLocked(sid); err != nil {
		return err
	}
	delete(s.Shelves, sid)
	return nil
}

// Lists the books on a bookstore shelf.
func (s *Store) ListBooks(sid int64) (BookList, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, err := s.getShelfLocked(sid)
	if err != nil {
		return BookList{}, err
	}

	shelfBooks := s.BookMaps[shelf.Id()]

	bl := BookList{Books: make([]Book, 0, len(shelfBooks))}
	for _, book := range shelfBooks {
		bl.Books = append(bl.Books, *book)
	}

	return bl, nil
}

// Creates a new book on a bookstore shelf.
func (s *Store) CreateBook(sid int64, book Book) (Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	shelf, err := s.getShelfLocked(sid)
	if err != nil {
		return Book{}, err
	}

	s.LastBookID++
	bid := s.LastBookID
	book.Name = fmt.Sprintf("%s/books/%d", shelf.Name, bid)
	s.checkBooksLocked(shelf)
	s.BookMaps[sid][bid] = &book

	log.Printf("CREATED AND SAVED BOOK %+v in shelf %+v", book, s.BookMaps[shelf.Id()])
	return book, nil
}

// Returns an existing book from a bookstore shelf.
func (s *Store) GetBook(sid int64, bid int64) (Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	_, book, err := getBookLocked(s, sid, bid)
	if err != nil {
		return Book{}, err
	}

	return *book, nil
}

// Deletes a book from a bookstore shelf.
func (s *Store) DeleteBook(sid int64, bid int64) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	delete(s.BookMaps[sid], bid)
	return nil
}
