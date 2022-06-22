package main

import (
	"crypto-blockchain/block"
	"crypto-blockchain/utils"
	"crypto-blockchain/wallet"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type Server struct {
	port uint16
}

func NewServer(port uint16) *Server {
	return &Server{port}
}

func (server *Server) Port() uint16 {
	return server.port
}

func (server *Server) GetBlockchain() *block.Blockchain {
	blockchain, ok := cache["blockchain"]

	if !ok {
		minersWallet := wallet.NewWallet()
		blockchain = block.NewBlockChain(minersWallet.Address(), server.Port())
		cache["blockchain"] = blockchain

		log.Printf("private_key %v", minersWallet.PrivateKeyStr())
		log.Printf("public_key %v", minersWallet.PublicKeyStr())
		log.Printf("address %v", minersWallet.Address())
	}

	return blockchain
}

func (server *Server) GetChain(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		writer.Header().Add("Content-Type", "application/json")

		blockchain := server.GetBlockchain()
		marshal, _ := blockchain.MarshalJSON()

		io.WriteString(writer, string(marshal[:]))
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("405 - Method not allowed"))
	}
}

func (server *Server) Transactions(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		writer.Header().Add("Content-Type", "application/json")
		blockchain := server.GetBlockchain()
		transactions := blockchain.TransactionPool()
		marshal, _ := json.Marshal(struct {
			Transactions []*block.Transaction `json:"transactions"`
			Length       int                  `json:"length"`
		}{
			Transactions: transactions,
			Length:       len(transactions),
		})
		io.WriteString(writer, string(marshal[:]))

	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var transactionRequest block.TransactionRequest

		err := decoder.Decode(&transactionRequest)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}
		if !transactionRequest.Validate() {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("An error occurred while validating your transaction"))
		}

		publicKey := utils.PublicKeyFromString(*transactionRequest.SenderPublicKey)
		signature := utils.SignatureFromString(*transactionRequest.Signature)

		blockchain := server.GetBlockchain()
		isCreated := blockchain.CreateTransaction(*transactionRequest.SenderAddress,
			*transactionRequest.RecipientAddress,
			*transactionRequest.Value,
			publicKey,
			signature)

		writer.Header().Add("Content-Type", "application/json")

		var marshal []byte
		if !isCreated {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("An error occurred while creating your transaction"))
		} else {
			writer.WriteHeader(http.StatusCreated)
		}

		io.WriteString(writer, string(marshal))

	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *Server) Mine(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockchain := server.GetBlockchain()
		isMined := blockchain.Mining()

		var marshal []byte
		if !isMined {
			writer.WriteHeader(http.StatusBadRequest)
			marshal = []byte("Block wasn't mined")
		} else {
			marshal = []byte("Block was mined")
		}

		writer.Header().Add("Content-Type", "apllication/json")
		io.WriteString(writer, string(marshal))
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *Server) StartMine(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockchain := server.GetBlockchain()
		blockchain.StartMining()

		marshal := []byte("Mining was ended successfully")
		writer.Header().Add("Content-Type", "apllication/json")
		io.WriteString(writer, string(marshal))
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *Server) Amount(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		address := req.URL.Query().Get("address")
		if address == "" {
			writer.WriteHeader(http.StatusBadRequest)
		}

		amount := server.GetBlockchain().CalculateTotalAmount(address)
		amountResponse := &block.AmountResponse{amount}
		marshal, _ := json.Marshal(amountResponse)

		writer.Header().Add("Content-Type", "application/json")
		io.WriteString(writer, string(marshal[:]))

	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *Server) Run() {
	http.HandleFunc("/", server.GetChain)
	http.HandleFunc("/transactions", server.Transactions)
	http.HandleFunc("/mine", server.Mine)
	http.HandleFunc("/mine/start", server.StartMine)
	http.HandleFunc("/amount", server.Amount)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(server.Port())), nil))
}
