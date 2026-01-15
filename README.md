# Assignment2 - Tasks API

Simple Tasks REST API using Gin and MongoDB.

Basic features:

- POST /tasks - create task (validates title, completed defaults to false) -> returns 201 with created resource
- GET /tasks - list tasks
- GET /tasks/:id - get specific task -> 404 if not found
- PUT /tasks/:id - update existing task (validates title if provided) -> returns updated resource
- DELETE /tasks/:id - delete task -> 404 if not found

Requirements:

- Go 1.20+
- MongoDB and env var MONGO_URI set (e.g. export MONGO_URI="mongodb://localhost:27017")

Run:

- export MONGO_URI="your-mongo-uri"
- go run main.go

Notes:

- API performs input validation and returns appropriate HTTP status codes (400/404/422/500).
- Graceful shutdown implemented; DB client is disconnected on exit.
