package database

import (
	"encoding/json"
	"errors"
	hd "github.com/btcsuite/btcutil/hdkeychain"
	iwallet "github.com/cpacia/wallet-interface"
	"time"
)

type CoinRecord struct {
	MasterPriv         string
	EncryptedMasterKey bool
	MasterPub          string
	Coin               string `gorm:"unique;not null"`
	BestBlockHeight    uint64
	BestBlockID        string
}

func (c *CoinRecord) MasterPrivateKey() (*hd.ExtendedKey, error) {
	if c.EncryptedMasterKey {
		return nil, errors.New("master private key is encrypted")
	}
	return hd.NewKeyFromString(c.MasterPriv)
}

func (c *CoinRecord) MasterPublicKey() (*hd.ExtendedKey, error) {
	return hd.NewKeyFromString(c.MasterPub)
}

func (c *CoinRecord) BlockchainInfo() iwallet.BlockchainInfo {
	return iwallet.BlockchainInfo{
		BestBlock: iwallet.BlockID(c.BestBlockID),
		Height:    c.BestBlockHeight,
	}
}

type AddressRecord struct {
	Addr     string `gorm:"unique;not null"`
	KeyIndex int
	Change   bool
	Used     bool
	Coin     string
}

func (ar *AddressRecord) Address() iwallet.Address {
	return iwallet.NewAddress(ar.Addr, iwallet.CoinType(ar.Coin))
}

type TransactionRecord struct {
	Txid                   string `gorm:"primary_key;unique;not null"`
	SerlializedTransaction []byte
	BlockHeight            uint64
	Timestamp              time.Time
	Coin                   string
}

func NewTransactionRecord(tx iwallet.Transaction, timestamp time.Time, coinType iwallet.CoinType) (*TransactionRecord, error) {
	out, err := json.MarshalIndent(&tx, "", "    ")
	if err != nil {
		return nil, err
	}
	return &TransactionRecord{
		Txid:                   tx.ID.String(),
		SerlializedTransaction: out,
		BlockHeight:            tx.Height,
		Timestamp:              timestamp,
		Coin:                   coinType.CurrencyCode(),
	}, nil
}

func (tr *TransactionRecord) Transaction() (iwallet.Transaction, error) {
	var tx iwallet.Transaction
	if err := json.Unmarshal(tr.SerlializedTransaction, &tx); err != nil {
		return iwallet.Transaction{}, err
	}
	return tx, nil
}

func (tr *TransactionRecord) TransactionID() iwallet.TransactionID {
	return iwallet.TransactionID(tr.Txid)
}

func (tr *TransactionRecord) Height() uint64 {
	return tr.BlockHeight
}

func (tr *TransactionRecord) CoinType() iwallet.CoinType {
	return iwallet.CoinType(tr.Coin)
}

type UtxoRecord struct {
	Outpoint  string `gorm:"primary_key;unique;not null"`
	Height    uint64
	Timestamp time.Time
	Coin      string
}