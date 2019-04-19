package main

import "bytes"

type TXInput struct {
	Txid []byte
	Vout int
	Signature []byte
	PubKey []byte
}
