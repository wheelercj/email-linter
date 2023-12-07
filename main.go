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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

var emailProtectionDomains []string = []string{"duck.com", "mozmail.com", "icloud.com"}
var verbose bool = false

func main() {
	godotenv.Load()
	token := os.Getenv("JMAP_TOKEN")
	if len(token) == 0 {
		fmt.Println(
			"Create a Fastmail API token and either create a JMAP_TOKEN env var and",
			"run this again, or enter the token here:",
		)
		_, err := fmt.Scanln(&token)
		if err != nil {
			panic(err)
		}
	}

	sessionReq, err := http.NewRequest(
		"GET",
		"https://api.fastmail.com/jmap/session",
		nil,
	)
	if err != nil {
		panic(err)
	}
	sessionReq.Header.Set("Content-Type", "application/json")
	sessionReq.Header.Set(
		"Authorization",
		fmt.Sprintf("Bearer %s", token),
	)

	sessionRes, err := http.DefaultClient.Do(sessionReq)
	if err != nil {
		panic(err)
	}
	sessionBytes, err := io.ReadAll(sessionRes.Body)
	if err != nil {
		panic(err)
	}
	if strings.EqualFold(string(sessionBytes), "Authorization header not a valid format\n") {
		panic("Authorization header in an invalid format or has an invalid token")
	}
	var session map[string]any
	err = json.Unmarshal(sessionBytes, &session)
	if err != nil {
		panic(err)
	}

	primaryAccounts := session["primaryAccounts"].(map[string]any)
	accountID := primaryAccounts["urn:ietf:params:jmap:mail"].(string)
	if verbose {
		fmt.Printf("account ID: %s\n", accountID)
	}
	url := session["apiUrl"].(string)

	// TODO: try to combine these into one web request.
	inboxID := getMailboxID(`{"role": "inbox"}`, accountID, url, token)
	spamID := getMailboxID(`{"name": "spam"}`, accountID, url, token)
	if verbose {
		fmt.Printf("inbox ID: %s\n", inboxID)
		fmt.Printf("spam folder ID: %s\n", spamID)
	}

	emailsReqBody := fmt.Sprintf(`
		{
			"using": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"],
			"methodCalls": [
				[
					"Email/query",
					{
						"accountId": "%s",
						"filter": {"inMailbox": "%s"}
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
						"properties": ["to"]
					},
					"1"
				]
			]
		}
	`, accountID, inboxID, accountID)
	// TODO: above, add properties:
	// "properties": ["from", "to", "cc", "bcc"]

	emailsList := getEmailsList(emailsReqBody, url, token)

	var singleUseAddresses []string
	for _, emailAny := range emailsList {
		email := emailAny.(map[string]any)
		// if ccList := email["cc"]; ccList != nil {
		// 	slog.Warn("CC field detected (and ignored): %s", email["cc"])
		// }
		// if bccListAny := email["bcc"]; bccListAny != nil {
		// 	slog.Warn("BCC field detected (and ignored): %s", email["bcc"])
		// }
		// if fromListAny := email["from"]; fromListAny != nil {
		// 	fromList := fromListAny.([]any)
		// 	singleUseAddresses = appendIfSingleUse(singleUseAddresses, fromList)
		// }
		if toListAny := email["to"]; toListAny != nil {
			toList := toListAny.([]any)
			singleUseAddresses = appendIfSingleUse(singleUseAddresses, toList)
		}
	}

	slices.Sort(singleUseAddresses)
	singleUseAddresses = slices.Compact(singleUseAddresses)
	if verbose {
		fmt.Printf("%d single-use addresses found:\n", len(singleUseAddresses))
		fmt.Println("\t" + strings.Join(singleUseAddresses, "\n\t"))
	}
	singleUseAddressesStr := strings.Join(singleUseAddresses, "\"}, {\"to\": \"")

	emailsReq2Body := fmt.Sprintf(`
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
	`, accountID, spamID, singleUseAddressesStr, accountID)

	emailsList2 := getEmailsList(emailsReq2Body, url, token)

	toAndFrom := make(map[string][]string)
	for _, emailAny := range emailsList2 {
		email := emailAny.(map[string]any)
		to := strings.ToLower(email["to"].([]any)[0].(map[string]any)["email"].(string))
		from := strings.ToLower(email["from"].([]any)[0].(map[string]any)["email"].(string))
		if _, ok := toAndFrom[to]; ok {
			toAndFrom[to] = append(toAndFrom[to], from)
		} else {
			toAndFrom[to] = []string{from}
		}
	}

	fmt.Printf(
		"Your inbox's %d single-use addresses and those they received from:\n",
		len(singleUseAddresses),
	)
	var printedYet bool
	for to := range toAndFrom {
		slices.Sort(toAndFrom[to])
		toAndFrom[to] = slices.Compact(toAndFrom[to])
		fmt.Println(to)
		for _, from := range toAndFrom[to] {
			fmt.Printf("\t%s\n", from)
		}
		printedYet = true
	}
	if !printedYet {
		fmt.Println("No single-use addresses with multiple recipients found.")
	}
}

func appendIfSingleUse(singleUseAddresses []string, emailDataList []any) []string {
	if len(emailDataList) > 1 {
		slog.Error("Multiple recipients currently not supported")
		return singleUseAddresses
	}
	to := emailDataList[0].(map[string]any)
	address := strings.ToLower(to["email"].(string))
	domain := strings.Split(address, "@")[1]
	if slices.Contains(emailProtectionDomains, domain) {
		singleUseAddresses = append(singleUseAddresses, address)
	}
	return singleUseAddresses
}

func getMailboxID(filterObjStr, accountID, url, token string) string {
	mailboxReqBody := fmt.Sprintf(`
		{
			"using": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"],
			"methodCalls": [
				[
					"Mailbox/query",
					{
						"accountId": "%s",
						"filter": %s
					},
					"0"
				]
			]
		}
	`, accountID, filterObjStr)
	mailboxRes, err := makeJMAPCall(url, token, mailboxReqBody)
	if err != nil {
		panic(err)
	}
	mailboxBytes, err := io.ReadAll(mailboxRes.Body)
	if err != nil {
		panic(err)
	}
	if bytes.Equal(mailboxBytes, []byte("Malformed JSON")) {
		panic("Malformed JSON")
	}
	var mailbox map[string]any
	err = json.Unmarshal(mailboxBytes, &mailbox)
	if err != nil {
		panic(err)
	}
	mailboxMethodRes := mailbox["methodResponses"].([]any)
	mailboxID := mailboxMethodRes[0].([]any)[1].(map[string]any)["ids"].([]any)[0].(string)
	return mailboxID
}

func getEmailsList(emailsReqBody, url, token string) []any {
	emailsRes, err := makeJMAPCall(url, token, emailsReqBody)
	if err != nil {
		panic(err)
	}
	emailsBytes, err := io.ReadAll(emailsRes.Body)
	if err != nil {
		panic(err)
	}
	if bytes.Equal(emailsBytes, []byte("Malformed JSON")) {
		panic("Malformed JSON")
	}

	var emails map[string]any
	err = json.Unmarshal(emailsBytes, &emails)
	if err != nil {
		panic(err)
	}
	emailsMethodRes := emails["methodResponses"].([]any)
	emailsGetRes := emailsMethodRes[1].([]any)
	emailsMap := emailsGetRes[1].(map[string]any)
	emailsList := emailsMap["list"].([]any)
	return emailsList
}

func makeJMAPCall(url, token, body string) (*http.Response, error) {
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(
		"POST",
		url,
		bodyReader,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return http.DefaultClient.Do(req)
}
