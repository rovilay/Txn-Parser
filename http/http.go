package parser_http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"txn-parser/service"
)

const (
	getCurrentBlockEndpoint = "/v1/blocks/latest"
	subscribeEndpoint       = "/v1/subscribe"
	getTransactionsEndpoint = "/v1/transactions/{address}"
)

type parserHttp struct {
	svc service.Parser
}

func NewParserHttp(svc service.Parser) *parserHttp {
	return &parserHttp{
		svc: svc,
	}
}

// Start starts the HTTP API server
func (p *parserHttp) Start(ctx context.Context, port uint) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: p.createAPIHandler(),
	}

	go func() {
		log.Println("Starting API server on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("API server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down API server")
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(timeout); err != nil {
		log.Fatalf("API server shutdown error: %v", err)
	}
}

// createAPIHandler creates the HTTP handler for the API endpoints
func (p *parserHttp) createAPIHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(getCurrentBlockEndpoint, p.handleGetCurrentBlock)
	mux.HandleFunc(subscribeEndpoint, p.handleSubscribe)
	mux.HandleFunc(getTransactionsEndpoint, p.handleGetTransactions)
	return mux
}

// handleGetCurrentBlock handles the /blocks/latest endpoint
func (p *parserHttp) handleGetCurrentBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	type AckJsonResponse struct {
		Data    interface{} `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	blockNumber := p.svc.GetCurrentBlock()
	data, err := json.Marshal(&AckJsonResponse{Message: "success", Data: blockNumber})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleSubscribe handles the /subscribe endpoint
func (p *parserHttp) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Address == "" {
		http.Error(w, "Missing address", http.StatusBadRequest)
		return
	}

	if !p.svc.Subscribe(request.Address) {
		http.Error(w, "Failed to subscribe", http.StatusBadRequest)
		return
	}

	type AckJsonResponse struct {
		Message string `json:"message"`
	}

	data, err := json.Marshal(&AckJsonResponse{Message: "subscription successful"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleGetTransactions handles the /transactions endpoint
func (p *parserHttp) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	address := r.PathValue("address")
	if address == "" {
		http.Error(w, "Missing address", http.StatusBadRequest)
		return
	}

	type AckJsonResponse struct {
		Data    interface{} `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	transactions := p.svc.GetTransactions(address)
	data, err := json.Marshal(&AckJsonResponse{Message: "success", Data: transactions})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
