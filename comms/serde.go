package comms

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

type CandidateBlockPayload struct {
	Profit       *big.Int `json:"profit"`
	BlockHeader  string   `json:"blockheader"`
	BlockUncles  []string `json:"blockuncles"`
	Transactions []string `json:"transactions"`
	Checksum     string   `json:"checksum"`
}

// Block represents an entire block in the Ethereum blockchain.
type Block struct {
	Header       *types.Header
	Uncles       []*types.Header
	Transactions types.Transactions

	// caches
	Hash atomic.Value
	Size atomic.Value

	// Td is used by package core to store the total difficulty
	// of the chain up to and including the block.
	Td *big.Int

	// These fields are used by package eth to track
	// inter-peer block relay.
	ReceivedAt   time.Time
	ReceivedFrom interface{}
}

func NewBlockVerbatim(header *types.Header, txs []*types.Transaction, uncles []*types.Header) *Block {
	b := &Block{Header: types.CopyHeader(header), Td: new(big.Int)}
	b.Transactions = make(types.Transactions, len(txs))
	copy(b.Transactions, txs)
	b.Uncles = make([]*types.Header, len(uncles))
	for i := range uncles {
		b.Uncles[i] = types.CopyHeader(uncles[i])
	}
	return b
}

func DecodeAsCB(cb CandidateBlockPayload) (*Block, *big.Int, error) {
	block := &Block{}
	profit := big.NewInt(0)

	// Check checksum
	descChecksum := cb.Checksum
	cb.Checksum = ""
	calcChecksum := AsSha256(cb)
	if descChecksum != calcChecksum {
		return block, profit, errors.New("mismatched checksum")
	}

	// Set profit
	profit.Set(cb.Profit)

	// Set block header
	headerBytes, err := hex.DecodeString(cb.BlockHeader)
	if err != nil {
		return block, profit, err
	}
	header := &types.Header{}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return block, profit, err
	}

	// Set Uncles
	uncles := make([]*types.Header, 0)
	for _, uncleEncoded := range cb.BlockUncles {
		uncleBytes, err := hex.DecodeString(uncleEncoded)
		if err != nil {
			return block, profit, err
		}
		uncle := &types.Header{}
		err = json.Unmarshal(uncleBytes, &uncle)
		if err != nil {
			return block, profit, err
		}
		uncles = append(uncles, uncle)
	}

	// Set Transactions
	transactions := make([]*types.Transaction, 0)
	for _, txEncoded := range cb.Transactions {
		uncleBytes, err := hex.DecodeString(txEncoded)
		if err != nil {
			return block, profit, err
		}
		tx := &types.Transaction{}
		err = json.Unmarshal(uncleBytes, &tx)
		if err != nil {
			return block, profit, err
		}
		transactions = append(transactions, tx)
	}

	candidateBlock := NewBlockVerbatim(header, transactions, uncles)

	return candidateBlock, profit, nil
}

func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
