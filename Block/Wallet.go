package Block

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto"
)

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
	RIPEMD160Hasher := crypto.RIPEMD160.New()
	RIPEMD160Hasher.Write(publicSHA256[:])
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

func checksum(payload []byte) []byte{
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}