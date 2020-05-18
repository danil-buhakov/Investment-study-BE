package main

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"os"
	"strings"
)

type (
	Contract struct {
		Abi     ethAbi.ABI
		Client  *ethclient.Client
		Account *AccountsContainer
	}

	SupportedToken struct {
		Name    string
		Address string
	}
)

var SupportedTokens []SupportedToken

func InitContact(acc *AccountsContainer) (*Contract, error) {
	file, err := os.Open("tokens.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&SupportedTokens)
	if err != nil {
		return nil, err
	}

	abi, err := ethAbi.JSON(strings.NewReader(ABI))
	if err != nil {
		return nil, err
	}

	conn, err := rpc.Dial("https://ropsten-rpc.linkpool.io/")
	if err != nil {
		return nil, err
	}

	ethClient := ethclient.NewClient(conn)

	return &Contract{
		Abi:     abi,
		Client:  ethClient,
		Account: acc,
	}, nil
}

func (c *Contract) GetTokenBalanceForAddress(tokenAddressHex, login string) (*big.Int, error) {
	addressHex, err := c.Account.Address(login)

	address := common.HexToAddress(addressHex)
	tokenAddress := common.HexToAddress(tokenAddressHex)

	contractData, err := c.Abi.Pack("balanceOf", address)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: contractData,
	}

	response, err := c.Client.CallContract(context.TODO(), msg, nil)
	if err != nil {
		return nil, err
	}

	balance := big.NewInt(0)
	err = c.Abi.Unpack(&balance, "balanceOf", response)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (c *Contract) GetTokenSymbol(tokenAddressHex string) (symbol string, err error) {
	tokenAddress := common.HexToAddress(tokenAddressHex)

	contractData, err := c.Abi.Pack("symbol")
	if err != nil {
		return symbol, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: contractData,
	}

	response, err := c.Client.CallContract(context.TODO(), msg, nil)
	if err != nil {
		return symbol, err
	}

	err = c.Abi.Unpack(&symbol, "symbol", response)
	if err != nil {
		return symbol, err
	}

	return symbol, nil
}

func (c *Contract) GetBalanceForAddress(addressHex string) (*big.Int, error) {
	address := common.HexToAddress(addressHex)
	bal, err := c.Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}

	return bal, nil
}

func (c *Contract) InvestOnToken(login string, amount int64, tokenAddressHex string) error {
	tokenAddress := common.HexToAddress(tokenAddressHex)

	privKey, err := crypto.HexToECDSA(c.Account.accounts[login])
	if err != nil {
		return err
	}

	contractData, err := c.Abi.Pack("buy")
	if err != nil {
		return err
	}

	gasPrice, err := c.Client.SuggestGasPrice(context.TODO())
	if err != nil {
		return err
	}

	nonce, err := c.Client.PendingNonceAt(context.TODO(), crypto.PubkeyToAddress(privKey.PublicKey))
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(amount), 400000, big.NewInt(gasPrice.Int64()*10), contractData)

	signer := &types.HomesteadSigner{}
	tx, err = types.SignTx(tx, signer, privKey)
	if err != nil {
		panic(err)
	}

	err = c.Client.SendTransaction(context.TODO(), tx)
	if err != nil {
		panic(err)
	}

	log.Print(tx.Hash().String())

	return nil
}
