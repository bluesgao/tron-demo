package main

import "time"

const (
	// UsdtContract is the USDT contract address on Tron
	UsdtContract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

	// TronGridUrl is the base URL for TronGrid API events endpoint
	TronGridUrl = "https://api.trongrid.io/v1/contracts/%s/events"

	// EventNameTransfer is the event name for Transfer events
	EventNameTransfer = "Transfer"

	// HeaderTronProAPIKey is the header name for TronGrid API key
	HeaderTronProAPIKey = "TRON-PRO-API-KEY"

	// DefaultLimit is the default number of events to fetch per request
	DefaultLimit = 200

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 10 * time.Second

	// IdempotencyKeySeparator is the separator used in idempotency keys
	IdempotencyKeySeparator = "#"

	// PollInterval is the interval between polling requests
	PollInterval = 5 * time.Second

	// LookbackWindow is the time window to look back from current time
	// Monitor transactions from (current time - LookbackWindow) to current time
	LookbackWindow = 1 * time.Minute

	// TronGridAPIKey is the API key for TronGrid API
	TronGridAPIKey = "2a99aae7-b9b6-4a4c-b3fd-71feb1d5a177"
)
