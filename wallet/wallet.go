package wallet

import (
	"crypto-blockchain/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    string
}

type Transaction struct {
	senderPublicKey  *ecdsa.PublicKey
	senderPrivateKey *ecdsa.PrivateKey
	senderAddress    string
	recipientAddress string
	value            float32
}

type TransactionRequest struct {
	SenderPublicKey  *string `json:"senderPublicKey"`
	SenderPrivateKey *string `json:"senderPrivateKey"`
	SenderAddress    *string `json:"senderAddress"`
	RecipientAddress *string `json:"recipientAddress"`
	Value            *string `json:"value"`
}

func NewWallet() *Wallet {
	wallet := new(Wallet)

	// Creating ECDSA private and public keys
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	wallet.privateKey = privateKey
	wallet.publicKey = &wallet.privateKey.PublicKey

	// Perform SHA-256 hashing on the public key
	sha256Hash := sha256.New()
	sha256Hash.Write(wallet.publicKey.X.Bytes())
	sha256Hash.Write(wallet.publicKey.Y.Bytes())
	sha256Digest := sha256Hash.Sum(nil)

	// Perform RIPEMD-160 hashing on the result of sha256Hash
	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(sha256Digest)
	ripemd160Digest := ripemd160Hash.Sum(nil)

	// Add version byte in front of ripemd160Hash (0x00 for Main Network)
	version := make([]byte, 21)
	version[0] = 0x00
	copy(version[1:], ripemd160Digest[:])

	// Perform SHA-256 hash on the extended result
	sha256Hash = sha256.New()
	sha256Hash.Write(version)
	sha256Digest = sha256Hash.Sum(nil)

	// One more SHA-256 hash :D
	sha256Hash = sha256.New()
	sha256Hash.Write(sha256Digest)
	sha256Digest = sha256Hash.Sum(nil)

	// Take the first 4 bytes of the summary SHA-256 hash for checksum
	checksum := sha256Digest[:4]

	// Add the 4 bytes from the checksum at the end of RIPEMD-160 hash
	result := make([]byte, 25)
	copy(result[:21], version[:])
	copy(result[21:], checksum[:])

	// Convert the result from a byte string into base58
	address := base58.Encode(result)
	wallet.address = address

	return wallet
}

func NewTransaction(
	privateKey *ecdsa.PrivateKey,
	publicKey *ecdsa.PublicKey,
	sender string,
	recipient string,
	value float32,
) *Transaction {
	return &Transaction{
		senderPrivateKey: privateKey,
		senderPublicKey:  publicKey,
		senderAddress:    sender,
		recipientAddress: recipient,
		value:            value,
	}
}

func (transaction *Transaction) GenerateSignature() *utils.Signature {
	marshal, _ := json.Marshal(transaction)
	hash := sha256.Sum256(marshal)
	r, s, _ := ecdsa.Sign(rand.Reader, transaction.senderPrivateKey, hash[:])

	return &utils.Signature{r, s}
}

// Validate checks that all fields are not nil
func (transactionRequest *TransactionRequest) Validate() bool {
	if transactionRequest.SenderPublicKey == nil ||
		transactionRequest.SenderPrivateKey == nil ||
		transactionRequest.SenderAddress == nil ||
		transactionRequest.RecipientAddress == nil ||
		transactionRequest.Value == nil {
		return false
	}

	return true
}

func (wallet *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return wallet.privateKey
}

func (wallet *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", wallet.privateKey.D.Bytes())
}

func (wallet *Wallet) PublicKey() *ecdsa.PublicKey {
	return wallet.publicKey
}

func (wallet *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", wallet.publicKey.X.Bytes(), wallet.publicKey.Y.Bytes())
}

func (wallet *Wallet) Address() string {
	return wallet.address
}

func (transaction *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"senderAddress"`
		Recipient string  `json:"recipientAddress"`
		Value     float32 `json:"value"`
	}{
		Sender:    transaction.senderAddress,
		Recipient: transaction.recipientAddress,
		Value:     transaction.value,
	})
}

func (wallet *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
		Address    string `json:"address"`
	}{
		PrivateKey: wallet.PrivateKeyStr(),
		PublicKey:  wallet.PublicKeyStr(),
		Address:    wallet.Address(),
	})
}
