# Students API

A simple REST API to manage student records (Create, Read, Update, Delete), built in Go with a PostgreSQL database.

This project is a good example of a clean, minimal Go backend: no heavy framework, just the standard library's HTTP server plus a few small, well-known packages.

---

## Tech Stack

| Piece | What it does |
|---|---|
| **Go (net/http)** | The web server itself. Go 1.22+ added routing with HTTP methods and path parameters (`{id}`) directly into `net/http`, so no external router library is needed. |
| **PostgreSQL** | The database that stores student records. |
| **pgx** | The Go driver used to talk to PostgreSQL. |
| **godotenv** | Loads configuration (like the database URL) from a `.env` file instead of hardcoding it. |
| **go-playground/validator** | Validates incoming JSON (e.g. "email must be a valid email", "age must be greater than 0"). |
| **log/slog** | Go's built-in structured logger, used to print readable logs (`INFO`, `ERROR`) to the terminal. |

No frontend, no framework — just Go talking to Postgres over HTTP.

---

## Project Structure

```
students-api/
├── cmd/
│   └── students-api/
│       └── main.go              # Entry point — starts the server
├── internal/
│   ├── config/
│   │   └── config.go             # Loads settings from .env
│   ├── http/
│   │   └── handlers/
│   │       └── student/
│   │           └── student.go    # HTTP handlers (Create, GetById, GetList, Update, Delete)
│   ├── logger/
│   │   └── logger.go             # Sets up logging
│   ├── storage/
│   │   ├── storage.go            # Defines the Storage interface + shared errors
│   │   └── postgres/
│   │       └── postgres.go       # Postgres implementation of Storage
│   ├── types/
│   │   └── types.go              # Data shapes: Student, StudentRequest
│   └── utils/
│       └── response/
│           └── response.go       # Helper for writing consistent JSON responses
├── .env                           # Your real secrets (never committed to git)
├── .env.example                   # Template showing what .env should look like
├── go.mod                         # Go module + dependency list
└── go.sum
```

**`internal/`** is a special Go folder name — code inside it can only be imported by code inside this same project. It's Go's way of saying "this is private to this app."

---

## How It's Organized (the "why")

This project follows a pattern often called **layered architecture**:

1. **`main.go`** — wires everything together and starts the server. It doesn't contain business logic.
2. **`handlers` (student.go)** — receives HTTP requests, validates input, calls storage, and writes a response. It doesn't know *how* data is stored (SQL, files, etc.) — it just calls a `Storage` interface.
3. **`storage.go` (interface)** — defines *what* storage must be able to do (`CreateStudent`, `GetStudentById`, etc.), without saying *how*.
4. **`postgres.go`** — the actual implementation: real SQL queries against PostgreSQL.

Why bother separating these? Because the handlers never talk to SQL directly. If you swapped PostgreSQL for MySQL or an in-memory store, you'd only need a new file that implements the same `Storage` interface — the handlers and routes wouldn't change at all.

---

## Data Model

A **Student** looks like this in JSON:

```json
{
  "id": 1,
  "name": "Abhishek",
  "email": "abhi@example.com",
  "age": 21
}
```

When **creating or updating** a student, you send a `StudentRequest` (no `id` — the database assigns that):

```json
{
  "name": "Abhishek",
  "email": "abhi@example.com",
  "age": 21
}
```

Validation rules (checked automatically before hitting the database):
- `name` — required
- `email` — required, must look like a real email
- `age` — required, must be greater than 0

If validation fails, the API responds with `400 Bad Request` and a message explaining what's wrong.

---

## API Endpoints

| Method | Route | Purpose |
|---|---|---|
| `POST` | `/api/students` | Create a new student |
| `GET` | `/api/students` | Get all students |
| `GET` | `/api/students/{id}` | Get one student by ID |
| `PUT` | `/api/students/{id}` | Update a student by ID |
| `DELETE` | `/api/students/{id}` | Delete a student by ID |

### Example: Create a student

**Request**
```
POST /api/students
Content-Type: application/json

{
  "name": "Priya Sharma",
  "email": "priya.sharma@example.com",
  "age": 23
}
```

**Response** (`201 Created`)
```json
{ "id": 1 }
```

### Example: Get a student

```
GET /api/students/1
```

**Response** (`200 OK`)
```json
{
  "id": 1,
  "name": "Priya Sharma",
  "email": "priya.sharma@example.com",
  "age": 23
}
```

If the ID doesn't exist, you get `404 Not Found`.

---

## How a Request Flows Through the Code

Take `GET /api/students/1` as an example:

1. **`main.go`** matches the route `GET /api/students/{id}` and calls `student.GetById(db)`.
2. **`student.go` → `GetById`** reads `{id}` from the URL, converts it to a number, and calls `store.GetStudentById(id)`.
3. **`postgres.go` → `GetStudentById`** runs a SQL `SELECT` query against the database.
   - If a row is found → returns the student.
   - If no row is found → returns a special error: `storage.ErrStudentNotFound`.
4. Back in the handler, if that specific error comes back, it responds with `404`. Any other error becomes a `500 Internal Server Error`. Otherwise, it writes the student back as JSON with `200 OK`.

This "check for a specific known error" pattern (using `errors.Is`) is how Go tells the difference between "nothing was found" (expected, not really a bug) and "something went wrong" (a real error).

---

## Configuration

Settings are loaded from a `.env` file in the project root (see `.env.example` as a template):

```
ENV=dev
ADDR=localhost:8082
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require
```

- `ENV` — `dev` or `prod`. Changes the log format (plain text in dev, JSON in prod).
- `ADDR` — the host and port the server listens on.
- `DATABASE_URL` — your PostgreSQL connection string. **Required** — the app won't start without it.

`.env` is listed in `.gitignore`, so your real database credentials are never committed to version control. Only `.env.example` (with fake placeholder values) is meant to be shared.

---

## Running the Project

1. Make sure you have Go installed and a PostgreSQL database ready (a free [Neon](https://neon.tech) database works fine).
2. Copy `.env.example` to `.env` and fill in your real `DATABASE_URL`.
3. Install dependencies:
   ```
   go mod tidy
   ```
4. Run the server:
   ```
   go run cmd/students-api/main.go
   ```
5. You should see logs like:
   ```
   database connected
   server started address=localhost:8082
   ```
6. Test it with Postman, curl, or any HTTP client, e.g.:
   ```
   curl -X POST http://localhost:8082/api/students \
     -H "Content-Type: application/json" \
     -d '{"name":"Abhishek","email":"abhi@example.com","age":21}'
   ```

The server keeps a table called `students` in your database, and creates it automatically on startup if it doesn't already exist.

---

## Graceful Shutdown

When you stop the server (Ctrl+C), it doesn't just die immediately. It:
1. Stops accepting new requests.
2. Waits up to 5 seconds for any in-flight requests to finish.
3. Then shuts down cleanly and closes the database connection.

This avoids cutting off a request halfway through, which could leave data in a half-written state.

---

## Possible Next Steps

Ideas if you want to keep building on this:
- Add authentication (e.g. API keys or JWT) so not just anyone can hit the API.
- Add pagination to `GET /api/students` for when the list gets large.
- Write automated tests for the handlers and storage layer.
- Add a Dockerfile to containerize the app.