package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"txn-parser/model"
	"txn-parser/repository"
)

type Parser interface {
	GetCurrentBlock() int
	Subscribe(address string) bool
	GetTransactions(address string) []model.Transaction
}

type ethereumParserService struct {
	repo repository.ParserRepository
}

const ethRPCURL = "https://ethereum-rpc.publicnode.com"

func NewEthereumParserService(repo repository.ParserRepository) *ethereumParserService {
	return &ethereumParserService{
		repo: repo,
	}
}

// GetCurrentBlock returns the last parsed block number
func (p *ethereumParserService) GetCurrentBlock() int {
	return p.repo.GetCurrentBlockNumber()
}

// Subscribe adds address to observer
func (p *ethereumParserService) Subscribe(address string) bool {
	lowerCaseAddress := strings.ToLower(address)
	if p.repo.GetAddress(lowerCaseAddress) {
		return false
	}

	return p.repo.AddAddress(lowerCaseAddress)
}

// GetTransactions returns observed txns to subscribed address
func (p *ethereumParserService) GetTransactions(address string) []model.Transaction {
	lowerCaseAddress := strings.ToLower(address)

	return p.repo.GetTransactions(lowerCaseAddress)
}

// fetchBlockNumber fetches the current block number from the Ethereum node
func (p *ethereumParserService) fetchBlockNumber() (int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	response, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("error marshaling payload: %v", err)
	}

	resp, err := http.Post(ethRPCURL, "application/json", bytes.NewBuffer(response))
	if err != nil {
		return 0, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// rate limit exceeded
	if resp.StatusCode == 429 {
		return 0, fmt.Errorf("rate limit exceeded: %v", err)
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	blockNumber, err := strconv.ParseInt(result.Result[2:], 16, 64) // Remove "0x" prefix
	if err != nil {
		return 0, fmt.Errorf("error parsing block number: %v", err)
	}

	return int(blockNumber), nil
}

// fetchBlock fetches the block by block number from the Ethereum node
func (p *ethereumParserService) fetchBlock(blockNumber int) (*model.Block, error) {
	blockNumberHex := "0x" + strconv.FormatInt(int64(blockNumber), 16)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumberHex, true},
		"id":      1,
	}
	response, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %v", err)
	}

	resp, err := http.Post(ethRPCURL, "application/json", bytes.NewBuffer(response))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// rate limit exceeded
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit exceeded: %v", err)
	}

	var result struct {
		Result *model.Block `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Check if the block was found
	if result.Result == nil {
		return nil, fmt.Errorf("block not found: %v", err)
	}

	return result.Result, nil
}

// processBlock processes a block to extract transactions for subscribed addresses
func (p *ethereumParserService) processBlock(blockNumber int) {
	block, err := p.fetchBlock(blockNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	var txns = make(map[string][]model.Transaction)

	// collate all txns for subscribed addresses
	for _, txn := range block.Transactions {
		from := strings.ToLower(txn.From)
		to := strings.ToLower(txn.To)

		// add outbound txn
		if p.repo.GetAddress(from) {
			txns[from] = append(txns[from], txn)
		}

		// add inbound txn
		if from != to && p.repo.GetAddress(to) {
			txns[to] = append(txns[to], txn)
		}
	}

	// save txns
	p.repo.AddTransactions(txns)
}

// StartParsing parses block txns at most 5 blocks from currentBlock at the specified interval
func (p *ethereumParserService) StartParsing(ctx context.Context, intervalInSecond uint) {
	if intervalInSecond == 0 {
		intervalInSecond = 5
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("parser done")
			return
		default:
			// check if any subscription
			if p.repo.GetAddressCount() == 0 {
				continue
			}

			currentBlock, err := p.fetchBlockNumber()
			if err != nil {
				fmt.Printf("Error fetching block number: %v", err)
			} else {
				// start 5 blocks from current
				startBlock := p.repo.GetCurrentBlockNumber()
				if startBlock == 0 {
					startBlock = currentBlock - 5
				}

				if currentBlock > startBlock {
					fmt.Printf("processing blockNumber %d - %d\n", startBlock+1, currentBlock)

					for i := startBlock + 1; i <= currentBlock; i++ {
						go func(i int) {
							p.processBlock(i)
						}(i)
					}

					p.repo.SetCurrentBlockNumber(uint(currentBlock))
				}
			}

			// sleep to avoid rate-limiting
			time.Sleep(time.Second * time.Duration(intervalInSecond))
		}
	}
}
