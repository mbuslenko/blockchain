package block

import (
	"crypto-blockchain/utils"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Transaction struct {
	senderAddress    string
	recipientAddress string
	value            float32
}

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
}

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
}

type TransactionRequest struct {
	SenderAddress    *string  `json:"senderAddress"`
	RecipientAddress *string  `json:"recipientAddress"`
	SenderPublicKey  *string  `json:"senderPublicKey"`
	Value            *float32 `json:"value"`
	Signature        *string  `json:"signature"`
}

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

const (
	MINING_DIFFICULTY = 3
	MINING_SENDER     = "BLOCKCHAIN"
	MINING_REWARD     = 1.0
	MINING_TIMER_SEC  = 20
)

// NewTransaction generates and returns new Transaction
func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{
		sender,
		recipient,
		value,
	}
}

// NewBlock generates and returns new Block
func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	return &Block{
		timestamp:    time.Now().UnixNano(),
		nonce:        nonce,
		previousHash: previousHash,
		transactions: transactions,
	}
}

// NewBlockChain starts new blockchain with init block
func NewBlockChain(blockchainAddress string, port uint16) *Blockchain {
	initBlock := &Block{}
	blockchain := new(Blockchain)
	blockchain.blockchainAddress = blockchainAddress
	blockchain.CreateBlock(0, initBlock.Hash())
	blockchain.port = port

	return blockchain
}

func (transaction *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderAddress    string  `json:"senderAddress"`
		RecipientAddress string  `json:"recipientAddress"`
		Value            float32 `json:"value"`
	}{
		SenderAddress:    transaction.senderAddress,
		RecipientAddress: transaction.recipientAddress,
		Value:            transaction.value,
	})
}

func (block *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previousHash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    block.timestamp,
		Nonce:        block.nonce,
		PreviousHash: fmt.Sprintf("%x", block.previousHash),
		Transactions: block.transactions,
	})
}

func (blockchain *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	}{
		Blocks: blockchain.chain,
	})
}

func (amountResponse *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: amountResponse.Amount,
	})
}

// Hash calculates the hash for the Block
func (block *Block) Hash() [32]byte {
	marshal, _ := json.Marshal(block)

	return sha256.Sum256([]byte(marshal))
}

// CreateBlock appends new Block to Blockchain
func (blockchain *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	block := NewBlock(nonce, previousHash, blockchain.transactionPool)
	blockchain.chain = append(blockchain.chain, block)
	blockchain.transactionPool = []*Transaction{}

	return block
}

// LastBlock returns the last Block from the Blockchain
func (blockchain *Blockchain) LastBlock() *Block {
	return blockchain.chain[len(blockchain.chain)-1]
}

func (blockchain *Blockchain) CreateTransaction(
	sender string,
	recipient string,
	value float32,
	senderPublicKey *ecdsa.PublicKey,
	signature *utils.Signature,
) bool {
	isTransacted := blockchain.AddTransaction(sender, recipient, value, senderPublicKey, signature)

	// TODO: Add sync

	return isTransacted
}

// AddTransaction appends new Transaction to transaction pool
func (blockchain *Blockchain) AddTransaction(
	sender string,
	recipient string,
	value float32,
	senderPublicKey *ecdsa.PublicKey,
	signature *utils.Signature,
) bool {
	transaction := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		blockchain.transactionPool = append(blockchain.transactionPool, transaction)
		return true
	}

	if blockchain.VerifyTransactionSignature(senderPublicKey, signature, transaction) {
		blockchain.transactionPool = append(blockchain.transactionPool, transaction)
		return true
	}

	log.Println("ERROR Validating transaction")
	return false
}

// VerifyTransactionSignature verifies the signature in r, s
// of hash using the senderPublicKey. Its return true
// whether the signature is valid
func (blockchain *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey,
	signature *utils.Signature,
	transaction *Transaction,
) bool {
	marshal, _ := json.Marshal(transaction)
	hash := sha256.Sum256([]byte(marshal))

	return ecdsa.Verify(senderPublicKey, hash[:], signature.R, signature.S)
}

