import XCTest
import Foundation
@testable import Bookstore

func Log(_ message : String) {
  FileHandle.standardError.write((message + "\n").data(using:.utf8)!)
}

let service = "http://localhost:8080"

class BookstoreTests: XCTestCase {
  
  func testBasic() {
    // create a client
    let b = Bookstore.Client(service:service)    
    // reset the service by deleting all shelves
    do {
        try b.deleteShelves()       
    } catch (let error) {
        XCTFail("\(error)")                 
    }    
    // verify that the service has no shelves
    do {
        let response = try b.listShelves()  
        XCTAssertEqual(response.shelves.count, 0)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // attempting to get a shelf should return an error
    do {
        let _ = try b.getShelf(shelf:1)
        XCTFail("server error")             
    } catch (let error) {
        print("\(error)")   
    }
    // attempting to get a book should return an error
    do { 
        let _ = try b.getBook(shelf:1, book:2)
    } catch (let error) {
        print("\(error)")   
    }       
    // add a shelf
    do {
        var shelf = Shelf()
        shelf.theme = "mysteries"
        let response = try b.createShelf(shelf:shelf)
        if (response.name != "shelves/1") ||
            (response.theme != "mysteries") {
             XCTFail("mismatch")                    
        }
    } catch (let error) {
        XCTFail("\(error)")                 
    }
   // add another shelf
   do {
       var shelf = Shelf()
       shelf.theme = "comedies"
       let response = try b.createShelf(shelf:shelf)
       if (response.name != "shelves/2") ||
           (response.theme != "comedies") {
            XCTFail("mismatch")                    
       }
   } catch (let error) {
       XCTFail("\(error)")                 
   }
    // get the first shelf that was added
    do {
        let response = try b.getShelf(shelf:1)
        if (response.name != "shelves/1") ||
            (response.theme != "mysteries") {
             XCTFail("mismatch")                    
        }
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // list shelves and verify that there are 2
    do {
        let response = try b.listShelves()  
        XCTAssertEqual(response.shelves.count, 2)
    } catch (let error) {
        XCTFail("\(error)")                 
    }	
    // delete a shelf
    do {
        try b.deleteShelf(shelf:2)       
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // list shelves and verify that there is only 1
    do {
        let response = try b.listShelves()  
        XCTAssertEqual(response.shelves.count, 1)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // list books on a shelf, verify that there are none
    do {
        let response = try b.listBooks(shelf:1)  
        XCTAssertEqual(response.books.count, 0)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // create a book
    do {
        var book = Book()
        book.author = "Agatha Christie"
        book.title = "And Then There Were None"
        let _ = try b.createBook(shelf:1, book:book)
    } catch (let error) {
        XCTFail("\(error)")                 
    }	
    // create another book
    do {
        var book = Book()
        book.author = "Agatha Christie"
        book.title = "Murder on the Orient Express"
        let _ = try b.createBook(shelf:1, book:book)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // get the first book that was added
    do {
        let response = try b.getBook(shelf:1, book:1)
        if (response.author != "Agatha Christie") ||
            (response.title != "And Then There Were None") {
             XCTFail("mismatch")                    
        }
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // list the books on a shelf and verify that there are 2
    do {
        let response = try b.listBooks(shelf:1)  
        XCTAssertEqual(response.books.count, 2)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // delete a book
    do {
        try b.deleteBook(shelf:1, book:2)       
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // list the books on a shelf and verify that is only 1
    do {
        let response = try b.listBooks(shelf:1)  
        XCTAssertEqual(response.books.count, 1)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
    // verify the handling of a badly-formed request
    do {
  	  var path = service 
  	  path = path + "/shelves"
  	  let url = URL(string:path)
  	  var request = URLRequest(url:url!)
  	  request.httpMethod = "POST"
  	  request.httpBody = "".data(using:.utf8)
  	  let (data, response, error) = URLSession.shared.fetch(request)	  
      // we expect a 400 (Bad Request) code  	  
      XCTAssertEqual((response as! HTTPURLResponse).statusCode, 400)
    } catch (let error) {
        XCTFail("\(error)")                 
    }
  }
}

extension BookstoreTests {
  static var allTests : [(String, (BookstoreTests) -> () throws -> Void)] {
    return [
      ("testBasic", testBasic),
    ]
  }
}
