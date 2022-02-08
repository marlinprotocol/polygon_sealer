package analytics

import (
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Analytics struct {
	Subject string
	Count   int
}

var AnalyticsChan chan *Analytics
var RecvCB = Analytics{Subject: "CB..Recv", Count: 1}

func ShowAnalytics(secs int) {
	log.Info("Started show analytics")
	for {
		time.Sleep(time.Duration(secs) * time.Second)
		aMap := map[string]int{"CB..Recv": 0}
		for len(AnalyticsChan) > 0 {
			a := <-AnalyticsChan
			if val, ok := aMap[a.Subject]; ok {
				aMap[a.Subject] = val + a.Count
			} else {
				aMap[a.Subject] = a.Count
			}
		}
		keys := make([]string, len(aMap))
		i := 0
		for k := range aMap {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		analyticsString := ""
		for _, k := range keys {
			analyticsString += "[" + k + ": " + strconv.Itoa(aMap[k]) + "] "
		}
		log.Info("Analytics (", secs, "secs): ", analyticsString)
	}
}
