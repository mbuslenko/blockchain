package main

import (
	"bytes"
	"crypto-blockchain/block"
	"crypto-blockchain/utils"
	"crypto-blockchain/wallet"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"
)

const tmplDir = "/Users/mac/Desktop/Work/blockchain/wallet_server/templates/"

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (walletServer *WalletServer) Index(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		tmpl, _ := template.ParseFiles(path.Join(tmplDir + "index.html"))
		tmpl.Execute(writer, "")
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("405 - Method not allowed"))
	}
}

func (walletServer *WalletServer) Port() uint16 {
	return walletServer.port
}

func (walletServer *WalletServer) Gateway() string {
	return walletServer.gateway
}

func (walletServer *WalletServer) Wallet(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		writer.Header().Add("Content-Type", "application/json")
		newWallet := wallet.NewWallet()

		marshal, _ := newWallet.MarshalJSON()
		io.WriteString(writer, string(marshal[:]))
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (walletServer *WalletServer) CreateTransaction(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		writer.Header().Add("Content-Type", "application/json")

		decoder := json.NewDecoder(req.Body)
		var transactionRequest wallet.TransactionRequest
		err := decoder.Decode(&transactionRequest)

		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if !transactionRequest.Validate() {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		publicKey := utils.PublicKeyFromString(*transactionRequest.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*transactionRequest.SenderPrivateKey, publicKey)

		value, err := strconv.ParseFloat(*transactionRequest.Value, 32)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		value32 := float32(value)

		transaction := wallet.NewTransaction(privateKey,
			publicKey,
			*transactionRequest.SenderAddress,
			*transactionRequest.RecipientAddress,
			value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bcTransactionRequest := &block.TransactionRequest{
			transactionRequest.SenderAddress,
			transactionRequest.RecipientAddress,
			transactionRequest.SenderPublicKey,
			&value32,
			&signatureStr,
		}
		marshal, _ := json.Marshal(bcTransactionRequest)
		buff := bytes.NewBuffer(marshal)

		resp, _ := http.Post(walletServer.Gateway()+"/transactions", "application/json", buff)
		writer.WriteHeader(resp.StatusCode)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (walletServer *WalletServer) Run() {
	http.HandleFunc("/", walletServer.Index)
	http.HandleFunc("/wallet", walletServer.Wallet)
	http.HandleFunc("/transaction", walletServer.CreateTransaction)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(int(walletServer.Port())), nil))
}
