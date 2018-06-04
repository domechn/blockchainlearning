package Block

import (
	"math/big"
	"fmt"
	"log"
	"encoding/hex"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"os"
)

const subsidy = 10

//交易
type Transaction struct {
	ID   []byte  //交易的唯一标识
	Vin  []TXInput	//交易所包含的输入 可以多条
	Vout []TXOutput  //交易所包含的输出 可以多条
}
//输入  输入及之前的交易的输出
type TXInput struct {
	Txid      []byte	//之前的ID
	Vout      int	//引用的输出在之前交易中的索引
	Signature []byte //
	PubKey    []byte  //
}
//输出
type TXOutput struct {
	Value        int //金额
	PubKeyHash []byte  //验证
}

type TXOutputs struct {
	Outputs []TXOutput
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
	txo := &TXOutput{value,nil}
	txo.Lock([]byte(address))
	return txo
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

func NewUTXOTransaction(wallet *Wallet, to string, amount int, utxoSet *UTXOSet) *Transaction {
	pubKeyHash := HashPubKey(wallet.PublickKey)

	toHash := Base58Decode([]byte(to))
	toHash = toHash[1:len(toHash)-4]
	
	if bytes.Compare(pubKeyHash,toHash) == 0 {
		fmt.Println("Cannot transacte to yourself")
		os.Exit(1)
	}
	var inputs []TXInput
	var outputs []TXOutput
	acc, validOutputs := utxoSet.FindSpendableOutputs(pubKeyHash, amount)
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

	from := fmt.Sprintf("%s", wallet.GetAddress())
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))

	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	utxoSet.Blockchain.SignTransaction(&tx,wallet.PrivateKey)
	return &tx
}


func (tx *Transaction) Sign(privKey ecdsa.PrivateKey,prevTXs map[string]Transaction){
	if tx.IsCoinbase() {
		return
	}
	
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
  
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}

}

func (tx *Transaction) TrimmedCopy() Transaction{
	var inputs []TXInput
	var outputs []TXOutput

	for _ , in := range tx.Vin{
		inputs = append(inputs,TXInput{in.Txid,in.Vout,nil,nil})
	}

	for _,out := range tx.Vout{
		outputs = append(outputs,TXOutput{out.Value,out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID,inputs,outputs}
	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool{
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	
	for inID,vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:sigLen/2])
		s.SetBytes(vin.Signature[sigLen/2:])
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:keyLen/2])
		y.SetBytes(vin.PubKey[keyLen/2:])
		dataToVerify := fmt.Sprintf("%x\n",txCopy)
		rawPubKey := ecdsa.PublicKey{Curve:curve,X:&x,Y:&y}
		if ecdsa.Verify(&rawPubKey,[]byte(dataToVerify),&r,&s) == false{
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}
	return true
}

//序列化所有输出
func (outs TXOutputs) Serialize() []byte{
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

//反序列化所有输出
func DeserializeOutputs(data []byte) TXOutputs{
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}


func (tx *Transaction) Serialize() []byte{
	var encoded bytes.Buffer

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction{
	var transaction Transaction
	decoded := gob.NewDecoder(bytes.NewReader(data))
	err := decoded.Decode(&transaction)
	if err != nil{
		log.Panic(err)
	}
	return transaction
}