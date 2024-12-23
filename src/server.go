package main

import (
	"encoding/json"
	"fmt"
	helpers "goserver/src/Helpers"
	index "goserver/src/Routes"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"github.com/joho/godotenv"
)

type Worker struct {
	ID      int
	JobChan chan Job
	Quit    chan bool
}

type Job struct {
	ID      string
	Payload map[string]interface{}
}

func NewWorker(id int) Worker {
	return Worker{
		ID:      id,
		JobChan: make(chan Job),
		Quit:    make(chan bool),
	}
}

func (w Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case job := <-w.JobChan:
			processJob(w.ID, job)
		case <-w.Quit:
			log.Printf("Worker %d shutting down...\n", w.ID)
			return
		}
	}
}

func processJob(workerID int, job Job) {
	log.Printf("Worker %d processing job ID %s with payload: %v\n", workerID, job.ID, job.Payload)
	time.Sleep(1 * time.Second)
	log.Printf("Worker %d completed job ID %s\n", workerID, job.ID)
}

type WorkerPool struct {
	Workers    []Worker
	JobQueue   chan Job
	WorkerQuit chan bool
}

func NewWorkerPool(numWorkers int, jobQueueSize int) *WorkerPool {
	jobQueue := make(chan Job, jobQueueSize)
	workers := make([]Worker, numWorkers)
	workerQuit := make(chan bool)

	for i := 0; i < numWorkers; i++ {
		workers[i] = NewWorker(i + 1)
	}

	return &WorkerPool{
		Workers:    workers,
		JobQueue:   jobQueue,
		WorkerQuit: workerQuit,
	}
}

func (wp *WorkerPool) Start() {
	var wg sync.WaitGroup
	for _, worker := range wp.Workers {
		wg.Add(1)
		go worker.Start(&wg)
	}
	go func() {
		wg.Wait()
		close(wp.WorkerQuit)
	}()
}

func (wp *WorkerPool) AssignJob(job Job) {
	wp.JobQueue <- job
}

func (wp *WorkerPool) Shutdown() {
	for _, worker := range wp.Workers {
		worker.Quit <- true
	}
	close(wp.JobQueue)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	secretKey := os.Getenv("SECRET_KEY")

	if secretKey == "" {
		log.Fatal("Database credentials are not set in environment variables")
	}

	port := 6002
	numWorkers := 2
	jobQueueSize := 1000

	logger := helpers.NewColorizedLogger(true)
	logger.Info(fmt.Sprintf("Starting Lunar Database Server on Port %d", port))

	workerPool := NewWorkerPool(numWorkers, jobQueueSize)
	workerPool.Start()
	mux := http.NewServeMux()

	index.RegisterRoutes(mux, logger)

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Error("Invalid Method Was Attempted To Lunar Database Server")
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("Invalid Payload Posted to Lunar Database Server")
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		job := Job{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Payload: payload,
		}

		select {
		case workerPool.JobQueue <- job:
			logger.Info(fmt.Sprintf("Job ID %s has been queued successfully", job.ID))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "queued", "jobId": job.ID})
		default:
			logger.Warn("Server is currently busy, try again later")
			http.Error(w, "Server busy, try again later", http.StatusTooManyRequests)
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Warn("Shutting down Lunar Server...")
		workerPool.Shutdown()
		server.Close()
	}()

	logger.Info("Lunar Database Server is up and running")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(fmt.Sprintf("Server error: %s", err))
	}
}
