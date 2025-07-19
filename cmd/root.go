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
	"strings"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

const serviceName = "email-linter"
const userName = "awesome-person"

var Verbose bool
var ApiSessionUrl string
var Domains string
var MaxFrom int
var PrintJson bool

func runFunc(cmd *cobra.Command, args []string) error {
	token, err := getApiToken()
	if err != nil {
		return err
	}

	accountId, url := getAccountIdAndApiUrl(token)
	inboxId, spamId := getInboxAndSpamIds(accountId, url, token)
	maskedAddrs := getMaskedAddrs(inboxId, accountId, url, token)
	if len(maskedAddrs) == 0 {
		return fmt.Errorf("no masked addresses found in your inbox")
	}

	toAndFrom := getSendersToMaskedAddrs(maskedAddrs, spamId, accountId, url, token)
	printAddrs(maskedAddrs, toAndFrom)

	return nil
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "email-linter",
	Version: "v0.0.8",
	RunE:    runFunc,
	Short:   "Easily find spam and phishing emails received at masked email addresses.",
}

// Execute adds all child commands to the root command and sets flags appropriately. This is called
// by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(logoutCmd)

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
		"fastmail.com duck.com mozmail.com icloud.com passmail.com passmail.net",
		"email protection service domains to search for",
	)
	rootCmd.Flags().IntVarP(
		&MaxFrom,
		"maxFrom",
		"f",
		5,
		"max unique senders to a masked email address; does not apply to JSON output",
	)
	rootCmd.Flags().BoolVarP(
		&PrintJson,
		"json",
		"j",
		false,
		"print output as JSON",
	)
}

// getApiToken returns a JMAP token that is retrieved from either the system's keyring or
// interactively from the user. If the user enters the token interactively, they are asked whether
// they want to save it into the system's keyring.
func getApiToken() (string, error) {
	token, err := keyring.Get(serviceName, userName)
	if err == nil && len(token) > 0 {
		return token, nil
	} else if err != nil && err != keyring.ErrNotFound {
		return "", err
	}

	// Ask for the token to be entered interactively.
	for len(token) == 0 {
		fmt.Print("Create a read-only JMAP API token and enter it here: ")
		_, err = fmt.Scanln(&token)
		if err != nil {
			return "", err
		}
		token = strings.TrimSpace(token)
	}

	// Ask whether to save the token into the system's keyring.
	var y_or_n string
	for y_or_n != "y" && y_or_n != "n" {
		fmt.Println("Would you like the token to be saved in your device's keyring? (y/n): ")
		_, err = fmt.Scanln(&y_or_n)
		if err != nil {
			return "", err
		}

		y_or_n = strings.ToLower(y_or_n)
	}

	if y_or_n == "y" {
		// Save the token into the system's keyring.
		err = keyring.Set(serviceName, userName, token)
		if err != nil {
			return "", err
		}
	}

	return token, nil
}
