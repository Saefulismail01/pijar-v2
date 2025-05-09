# Pijar API Documentation

## Base URL
`http://localhost:8080/pijar`

## Authentication
All endpoints (except `/register` and `/login`) require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
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
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "birth_year": 1990,
  "phone": "081234567890"
}
```

**Response:**
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890"
  }
}
```

#### Login
```
POST /pijar/login
```

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "jwt_token_here"
}
```

### AI Coach

Endpoints untuk berinteraksi dengan AI Coach yang memberikan bimbingan karir dan pengembangan diri.

#### Chat with AI Coach
```
POST /pijar/coach/:user_id
```

**Deskripsi:**
Mengirim pesan ke AI Coach dan mendapatkan respon bimbingan. AI akan memberikan saran berdasarkan input pengguna dengan konteks percakapan sebelumnya.

**Kegunaan:**
- Mendapatkan saran karir
- Konsultasi pengembangan diri
- Bimbingan dalam menghadapi tantangan pekerjaan
- Rekomendasi langkah-langkah pengembangan karir

**Request Body:**
```json
{
  "user_input": "Apa yang harus saya lakukan untuk memulai karir di bidang teknologi?"
}
```

**Response:**
```json
{
  "ai_response": "Untuk memulai karir di bidang teknologi, Anda bisa memulainya dengan..."
}
```

#### Get User's Chat History
```
GET /pijar/coach/:user_id
```

**Deskripsi:**
Mendapatkan riwayat percakapan pengguna dengan AI Coach.

**Kegunaan:**
- Melihat kembali percakapan sebelumnya
- Melacak perkembangan konsultasi
- Menganalisis riwayat bimbingan yang telah diberikan

**Response:**
```json
{
  "sessions": [
    {
      "id": 1,
      "user_id": 1,
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
  ]
}
```

#### Delete User's Chat History
```
DELETE /pijar/coach/:user_id
```

**Deskripsi:**
Menghapus semua riwayat percakapan pengguna dengan AI Coach.

**Kegunaan:**
- Menghapus riwayat percakapan lama
- Memulai percakapan baru yang lebih fokus
- Membersihkan data percakapan yang tidak diperlukan lagi

**Response:**
```json
{
  "message": "Session untuk user ID 1 berhasil dihapus"
}
```

### Journal

Endpoints untuk manajemen jurnal pribadi pengguna.

#### Create Journal
```
POST /pijar/journals/
```

**Deskripsi:**
Membuat entri jurnal baru.

**Kegunaan:**
- Mencatat pengalaman harian
- Melacak perkembangan pribadi
- Mencatat pencapaian atau tantangan
- Menulis refleksi diri

**Request Body:**
```json
{
  "user_id": 1,
  "judul": "Hari Pertama Kerja",
  "isi": "Hari ini adalah hari pertama saya bekerja...",
  "perasaan": "senang"
}
```

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

#### Get User's Journals
```
GET /pijar/journals/user/:userID
```

**Deskripsi:**
Mendapatkan semua jurnal yang dimiliki oleh pengguna.

**Kegunaan:**
- Melihat riwayat jurnal
- Melacak perkembangan dari waktu ke waktu
- Menganalisis pola emosi dan pengalaman

**Response:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "judul": "Hari Pertama Kerja",
    "isi": "Hari ini adalah hari pertama saya bekerja...",
    "perasaan": "senang",
    "created_at": "2023-01-01T00:00:00Z"
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
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "birth_year": 1990,
    "phone": "081234567890",
    "role": "user"
  }
}
```

#### Update User
```
PUT /pijar/users
```

**Deskripsi:**
Memperbarui informasi pengguna.

**Kegunaan:**
- Memperbarui profil pengguna
- Memperbaiki data yang salah
- Mengubah peran atau status pengguna

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
  "error": "User not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```
