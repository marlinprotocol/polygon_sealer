package comms

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marlinprotocol/polygon_sealer/analytics"
	log "github.com/sirupsen/logrus"
)

func InitHttpListener(listenAddr string) chan *CandidateBlock {
	ch := make(chan *CandidateBlock, 100)
	go serveHttp(listenAddr, ch)
	return ch
}

func serveHttp(listenAddr string, ch chan *CandidateBlock) {
	analyticsBody := analytics.Analytics{
		Subject: "recieved candidate blocks",
		Count:   1,
	}
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		var cbp CandidateBlockPayload
		err := json.NewDecoder(r.Body).Decode(&cbp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			log.Error("Bad CB recv", "error", err.Error())
			return
		}
		cb, profit, err := DecodeAsCB(cbp)
		if err != nil {
			log.Error("Bad CB decode", "error", err)
		}

		ch <- &CandidateBlock{
			Block:  cb,
			Profit: profit,
		}

		analytics.AnalyticsChan <- &analyticsBody
	})
	go http.ListenAndServe(listenAddr, nil)
	log.Info("Listening for candidates on [", listenAddr, "]")
}

func showAnalytics(analytics chan bool) {
	for {
		time.Sleep(10 * time.Second)
		count := 0
		for len(analytics) > 0 {
			count++
			<-analytics
		}
		log.Info("Listener: Received ", count, " candidate blocks in last 10s")
	}
}
