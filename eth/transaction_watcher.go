package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	infuraURL     = "https://mainnet.infura.io/v3/17765d4ea21943b09d2b02370131ac3a" // 替换为你的
	usdtAddress   = "0xdAC17F958D2ee523a2206206994597C13D831ec7"                    // ERC20 USDT
	targetAddress = "0xc8Fb0Ec6C8331cE5e014a34E7e2adc85BC9C701A"                    // 替换为你监听的地址
)

func main() {
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("Failed to connect to Infura: %v", err)
	}
	defer client.Close()

	contractAddress := common.HexToAddress(usdtAddress)
	toAddress := common.HexToAddress(targetAddress)

	// ERC20 Transfer(address indexed from, address indexed to, uint256 value)
	transferEventSig := []byte("Transfer(address,address,uint256)")
	transferSigHash := common.BytesToHash(cryptoKeccak256(transferEventSig))

	// topic[0]: event signature hash
	// topic[1]: from (optional)
	// topic[2]: to (我们关心的地址)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{transferSigHash}, nil, {common.BytesToHash(toAddress.Bytes())}},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to fetch logs: %v", err)
	}

	// 解析 ABI
	erc20ABI, err := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	for _, vLog := range logs {
		// 解析日志
		var transferEvent struct {
			From  common.Address
			To    common.Address
			Value *big.Int
		}
		err := erc20ABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
		if err != nil {
			log.Printf("Failed to unpack log data: %v", err)
			continue
		}
		transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
		transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

		fmt.Printf("✅ 收到转账:\n")
		fmt.Printf("From: %s\n", transferEvent.From.Hex())
		fmt.Printf("To:   %s\n", transferEvent.To.Hex())
		fmt.Printf("Amount: %s USDT (raw: %s)\n", new(big.Float).Quo(new(big.Float).SetInt(transferEvent.Value), big.NewFloat(1e6)).Text('f', 6), transferEvent.Value.String())
		fmt.Printf("TxHash: %s\n\n", vLog.TxHash.Hex())
	}
}

// keccak256 hash（同 go-ethereum 的 crypto/sha3）
func cryptoKeccak256(data []byte) []byte {
	hash := make([]byte, 32)
	copy(hash, crypto.Keccak256(data))
	return hash
}
