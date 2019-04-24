package main

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode



	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum) //最底层的交易数据节点 hash后
		nodes = append(nodes, *node)
	}

	for ; len(nodes)>1; {//直到剩一个根节点
		if len(nodes) %2 !=0 {
			nodes = append(nodes, nodes[len(nodes)-1])//如果是奇数个节点 将最后一个交易复制一份
		}
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {//两个，两个组队
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)//生成上面一层的节点
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel//跑到上面一层继续合并

	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]//取出来的应该不是hash原来的类型
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}