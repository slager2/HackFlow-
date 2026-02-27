# HackFlow API (Backend)

The backend service for the **HackFlow** project — an AI-aggregator for IT events and hackathons. Built with Go 1.22+, Gin, and PostgreSQL.

## 🚀 Overview

The backend is built using a clean, layered architecture and provides a REST API for the Next.js frontend to fetch verified (non-hallucinated) hackathons.

### Tech Stack
- **Go 1.22+**: Core language.
- **Gin**: HTTP web framework for routing and middleware.
- **GORM (PostgreSQL)**: ORM library for database interactions and migrations.
- **Docker & Docker Compose**: Containerization and local development environment.
- **slog**: Native Go structured logging.

## 📂 Project Structure

```text
backend/
├── cmd/
│   └── api/
│       └── main.go         # Application entrypoint
├── internal/
│   ├── config/             # Environment configuration (godotenv)
│   ├── database/           # PostgreSQL connection and auto-migrations
│   ├── handlers/           # Gin HTTP route handlers
│   ├── logger/             # Structured logging configuration
│   └── models/             # Database schemas (GORM models)
├── .env                    # Environment variables (ignored in version control)
├── docker-compose.yaml     # Container orchestration (API + PostgreSQL)
├── Dockerfile              # Multi-stage build for Go binary
└── go.mod & go.sum         # Go module dependencies
```

## 🛠 Prerequisites

- Docker Desktop installed and running.
- (Optional) Go 1.22+ installed locally if you want to run it without Docker.

## ⚙️ Configuration

Create a `.env` file in the `backend/` directory based on the following template. Since we are using Docker Compose, we connect to the PostgreSQL service named `db` (or `host.docker.internal` if testing against a host DB).

```env
# Example .env configuration
DB_HOST=host.docker.internal
DB_USER=hackflow_user
DB_PASSWORD=supersecretpassword
DB_NAME=hackflow
DB_PORT=5432
PORT=8080
```

## 🚀 Running the project

The easiest way to run the entire backend infrastructure (Go API and PostgreSQL database) is using Docker Compose:

1. Open your terminal in the `backend` directory.
2. Run the following command:

```bash
docker-compose up -d --build
```

This will:
- Pull the PostgreSQL image and start the database.
- Build the Go application into a lightweight Alpine image.
- Start the API backend on port `8080`.

**To stop the containers:**
```bash
docker-compose down
```

## 🌐 API Endpoints

### `GET /api/hackathons`

Fetches a list of IT events.

**Query Parameters:**
- `q` (optional): Case-insensitive search query to filter events by `title` or `city`.

**Response (JSON):**

```json
[
  {
    "id": 1,
    "CreatedAt": "2026-02-27T00:00:00Z",
    "UpdatedAt": "2026-02-27T00:00:00Z",
    "DeletedAt": null,
    "title": "Decentrathon",
    "date": "15 Марта 2026",
    "format": "Офлайн",
    "city": "Астана",
    "ageLimit": "18+",
    "link": "https://decentrathon.io"
  }
]
```
