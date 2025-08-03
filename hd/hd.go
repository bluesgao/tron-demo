package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
	"log"
)

func generateMnemonic() string {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		log.Fatal(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		log.Fatal(err)
	}
	return mnemonic
}

// tron地址转换
func EthAddressToTron(evmAddr string) string {
	addr := crypto.Keccak256([]byte(evmAddr)) // dummy, not needed here
	_ = addr
	return "" // 非真实转换，只是提示错误逻辑
}

func TronAddressFromPrivateKey(pk *ecdsa.PrivateKey) string {
	pubKey := pk.PublicKey
	pubBytes := crypto.FromECDSAPub(&pubKey)[1:] // 去掉 0x04 前缀

	// 1. TRON 使用 keccak256
	hash := crypto.Keccak256(pubBytes)
	// 2. 取最后20字节，加 0x41 前缀
	addr := append([]byte{0x41}, hash[12:]...)
	// 3. Base58Check 编码
	tronAddr := base58.CheckEncode(addr, 0x00)
	return tronAddr
}

func main() {
	// 示例助记词
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"

	// 创建钱包
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// ETH 地址
	ethPath := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	ethAcc, _ := wallet.Derive(ethPath, false)
	fmt.Println("ETH地址:", ethAcc.Address.Hex())

	// TRON 派生
	tronPath := hdwallet.MustParseDerivationPath("m/44'/195'/0'/0/0")
	tronAcc, _ := wallet.Derive(tronPath, false)

	// 获取 TRON 私钥
	priv, _ := wallet.PrivateKey(tronAcc)
	tronAddr := TronAddressFromPrivateKey(priv)
	fmt.Println("TRON地址:", tronAddr)
}
