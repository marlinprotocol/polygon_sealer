/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/marlinprotocol/polygon_sealer/analytics"
	"github.com/marlinprotocol/polygon_sealer/comms"
	"github.com/marlinprotocol/polygon_sealer/sealer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listenAddr  string
	sendAddr    string
	keystoreLoc string
	passwordLoc string
)

// startCmd represents the generate command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start signer",
	Long:  `Start signing operations`,
	Run: func(cmd *cobra.Command, args []string) {
		// Basic sanity
		if listenAddr == "" || sendAddr == "" || keystoreLoc == "" || passwordLoc == "" {
			log.Error("Need further parameters to start the signer. Consult {polygon_sealer start -h}")
		}

		// Initialise analytics
		analytics.AnalyticsChan = make(chan *analytics.Analytics, 10000)
		go analytics.ShowAnalytics(10)

		// Initialise communication routines
		cbChan := comms.InitHttpListener(listenAddr)
		sbChan := comms.InitHttpSender(sendAddr) // Maybe we want to send to multiple guys?

		// Initialise sealing routine
		sealer.InitSealer(keystoreLoc, passwordLoc, cbChan, sbChan)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&listenAddr, "listenaddr", "l", "0.0.0.0:6553", "Exposed endpoint to recieve candidate blocks")
	startCmd.Flags().StringVarP(&sendAddr, "sendaddr", "s", "127.0.0.1:1320", "Node endpoint to send sealed blocks")
	startCmd.Flags().StringVarP(&keystoreLoc, "keystore", "k", "/root/bor_keystore", "Keystore file to use for signing")
	startCmd.Flags().StringVarP(&passwordLoc, "password", "p", "/root/bor_password", "Keystore file password to use for signing")
}
