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
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

// getDisposableAddresses makes a web request for emails in the inbox, and from them
// finds the receiving disposable addresses. The email addresses are sorted
// alphabetically and deduplicated.
func getDisposableAddresses(inboxId, accountId, url, token string) []string {
	emailsList := getInboxEmailsRecipients(inboxId, accountId, url, token)
	if len(emailsList) == 0 {
		return nil
	}

	var disposableAddresses []string
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		// if ccList := email["cc"]; ccList != nil {
		// 	slog.Warn("CC field detected and ignored: %s", email["cc"])
		// }
		// if bccListAny := email["bcc"]; bccListAny != nil {
		// 	slog.Warn("BCC field detected and ignored: %s", email["bcc"])
		// }
		// if fromListAny := email["from"]; fromListAny != nil {
		// 	slog.Info("from field ignored")
		// }
		if toListAny := email["to"]; toListAny != nil {
			toList := toListAny.([]any)
			disposableAddresses = appendIfDisposable(disposableAddresses, toList)
		}
	}
	if len(disposableAddresses) == 0 {
		return nil
	}

	slices.Sort(disposableAddresses)
	disposableAddresses = slices.Compact(disposableAddresses)
	if Verbose {
		fmt.Printf("%d disposable addresses found:\n", len(disposableAddresses))
		fmt.Println("\t" + strings.Join(disposableAddresses, "\n\t"))
	}

	return disposableAddresses
}

// getSendersToDisposableAddresses makes a web request for the "to" and "from" fields of
// all emails outside the spam folder received through disposable addresses. Each item
// of the returned map has keys of the "to" addresses, and values of slices of the
// corresponding "from" addresses.
func getSendersToDisposableAddresses(
	disposableAddresses []string, spamId, accountId, url, token string,
) map[string][]string {
	disposableAddressesStr := strings.Join(disposableAddresses, "\"}, {\"to\": \"")
	emailsReqBody := fmt.Sprintf(`
		{
			"using": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"],
			"methodCalls": [
				[
					"Email/query",
					{
						"accountId": "%s",
						"filter": {
							"operator": "AND",
							"conditions": [
								{
									"inMailboxOtherThan": ["%s"]
								},
								{
									"operator": "OR",
									"conditions": [{"to": "%s"}]
								}
							]
						}
					},
					"0"
				],
				[
					"Email/get",
					{
						"accountId": "%s",
						"#ids": {
							"resultOf": "0",
							"name": "Email/query",
							"path": "/ids"
						},
						"properties": ["to", "from"]
					},
					"1"
				]
			]
		}
	`, accountId, spamId, disposableAddressesStr, accountId)

	emailsList := getEmailsList(emailsReqBody, url, token)

	toAndFrom := make(map[string][]string)
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		to := strings.ToLower(email["to"].([]any)[0].(map[string]any)["email"].(string))
		if slices.Contains(disposableAddresses, to) {
			from := strings.ToLower(email["from"].([]any)[0].(map[string]any)["email"].(string))
			toAndFrom[to] = append(toAndFrom[to], from)
		}
	}

	return toAndFrom
}

// appendIfDisposable finds the recipient email address in emailDataList, which is a
// slice containing one map with keys "name" and "email". If the email address's domain
// is that of an email protection service, the address is added to the slice of
// disposable addresses. All email addresses are lowercased.
func appendIfDisposable(disposableAddresses []string, emailDataList []any) []string {
	if len(emailDataList) > 1 {
		panic("Multiple recipients currently not supported")
	}
	to := emailDataList[0].(map[string]any)
	address := strings.ToLower(to["email"].(string))
	domain := strings.Split(address, "@")[1]
	if strings.Contains(Domains, domain) {
		disposableAddresses = append(disposableAddresses, address)
	}
	return disposableAddresses
}

// printAddresses prints all the disposable email addresses and the addresses they
// received emails from. The "from" addresses are sorted alphabetically and
// deduplicated. If JSON is not being printed and there are more than a certain number
// of unique senders to one address, the number of senders is printed instead of their
// addresses.
func printAddresses(disposableAddresses []string, toAndFrom map[string][]string) {
	if len(disposableAddresses) == 0 {
		fmt.Fprint(os.Stderr, "No disposable addresses found in your inbox")
		os.Exit(0)
	}

	for to := range toAndFrom {
		slices.Sort(toAndFrom[to])
		toAndFrom[to] = slices.Compact(toAndFrom[to])
	}

	if PrintJson {
		bytes, err := json.Marshal(toAndFrom)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bytes))
	} else {
		if len(disposableAddresses) == 1 {
			fmt.Printf(
				"Your inbox's 1 disposable address and those it received from:\n",
			)
		} else {
			fmt.Printf(
				"Your inbox's %d disposable addresses and those they received from:\n",
				len(disposableAddresses),
			)
		}
		for to := range toAndFrom {
			fmt.Println(to)
			froms := toAndFrom[to]
			if len(froms) > MaxFrom {
				fmt.Printf(
					"\tReceived emails from %d unique addresses. Use `-f %d` if you want to see them.\n",
					len(froms),
					len(froms),
				)
			} else {
				for _, from := range froms {
					fmt.Printf("\t%s\n", from)
				}
			}
		}
	}
}
