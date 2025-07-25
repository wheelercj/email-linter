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

// getMaskedAddrs makes a web request for emails in the inbox, and from them finds the receiving
// masked addresses. The email addresses are sorted alphabetically and deduplicated. If the total
// number of email threads in the inbox is greater than the number of emails retrieved and the
// output format is not JSON, a message is printed explaining this.
func getMaskedAddrs(inboxId, accountId, url, token string) []string {
	emailsList, totalThreads := getInboxEmailsRecipients(inboxId, accountId, url, token)
	if len(emailsList) == 0 {
		return nil
	}
	if !PrintJson && totalThreads > len(emailsList) {
		fmt.Printf(
			"Looking for masked addresses in the newest %d of %d email threads.\n",
			len(emailsList),
			totalThreads,
		)
	}

	var maskedAddrs []string
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		if toListAny := email["to"]; toListAny != nil {
			maskedAddrs = appendIfMasked(maskedAddrs, toListAny.([]any))
		}
		if ccListAny := email["cc"]; ccListAny != nil {
			maskedAddrs = appendIfMasked(maskedAddrs, ccListAny.([]any))
		}
		if bccListAny := email["bcc"]; bccListAny != nil {
			maskedAddrs = appendIfMasked(maskedAddrs, bccListAny.([]any))
		}
	}
	if len(maskedAddrs) == 0 {
		return nil
	}

	slices.Sort(maskedAddrs)
	maskedAddrs = slices.Compact(maskedAddrs)
	if Verbose {
		fmt.Printf("%d masked addresses found:\n", len(maskedAddrs))
		fmt.Println("\t" + strings.Join(maskedAddrs, "\n\t"))
	}

	return maskedAddrs
}

// getSendersToMaskedAddrs makes a web request for the "to", "cc", "bcc", and "from" fields of all
// emails outside the spam folder received through masked addresses. Each item of the returned map
// has keys of the recipient addresses, and values of slices of the corresponding "from" addresses.
// If the number of matching emails goes over a limit and the output format is not JSON, a message
// is printed explaining this.
func getSendersToMaskedAddrs(
	maskedAddrs []string, spamId, accountId, url, token string,
) map[string][]string {
	toDispAddrsStr := strings.Join(maskedAddrs, "\"}, {\"to\": \"")
	ccDispAddrsStr := strings.Join(maskedAddrs, "\"}, {\"cc\": \"")
	bccDispAddrsStr := strings.Join(maskedAddrs, "\"}, {\"bcc\": \"")
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
									"conditions": [{"to": "%s"}, {"cc": "%s"}, {"bcc": "%s"}]
								}
							]
						},
						"sort": [{
							"isAscending": false,
							"property": "receivedAt"
						}],
						"collapseThreads": true,
						"position": 0,
						"limit": 100,
						"calculateTotal": true
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
						"properties": ["to", "cc", "bcc", "from"]
					},
					"1"
				]
			]
		}
	`, accountId, spamId, toDispAddrsStr, ccDispAddrsStr, bccDispAddrsStr, accountId)

	emailsList, totalMatches := getEmailsList(emailsReqBody, url, token)

	if !PrintJson && totalMatches > len(emailsList) {
		fmt.Printf(
			"Retrieved %d of %d emails received at masked addresses.\n",
			len(emailsList),
			totalMatches,
		)
	}

	toAndFrom := make(map[string][]string)
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		tos := getAddrs(email, "to")
		ccs := getAddrs(email, "cc")
		bccs := getAddrs(email, "bcc")
		froms := getAddrs(email, "from")

		for _, to := range tos {
			if slices.Contains(maskedAddrs, to) {
				toAndFrom[to] = append(toAndFrom[to], froms...)
			}
		}
		for _, cc := range ccs {
			if slices.Contains(maskedAddrs, cc) {
				toAndFrom[cc] = append(toAndFrom[cc], froms...)
			}
		}
		for _, bcc := range bccs {
			if slices.Contains(maskedAddrs, bcc) {
				toAndFrom[bcc] = append(toAndFrom[bcc], froms...)
			}
		}
	}

	return toAndFrom
}

// getAddrs gets email addresses from an email's sender and recipient data. The category determines
// which email addresses; it can be "to", "cc", "bcc", or "from".
func getAddrs(email map[string]any, category string) []string {
	cat := email[category]
	if cat == nil {
		return nil
	}

	var addrs []string
	for _, person := range cat.([]any) {
		addrs = append(addrs, person.(map[string]any)["email"].(string))
	}

	return addrs
}

// appendIfMasked finds in one email's recipients any and all email addresses that are masked email
// addresses. recipientMaps is a slice of maps each with keys "name" and "email" representing one
// recipient. This function lowercases all email addresses. If multiple masked recipient addresses
// are found and attempts to determine which one belongs to the user do not completely succeed, a
// warning is printed and multiple addresses are added to the return.
func appendIfMasked(maskedAddrs []string, recipientMaps []any) []string {
	var newDispAddrs []string
	for _, recipientMap := range recipientMaps {
		recipient := recipientMap.(map[string]any)
		address := strings.ToLower(recipient["email"].(string))
		domain := strings.Split(address, "@")[1]
		if strings.Contains(Domains, domain) {
			newDispAddrs = append(newDispAddrs, address)
		}
	}

	if len(newDispAddrs) > 1 {
		newDispAddrs = removeNonUserAddrs(newDispAddrs)
	}

	maskedAddrs = append(maskedAddrs, newDispAddrs...)

	return maskedAddrs
}

// removeNonUserAddrs attempts to remove from dispAddrs any email addresses that do not belong to
// the user. dispAddrs is the masked addresses in one email's recipient addresses. If multiple
// masked recipient addresses remain, a warning is printed.
func removeNonUserAddrs(dispAddrs []string) []string {
	// If the email was forwarded to a duck address, the recipient addresses will all have the same
	// ending: the user's duck address.
	var unique []string
	for _, addr := range dispAddrs {
		if strings.HasSuffix(addr, "@duck.com") && strings.Contains(addr, "_at_") {
			tokens := strings.Split(addr, "_")
			addr = tokens[len(tokens)-1]
		}
		if !slices.Contains(unique, addr) {
			unique = append(unique, addr)
		}
	}

	dispAddrs = unique

	if len(dispAddrs) > 1 {
		fmt.Printf(
			"Warning: multiple masked addresses found in one email: %v\n",
			dispAddrs,
		)
	}

	return dispAddrs
}

// printAddrs prints all the masked email addresses and the addresses they received emails from.
// The "from" addresses are sorted alphabetically and deduplicated. If JSON is not being printed
// and there are more than a certain number of unique senders to one address, the number of senders
// is printed instead of their addresses.
func printAddrs(maskedAddrs []string, toAndFrom map[string][]string) {
	if len(maskedAddrs) == 0 {
		fmt.Fprintln(os.Stderr, "No masked addresses found in your inbox")
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
		if len(maskedAddrs) == 1 {
			fmt.Printf(
				"Your inbox's 1 masked address and those it received from:\n",
			)
		} else {
			fmt.Printf(
				"Your inbox's %d masked addresses and those they received from:\n",
				len(maskedAddrs),
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
