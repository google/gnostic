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
	"log"

	"github.com/googleapis/openapi-compiler/plugins/go/openapi_go_generator/examples/bookstore/bookstore"
)

type Service struct {
	store *Store
}

func NewService() *Service {
	return &Service{store: &Store{}}
}

func (service *Service) ListShelves(responses *bookstore.ListShelvesResponses) (err error) {
	log.Printf("ListShelves")
	shelfList := service.store.ListShelves()
	response := &bookstore.ListShelvesResponse{}
	response.Shelves = shelfList.Shelves
	(*responses).OK = response
	return nil
}

func (service *Service) CreateShelf(parameters *bookstore.CreateShelfParameters, responses *bookstore.CreateShelfResponses) (err error) {
	log.Printf("CreateShelf %+v", parameters)
	shelf := service.store.CreateShelf(parameters.Shelf)
	(*responses).OK = &shelf
	return nil
}

func (service *Service) DeleteShelves() (err error) {
	log.Printf("DeleteShelves")
	service.store = &Store{}
	return nil
}

func (service *Service) GetShelf(parameters *bookstore.GetShelfParameters, responses *bookstore.GetShelfResponses) (err error) {
	log.Printf("GetShelf %+v", parameters)
	shelf, err := service.store.GetShelf(parameters.Shelf)
	(*responses).OK = &shelf
	return nil
}

func (service *Service) DeleteShelf(parameters *bookstore.DeleteShelfParameters) (err error) {
	log.Printf("DeleteShelf %+v", parameters)
	return service.store.DeleteShelf(parameters.Shelf)
}

func (service *Service) ListBooks(parameters *bookstore.ListBooksParameters, responses *bookstore.ListBooksResponses) (err error) {
	log.Printf("ListBooks %+v", parameters)
	bl, err := service.store.ListBooks(parameters.Shelf)
	response := &bookstore.ListBooksResponse{}
	response.Books = bl.Books
	(*responses).OK = response
	return nil
}

func (service *Service) CreateBook(parameters *bookstore.CreateBookParameters, responses *bookstore.CreateBookResponses) (err error) {
	log.Printf("CreateBook %+v", parameters)
	book, err := service.store.CreateBook(parameters.Shelf, parameters.Book)
	(*responses).OK = &book
	return err
}

func (service *Service) GetBook(parameters *bookstore.GetBookParameters, responses *bookstore.GetBookResponses) (err error) {
	log.Printf("GetBook %+v", parameters)
	book, err := service.store.GetBook(parameters.Shelf, parameters.Book)
	(*responses).OK = &book
	return nil
}

func (service *Service) DeleteBook(parameters *bookstore.DeleteBookParameters) (err error) {
	log.Printf("DeleteBook %+v", parameters)
	return service.store.DeleteBook(parameters.Shelf, parameters.Book)
}
