package repository

import (
	"sync"
	"txn-parser/model"
)

type ParserRepository interface {
	GetCurrentBlockNumber() int
	SetCurrentBlockNumber(num uint) bool
	AddAddress(address string) bool
	GetAddress(address string) bool
	GetAddressCount() int
	GetTransactions(address string) []model.Transaction
	AddTransactions(map[string][]model.Transaction) bool
}

type inMemoryParserRepository struct {
	currentBlock int
	addresses    map[string]bool
	transactions map[string][]model.Transaction
	mu           sync.RWMutex
}

func NewInMemoryParserRepository() *inMemoryParserRepository {
	return &inMemoryParserRepository{
		addresses:    make(map[string]bool),
		transactions: make(map[string][]model.Transaction),
	}
}

func (r *inMemoryParserRepository) GetCurrentBlockNumber() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.currentBlock
}

func (r *inMemoryParserRepository) SetCurrentBlockNumber(num uint) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.currentBlock = int(num)
	return true
}

func (r *inMemoryParserRepository) AddAddress(address string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.addresses[address] = true
	return true
}

func (r *inMemoryParserRepository) GetAddress(address string) bool {
	return r.addresses[address]
}

func (r *inMemoryParserRepository) GetAddressCount() int {
	return len(r.addresses)
}

func (r *inMemoryParserRepository) GetTransactions(address string) []model.Transaction {
	r.mu.Lock()
	defer r.mu.Unlock()

	if transactions, exists := r.transactions[address]; exists {
		return transactions
	}

	return []model.Transaction{}
}

func (r *inMemoryParserRepository) AddTransactions(txns map[string][]model.Transaction) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range txns {
		r.transactions[k] = append(r.transactions[k], v...)
	}

	return true
}
