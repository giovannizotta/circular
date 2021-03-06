package node

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
)

type PreimageHashPair struct {
	Preimage string `json:"preimage"`
	Hash     string `json:"hash"`
}

func NewPreimageHashPair() PreimageHashPair {
	preimage := make([]byte, 32)
	//fill the slice with random bytes
	rand.Read(preimage)
	hash := sha256.Sum256(preimage)

	return PreimageHashPair{
		Preimage: hex.EncodeToString(preimage),
		Hash:     hex.EncodeToString(hash[:]),
	}
}

func (n *Node) GeneratePreimageHashPair() (string, error) {
	pair := NewPreimageHashPair()
	err := n.DB.Set(pair.Hash, []byte(pair.Preimage))
	if err != nil {
		return "", err
	}
	return pair.Hash, nil
}
