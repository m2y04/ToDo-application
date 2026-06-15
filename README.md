# ToDo Application

Modular ToDo application with authentication for the AFS x Reboot Tech Challenge.

## Stack

*   Frontend: React + Vite
*   Backend: Go `net/http`
*   Database: PostgreSQL
*   Auth: username/password, bcrypt password hashing, JWT
*   Deployment: Docker Compose

## Architecture

```
frontend  ->  backend API  ->  PostgreSQL
React         Go REST API       users + todos
```

*   The frontend stores the JWT in `localStorage`.
*   Protected requests use `Authorization: Bearer TOKEN`.
*   The backend verifies JWTs before allowing access to ToDo routes.
*   PostgreSQL runs in its own Docker container.
*   Each ToDo belongs to one user.

## Project Structure

```
.
├── backend/
│   ├── main.go
│   └── src/
│       ├── auth/
│       ├── db/
│       ├── handlers/
│       ├── middleware/
│       └── models/
├── frontend/
│   └── src/
│       ├── components/
│       ├── api.js
│       └── App.jsx
├── db/init/001_schema.sql
├── docker-compose.yml
├── .env.example
└── test_backend.sh
```

## Environment

Create a local `.env` from the example:

```
cp .env.example .env
```

Set a real JWT secret in `.env`:

```
JWT_SECRET=replace_with_a_long_random_secret
```

`.env` is ignored by Git.

## Run Locally

Start PostgreSQL:

```
docker compose up -d postgres
```

Run the backend:

```
cd backend
set -a
. ../.env
set +a
go run .
```

Run the frontend in another terminal:

```
cd frontend
npm install
npm run dev
```

Open:

```
http://localhost:3000
```

Backend health:

```
http://localhost:5000/health
```

## Run With Docker Compose

```
docker compose up --build
```

Services:

*   Frontend: `http://localhost:3000`
*   Backend: `http://localhost:5000`
*   PostgreSQL: exposed locally on `localhost:5433`

## Database

Tables:

*   `users`
*   `todos`

Schema is initialized from:

```
db/init/001_schema.sql
```

The ToDo table includes:

*   `user_id` ownership
*   `completed`
*   `created_at`
*   `updated_at`
*   index on `todos.user_id`

## API Endpoints

Auth:

```
POST /auth/register
POST /auth/login
GET  /auth/me
```

Todos:

```
GET    /todos
POST   /todos
PUT    /todos/:id
DELETE /todos/:id
```

Health:

```
GET /health
```

## Backend Tests

Start PostgreSQL and the backend, then run:

```
./test_backend.sh
```

The script covers:

*   health check
*   CORS preflight
*   register/login
*   malformed JSON
*   duplicate username
*   invalid, missing, and expired JWT
*   ToDo create/list/update/delete
*   mark and unmark completed
*   cross-user access protection

## Notes

*   Passwords are stored as bcrypt hashes.
*   JWT secret is required through environment configuration.
*   Users can only access their own ToDos.
*   `node_modules/`, `dist/`, `.env`, and local tool settings are ignored.
