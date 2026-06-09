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

### 4. Cara Uji Coba (Hit Callback)

Sebelum mencoba *hit* callback, pastikan Anda telah membuat **Destination** (tujuan *routing*) melalui Dashboard Frontend di `http://127.0.0.1:3131` atau via API. Anda membutuhkan:
- **Routing Code** (misal: `SHOP`)
- **Target URL** (misal: `https://webhook.site/xxxx`)
- **Provider Token** (misal: `token_rahasia_anda`)

Berikut adalah contoh untuk mengirim simulasi callback (jangan lupa sesuaikan alamat jika menggunakan *ngrok*).

**Untuk Flip (`/flip/callback`)**
Endpoint ini menerima data dalam format `application/x-www-form-urlencoded`. Parameternya adalah `token` dan `data` (berisi JSON dengan `reference_id` yang diawali Routing Code).

```bash
curl -X POST http://127.0.0.1:3131/flip/callback \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=token_rahasia_anda" \
  -d 'data={"reference_id": "SHOP-INV-001", "status": "SUCCESS", "amount": 50000}'
```

**Untuk Midtrans (`/midtrans/callback`)**
Endpoint ini menerima `application/json` mentah, dan token di-*passing* via Header `Authorization`. Di dalam payload, harus terdapat `order_id` yang diawali Routing Code.

```bash
curl -X POST http://127.0.0.1:3131/midtrans/callback \
  -H "Content-Type: application/json" \
  -H "Authorization: token_rahasia_anda" \
  -d '{
    "order_id": "SHOP-ORD-002",
    "transaction_status": "settlement",
    "gross_amount": "100000.00"
  }'
```

### 5. Contoh Code Snippet Aplikasi Penerima (Destination Service)

Berikut adalah contoh kode untuk aplikasi internal Anda yang menerima *forwarding* data dari Gateway ini. 
Aplikasi Anda wajib memvalidasi header `X-Gateway-Auth` (sesuai `SecretToken` yang Anda atur di Dashboard) dan mengembalikan HTTP Status `200 OK`.

**1. PHP (Native)**
```php
<?php
// Menerima Header X-Gateway-Auth
$headers = getallheaders();
$gatewayAuth = isset($headers['X-Gateway-Auth']) ? $headers['X-Gateway-Auth'] : '';

// Validasi Token Rahasia
if ($gatewayAuth !== 'secret_token_dari_dashboard') {
    http_response_code(401);
    echo "Unauthorized";
    exit;
}

// Mengambil raw payload yang diteruskan oleh Gateway
$rawPayload = file_get_contents('php://input');

// Payload ini akan sama persis dengan yang dikirimkan oleh Flip/Midtrans
// Lakukan proses update transaksi di database Anda berdasarkan payload tersebut...

// Wajib merespon 200 OK secepatnya agar Gateway berhenti melakukan Auto-Retry
http_response_code(200);
echo "OK";
?>
```

**2. Node.js (Express)**
```javascript
const express = require('express');
const app = express();

// Tambahkan middleware untuk membaca body
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.post('/terima-callback', (req, res) => {
  // 1. Validasi Header X-Gateway-Auth
  const gatewayAuth = req.headers['x-gateway-auth'];
  
  if (gatewayAuth !== 'secret_token_dari_dashboard') {
    return res.status(401).send('Unauthorized');
  }

  // 2. Ambil payload yang diteruskan oleh Gateway
  // Payload ini sama persis dengan yang dikirimkan oleh Flip/Midtrans
  const payload = req.body;
  console.log("Menerima data:", payload);

  // 3. Lakukan proses update transaksi di database internal Anda...

  // 4. Wajib merespon 200 OK agar Gateway berhenti melakukan Auto-Retry
  res.status(200).send('OK');
});

app.listen(8080, () => {
  console.log('Aplikasi penerima berjalan di port 8080');
});
```

## Pengujian (Testing)

Aplikasi ini sudah dilengkapi unit test menggunakan **SQLite In-Memory**, sehingga Anda bisa menguji fungsionalitas tanpa risiko mengotori atau merusak database PostgreSQL Anda.

```bash
# Menjalankan semua Test
go test ./... -v
```
