# Pijar API

A comprehensive backend API for the Pijar application built with Golang, providing user management, coaching sessions, learning content, journaling, goal tracking, and payment processing capabilities.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [API Endpoints](#api-endpoints)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Technologies](#technologies)
- [License](#license)

## Overview

Pijar API is a robust backend service that powers the Pijar application, providing a platform for personal development through coaching sessions, goal tracking, journaling, and curated learning content. The API handles user management, authentication, payment processing via Midtrans, and AI-powered coaching sessions.

## Features

- **User Management**: Registration, authentication, and profile management
- **Goal Tracking**: Create and manage daily goals and progress
- **Journaling**: Personal journal entries with export capabilities
- **Learning Content**: Topics and articles for personal development
- **AI Coaching**: Interactive coaching sessions with history tracking
- **Payment Processing**: Secure payment handling via Midtrans integration

## API Endpoints

### User Management

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/pijar/register` | Register new user | Public |
| POST | `/pijar/login` | User login | Public |
| GET | `/pijar/users` | Get all users | Admin |
| GET | `/pijar/users/:id` | Get user by ID | Admin |
| PUT | `/pijar/users/:id` | Update user | Admin |
| DELETE | `/pijar/users/:id` | Delete user | Admin |
| GET | `/pijar/users/email/:email` | Find user by email | Admin |
| GET | `/pijar/profile` | Get own profile | User |
| PUT | `/pijar/profile` | Update own profile | User |

### Goals Management

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| GET | `/pijar/goals/:user_id` | Get user goals | Admin |
| POST | `/pijar/goals/:user_id` | Create new goal | User |
| PUT | `/pijar/goals/:user_id/:id` | Update goal | User |
| PUT | `/pijar/goals/complete-article` | Update goal progress | User |
| DELETE | `/pijar/goals/:user_id/:id` | Delete goal | User |

### Payment Processing

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/pijar/payments` | Create new payment | User |
| GET | `/pijar/payments/:id` | Check payment status | User |

### Journal Management

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/pijar/journals` | Create new journal | User |
| GET | `/pijar/journals/user/:userID` | Get journals by user ID | User |
| PUT | `/pijar/journals/:journalID` | Update journal | User |
| DELETE | `/pijar/journals/:journalID` | Delete journal | User |
| GET | `/pijar/journals/user/:userID/export` | Export journals to PDF | User |
| GET | `/pijar/journals` | Get all journals | Admin |
| GET | `/pijar/journals/:journalID` | Get journal by ID | Admin |

### Topic Management

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/pijar/topics` | Create new topic | Admin |
| GET | `/pijar/topics` | Get all topics | Public |
| PUT | `/pijar/topics/:id` | Update topic | Admin |
| DELETE | `/pijar/topics/:id` | Delete topic | Admin |
| GET | `/pijar/topics/:id` | Get topic by ID | Admin |

### Article Management

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| GET | `/pijar/articles` | Get all articles with pagination | Public |
| GET | `/pijar/articles/all` | Get all articles without pagination | Public |
| POST | `/pijar/articles/generate` | Generate new article | Admin |
| POST | `/pijar/articles/search` | Search articles by title | Public |
| GET | `/pijar/articles/:id` | Get article by ID | Admin |
| DELETE | `/pijar/articles/:id` | Delete article | Admin |

### AI Coach Session

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/pijar/sessions/start/:user_id` | Start new coaching session | User |
| POST | `/pijar/sessions/continue/:sessionId/:user_id` | Continue coaching session | User |
| GET | `/pijar/sessions/history/:sessionId/:user_id` | Get session history | User |
| GET | `/pijar/sessions/user/:user_id` | Get all user sessions | Admin |

## Installation

### Prerequisites

- Go   
- PostgreSQL 

### Clone the Repository

```bash
git clone https://github.com/enigma-camp/pijar.git
cd pijar
```

### Install Dependencies

```bash
go mod download
```

## Configuration

Create a `.env` file in the root directory with the following variables:

```
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pijar_db

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h

# Midtrans
MIDTRANS_SERVER_KEY=your_midtrans_server_key
MIDTRANS_CLIENT_KEY=your_midtrans_client_key
MIDTRANS_MERCHANT_ID=your_midtrans_merchant_id
MIDTRANS_IS_PRODUCTION=false

# AI Service (if applicable)
AI_SERVICE_URL=your_ai_service_url
AI_SERVICE_API_KEY=your_ai_service_api_key
```

## Running the Application

### Development Mode

```bash
go run main.go
```

### Production Mode

```bash
go build -o pijar-api
./pijar-api
```

### Using Docker

```bash
# Build the Docker image
docker build -t pijar-api .

# Run the container
docker run -p 8884:8080 --env-file .env pijar-api
```

## Development

### Project Structure

```
├── config/              # Configuration setup
├── delivery/
│   └── controller/      # API controllers
│   └── server.go        # API handlers
├── middleware/          # Middleware functions
├── models/              # Data models
│   └── dto/             # Data transfer objects
├── repository/          # Database operations
├── usecase/             # Business logic
├── utils/ 
│   └── model_util/      # Helper functions
│   └── service/         # Helper functions
├── .env                 # Environment variables
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

### Running Tests

```bash
go test ./...
```

## API Documentation

Detailed API documentation is available:

1. Run the server locally
2. Visit `http://localhost:8080/swagger/index.html` for Swagger documentation
3. Alternatively, refer to the API documentation file in the project

## Technologies

- **Backend**: Go
- **Database**: PostgreSQL/MySQL
- **Authentication**: JWT
- **Payment Gateway**: Midtrans
- **Documentation**: Swagger
- **Containerization**: Docker

## License

Developer by Team 3 Incubation Class Enigma Camp
- Saeful Ismail
- Andika Prasetya
- Mhammad Hammas
- Indiyana Noviyanti

---