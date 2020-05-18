package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"os"
	"sync"
)

type AccountsContainer struct {
	mu       *sync.Mutex
	accounts map[string]string
}

var Accounts map[string]string

func InitAccount() (*AccountsContainer, error) {
	file, err := os.Open("accounts.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Accounts)
	if err != nil {
		return nil, err
	}

	return &AccountsContainer{
			mu:       &sync.Mutex{},
			accounts: Accounts,
		},
		nil
}

func (c *AccountsContainer) Address(login string) (string, error) {
	privKey, err := crypto.HexToECDSA(c.accounts[login])
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(privKey.PublicKey).String(), nil
}

func (c *AccountsContainer) Add(login string) (string, error) {

	if _, ok := c.accounts[login]; ok {
		priv, err := crypto.HexToECDSA(c.accounts[login])
		if err != nil {
			return "", err
		}

		return crypto.PubkeyToAddress(priv.PublicKey).String(), nil
	}

	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.accounts[login] = hex.EncodeToString(crypto.FromECDSA(privKey))

	bt, err := json.Marshal(c.accounts)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile("accounts.json", bt, 0644)
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(privKey.PublicKey).String(), nil
}
