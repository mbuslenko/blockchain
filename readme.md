
# Crypto Blockchain on Go

Cryptocurrency blockchain using the Proof of Work concept, with the ability to mine, send & receive transactions etc.




## TODO

- Finilize API to separate modules

- Create Proof of Stake version

- Non-crypto version



## Run Locally

Clone the project

```bash
  git clone https://github.com/mbuslenko/blockchain.git
```

Go to the project directory

```bash
  cd blockchain
```

Install dependencies

```bash
  go get
```

Start the blockchain server

```bash
  cd server
  go run main.go server.go
```

Start the wallet server
```bash
  cd wallet_server
  go run main.go server.go
```


## Related

Other implementation on Node

[node-blockchain](https://github.com/mbuslenko/node-blockchain)

