package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Signature consist of two numbers (integers): r and s .
//Ethereum also uses an additional v (recovery identifier) variable.
type Signature struct {
	R *big.Int
	S *big.Int
}

func (signature *Signature) String() string {
	return fmt.Sprintf("%064x%064x", signature.R, signature.S)
}

func StringToBigIntTuple(str string) (big.Int, big.Int) {
	bytesX, _ := hex.DecodeString(str[:64])
	bytesY, _ := hex.DecodeString(str[64:])

	var bigX big.Int
	var bigY big.Int

	_ = bigX.SetBytes(bytesX)
	_ = bigY.SetBytes(bytesY)

	return bigX, bigY
}

func PublicKeyFromString(str string) *ecdsa.PublicKey {
	x, y := StringToBigIntTuple(str)

	return &ecdsa.PublicKey{elliptic.P256(), &x, &y}
}

func PrivateKeyFromString(str string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {
	bytes, _ := hex.DecodeString(str[:64])

	var bi big.Int
	_ = bi.SetBytes(bytes)

	return &ecdsa.PrivateKey{*publicKey, &bi}
}

func SignatureFromString(str string) *Signature {
	r, s := StringToBigIntTuple(str)
	return &Signature{&r, &s}
}
