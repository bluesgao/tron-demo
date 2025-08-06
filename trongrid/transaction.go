package trongrid

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// TransactionData 表示交易数据的结构
type TransactionData struct {
	Ret                  []Ret         `json:"ret"`
	Signature            []string      `json:"signature"`
	TxID                 string        `json:"txID"`
	NetUsage             int           `json:"net_usage"`
	RawDataHex           string        `json:"raw_data_hex"`
	NetFee               int           `json:"net_fee"`
	EnergyUsage          int           `json:"energy_usage"`
	BlockNumber          int           `json:"blockNumber"`
	BlockTimestamp       int64         `json:"block_timestamp"`
	EnergyFee            int           `json:"energy_fee"`
	EnergyUsageTotal     int           `json:"energy_usage_total"`
	RawData              RawData       `json:"raw_data,omitempty"`
	InternalTransactions []interface{} `json:"internal_transactions"`
}

// Ret 表示交易返回结果
type Ret struct {
	ContractRet string `json:"contractRet"`
	Fee         int    `json:"fee"`
}

// Value 表示合约参数值
type Value struct {
	Balance         int64                  `json:"balance"`
	Resource        string                 `json:"resource"`
	ReceiverAddress string                 `json:"receiver_address"`
	OwnerAddress    string                 `json:"owner_address"`
	ContractAddress string                 `json:"contract_address"`
	Data            string                 `json:"data"`   // 合约调用 data
	Other           map[string]interface{} `json:"-"`      // 备用
	Amount          json.Number            `json:"amount"` // 备用字段
}

// Parameter 表示合约参数
type Parameter struct {
	Value   Value  `json:"value"`
	TypeURL string `json:"type_url"`
}

// Contract 表示智能合约
type Contract struct {
	Parameter Parameter `json:"parameter"`
	Type      string    `json:"type"`
}

// RawData 表示原始交易数据
type RawData struct {
	Contract      []Contract `json:"contract"`
	RefBlockBytes string     `json:"ref_block_bytes"`
	RefBlockHash  string     `json:"ref_block_hash"`
	Expiration    int64      `json:"expiration"`
	Timestamp     int64      `json:"timestamp"`
}

// JSONResponse 表示API响应结构
type JSONResponse struct {
	Data    []TransactionData `json:"data"`
	Success bool              `json:"success"`
	Meta    Meta              `json:"meta"`
}

// Links 表示分页链接
type Links struct {
	Next string `json:"next"`
}

// Meta 表示元数据
type Meta struct {
	At          int64  `json:"at"`
	Fingerprint string `json:"fingerprint"`
	Links       Links  `json:"links"`
	PageSize    int    `json:"page_size"`
}

// TRC20Transfer 表示TRC20转账信息
type TRC20Transfer struct {
	TxID            string
	FromAddress     string
	ToAddress       string
	ContractAddress string
	Amount          *big.Int
	BlockNumber     int
	Timestamp       time.Time
	Status          string
	// 以下字段用于计算手续费
	EnergyFee        int
	EnergyUsageTotal int
	NetUsage         int
	NetFee           int
	EnergyUsage      int
}

// Config 表示配置信息
type Config struct {
	BaseURL  string
	Address  string
	PageSize int
	APIKey   string // 可选的API密钥
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BaseURL:  "https://api.trongrid.io/v1/accounts/%s/transactions",
		Address:  "TUk2k7gSZGs9xquWH6XMGa8BWWvq2m6hbd",
		PageSize: 50, // 增加页面大小以获取更多数据
	}
}

