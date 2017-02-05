import Bookstore

class Server : Service {
  // Return all shelves in the bookstore.
  func listShelves () throws -> ListShelvesResponses {
	  return ListShelvesResponses()
  }
  // Create a new shelf in the bookstore.
  func createShelf (parameters : CreateShelfParameters) throws -> CreateShelfResponses {
	  return CreateShelfResponses()
  }
  // Delete all shelves.
  func deleteShelves () throws {
  	
  }
  // Get a single shelf resource with the given ID.
  func getShelf (parameters : GetShelfParameters) throws -> GetShelfResponses {
  	return GetShelfResponses()
  }
  // Delete a single shelf with the given ID.
  func deleteShelf (parameters : DeleteShelfParameters) throws {
  	
  }
  // Return all books in a shelf with the given ID.
  func listBooks (parameters : ListBooksParameters) throws -> ListBooksResponses {
	  return ListBooksResponses()
  }
  // Create a new book on the shelf.
  func createBook (parameters : CreateBookParameters) throws -> CreateBookResponses {
	  return CreateBookResponses()
  }
  // Get a single book with a given ID from a shelf.
  func getBook (parameters : GetBookParameters) throws -> GetBookResponses {
	  return GetBookResponses()
  }
  // Delete a single book with a given ID from a shelf.
  func deleteBook (parameters : DeleteBookParameters) throws {
  	
  }
}

initialize(service:Server(), port:8090)

run()

