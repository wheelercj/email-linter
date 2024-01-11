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
	"slices"
	"strings"
)

// getSingleUseAddresses makes a web request for emails in the inbox, and from them
// finds the receiving single-use addresses. The email addresses are sorted
// alphabetically and deduplicated.
func getSingleUseAddresses(inboxId, accountId, url, token string) []string {
	emailsList := getInboxEmailsRecipients(inboxId, accountId, url, token)

	var singleUseAddresses []string
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
			singleUseAddresses = appendIfSingleUse(singleUseAddresses, toList)
		}
	}

	slices.Sort(singleUseAddresses)
	singleUseAddresses = slices.Compact(singleUseAddresses)
	if Verbose {
		fmt.Printf("%d single-use addresses found:\n", len(singleUseAddresses))
		fmt.Println("\t" + strings.Join(singleUseAddresses, "\n\t"))
	}

	return singleUseAddresses
}

// getSendersToSingleUseAddresses makes a web request for the "to" and "from" fields of
// all emails outside the spam folder received through single-use addresses. Each item
// of the returned map has keys of the "to" addresses, and values of slices of the
// corresponding "from" addresses.
func getSendersToSingleUseAddresses(
	singleUseAddresses []string, spamId, accountId, url, token string,
) map[string][]string {
	singleUseAddressesStr := strings.Join(singleUseAddresses, "\"}, {\"to\": \"")
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
	`, accountId, spamId, singleUseAddressesStr, accountId)

	emailsList := getEmailsList(emailsReqBody, url, token)

	toAndFrom := make(map[string][]string)
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		to := strings.ToLower(email["to"].([]any)[0].(map[string]any)["email"].(string))
		from := strings.ToLower(email["from"].([]any)[0].(map[string]any)["email"].(string))
		toAndFrom[to] = append(toAndFrom[to], from)
	}

	return toAndFrom
}

// appendIfSingleUse finds the recipient email address in emailDataList, which is a
// slice containing one map with keys "name" and "email". If the email address's domain
// is that of an email protection service, the address is added to the slice of
// single-use addresses. All email addresses are lowercased.
func appendIfSingleUse(singleUseAddresses []string, emailDataList []any) []string {
	if len(emailDataList) > 1 {
		panic("Multiple recipients currently not supported")
	}
	to := emailDataList[0].(map[string]any)
	address := strings.ToLower(to["email"].(string))
	domain := strings.Split(address, "@")[1]
	if strings.Contains(Domains, domain) {
		singleUseAddresses = append(singleUseAddresses, address)
	}
	return singleUseAddresses
}

// printAddresses prints all the single-use email addresses and the addresses they
// received emails from. The "from" addresses are sorted alphabetically and
// deduplicated.
func printAddresses(singleUseAddresses []string, toAndFrom map[string][]string) {
	if len(singleUseAddresses) == 0 {
		fmt.Println("No single-use addresses found.")
		return
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
		fmt.Printf(
			"Your inbox's %d single-use addresses and those they received from:\n",
			len(singleUseAddresses),
		)
		for to := range toAndFrom {
			fmt.Println(to)
			for _, from := range toAndFrom[to] {
				fmt.Printf("\t%s\n", from)
			}
		}
	}
}
