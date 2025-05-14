# Pijar API Documentation

## Base URL
`http://103.196.152.162:8884/pijar`

## Authentication
All endpoints (except `/register` and `/login`) require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Error message here"
}
```

### 401 Unauthorized
```json
{
  "error": "Missing or invalid token"
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden: role not allowed"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

## Endpoints

### Authentication

#### Register User
```
POST /pijar/register
```

**Request Body:**
```json
{
  "name": "Deny",
  "email": "deny@example.com",
  "password": "password123",
  "birth_year": 2005,
  "phone": "081234567890"
}
```

**Response (Success - 201 Created):**
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "name": "Deny",
    "email": "deny@example.com",
    "birth_year": 2005,
    "phone": "081234567890",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T00:00:00Z",
    "role": "USER"
  }
}
```

**Error Responses:**
- 400 Bad Request: Invalid input data
- 500 Internal Server Error: Failed to process registration

#### Login
```
POST /pijar/login
```

**Request Body:**
```json
{
  "email": "hasan3@gmail.com",
  "password": "password123"
}
```

**Response (Success - 200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJQSUpBUi1BUFAiLCJleHAiOjE3NDcyODg4MTgsIlVzZXJJZCI6IjUiLCJSb2xlIjoiVVNFUiJ9.5Ijy1rauMPPF7dIaB_PB3PrIOgVJ6_zi5I4a1xkgpRU",
  "user": {
    "id": 5,
    "name": "hasan",
    "email": "hasan3@gmail.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-14T05:09:30.738295Z",
    "updated_at": "2025-05-14T05:09:30.738295Z",
    "role": "USER"
  }
}
```

**Error Responses:**
- 400 Bad Request: Invalid input format
- 401 Unauthorized: Invalid credentials

### User Management

#### Get All Users (Admin Only)
```
GET /pijar/users
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Deny Caknan",
      "email": "deny@example.com",
      "birth_year": 1990,
      "phone": "081234567890",
      "created_at": "2025-05-13T00:00:00Z",
      "updated_at": "2025-05-13T00:00:00Z",
      "role": "USER"
    }
  ]
}
```

#### Get User by ID (Admin Only)
```
GET /pijar/users/:id
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "data": {
    "id": 4,
    "name": "Hasan updated",
    "email": "HasanUpdate@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-14T03:05:21.514087Z",
    "updated_at": "2025-05-14T05:45:33.763078Z",
    "role": ""
  }
}
```

#### Update User (Admin Only)
```
PUT /pijar/users/:id
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Hasan updated",
  "email": "HasanUpdate@example.com",
  "birth_year": 1990,
  "phone": "081234567890",
  "role": "USER"
}
```

**Response (Success - 200 OK):**
```json
{
  "data": {
    "id": 4,
    "name": "Hasan updated",
    "email": "HasanUpdate@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-14T03:05:21.514087Z",
    "updated_at": "2025-05-14T05:45:33.763078Z",
    "role": "USER"
  }
}
```

#### Delete User (Admin Only)
```
DELETE /pijar/users/:id
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

#### Find User by Email (Admin Only)
```
GET /pijar/users/email/:email
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "data": {
    "id": 1,
    "name": "Deny Caknan",
    "email": "deny@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T00:00:00Z",
    "role": "USER"
  }
}
```

### User Profile

#### Get Own Profile
```
GET /pijar/profile
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "data": {
    "id": 5,
    "name": "hasan",
    "email": "hasan3@gmail.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-14T05:09:30.738295Z",
    "updated_at": "2025-05-14T05:09:30.738295Z",
    "role": "USER"
  }
}
```

#### Update Own Profile
```
PUT /pijar/profile/:id
```

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Hasan Updated",
  "email": "hasan.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567899"
}
```

**Response (Success - 200 OK):**
```json
{
  "data": {
    "id": 5,
    "name": "Hasan Updated",
    "email": "hasan.updated@example.com",
    "birth_year": 1991,
    "phone": "081234567899",
    "created_at": "2025-05-14T05:09:30.738295Z",
    "updated_at": "2025-05-14T05:45:33.763078Z",
    "role": "USER"
  }
}
```

### AI Coach Sessions

Endpoints untuk berinteraksi dengan AI Coach yang memberikan bimbingan karir dan pengembangan diri.

