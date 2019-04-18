package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

func IntToHex(num int64) []byte {//转化成16进制
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func ReverseBytes(data []byte) {
	for i, j:=0, len(data)-1; i < j; i,j=i+1,j-1{
		data[i], data[j] = data[j], data[i]
	}
}