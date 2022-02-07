package analytics

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Analytics struct {
	Subject string
	Count   int
}

var AnalyticsChan chan *Analytics

func ShowAnalytics(secs int) {
	log.Info("Started show analytics")
	for {
		time.Sleep(time.Duration(secs) * time.Second)
		aMap := map[string]int{"recieved candidate blocks": 0, "sealed blocks": 0, "sent sealed blocks": 0}
		for len(AnalyticsChan) > 0 {
			a := <-AnalyticsChan
			if val, ok := aMap[a.Subject]; ok {
				aMap[a.Subject] = val + a.Count
			} else {
				aMap[a.Subject] = a.Count
			}
		}
		analyticsString := ""
		for k, v := range aMap {
			analyticsString += "[" + k + ": " + strconv.Itoa(v) + "] "
		}
		log.Info("Analytics (", secs, "secs): ", analyticsString)
	}
}
