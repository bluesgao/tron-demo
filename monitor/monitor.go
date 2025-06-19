package monitor

import (
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"time"
)

type TransactionEvent struct {
	Address string
	Topics  string
	Data    string
}

func monitorTransactionEvents(c *client.GrpcClient, txID string) error {
	// Get transaction info which includes events
	txInfo, err := c.GetTransactionInfoByID(txID)
	if err != nil {
		return err
	}

	// Check transaction status
	if txInfo.Result != core.TransactionInfo_SUCESS {
		return fmt.Errorf("transaction failed: %s", txInfo.Result.String())
	}

	// Process contract events
	for _, event := range txInfo.Log {
		fmt.Printf("Event Topics:\n")
		for i, topic := range event.Topics {
			fmt.Printf("  Topic[%d]: %x\n", i, topic)
		}
		fmt.Printf("Event Data: %x\n", event.Data)
		fmt.Printf("Contract Address: %x\n", event.Address)

		// Parse event data based on ABI
		// Example: Transfer event
		// topic[0] = keccak256("Transfer(address,address,uint256)")
		// topic[1] = from address
		// topic[2] = to address
		// data = amount
	}

	// Process internal transactions
	for _, internal := range txInfo.InternalTransactions {
		fmt.Printf("Internal TX: %x\n", internal.Hash)
		fmt.Printf("  From: %x\n", internal.CallerAddress)
		fmt.Printf("  To: %x\n", internal.TransferToAddress)
		fmt.Printf("  Amount: %d\n", internal.CallValueInfo)
	}

	return nil
}

// Monitor new blocks for events
func MonitorBlockEvents(c *client.GrpcClient, startBlock int64) error {
	currentBlock := startBlock

	for {
		// Get block
		block, err := c.GetBlockByNum(currentBlock)
		if err != nil {
			// Block might not exist yet
			time.Sleep(3 * time.Second)
			continue
		}

		// Process transactions in block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.Txid)

			// Get transaction info
			txInfo, err := c.GetTransactionInfoByID(txID)
			if err != nil {
				continue
			}

			// Check if transaction has events
			if len(txInfo.Log) > 0 {
				fmt.Printf("Transaction %s has %d events\n", txID, len(txInfo.Log))
				monitorTransactionEvents(c, txID)
			}
		}

		currentBlock++
		time.Sleep(3 * time.Second) // TRON block time
	}
}
