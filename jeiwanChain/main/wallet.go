package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"golang.org/x/crypto/ripemd160"
)
const version = byte(0x00)
const addressChecksumLen = 4
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}



func NewWallet() *Wallet {//*wallet 才可以修改吧
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {//生成公钥、私钥对
	curve := elliptic.P256()
	private, err:=ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}

func HashPubKey(pubKey []byte) []byte { // 先SHA256后RIPEMD160 对公钥hash
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := crypto.RIPEMD160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(payload []byte) []byte {//哈希，计算校验和。校验和是结果哈希的前四个字节
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}



func (w Wallet) GetAddress() []byte { // 公钥转base58地址
	pubKeyHash := HashPubKey(w.PublicKey)//使用 RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次

	versionedPayload := append([]byte{version}, pubKeyHash...)//给哈希加上地址生成算法版本的前缀 。
	checksum := checksum(versionedPayload)//使用 SHA256(SHA256(payload)) 再哈希，计算校验和。校验和是结果哈希的前四个字节

	fullPayload := append(versionedPayload, checksum...)//将校验和附加到 version+PubKeyHash 的组合中
	address := Base58Encode(fullPayload)//使用 Base58 对 version+PubKeyHash+checksum 组合进行编码


	return address
}
// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
