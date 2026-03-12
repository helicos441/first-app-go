### Check the app health
```bash
curl -i -X GET http://localhost:8080/healthz
```

### List the books
```bash
curl -i -X GET http://localhost:8080/books
```

### Get a book by id
```bash
curl -i -X GET http://localhost:8080/books/1
```

### Create a new book
```bash
curl -i -X POST http://localhost:8080/books \
-H "Content-Type: application/json" \
-d '{"title":"The Go Work","author":"Test","year":2025}'
```