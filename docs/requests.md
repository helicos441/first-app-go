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