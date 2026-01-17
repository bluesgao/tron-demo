package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/fbsobreira/gotron-sdk/pkg/account"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/store"
	"github.com/yourname/tron-demo/block"
	"github.com/yourname/tron-demo/monitor"
	"google.golang.org/grpc"
)

func main() {
	// 主网地址
	tornEndpoint := "grpc.trongrid.io:50051"
	gRPCWalletClient := client.NewGrpcClient(tornEndpoint)
	err := gRPCWalletClient.Start(grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to start grpc client: %v", err)
		return
	}

	//err = getAccountInfo(gRPCWalletClient, "TRvzGHTsfgbVkrFjCovFtyLU4HBd3u6Fdw")
	//if err != nil {
	//	log.Fatalf("获取账户信息失败: %v", err)
	//	return
	//}
	//
	//设置合约地址（USDT合约在主网的地址）
	contractAddr := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" // TRC20 USDT
	// 设置用户地址
	userAddress := "TRvzGHTsfgbVkrFjCovFtyLU4HBd3u6Fdw"

	balance, err := getTRC20Balance(gRPCWalletClient, userAddress, contractAddr)
	if err != nil {
		log.Fatalf("获取USDT余额失败: %v", err)
	}

	fmt.Printf("USDT余额: %s (单位：6位精度)\n", balance)

	//TBk1CRZnfBqpY7Wr9DbfdJfpdVf3nGgk16
	//behind sound trust make tray steak game jeans regret three coil dog hole cinnamon flat cart antique valley canyon laundry dinosaur real fuel potato
	//createAccount()
	//err = getAccountInfo(gRPCWalletClient, "TBk1CRZnfBqpY7Wr9DbfdJfpdVf3nGgk16")
	//if err != nil {
	//	log.Fatalf("获取账户信息失败: %v", err)
	//	return
	//}

	num, err := block.GetCurrentBlockNum(gRPCWalletClient)
	if err != nil {
		return
	}
	fmt.Println("当前区块高度：", num)

	//从当前最新高度开始监听
	startBlock := num // 例如你查询过的某个起始区块高度
	err = monitor.MonitorBlockEvents(gRPCWalletClient, startBlock)
	if err != nil {
		fmt.Println("监听失败:", err)
	}

	// 使用自定义配置
	//fmt.Println("\n=== 使用自定义配置获取转账交易 ===")
	//config := &trongrid.Config{
	//	BaseURL:  "https://api.trongrid.io/v1/accounts/%s/transactions",
	//	Address:  "TZAw4M78JonPirHnA1r5dfx5L954nay3DQ",
	//	PageSize: 100,
	//	// APIKey: "your-api-key-here", // 如果需要API密钥，请取消注释并填入
	//}
	//
	//transfers, err := trongrid.GetTRC20Transactions(config)
	//if err != nil {
	//	fmt.Printf("获取TRC20交易失败: %v\n", err)
	//} else {
	//	fmt.Printf("找到 %d 笔转账交易\n", len(transfers))
	//	trongrid.PrintTransfers(transfers)
	//}
	//
	//// 汇总 转入，转出
	//sumIn, sumOut := trongrid.SumInOut(transfers)
	//fmt.Printf("转入: %d, 转出: %d\n", sumIn, sumOut)

	//block.GetBlockByNum(gRPCWalletClient, int64(73216128))

}

// 获取当前区块
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
func getTRC20Balance(c *client.GrpcClient, account, contract string) (*big.Int, error) {
	contractAddr, err := address.Base58ToAddress(contract)
	if err != nil {
		return nil, fmt.Errorf("合约地址无效: %w", err)
	}

	accountAddr, err := address.Base58ToAddress(account)
	if err != nil {
		return nil, fmt.Errorf("账户地址无效: %w", err)
	}

	result, err := c.TRC20ContractBalance(accountAddr.String(), contractAddr.String())
	log.Printf("result:%s, err:%s", result, err)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int).SetBytes(result.Bytes())
	return balance, nil
}

// 创建账户
func createAccount() {
	// 创建本地账户（助记词 + keystore 存储）
	acc := &account.Creation{
		Name:       "myAccount",
		Passphrase: "StrongPassword123!",
	}

	if err := account.CreateNewLocalAccount(acc); err != nil {
		log.Fatalf("创建账户失败: %v", err)
	}

	// 获取地址
	addr, err := store.AddressFromAccountName(acc.Name)
	if err != nil {
		log.Fatalf("获取地址失败: %v", err)
	}

	fmt.Println("✅ 创建成功！")
	fmt.Printf("地址 Base58: %s\n", addr)
	fmt.Printf("助记词: %s\n", acc.Mnemonic)
}
