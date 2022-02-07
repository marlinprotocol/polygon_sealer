package comms

import "math/big"

type CandidateBlock struct {
	Block  *Block
	Profit *big.Int
}
