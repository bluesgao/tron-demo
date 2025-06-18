package main

import (
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"log"
	"math/big"
)

func main() {
	tornEndpoint := "grpc.trongrid.io:50051"
	//credential := "your-credential"
	//caPath := "your-ca.crt-file-path"
	//creds, err := credentials.NewClientTLSFromFile(caPath, "")
	//if err != nil {
	//	fmt.Printf("failed to load credentials: %v\n", err)
	//}
	gRPCWalletClient := client.NewGrpcClient(tornEndpoint)
	//gRPCWalletClient.SetAPIKey(credential)
	gRPCWalletClient.Start(grpc.WithInsecure())

	getAccountInfo(gRPCWalletClient, "TRvzGHTsfgbVkrFjCovFtyLU4HBd3u6Fdw")

	// 你的 TRON 地址（Base58 格式）

	tokenContract := "TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj" // TRC20 USDT

	userAddress := "TRvzGHTsfgbVkrFjCovFtyLU4HBd3u6Fdw"

	balance, err := getTRC20Balance(gRPCWalletClient, tokenContract, userAddress)
	if err != nil {
		log.Fatalf("获取USDT余额失败: %v", err)
	}

	fmt.Printf("USDT余额: %s (单位：6位精度)\n", balance.String())
}

func getNowBlock(c *client.GrpcClient) {
	resp, err := c.GetNowBlock()
	if err != nil {
		fmt.Printf("failed to get now block: %v\n", err)
	}
	fmt.Println("wallet resp: ", resp)
}

// 只能获取trc20地址的信息（其中余额是TRX）
func getAccountInfo(c *client.GrpcClient, addr string) error {
	// Convert address
	tronAddr, err := address.Base58ToAddress(addr)
	if err != nil {
		return err
	}

	// Get account
	account, err := c.GetAccount(tronAddr.String())
	if err != nil {
		return err
	}

	// Display information
	fmt.Printf("Address: %s\n", addr)
	fmt.Printf("Balance: %d sun (%f TRX)\n",
		account.Balance,
		float64(account.Balance)/1e6)
	fmt.Printf("Created: %v\n", account.CreateTime)

	// Check resources
	res, err := c.GetAccountResource(tronAddr.String())
	if err == nil {
		fmt.Printf("Bandwidth: %d/%d\n",
			res.FreeNetUsed,
			res.FreeNetLimit)
		fmt.Printf("Energy: %d/%d\n",
			res.EnergyUsed,
			res.EnergyLimit)
	}

	return nil
}

// 获取usdt余额
func getTRC20Balance(c *client.GrpcClient, tokenContract, account string) (*big.Int, error) {
	contractAddr, err := address.Base58ToAddress(tokenContract)
	if err != nil {
		return nil, fmt.Errorf("合约地址无效: %w", err)
	}

	accountAddr, err := address.Base58ToAddress(account)
	if err != nil {
		return nil, fmt.Errorf("账户地址无效: %w", err)
	}

	// balanceOf(address) 方法签名
	methodID := "70a08231"
	//addrHex := accountAddr.Hex()[2:] // 去掉 "41" 前缀
	//addrBytes, err := hex.DecodeString(addrHex)
	//if err != nil {
	//	return nil, fmt.Errorf("地址 hex 解码失败: %w", err)
	//}

	//data := append(common.Hex2Bytes(methodID), common.LeftPadBytes(addrBytes, 32)...)

	// 关键调用（TriggerConstantContract）
	result, err := c.TriggerConstantContract(accountAddr.String(), contractAddr.String(), methodID, "")
	if err != nil {
		return nil, fmt.Errorf("TriggerConstantContract 调用失败: %w", err)
	}

	if !result.GetResult().GetResult() {
		return nil, fmt.Errorf("TRC20 合约调用失败: %s", result.GetResult().GetMessage())
	}

	if len(result.GetConstantResult()) == 0 {
		return nil, fmt.Errorf("合约未返回数据")
	}

	balance := new(big.Int).SetBytes(result.GetConstantResult()[0])
	return balance, nil
}