// GetTRC20Transactions 获取所有类型的转账交易（包括TRC20、TRX、TRC10）
func GetTRC20Transactions(config *Config) ([]TRC20Transfer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var transfers []TRC20Transfer
	fingerprint := ""
	pageCount := 0

	for {
		pageCount++
		url := fmt.Sprintf(config.BaseURL, config.Address) + fmt.Sprintf("?limit=%d&only_confirmed=true", config.PageSize)
		if fingerprint != "" {
			url += fmt.Sprintf("&fingerprint=%s", fingerprint)
		}

		fmt.Printf("正在获取第 %d 页数据...\n", pageCount)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
		}
		req.Header.Set("accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TronDemo/1.0)")

		// 如果提供了API密钥，添加到请求头
		if config.APIKey != "" {
			req.Header.Set("TRON-PRO-API-KEY", config.APIKey)
		}

		fmt.Printf("请求URL: %s\n", url)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP请求失败: %w", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}

		var result JSONResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %w", err)
		}

		// 打印API响应信息
		fmt.Printf("第 %d 页API响应: Success=%v, 数据条数=%d\n", pageCount, result.Success, len(result.Data))

		// 如果没有数据，说明已经到达最后一页
		if len(result.Data) == 0 {
			fmt.Printf("第 %d 页没有数据，已到达最后一页\n", pageCount)
			break
		}

		fmt.Printf("第 %d 页获取到 %d 笔交易\n", pageCount, len(result.Data))

		// 处理交易数据
		pageTransfers := 0
		for _, tx := range result.Data {
			transfer, err := parseTransaction(tx)
			if err != nil {
				fmt.Printf("[TxID %s] 解析交易失败: %v\n", tx.TxID, err)
				continue
			}
			if transfer != nil {
				transfers = append(transfers, *transfer)
				pageTransfers++
			}
		}

		fmt.Printf("第 %d 页解析出 %d 笔TRC20转账\n", pageCount, pageTransfers)

		// 检查是否有下一页（fingerprint为空表示没有下一页）
		if result.Meta.Fingerprint == "" {
			fmt.Printf("第 %d 页的fingerprint为空，已到达最后一页\n", pageCount)
			break
		}

		// 更新fingerprint用于下一页请求
		fingerprint = result.Meta.Fingerprint
		fmt.Printf("下一页fingerprint: %s\n", fingerprint)

		// 添加延迟避免请求过于频繁
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("总共获取到 %d 笔TRC20转账交易\n", len(transfers))
	return transfers, nil
}

// parseTransaction 解析单个交易
func parseTransaction(tx TransactionData) (*TRC20Transfer, error) {
	if len(tx.RawData.Contract) == 0 {
		fmt.Printf("[TxID %s] 跳过：无合约数据\n", tx.TxID)
		return nil, nil
	}

	contract := tx.RawData.Contract[0]
	fmt.Printf("[TxID %s] 合约类型: %s\n", tx.TxID, contract.Type)

	// 根据不同的合约类型进行解析
	switch contract.Type {
	case "TriggerSmartContract":
		return parseTriggerSmartContract(tx, contract)
	case "TransferContract":
		//return parseTransferContract(tx, contract)
		return nil, nil
	case "TransferAssetContract":
		//return parseTransferAssetContract(tx, contract)
		return nil, nil
	default:
		fmt.Printf("[TxID %s] 跳过：不支持的合约类型: %s\n", tx.TxID, contract.Type)
		return nil, nil
	}
}

// parseTriggerSmartContract 解析智能合约触发交易
func parseTriggerSmartContract(tx TransactionData, contract Contract) (*TRC20Transfer, error) {
	v := contract.Parameter.Value

	// 检查是否有合约调用数据
	var contractData string
	if v.Data != "" {
		contractData = v.Data
		fmt.Printf("[TxID %s] 使用Data字段: %s\n", tx.TxID, contractData)
	} else if v.Resource != "" {
		contractData = v.Resource
		fmt.Printf("[TxID %s] 使用Resource字段: %s\n", tx.TxID, contractData)
	} else {
		fmt.Printf("[TxID %s] 跳过：无合约调用数据\n", tx.TxID)
		return nil, nil
	}

	// 检查数据长度
	if len(contractData) < 8 {
		fmt.Printf("[TxID %s] 跳过：合约数据太短: %s\n", tx.TxID, contractData)
		return nil, nil
	}

	// 检查前4字节是否为transfer方法签名
	methodSig := contractData[:8]
	// a9059cbb，即 transfer(address,uint256) 方法的签名
	// 095ea7b3，这是 approve(address,uint256) 的方法签名
	// approve 是 TRC20（和 ERC20）标准中非常重要的一个方法，它不是用来转账，而是授权第三方账户可以代表你转账指定数量的代币。
	if methodSig != "a9059cbb" {
		fmt.Printf("[TxID %s] 跳过：不是TRC20 transfer方法 (期望: a9059cbb, 实际: %s)\n", tx.TxID, methodSig)
		return nil, nil
	}

	fmt.Printf("[TxID %s] 发现TRC20 transfer交易\n", tx.TxID)

	toAddress, amount, err := parseTRC20Data(contractData)
	if err != nil {
		return nil, fmt.Errorf("TRC20数据解析失败: %w", err)
	}

	fromAddress := hexToBase58Check(v.OwnerAddress)
	contractAddress := hexToBase58Check(v.ContractAddress)
	timestamp := time.UnixMilli(tx.RawData.Timestamp)

	status := "UNKNOWN"
	if len(tx.Ret) > 0 {
		status = tx.Ret[0].ContractRet
	}

	fmt.Printf("[TxID %s] 解析成功: From=%s, To=%s, Contract=%s, Amount=%s\n",
		tx.TxID, fromAddress, toAddress, contractAddress, amount.String())

	return &TRC20Transfer{
		TxID:             tx.TxID,
		FromAddress:      fromAddress,
		ToAddress:        toAddress,
		ContractAddress:  contractAddress,
		Amount:           amount,
		BlockNumber:      tx.BlockNumber,
		Timestamp:        timestamp,
		Status:           status,
		EnergyFee:        tx.EnergyFee,
		EnergyUsageTotal: tx.EnergyUsageTotal,
		NetUsage:         tx.NetUsage,
		NetFee:           tx.NetFee,
		EnergyUsage:      tx.EnergyUsage,
	}, nil
}

