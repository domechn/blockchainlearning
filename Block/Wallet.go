package Block

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"encoding/gob"
	"bytes"
	"fmt"
	"os"
	"log"
)

const walletFile = "wallet_%s.dat"
const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublickKey []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet{
	private ,public := newKeyPair()
	wallet := Wallet{private,public}
	return &wallet
}


func newKeyPair() (ecdsa.PrivateKey,[]byte){
	curve := elliptic.P256()
	private ,_ := ecdsa.GenerateKey(curve,rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)
	return *private,pubKey
}

func (w Wallet) getAddress()[]byte{
	pubKeyHash := HashPubKey(w.PublickKey)
	versionedPayload := append([]byte{version},pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload,checksum...)
	address := Base58Encode(fullPayload)
	return []byte(address)
}

func HashPubKey(pubKey []byte)[]byte  {
	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_ ,err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

func checksum(payload []byte) []byte{
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func NewWallets(nodeID string) (*Wallets,error){
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets,err
}

func (ws Wallets) LoadFromFile(nodeID string) error{
	walletFile := fmt.Sprint(walletFile,nodeID)
	if _,err := os.Stat(walletFile);os.IsNotExist(err){
		return err
	}
	fileContent ,err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256)
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets = wallets.Wallets
	return nil
}

func (wallet Wallet) GetAddress() []byte{
	pubHashKey := HashPubKey(wallet.PublickKey)
	versionedPayload := append([]byte{version},pubHashKey...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)
	return address
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string 
	for address := range ws.Wallets{
		addresses = append(addresses,address)
	}
	return addresses
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

func (ws *Wallets) SaveToFile(nodeID string){
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile,nodeID)
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletFile,content.Bytes(),0644)
	if err != nil {
		log.Panic(err)
	}
}