package services

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"payment-gateway/internal/models"
	"payment-gateway/internal/repositories"

	"github.com/avast/retry-go/v4"
)

// CallbackJob mendefinisikan struktur data untuk antrean worker
type CallbackJob struct {
	ReferenceID string
	Payload     []byte
}

// JobQueue adalah channel Go dengan buffer (kapasitas antrean)
var JobQueue = make(chan CallbackJob, 1000)

// httpClient digunakan secara global (reuse) agar hemat memory (mencegah memory leak akibat koneksi menumpuk)
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// StartWorkerPool menjalankan sejumlah goroutine tetap (worker) di latar belakang
func StartWorkerPool(numWorkers int) {
	for i := 1; i <= numWorkers; i++ {
		go worker(i)
	}
	log.Printf("Worker Pool dimulai dengan %d Pekerja aktif\n", numWorkers)
}

// worker adalah goroutine yang senantiasa menunggu job baru dari JobQueue
func worker(id int) {
	for job := range JobQueue {
		ProcessCallback(job.ReferenceID, job.Payload)
	}
}

func ProcessCallback(referenceID string, rawPayload []byte) {
	// Ekstrak Routing Code (Misal: "SHOP-INV-001" -> "SHOP")
	parts := strings.SplitN(referenceID, "-", 2)
	if len(parts) < 2 {
		log.Println("Format reference_id tidak memiliki kode routing:", referenceID)
		return
	}
	routingCode := parts[0]

	// Cari konfigurasi tujuan menggunakan Service
	destService := NewDestinationService()
	dest, err := destService.GetByRoutingCode(routingCode)
	if err != nil {
		log.Printf("Tujuan untuk routing code '%s' tidak ditemukan\n", routingCode)
		return
	}

	// Proses Forwarding dengan Retry
	forwardRequest(referenceID, dest, rawPayload)
}

func forwardRequest(refID string, dest models.Destination, payload []byte) {
	var statusCode int

	// Blok retry-go: Akan diulang otomatis jika mengembalikan error
	err := retry.Do(
		func() error {
			req, err := http.NewRequest("POST", dest.TargetURL, bytes.NewBuffer(payload))
			if err != nil {
				return retry.Unrecoverable(err) // Tidak usah retry jika gagal bikin request
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Gateway-Auth", dest.SecretToken)

			resp, err := httpClient.Do(req)

			if err != nil {
				return err // Jaringan bermasalah/RTO -> akan di retry otomatis
			}
			defer resp.Body.Close()
			
			statusCode = resp.StatusCode

			// Jika target membalas dengan status 50x (Server Error), anggap gagal dan minta retry
			if statusCode >= 500 {
				return errors.New("target server merespon 50x error")
			}

			return nil // Berhasil
		},
		retry.Attempts(5), // Maksimal coba 5 kali
		retry.Delay(2*time.Second), // Jeda awal 2 detik sebelum retry pertama
		retry.MaxDelay(30*time.Second), // Maksimal jeda 30 detik
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry %d ke %s karena error: %v\n", n+1, dest.TargetURL, err)
		}),
	)

	if err != nil {
		log.Println("Gagal meneruskan ke URL secara permanen:", dest.TargetURL, err)
	} else {
		log.Printf("Forward berhasil! %s merespon HTTP %d\n", dest.TargetURL, statusCode)
	}

	// Simpan Log Audit
	logEntry := models.CallbackLog{
		ReferenceID: refID,
		RoutingCode: dest.RoutingCode,
		TargetURL:   dest.TargetURL,
		Payload:     string(payload),
		StatusCode:  statusCode,
	}
	logRepo := repositories.NewLogRepository()
	logRepo.Create(&logEntry)
}

// ForwardSync melakukan HTTP POST secara sinkron (tanpa retry) dan mengembalikan respons ke pemanggil
func ForwardSync(refID string, dest models.Destination, payload []byte) (int, []byte, error) {
	req, err := http.NewRequest("POST", dest.TargetURL, bytes.NewBuffer(payload))
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Gateway-Auth", dest.SecretToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	// Fallback membaca body dengan buffer standar
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	bodyBytes = buf.Bytes()

	// Simpan Log Audit
	logEntry := models.CallbackLog{
		ReferenceID: refID,
		RoutingCode: dest.RoutingCode,
		TargetURL:   dest.TargetURL,
		Payload:     string(payload),
		StatusCode:  resp.StatusCode,
	}
	logRepo := repositories.NewLogRepository()
	logRepo.Create(&logEntry)

	return resp.StatusCode, bodyBytes, nil
}