// CopyTransactionPool returns current transaction pool
func (blockchain *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, transaction := range blockchain.transactionPool {
		transactions = append(transactions, NewTransaction(
			transaction.senderAddress,
			transaction.recipientAddress,
			transaction.value,
		))
	}

	return transactions
}

// ValidProof checks whether first three numbers in nonce are zeros
func (blockchain *Blockchain) ValidProof(
	nonce int,
	previousHash [32]byte,
	transactions []*Transaction,
	difficulty int,
) bool {
	zeros := strings.Repeat("0", difficulty)

	guessBlock := Block{nonce, previousHash, 0, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())

	return guessHashStr[:difficulty] == zeros
}

// ProofOfWork iterates through the nonce set until the nonce is correct,
// and returns it. Correct nonce = nonce which starts from three zeros
func (blockchain *Blockchain) ProofOfWork() int {
	transactions := blockchain.CopyTransactionPool()
	previousHash := blockchain.LastBlock().Hash()

	// While nonce is not correct, we will continue
	nonce := 0
	for !blockchain.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}

	return nonce
}

// Mining creates new block in the Blockchain
func (blockchain *Blockchain) Mining() bool {
	blockchain.mux.Lock()
	defer blockchain.mux.Unlock()

	if len(blockchain.transactionPool) == 0 {
		return false
	}

	blockchain.AddTransaction(MINING_SENDER, blockchain.blockchainAddress, MINING_REWARD, nil, nil)

	nonce := blockchain.ProofOfWork()
	previousHash := blockchain.LastBlock().Hash()

	blockchain.CreateBlock(nonce, previousHash)
	return true
}

func (blockchain *Blockchain) StartMining() {
	blockchain.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, blockchain.StartMining)
}

// CalculateTotalAmount iterates through all transactions in the Blockchain
// and returns total amount of user's coins
func (blockchain *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, block := range blockchain.chain {
		for _, transaction := range block.transactions {
			value := transaction.value

			if blockchainAddress == transaction.recipientAddress {
				totalAmount += value
			}

			if blockchainAddress == transaction.senderAddress {
				totalAmount -= value
			}
		}
	}

	return totalAmount
}

// Validate checks that all fields are not nil
func (transactionRequest *TransactionRequest) Validate() bool {
	if transactionRequest.SenderPublicKey == nil ||
		transactionRequest.SenderAddress == nil ||
		transactionRequest.RecipientAddress == nil ||
		transactionRequest.Value == nil ||
		transactionRequest.Signature == nil {
		return false
	}

	return true
}

// TransactionPool returns blockchain transaction pool
func (blockchain *Blockchain) TransactionPool() []*Transaction {
	return blockchain.transactionPool
}

// Print is built-in function to print the Transaction
func (transaction *Transaction) Print() {
	separator := strings.Repeat("-", 50)

	log.Printf("     sender_address    %s\n", transaction.senderAddress)
	log.Printf("     recipient_address %s\n", transaction.recipientAddress)
	log.Printf("     value             %.1f\n", transaction.value)
	fmt.Printf("%s\n", separator)
}

// Print is built-in function to print the Block
func (block *Block) Print() {
	separator := strings.Repeat("-", 20)

	log.Printf(" timestamp         %d\n", block.timestamp)
	log.Printf(" nonce             %d\n", block.nonce)
	log.Printf(" previousHash      %x\n", block.previousHash)

	if len(block.transactions) > 0 {
		fmt.Printf("%s Transactions %s\n", separator, separator)
	}

	for _, transaction := range block.transactions {
		transaction.Print()
	}
}

// Print is built-in function to print all blocks in the Blockchain
func (blockchain *Blockchain) Print() {
	separator := strings.Repeat("=", 25)
	for i, block := range blockchain.chain {
		fmt.Printf("%s Chain %d %s\n", separator, i, separator)
		block.Print()
	}
}