// parseTransferContract 解析TRX转账交易
func parseTransferContract(tx TransactionData, contract Contract) (*TRC20Transfer, error) {
	v := contract.Parameter.Value

	fmt.Printf("[TxID %s] 发现TRX转账交易\n", tx.TxID)

	fromAddress := hexToBase58Check(v.OwnerAddress)
	toAddress := hexToBase58Check(v.ReceiverAddress)

	// TRX转账使用Amount字段
	var amount *big.Int
	if v.Amount != "" {
		if amountInt, err := v.Amount.Int64(); err == nil {
			amount = big.NewInt(amountInt)
		} else {
			amount = big.NewInt(0)
		}
	} else {
		amount = big.NewInt(0)
	}

	blockNumber := tx.BlockNumber
	timestamp := time.UnixMilli(tx.RawData.Timestamp)

	status := "UNKNOWN"
	if len(tx.Ret) > 0 {
		status = tx.Ret[0].ContractRet
	}

	fmt.Printf("[TxID %s] TRX转账: From=%s, To=%s, Amount=%s\n",
		tx.TxID, fromAddress, toAddress, amount.String())

	return &TRC20Transfer{
		TxID:             tx.TxID,
		FromAddress:      fromAddress,
		ToAddress:        toAddress,
		ContractAddress:  "TRX", // TRX原生转账
		Amount:           amount,
		BlockNumber:      blockNumber,
		Timestamp:        timestamp,
		Status:           status,
		EnergyFee:        tx.EnergyFee,
		EnergyUsageTotal: tx.EnergyUsageTotal,
		NetUsage:         tx.NetUsage,
		NetFee:           tx.NetFee,
		EnergyUsage:      tx.EnergyUsage,
	}, nil
}

// parseTransferAssetContract 解析TRC10代币转账交易
func parseTransferAssetContract(tx TransactionData, contract Contract) (*TRC20Transfer, error) {
	v := contract.Parameter.Value

	fmt.Printf("[TxID %s] 发现TRC10代币转账交易\n", tx.TxID)

	fromAddress := hexToBase58Check(v.OwnerAddress)
	toAddress := hexToBase58Check(v.ReceiverAddress)

	// TRC10代币转账使用Amount字段
	var amount *big.Int
	if v.Amount != "" {
		if amountInt, err := v.Amount.Int64(); err == nil {
			amount = big.NewInt(amountInt)
		} else {
			amount = big.NewInt(0)
		}
	} else {
		amount = big.NewInt(0)
	}

	blockNumber := tx.BlockNumber
	timestamp := time.UnixMilli(tx.RawData.Timestamp)

	status := "UNKNOWN"
	if len(tx.Ret) > 0 {
		status = tx.Ret[0].ContractRet
	}

	fmt.Printf("[TxID %s] TRC10转账: From=%s, To=%s, Amount=%s\n",
		tx.TxID, fromAddress, toAddress, amount.String())

	return &TRC20Transfer{
		TxID:             tx.TxID,
		FromAddress:      fromAddress,
		ToAddress:        toAddress,
		ContractAddress:  "TRC10", // TRC10代币转账
		Amount:           amount,
		BlockNumber:      blockNumber,
		Timestamp:        timestamp,
		Status:           status,
		EnergyFee:        tx.EnergyFee,
		EnergyUsageTotal: tx.EnergyUsageTotal,
		NetUsage:         tx.NetUsage,
		NetFee:           tx.NetFee,
		EnergyUsage:      tx.EnergyUsage,
	}, nil
}

