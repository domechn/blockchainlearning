package Block

import (
	"fmt"
	"log"
	"encoding/hex"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"os"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

type TXOutput struct {
	Value        int
	PubKeyHash []byte
}

func (in *TXInput) UsesKey(pubHashKey []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash,pubHashKey) == 0
}

func (out *TXOutput) IsLockedWithKey(pubHashKey []byte) bool{
	return bytes.Compare(out.PubKeyHash,pubHashKey) == 0
}

func (out *TXOutput) Lock(address []byte)  {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
	
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewTXOutput(value int ,address string) *TXOutput  {
	txo := TXOutput{value,nil}
	txo.Lock([]byte(address))
	return &txo
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1,nil,[]byte(data)}
	txout := NewTXOutput(subsidy,to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.SetID()
	return &tx
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(wallet *Wallet,from, to string, amount int, bc *BlockChain) *Transaction {
	if from == to {
		fmt.Println("Cannot transacte to yourself")
		os.Exit(1)
	}
	var inputs []TXInput
	var outputs []TXOutput
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil,wallet.PublickKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))

	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
