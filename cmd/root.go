// Copyright 2023 Chris Wheeler

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Verbose bool
var ApiSessionUrl string
var Domains string
var PrintJson bool

func runFunc(cmd *cobra.Command, args []string) {
	token := getApiToken()
	accountId, url := getAccountIdAndApiUrl(token)
	inboxId, spamId := getInboxAndSpamIds(accountId, url, token)
	singleUseAddresses := getSingleUseAddresses(inboxId, accountId, url, token)
	toAndFrom := getSendersToSingleUseAddresses(singleUseAddresses, spamId, accountId, url, token)
	printAddresses(singleUseAddresses, toAndFrom)
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "email-linter",
	Version: "0.0.3",
	Run:     runFunc,
	Short:   "Easily find spam and phishing emails received at single-use email addresses.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(
		&Verbose,
		"verbose",
		false,
		"display extra info while running",
	)
	rootCmd.Flags().StringVar(
		&ApiSessionUrl,
		"url",
		"https://api.fastmail.com/jmap/session",
		"the API URL to request a session from",
	)
	rootCmd.Flags().StringVarP(
		&Domains,
		"domains",
		"d",
		"duck.com mozmail.com icloud.com",
		"email protection service domains to search for",
	)
	rootCmd.Flags().BoolVarP(
		&PrintJson,
		"json",
		"j",
		false,
		"print output as JSON",
	)
}

// getApiToken looks for a JMAP_TOKEN environment variable, or asks for the token to be
// entered interactively as a fallback.
func getApiToken() string {
	token := os.Getenv("JMAP_TOKEN")
	if len(token) == 0 {
		fmt.Println(
			"Create an API token and either create a JMAP_TOKEN env var and run this",
			"again, or enter the token here:",
		)
		_, err := fmt.Scanln(&token)
		if err != nil {
			panic(err)
		}
	}
	return token
}
