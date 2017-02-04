import Bookstore

func main() throws {


  let c = Bookstore.Client(service:"http://localhost:8080")
  let shelves = try c.listShelves()

  print("Hello")
}

try main()