#### Start New Session
```
POST /pijar/sessions/start/:user_id
```

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_input": "Apa yang harus saya lakukan untuk memulai karir di bidang teknologi?"
}
```

**Response (Success - 201 Created):**
```json
{
  "session_id": "unique-session-id",
  "response": "Untuk memulai karir di bidang teknologi, Anda bisa memulainya dengan..."
}
```

**Error Responses:**
- 400 Bad Request: Invalid input format
- 401 Unauthorized: Missing or invalid token
- 500 Internal Server Error: Failed to start session

#### Continue Session
```
POST /pijar/sessions/continue/:sessionId/:user_id
```

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_input": "Terima kasih atas sarannya. Bagaimana dengan skill yang harus saya kuasai?"
}
```

**Response (Success - 200 OK):**
```json
{
  "session_id": "unique-session-id",
  "response": "Berikut adalah beberapa skill penting yang bisa Anda pelajari..."
}
```

#### Get Session History
```
GET /pijar/sessions/history/:sessionId/:user_id
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "session_id": "unique-session-id",
  "messages": [
    {
      "role": "user",
      "content": "Apa yang harus saya lakukan?"
    },
    {
      "role": "assistant",
      "content": "Anda bisa memulai dengan..."
    }
  ]
}
```

#### Delete Session
```
DELETE /pijar/sessions/:sessionId/:user_id
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "message": "Session deleted successfully"
}
```

#### Get All User Sessions (Admin Only)
```
GET /pijar/sessions/user/:user_id
```

**Headers:**
```
Authorization: Bearer <admin_jwt_token>
```

**Response (Success - 200 OK):**
```json
[
  {
    "id": 1,
    "session_id": "unique-session-id-1",
    "user_id": 1,
    "timestamp": "2025-05-13T10:30:00Z"
  },
  {
    "id": 2,
    "session_id": "unique-session-id-2",
    "user_id": 1,
    "timestamp": "2025-05-13T11:45:00Z"
  }
]
```

### Journal

Endpoints untuk manajemen jurnal pribadi pengguna.

#### Create Journal Entry
```
POST /pijar/journals
```

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_id": 1,
  "title": "Hari Pertama Kerja",
  "content": "Hari ini adalah hari pertama saya bekerja...",
  "mood": "happy"
}
```

**Response (Success - 201 Created):**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Hari Pertama Kerja",
  "content": "Hari ini adalah hari pertama saya bekerja...",
  "mood": "happy",
  "created_at": "2025-05-13T00:00:00Z"
}
```

**Error Responses:**
- 400 Bad Request: Invalid input data
- 401 Unauthorized: Missing or invalid token
- 500 Internal Server Error: Failed to create journal entry

#### Get User's Journals
```
GET /pijar/journals/user/:userID
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "title": "Hari Pertama Kerja",
    "content": "Hari ini adalah hari pertama saya bekerja...",
    "mood": "happy",
    "created_at": "2025-05-13T00:00:00Z"
  },
  {
    "id": 2,
    "user_id": 1,
    "title": "Mencoba Hal Baru",
    "content": "Hari ini saya mencoba sesuatu yang baru...",
    "mood": "excited",
    "created_at": "2025-05-12T00:00:00Z"
  }
]
```

#### Get Journal by ID
```
GET /pijar/journals/:journalID
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Hari Pertama Kerja",
  "content": "Hari ini adalah hari pertama saya bekerja...",
  "mood": "happy",
  "created_at": "2025-05-13T00:00:00Z"
}
```

#### Update Journal
```
PUT /pijar/journals/:journalID
```

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Hari Pertama Kerja (Revisi)",
  "content": "Hari ini adalah hari pertama saya bekerja... Sangat menyenangkan!",
  "mood": "very happy"
}
```

**Response (Success - 200 OK):**
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Hari Pertama Kerja (Revisi)",
  "content": "Hari ini adalah hari pertama saya bekerja... Sangat menyenangkan!",
  "mood": "very happy",
  "created_at": "2025-05-13T00:00:00Z",
  "updated_at": "2025-05-14T00:00:00Z"
}
```

#### Delete Journal
```
DELETE /pijar/journals/:journalID
```

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success - 200 OK):**
```json
{
  "message": "Journal deleted successfully"
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Error message here"
}
```

### 401 Unauthorized
```json
{
  "error": "Missing Authorization header"
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden: role not allowed"
}
```

### 404 Not Found
```json
{
  "error": "User not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```