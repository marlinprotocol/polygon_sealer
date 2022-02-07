package sealer

import (
	"time"

	"github.com/marlinprotocol/polygon_sealer/comms"
	log "github.com/sirupsen/logrus"
)

type Sealer struct {
	keystore string
	password string
	cbChan   chan *comms.CandidateBlock
	sbChan   chan *comms.Block
}

func InitSealer(keystoreLoc string, passwordLoc string,
	candidateChan chan *comms.CandidateBlock, sealedChan chan *comms.Block) {
	log.Info("Init event driven candidate block sealing routine")
	time.Sleep(time.Second * 100)
}
