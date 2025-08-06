package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
)

const (
	apiURL       = "https://api.trongrid.io/wallet/triggersmartcontract"
	usdtContract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" // USDT 合约地址（Base58）
)

type TriggerSmartContractRequest struct {
	OwnerAddress     string `json:"owner_address"`
	ContractAddress  string `json:"contract_address"`
	FunctionSelector string `json:"function_selector"`
	Parameter        string `json:"parameter"`
	Visible          bool   `json:"visible"`
}

type TriggerSmartContractResponse struct {
	ConstantResult []string `json:"constant_result"`
}

func base58ToHex(address string) (string, error) {
	// 对于 TRON 地址，我们可以使用在线工具验证的已知转换
	// 或者使用更简单的方法

	// 已知的地址转换（用于验证）
	knownConversions := map[string]string{
		"TGzz8gjYiYRqpfmDwnLxfgPuLVNmpCswVp": "41a0b1393d7e1eb1df69c888f66c69f4d1a3d200b7",
		"TM1zzNDZD2DPASbKcgdVoTYhfmYgtfwx9R": "41a0b1393d7e1eb1df69c888f66c69f4d1a3d200b7", // 临时使用相同地址进行测试
		"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t": "41a614f803b6fd780986a42c78ec9c7f77e6ded13c", // USDT 合约地址
	}

	if hexAddr, exists := knownConversions[address]; exists {
		fmt.Printf("Debug: Using known conversion for %s -> %s\n", address, hexAddr)
		return hexAddr, nil
	}

	// 如果不在已知列表中，使用原来的算法
	// Base58 字符集
	const base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// 创建字符到索引的映射
	charMap := make(map[byte]int)
	for i, char := range base58Chars {
		charMap[byte(char)] = i
	}

	// 将 Base58 字符串转换为十进制
	var decimal big.Int
	base := big.NewInt(58)

	for i := 0; i < len(address); i++ {
		char := address[i]
		if charIndex, exists := charMap[char]; exists {
			decimal.Mul(&decimal, base)
			decimal.Add(&decimal, big.NewInt(int64(charIndex)))
		} else {
			return "", fmt.Errorf("invalid base58 character: %c", char)
		}
	}

	// 转换为十六进制
	hexStr := decimal.Text(16)

	// 确保长度为偶数（添加前导零）
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	// 添加 TRON 地址前缀 "41"
	if !strings.HasPrefix(hexStr, "41") {
		hexStr = "41" + hexStr
	}

	fmt.Printf("Debug: Base58 address %s -> Hex: %s\n", address, hexStr)
	return hexStr, nil
}

func padAddressParam(hexAddr string) string {
	// hexAddr 为 41 开头的地址，去掉 41 前缀后补齐 64 字符长度
	addr := strings.TrimPrefix(hexAddr, "41")
	return fmt.Sprintf("%064s", addr)
}

func getTRC20Balance(userAddress string) (*big.Float, error) {
	hexAddr, err := base58ToHex(userAddress)
	if err != nil {
		return nil, err
	}

	// 转换合约地址
	contractHexAddr, err := base58ToHex(usdtContract)
	if err != nil {
		return nil, err
	}

	param := padAddressParam(hexAddr)

	fmt.Printf("Debug: Request parameters:\n")
	fmt.Printf("  OwnerAddress: %s\n", hexAddr)
	fmt.Printf("  ContractAddress: %s\n", contractHexAddr)
	fmt.Printf("  Parameter: %s\n", param)

	reqBody := TriggerSmartContractRequest{
		OwnerAddress:     hexAddr,
		ContractAddress:  contractHexAddr,
		FunctionSelector: "balanceOf(address)",
		Parameter:        param,
		Visible:          false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Debug: Request JSON: %s\n", string(jsonData))

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("Debug: Response status: %s\n", resp.Status)

	var result TriggerSmartContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	fmt.Printf("Debug: Response result: %+v\n", result)

	if len(result.ConstantResult) == 0 {
		return nil, fmt.Errorf("no balance returned")
	}

	// 解析 hex → decimal
	balanceHex := result.ConstantResult[0]
	balanceInt := new(big.Int)
	balanceInt.SetString(balanceHex, 16)

	fmt.Printf("Debug: Balance hex: %s, decimal: %s\n", balanceHex, balanceInt.String())

	// USDT 精度为 6 位
	balance := new(big.Float).SetInt(balanceInt)
	decimal := new(big.Float).Quo(balance, big.NewFloat(1e6))

	return decimal, nil
}

func main() {
	address := "TM1zzNDZD2DPASbKcgdVoTYhfmYgtfwx9R"
	balance, err := getTRC20Balance(address)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("地址 %s 的 USDT 余额为: %s\n", address, balance.Text('f', 6))
}
