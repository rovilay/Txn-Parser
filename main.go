package main

import (
	"context"
	"os"
	"os/signal"

	httpService "txn-parser/http"
	"txn-parser/repository"
	"txn-parser/service"
)

const PORT = 8080

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	repo := repository.NewInMemoryParserRepository()
	parser := service.NewEthereumParserService(repo)
	server := httpService.NewParserHttp(parser)

	// parse ethereum blocks every 10 seconds
	// Note: parsing starts once there is at least one subscription
	go parser.StartParsing(ctx, 10)

	// start API server
	server.Start(ctx, PORT)
}
