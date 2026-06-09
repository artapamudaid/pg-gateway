# Build Stage: Menggunakan golang alpine untuk mengkompilasi aplikasi Go
FROM golang:1.25-alpine AS builder

# Set working directory di dalam container
WORKDIR /app

# Install git (dibutuhkan untuk fetch dependencies) dan tzdata (untuk timezone)
RUN apk add --no-cache git tzdata

# Menyalin file go.mod dan go.sum terlebih dahulu.
# Hal ini penting untuk memanfaatkan layer caching Docker.
COPY go.mod go.sum ./
RUN go mod download

# Menyalin seluruh source code
COPY . .

# Kompilasi aplikasi Go. 
# CGO_ENABLED=0 digunakan agar binary bersifat statis penuh (sangat aman & ringan).
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-gateway ./cmd/main.go

# Production Stage: Menggunakan Alpine yang sangat kecil (~5MB)
FROM alpine:latest

# Install sertifikat CA untuk mendukung HTTPS calls dan set zona waktu
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Jakarta

WORKDIR /root/

# Copy binary 'payment-gateway' dari tahap builder sebelumnya
COPY --from=builder /app/payment-gateway .

# Copy folder frontend agar bisa diakses oleh Fiber app.Static
COPY --from=builder /app/frontend ./frontend

# Terapkan GOMEMLIMIT untuk menjamin batasan memori (misal 250MB) 
# agar aplikasi tidak mengonsumsi RAM berlebih secara mendadak.
ENV GOMEMLIMIT=250MiB

# Mengekspos port 3000
EXPOSE 3000

# Eksekusi aplikasi
CMD ["./payment-gateway"]
