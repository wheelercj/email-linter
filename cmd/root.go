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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path"
	"strings"

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
	if len(singleUseAddresses) == 0 {
		fmt.Fprint(os.Stderr, "No single-use addresses found")
		os.Exit(0)
	}
	toAndFrom := getSendersToSingleUseAddresses(singleUseAddresses, spamId, accountId, url, token)
	printAddresses(singleUseAddresses, toAndFrom)
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "email-linter",
	Version: "v0.0.5",
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

// getApiToken looks for and returns a JMAP token. It first looks for
// `~/.config/email-linter/jmap_token`. If a token is not found there, it checks for a
// JMAP_TOKEN environment variable. If this does not exist either, it asks for the token
// to be entered interactively.
func getApiToken() string {
	var token string

	// Look for a token file named `~/.config/email-linter/jmap_token`.
	tokenFilePath, err := expandTilde("~/.config/email-linter/jmap_token")
	if err != nil {
		slog.Warn(err.Error())
	} else {
		isFile, err := fileExists(tokenFilePath)
		if err != nil {
			slog.Warn(err.Error())
		} else if isFile {
			bytes, err := os.ReadFile(tokenFilePath)
			if err != nil {
				slog.Warn(err.Error())
			} else if len(bytes) > 0 {
				token = string(bytes)
			}
		}
	}

	// Look for a token env var named `JMAP_TOKEN`.
	if len(token) == 0 {
		token = os.Getenv("JMAP_TOKEN")
	}

	// Ask for the token to be entered interactively.
	if len(token) == 0 {
		fmt.Print(
			`Create a read-only JMAP API token and either:
  * put it in a file named ~/.config/email-linter/jmap_token
  * or put it in a environment variable named JMAP_TOKEN
  * or enter the token here: `,
		)
		_, err := fmt.Scanln(&token)
		if err != nil {
			panic(err)
		}
	}

	return token
}

// fileExists determines whether a file exists.
func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// expandTilde replaces any leading `~/` in filePath with the current user's home
// folder. Any backslashes are replaced with forward slashes. If filePath does not start
// with `~/`, it is returned unchaged (unless it had backslashes replaced).
func expandTilde(filePath string) (string, error) {
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	if !strings.HasPrefix(filePath, "~/") {
		return filePath, nil
	}
	filePath = strings.TrimPrefix(filePath, "~/")
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	filePath = path.Join(u.HomeDir, filePath)
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	return filePath, nil
}
