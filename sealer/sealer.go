package sealer

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/marlinprotocol/polygon_sealer/analytics"
	"github.com/marlinprotocol/polygon_sealer/comms"
	log "github.com/sirupsen/logrus"
)

type Sealer struct {
	keystore  string
	password  string
	cbChan    chan *comms.CandidateBlock
	sbChan    chan *comms.Block
	height    *big.Int
	maxProfit *big.Int
	stopCh    chan struct{}
	workCh    chan *comms.Block
}

func (s *Sealer) start() {
	go s.sealerRoutine()
	for {
		cb := <-s.cbChan

		height := cb.Block.Header.Number
		profit := cb.Profit

		analytics.AnalyticsChan <- &analytics.Analytics{Subject: "CB." + height.String(), Count: 1}

		cmp := height.Cmp(s.height)
		if cmp == -1 {
			// CB for lower blockheight
			continue
		} else if cmp == 0 {
			if profit.Cmp(s.maxProfit) != 1 {
				// CB with same height, though less or equal profit to
				// already seen CB
				continue
			}
			// CB with same height, better profit, process
			s.maxProfit.Set(profit)
			s.process(cb.Block)
		} else {
			// CB with higher block height
			s.height.Set(height)
			s.maxProfit.Set(profit)
			s.process(cb.Block)
		}
	}
}

func (s *Sealer) sealerRoutine() {
	log.Info("Started sealer routine")
	for {
		block := <-s.workCh
		header := block.Header

		// Extradata seal
		sighash, err := s.signFn(accounts.Account{Address: s.signer},
			"application/x-bor-header",
			BorRLP(header, c.config))
		copy(header.Extra[len(header.Extra)-extraSeal:], sighash)

		delay := time.Unix(int64(block.Header.Time), 0).Sub(time.Now())
		// Reduce wait delay by 60ms
		// To account for delay in sending sealed block
		delay -= 60 * time.Millisecond

		select {
		case <-s.stopCh:
			// Stop current wait to disburse sealed block
			// New, better or higher height sealing job underway?
			continue
		case <-time.After(delay):
			// Waited for turn, send across the
			s.sendSealedBlock(block, header)
		}
	}
}

func (s *Sealer) process(block *comms.Block) {
	s.stopCh <- struct{}{}
	s.workCh <- block
}

func InitSealer(keystoreLoc string, passwordLoc string,
	candidateChan chan *comms.CandidateBlock, sealedChan chan *comms.Block) {
	sealer := &Sealer{
		keystore:  keystoreLoc,
		password:  passwordLoc,
		cbChan:    candidateChan,
		sbChan:    sealedChan,
		height:    big.NewInt(0),
		maxProfit: big.NewInt(0),
		stopCh:    make(chan struct{}),
		workCh:    make(chan *comms.Block),
	}

	sealer.start()
}