// parseTRC20Data 解析TRC20转账数据
func parseTRC20Data(data string) (string, *big.Int, error) {
	data = strings.TrimPrefix(data, "0x")
	if len(data) < 8+64+64 { // 方法签名+to(32字节)+amount(32字节)
		return "", nil, fmt.Errorf("数据长度太短: %d", len(data))
	}

	bytes, err := hex.DecodeString(data)
	if err != nil {
		return "", nil, fmt.Errorf("解码十六进制数据失败: %w", err)
	}

	// to地址在第16~36字节，截取20字节
	if len(bytes) < 68 {
		return "", nil, fmt.Errorf("数据长度不足，无法解析地址和金额")
	}

	toBytes := bytes[16:36]
	toAddress := hexToBase58Check(hex.EncodeToString(toBytes))

	// amount 是后面32字节
	amount := new(big.Int).SetBytes(bytes[36:68])

	return toAddress, amount, nil
}

// hexToBase58Check 将十六进制地址转换为Base58Check格式
func hexToBase58Check(hexAddr string) string {
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	raw, err := hex.DecodeString(hexAddr)
	if err != nil {
		return ""
	}

	if len(raw) == 20 {
		raw = append([]byte{0x41}, raw...)
	} else if len(raw) == 21 && raw[0] != 0x41 {
		raw[0] = 0x41
	}

	checksum := sha256Twice(raw)[:4]
	full := append(raw, checksum...)
	return base58Encode(full)
}

// base58Encode Base58编码
func base58Encode(b []byte) string {
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var result []byte
	x := new(big.Int).SetBytes(b)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for x.Cmp(zero) > 0 {
		x.QuoRem(x, base, mod)
		result = append(result, alphabet[mod.Int64()])
	}

	// 反转
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	// 前导0映射为1
	for _, v := range b {
		if v == 0 {
			result = append([]byte{'1'}, result...)
		} else {
			break
		}
	}

	return string(result)
}

// sha256Twice 双重SHA256哈希
func sha256Twice(b []byte) []byte {
	return sha256Sum(sha256Sum(b))
}

// sha256Sum SHA256哈希
func sha256Sum(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	return h.Sum(nil)
}

// PrintTransfers 打印转账信息
func PrintTransfers(transfers []TRC20Transfer) {
	if len(transfers) == 0 {
		fmt.Println("没有找到任何转账交易")
		return
	}

	// 按类型统计
	trc20Count := 0
	trxCount := 0
	trc10Count := 0

	for _, transfer := range transfers {
		switch transfer.ContractAddress {
		case "TRX":
			trxCount++
		case "TRC10":
			trc10Count++
		default:
			trc20Count++
		}
	}

	fmt.Printf("=== 转账统计 ===\n")
	fmt.Printf("TRC20代币转账: %d 笔\n", trc20Count)
	fmt.Printf("TRX原生转账: %d 笔\n", trxCount)
	fmt.Printf("TRC10代币转账: %d 笔\n", trc10Count)
	fmt.Printf("总计: %d 笔\n\n", len(transfers))

	for i, transfer := range transfers {
		fmt.Printf("[%d] TxID: %s\n", i+1, transfer.TxID)
		fmt.Printf("     From:     %s\n", transfer.FromAddress)
		fmt.Printf("     To:       %s\n", transfer.ToAddress)
		fmt.Printf("     Type:     %s\n", getTransferType(transfer.ContractAddress))
		fmt.Printf("     Contract: %s\n", transfer.ContractAddress)
		fmt.Printf("     Amount:   %s\n", transfer.Amount.String())
		fmt.Printf("     Block:    %d\n", transfer.BlockNumber)
		fmt.Printf("     Time:     %s\n", transfer.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("     Status:   %s\n", transfer.Status)
		fmt.Printf("     EnergyFee:      %d (sun)\n", transfer.EnergyFee)
		fmt.Printf("     EnergyUsageTotal:      %d\n", transfer.EnergyUsageTotal)
		fmt.Printf("     NetUsage:      %d \n", transfer.NetUsage)
		fmt.Printf("     NetFee:      %d (sun)\n", transfer.NetFee)
		fmt.Printf("     EnergyUsage:      %d \n", transfer.EnergyUsage)
		fmt.Printf("----------------------------------\n\n")
	}
}

// getTransferType 获取转账类型描述
func getTransferType(contractAddress string) string {
	switch contractAddress {
	case "TRX":
		return "TRX原生转账"
	case "TRC10":
		return "TRC10代币转账"
	default:
		return "TRC20代币转账"
	}
}

func SumInOut(transfers []TRC20Transfer) (int64, int64) {
	sumIn := int64(0)
	sumOut := int64(0)
	for _, transfer := range transfers {
		if transfer.FromAddress == "TZAw4M78JonPirHnA1r5dfx5L954nay3DQ" {
			sumIn += transfer.Amount.Int64()
		}
		if transfer.ToAddress == "TZAw4M78JonPirHnA1r5dfx5L954nay3DQ" {
			sumOut += transfer.Amount.Int64()
		}
	}
	return sumIn, sumOut
}
