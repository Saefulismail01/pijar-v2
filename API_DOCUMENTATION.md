# Pijar API Documentation

## Base URL
`http://localhost:8884/pijar`

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
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "birth_year": 1990,
      "phone": "081234567890",
      "role": "user",
      "created_at": "2025-05-13T00:00:00Z"
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
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "role": "user",
    "created_at": "2025-05-13T00:00:00Z"
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
  "name": "John Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567899",
  "role": "admin"
}
```

**Response (Success - 200 OK):**
```json
{
  "message": "User updated successfully",
  "data": {
    "id": 1,
    "name": "John Updated",
    "email": "john.updated@example.com",
    "birth_year": 1991,
    "phone": "081234567899",
    "role": "admin",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T01:00:00Z"
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
  "message": "User found",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "role": "user",
    "created_at": "2025-05-13T00:00:00Z"
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
  "message": "Profile retrieved successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "role": "user",
    "created_at": "2025-05-13T00:00:00Z"
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
  "name": "John Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567899"
}
```

**Response (Success - 200 OK):**
```json
{
  "message": "Profile updated successfully",
  "data": {
    "id": 1,
    "name": "John Updated",
    "email": "john.updated@example.com",
    "birth_year": 1991,
    "phone": "081234567899",
    "role": "user",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T01:00:00Z"
  }
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
  "email": "deny@example.com",
  "password": "password123"
}
```

**Response (Success - 200 OK):**
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
- 400 Bad Request: Invalid input format
- 401 Unauthorized: Invalid credentials

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

**Deskripsi:**
Mendapatkan detail jurnal tertentu berdasarkan ID.

**Kegunaan:**
- Melihat isi jurnal spesifik
- Merefleksikan kembali entri tertentu
- Menggunakan data jurnal untuk analisis lebih lanjut

**Response:**
```json
{
  "id": 1,
  "user_id": 1,
  "judul": "Hari Pertama Kerja",
  "isi": "Hari ini adalah hari pertama saya bekerja...",
  "perasaan": "senang",
  "created_at": "2023-01-01T00:00:00Z"
}
```

#### Update Journal
```
PUT /pijar/journals/:journalID
```

**Deskripsi:**
Memperbarui isi jurnal yang sudah ada.

**Kegunaan:**
- Memperbaiki kesalahan penulisan
- Menambahkan informasi baru
- Memperbarui refleksi atau pemikiran

**Request Body:**
```json
{
  "judul": "Hari Pertama Kerja (Revisi)",
  "isi": "Hari ini adalah hari pertama saya bekerja... Sangat menyenangkan!",
  "perasaan": "sangat senang"
}
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1,
  "judul": "Hari Pertama Kerja (Revisi)",
  "isi": "Hari ini adalah hari pertama saya bekerja... Sangat menyenangkan!",
  "perasaan": "sangat senang",
  "created_at": "2023-01-01T00:00:00Z"
}
```

#### Delete Journal
```
DELETE /pijar/journals/:journalID
```

**Deskripsi:**
Menghapus jurnal tertentu.

**Kegunaan:**
- Menghapus entri yang tidak diperlukan
- Membersihkan data yang sudah tidak relevan
- Menjaga privasi dengan menghapus catatan sensitif

**Response:**
```json
{
  "message": "Journal 1 deleted successfully"
}
```

### User Management (Admin Only)

Endpoints untuk manajemen pengguna (hanya dapat diakses oleh admin).

#### Get All Users
```
GET /pijar/users
```

**Deskripsi:**
Mendapatkan daftar semua pengguna yang terdaftar di sistem.

**Kegunaan:**
- Melihat daftar pengguna
- Monitoring aktivitas pengguna
- Manajemen akun pengguna

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "birth_year": 1990,
      "phone": "081234567890",
      "role": "user"
    }
  ]
}
```

#### Get User by ID
```
GET /pijar/users/:id
```

**Deskripsi:**
Mendapatkan detail pengguna berdasarkan ID.

**Kegunaan:**
- Melihat profil pengguna tertentu
- Verifikasi data pengguna
- Administrasi akun

**Response:**
```json
{
  "message": "User created successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T00:00:00Z",
    "role": "USER"
  }
}
```

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
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "birth_year": 1990,
      "phone": "081234567890",
      "created_at": "2025-05-13T00:00:00Z",
      "updated_at": "2025-05-13T00:00:00Z",
      "role": "USER"
    },
    {
      "id": 2,
      "name": "Admin User",
      "email": "admin@example.com",
      "birth_year": 1985,
      "phone": "081234567891",
      "created_at": "2025-05-12T00:00:00Z",
      "updated_at": "2025-05-12T00:00:00Z",
      "role": "ADMIN"
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
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T00:00:00Z",
    "role": "USER"
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
  "name": "John Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567899",
  "role": "ADMIN"
}
```

**Response (Success - 200 OK):**
```json
{
  "message": "User updated successfully",
  "data": {
    "id": 1,
    "name": "John Updated",
    "email": "john.updated@example.com",
    "birth_year": 1991,
    "phone": "081234567899",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T01:00:00Z",
    "role": "ADMIN"
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
  "message": "Profile retrieved successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T00:00:00Z",
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
  "name": "John Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567899"
}
```

**Response (Success - 200 OK):**
```json
{
  "message": "Profile updated successfully",
  "data": {
    "id": 1,
    "name": "John Updated",
    "email": "john.updated@example.com",
    "birth_year": 1991,
    "phone": "081234567899",
    "created_at": "2025-05-13T00:00:00Z",
    "updated_at": "2025-05-13T01:00:00Z",
    "role": "USER"
  }
}
```

**Request Body:**
```json
{
  "id": 1,
  "name": "John Doe Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567891"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "John Doe Updated",
  "email": "john.updated@example.com",
  "birth_year": 1991,
  "phone": "081234567891",
  "role": "user"
}
```

#### Delete User
```
DELETE /pijar/users/:id
```

**Deskripsi:**
Menghapus akun pengguna dari sistem.

**Kegunaan:**
- Menonaktifkan akun yang tidak aktif
- Memenuhi permintaan penghapusan data
- Membersihkan data pengguna yang sudah tidak diperlukan

**Response:**
```json
{
  "message": "User deleted successfully"
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
  "error": "User not found",
  "error": "Journal not found",
  "error": "Session not found",
  "error": "User not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```
