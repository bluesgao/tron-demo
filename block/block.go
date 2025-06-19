package block

import (
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"log"
)

// 获取当前区块高度
func GetCurrentBlockNum(c *client.GrpcClient) (int64, error) {
	latest, err := c.GetBlockByLatestNum(1)
	if err != nil {
		log.Fatalf("获取最新块失败: %v", err)
	}
	block := latest.Block[0]
	fmt.Printf("✅ 当前 TRON 区块高度： %d\n", block.BlockHeader.RawData.Number)
	return block.BlockHeader.RawData.Number, nil
}

// 根据num获取区块信息
func GetBlockByNum(c *client.GrpcClient, num int64) {
	block, err := c.GetBlockByNum(num)
	if err != nil {
		log.Fatalf("获取块失败: %v", err)
	}
	fmt.Printf("区块 %d , ID %s 中的交易数量：%d\n", block.BlockHeader.RawData.Number, string(block.GetBlockid()), len(block.Transactions))
}
