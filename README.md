# Payment Gateway Smart Gateway

Payment Gateway Smart Gateway adalah gateway HTTP berkinerja tinggi yang dirancang untuk menerima webhook/callback dari Payment Gateway dan meneruskannya (forwarding) secara andal ke berbagai layanan internal tujuan (destination) berdasarkan konfigurasi _routing_.

## Fitur Utama

- **High-Throughput Callback Handling**: Dibangun dengan Go dan Fiber untuk konkurensi maksimal dengan latensi rendah.
- **Worker Pool & Job Queue**: Menggunakan Go Channel secara bawaan untuk menampung lonjakan traffic callback mendadak tanpa menyebabkan _Out of Memory_ (OOM).
- **Auto-Retry Mechanism**: Secara otomatis mengulang (retry) pengiriman data ke URL Tujuan (Target URL) dengan pola _exponential backoff_ jika server target sedang mati.
- **Database Connection Pooling**: Mencegah putusnya koneksi database (PostgreSQL) saat server menerima ribuan hit mendadak.
- **Docker-Ready**: Dilengkapi dengan arsitektur _Multi-stage Build_ Dockerfile (berbasis Alpine) yang menghasilkan file binary statis berukuran sangat kecil (~20MB).

## Tech Stack

- **Bahasa**: Go 1.25+
- **Framework Web**: [Fiber v2](https://gofiber.io/)
- **Database ORM**: [GORM](https://gorm.io/) (PostgreSQL & SQLite in-memory untuk testing)
- **Retry Logic**: `github.com/avast/retry-go/v4`

## Cara Penggunaan

### 1. Konfigurasi Environment

Buat atau ubah file konfigurasi environment dari file contoh (`.env.example`):

```bash
cp .env.example .env
```

Pastikan Anda memasukkan detail koneksi database dan kredensial Token rahasia Provider Payment Gateway:

```dotenv
APP_PORT=:3131
DB_DSN=host=localhost user=postgres password=rahasia dbname=flip_gateway port=5432 sslmode=disable TimeZone=Asia/Jakarta
JWT_SECRET=super_secret_key_industri
```

### 2. Menjalankan secara Lokal

Pastikan Anda memiliki Go yang ter-install di sistem operasi Anda.

```bash
# Mengunduh modul dependensi
go mod tidy

# Menjalankan aplikasi
go run cmd/main.go
```

### 3. Menjalankan via Docker

Untuk menjalankan aplikasi ini menggunakan container Docker yang terisolasi dan sudah siap untuk environment production:

```bash
# Build Docker image
docker build -t payment-gateway:latest .

# Run Docker container (dengan host network agar bisa terhubung ke database local, baca variabel dari .env)
docker run -d --name payment-gateway-app --env-file .env --network="host" payment-gateway:latest
```

## Pengujian (Testing)

Aplikasi ini sudah dilengkapi unit test menggunakan **SQLite In-Memory**, sehingga Anda bisa menguji fungsionalitas tanpa risiko mengotori atau merusak database PostgreSQL Anda.

```bash
# Menjalankan semua Test
go test ./... -v
```
