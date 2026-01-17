package main

import "strings"

// TronGridResp represents the response from TronGrid API
type TronGridResp struct {
	Data []Event `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
		PageSize    int    `json:"page_size"`
	} `json:"meta"`
	Success bool `json:"success"`
}

// Event represents a TronGrid event
type Event struct {
	BlockNumber    int64                  `json:"block_number"`
	BlockTimestamp int64                  `json:"block_timestamp"`
	EventIndex     int64                  `json:"event_index"`
	EventName      string                 `json:"event_name"`
	TransactionID  string                 `json:"transaction_id"`
	Unconfirmed    bool                   `json:"_unconfirmed"`
	Result         map[string]interface{} `json:"result"`
}

// ToHex returns the "to" address in hex format (lowercase)
func (e Event) ToHex() string {
	v, _ := e.Result["to"].(string)
	return strings.ToLower(v)
}

// FromHex returns the "from" address in hex format (lowercase)
func (e Event) FromHex() string {
	v, _ := e.Result["from"].(string)
	return strings.ToLower(v)
}

// ValueStr returns the "value" as a string
func (e Event) ValueStr() string {
	v, _ := e.Result["value"].(string)
	return v
}
