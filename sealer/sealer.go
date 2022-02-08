package sealer

import (
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/marlinprotocol/polygon_sealer/analytics"
	"github.com/marlinprotocol/polygon_sealer/comms"
	log "github.com/sirupsen/logrus"
)

// Bor protocol constants.
const (
	extraVanity = 32 // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = 65 // Fixed number of extra-data suffix bytes reserved for signer seal
)

type Sealer struct {
	keysdir   string
	keystore  *keystore.KeyStore
	passfile  string
	password  string
	cbChan    chan *comms.CandidateBlock
	sbChan    chan *comms.Block
	height    *big.Int
	maxProfit *big.Int
	stopCh    chan struct{}
	workCh    chan *comms.Block
}

func (s *Sealer) start() {
	// Setup keystore
	s.keystore = keystore.NewKeyStore(s.keysdir, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(s.keystore.Accounts()) != 1 {
		log.Error("Keystore either not found, or found multiple accounts. Count: ", len(s.keystore.Accounts()))
		os.Exit(1)
	}

	// Extract wallets
	wallet := s.keystore.Wallets()[0]

	passwordBytes, err := ioutil.ReadFile(s.passfile) // just pass the file name
	if err != nil {
		log.Error("Error reading password file: ", err)
		os.Exit(1)
	}

	// Sanity checks
	s.password = string(passwordBytes)
	err = wallet.Open(s.password)
	if err != nil {
		log.Error("Error while opening wallet: ", err)
		os.Exit(1)
	}

	log.Info("Successfully loaded wallet with sealer address: ", s.keystore.Accounts()[0].Address)

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
			// s.process(cb.Block)
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
	anaItp := &analytics.Analytics{Subject: "Seal.Interrupt", Count: 1}
	anaDone := &analytics.Analytics{Subject: "Seal.WaitSuccess", Count: 1}
	for {
		block := <-s.workCh
		header := block.Header
		wallet := s.keystore.Wallets()[0]
		// _ = header
		// Extradata seal
		borrlp := BorRLP(header, &BorMainnetParams)
		sighash, err := wallet.SignDataWithPassphrase(accounts.Account{Address: wallet.Accounts()[0].Address},
			s.password,
			"application/x-bor-header",
			borrlp)
		if err != nil {
			log.Error("Error while sealing block: ", err)
			continue
		}
		copy(header.Extra[len(header.Extra)-extraSeal:], sighash)

		delay := time.Unix(int64(block.Header.Time), 0).Sub(time.Now())
		// Reduce wait delay by 60ms
		// To account for delay in sending sealed block
		delay -= 60 * time.Millisecond

		select {
		case <-s.stopCh:
			// Stop current wait to disburse sealed block
			// New, better or higher height sealing job underway?
			analytics.AnalyticsChan <- anaItp
			continue
		case <-time.After(delay):
			// Waited for turn, send across the
			// s.sendSealedBlock(block, header)
			analytics.AnalyticsChan <- anaDone
			s.sbChan <- block
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
		keysdir:   keystoreLoc,
		keystore:  nil,
		passfile:  passwordLoc,
		cbChan:    candidateChan,
		sbChan:    sealedChan,
		height:    big.NewInt(0),
		maxProfit: big.NewInt(0),
		stopCh:    make(chan struct{}, 1),
		workCh:    make(chan *comms.Block, 1),
	}

	sealer.start()
}
