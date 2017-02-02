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

type BookList struct {
	Books []bookstore.Book
}

type Service struct {
	Shelves     map[int64]*bookstore.Shelf
	BookMaps    map[int64]map[int64]*bookstore.Book
	LastShelfID int64
	LastBookID  int64
	Mutex       sync.Mutex
}

func NewService() *Service {
	return &Service{
		Shelves:  make(map[int64]*bookstore.Shelf),
		BookMaps: make(map[int64]map[int64]*bookstore.Book),
	}
}

func (service *Service) ListShelves(responses *bookstore.ListShelvesResponses) (err error) {
	log.Printf("ListShelves")
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	shelves := make([]bookstore.Shelf, 0, len(service.Shelves))
	for _, shelf := range service.Shelves {
		shelves = append(shelves, *shelf)
	}
	response := &bookstore.ListShelvesResponse{}
	response.Shelves = shelves
	(*responses).OK = response
	return err
}

func (service *Service) CreateShelf(parameters *bookstore.CreateShelfParameters, responses *bookstore.CreateShelfResponses) (err error) {
	log.Printf("CreateShelf %+v", parameters)
	shelf := parameters.Shelf
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	service.LastShelfID++
	sid := service.LastShelfID
	shelf.Name = fmt.Sprintf("shelves/%d", sid)
	service.Shelves[sid] = &shelf
	(*responses).OK = &shelf
	return err
}

func (service *Service) DeleteShelves() (err error) {
	log.Printf("DeleteShelves")
	service.Shelves = make(map[int64]*bookstore.Shelf)
	service.BookMaps = make(map[int64]map[int64]*bookstore.Book)
	service.LastShelfID = 0
	service.LastBookID = 0
	return nil
}

func (service *Service) GetShelf(parameters *bookstore.GetShelfParameters, responses *bookstore.GetShelfResponses) (err error) {
	log.Printf("GetShelf %+v", parameters)
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	shelf, status, err := service.getShelfLocked(parameters.Shelf)
	if err != nil {
		handlerError := &bookstore.Error{}
		handlerError.Code = int32(status)
		handlerError.Message = err.Error()
		(*responses).Default = handlerError
		return nil
	} else {
		(*responses).OK = shelf
		return nil
	}
}

func (service *Service) DeleteShelf(parameters *bookstore.DeleteShelfParameters) (err error) {
	log.Printf("DeleteShelf %+v", parameters)
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	if _, _, err := service.getShelfLocked(parameters.Shelf); err != nil {
		return err
	}
	delete(service.Shelves, parameters.Shelf)
	return nil
}

func (service *Service) ListBooks(parameters *bookstore.ListBooksParameters, responses *bookstore.ListBooksResponses) (err error) {
	log.Printf("ListBooks %+v", parameters)
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	shelf, _, err := service.getShelfLocked(parameters.Shelf)
	if err != nil {
		return err
	}
	shelfBooks := service.BookMaps[shelf_id(shelf)]
	bookList := make([]bookstore.Book, 0, len(shelfBooks))
	for _, book := range shelfBooks {
		bookList = append(bookList, *book)
	}
	response := &bookstore.ListBooksResponse{}
	response.Books = bookList
	(*responses).OK = response
	return nil
}

func (service *Service) CreateBook(parameters *bookstore.CreateBookParameters, responses *bookstore.CreateBookResponses) (err error) {
	log.Printf("CreateBook %+v", parameters)
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	shelf, _, err := service.getShelfLocked(parameters.Shelf)
	if err != nil {
		return err
	}
	service.LastBookID++
	bid := service.LastBookID
	book := parameters.Book
	book.Name = fmt.Sprintf("%s/books/%d", shelf.Name, bid)
	if service.BookMaps[shelf_id(shelf)] == nil {
		service.BookMaps[shelf_id(shelf)] = make(map[int64]*bookstore.Book)
	}
	service.BookMaps[parameters.Shelf][bid] = &book
	log.Printf("CREATED AND SAVED BOOK %+v in shelf %+v", book, service.BookMaps[shelf_id(shelf)])
	(*responses).OK = &book
	return err
}

func (service *Service) GetBook(parameters *bookstore.GetBookParameters, responses *bookstore.GetBookResponses) (err error) {
	log.Printf("GetBook %+v", parameters)
	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	_, book, status, err := service.getBookLocked(parameters.Shelf, parameters.Book)
	if err != nil {
		return err
	}
	if err != nil {
		handlerError := &bookstore.Error{}
		handlerError.Code = int32(status)
		handlerError.Message = err.Error()
		(*responses).Default = handlerError
		return nil
	} else {
		(*responses).OK = book
		return nil
	}
}

func (service *Service) DeleteBook(parameters *bookstore.DeleteBookParameters) (err error) {
	log.Printf("DeleteBook %+v", parameters)

	service.Mutex.Lock()
	defer service.Mutex.Unlock()
	delete(service.BookMaps[parameters.Shelf], parameters.Book)

	return nil
}

// helpers

func (service *Service) getShelfLocked(sid int64) (shelf *bookstore.Shelf, status int, err error) {
	shelf, ok := service.Shelves[sid]
	if !ok {
		return nil, status, errors.New(fmt.Sprintf("Couldn't find shelf %q", sid))
	}
	return shelf, http.StatusOK, nil
}

func (service *Service) getBookLocked(sid int64, bid int64) (shelf *bookstore.Shelf, book *bookstore.Book, status int, err error) {
	shelf, status, err = service.getShelfLocked(sid)
	if err != nil {
		return nil, nil, status, err
	}
	book, ok := service.BookMaps[sid][bid]
	if !ok {
		return nil, nil, http.StatusNotFound, errors.New(fmt.Sprintf("Couldn't find book %q on shelf %q", bid, sid))
	}
	return shelf, book, status, nil
}
