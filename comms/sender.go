package comms

import (
	log "github.com/sirupsen/logrus"
)

func InitHttpSender(sendAddr string) chan *Block {
	ch := make(chan *Block, 100)
	log.Info("TODO: Init HTTP sealed block sending routine")
	return ch
}
