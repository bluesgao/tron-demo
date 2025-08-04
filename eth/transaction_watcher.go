package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/17765d4ea21943b09d2b02370131ac3a")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// USDT 合约地址（Ethereum 主网）
	contractAddress := common.HexToAddress("0x49719D256A5eA16bFa579ea16E95bea9fa41A452")

	// 要监听的用户地址
	targetAddress := common.HexToAddress("0x1EaCa1277BcDFa83E60658D8938B3D63cD3E63C1")

	// 获取当前区块高度
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fromBlock := big.NewInt(0).Sub(header.Number, big.NewInt(100)) // 最近 100 个块
	toBlock := header.Number

	// 事件签名：Transfer(address,address,uint256)
	eventSignature := []byte("Transfer(address,address,uint256)")
	eventSigHash := crypto.Keccak256Hash(eventSignature)

	// 地址需要 padding 为 32 字节
	topicTo := common.LeftPadBytes(targetAddress.Bytes(), 32)

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{eventSigHash},                // Transfer event
			nil,                           // from 可选
			{common.BytesToHash(topicTo)}, // to 为目标地址
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	// 打印日志
	for _, vLog := range logs {
		fmt.Printf("TxHash: %s\n", vLog.TxHash.Hex())
		fmt.Printf("From: %s\n", common.BytesToAddress(vLog.Topics[1].Bytes()))
		fmt.Printf("To: %s\n", common.BytesToAddress(vLog.Topics[2].Bytes()))
		amount := new(big.Int).SetBytes(vLog.Data)
		fmt.Printf("Amount (USDT smallest unit): %s\n", amount.String())
		fmt.Println("-----")
	}
}
