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
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

// emailProtectionDomains is a slice of the parts after `@` of email addresses from
// email protection services.
var emailProtectionDomains []string = []string{"duck.com", "mozmail.com", "icloud.com"}
var verbose bool = false
var ApiSessionUrl string = "https://api.fastmail.com/jmap/session"

func main() {
	godotenv.Load()
	token := getApiToken()
	accountId, url := getAccountIdAndApiUrl(token)
	inboxId, spamId := getInboxAndSpamIds(accountId, url, token)
	singleUseAddresses := getSingleUseAddresses(inboxId, accountId, url, token)
	toAndFrom := getSendersToSingleUseAddresses(singleUseAddresses, spamId, accountId, url, token)
	printAddresses(singleUseAddresses, toAndFrom)
}

// getApiToken looks for a JMAP_TOKEN environment variable, and asks for the token to be
// entered interactively as a fallback.
func getApiToken() string {
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
	return token
}

// getApiSession makes a web request to create an API session.
func getApiSession(token string) map[string]any {
	sessionRes, err := makeJmapCall("GET", ApiSessionUrl, token, "")
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
	return session
}

// getAccountIdAndApiUrl makes a web request to get the user's account ID and the API
// session URL to use for all further web requests.
func getAccountIdAndApiUrl(token string) (string, string) {
	session := getApiSession(token)
	primaryAccounts := session["primaryAccounts"].(map[string]any)
	accountId := primaryAccounts["urn:ietf:params:jmap:mail"].(string)
	if verbose {
		fmt.Printf("account ID: %s\n", accountId)
	}
	url := session["apiUrl"].(string)
	return accountId, url
}

// getInboxEmailsRecipients makes a web request for the names and addresses in the "to"
// fields of all emails in the inbox.
func getInboxEmailsRecipients(inboxId, accountId, url, token string) []any {
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
	`, accountId, inboxId, accountId)

	emailsList := getEmailsList(emailsReqBody, url, token)
	return emailsList
}

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
	if verbose {
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
	`, accountId, spamId, singleUseAddressesStr, accountId)

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

	return toAndFrom
}

// printAddresses prints all the single-use email addresses and the addresses they
// received emails from. The "from" addresses are sorted alphabetically and
// deduplicated.
func printAddresses(singleUseAddresses []string, toAndFrom map[string][]string) {
	if len(singleUseAddresses) == 0 {
		fmt.Println("No single-use addresses found.")
		return
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
		fmt.Println("None of the single-use addresses have received emails.")
	}
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
	if slices.Contains(emailProtectionDomains, domain) {
		singleUseAddresses = append(singleUseAddresses, address)
	}
	return singleUseAddresses
}

// getInboxAndSpamIds makes a web request for the IDs of the inbox and the spam folder.
func getInboxAndSpamIds(accountId, url, token string) (string, string) {
	mailboxesReqBody := fmt.Sprintf(`
		{
			"using": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"],
			"methodCalls": [
				[
					"Mailbox/query",
					{
						"accountId": "%s",
						"filter": {
							"operator": "OR",
							"conditions": [
								{"role": "inbox"},
								{"name": "spam"}
							]
						}
					},
					"0"
				],
				[
					"Mailbox/get",
					{
						"accountId": "%s",
						"#ids": {
							"resultOf": "0",
							"name": "Mailbox/query",
							"path": "/ids"
						},
						"properties": ["id", "role", "name"]
					},
					"1"
				]
			]
		}
	`, accountId, accountId)
	mailboxesRes, err := makeJmapCall("POST", url, token, mailboxesReqBody)
	if err != nil {
		panic(err)
	}
	mailboxesBytes, err := io.ReadAll(mailboxesRes.Body)
	if err != nil {
		panic(err)
	}
	if bytes.Equal(mailboxesBytes, []byte("Malformed JSON")) {
		panic("Malformed JSON")
	}
	var mailboxes map[string]any
	err = json.Unmarshal(mailboxesBytes, &mailboxes)
	if err != nil {
		panic(err)
	}
	mailboxesMethodRes := mailboxes["methodResponses"].([]any)
	inboxAndSpam := mailboxesMethodRes[1].([]any)[1].(map[string]any)["list"].([]any)

	var inboxId, spamId string
	ufo := inboxAndSpam[0].(map[string]any) // unidentified folder object
	otherUfo := inboxAndSpam[1].(map[string]any)
	ufoName := strings.ToLower(ufo["name"].(string))
	ufoRole := strings.ToLower(ufo["role"].(string))
	if ufoName == "inbox" || ufoRole == "inbox" {
		inboxId = ufo["id"].(string)
		spamId = otherUfo["id"].(string)
	} else {
		inboxId = otherUfo["id"].(string)
		spamId = ufo["id"].(string)
	}

	if verbose {
		fmt.Printf("inbox ID: %s\n", inboxId)
		fmt.Printf("spam folder ID: %s\n", spamId)
	}

	return inboxId, spamId
}

// getEmailsList makes a web request with a JMAP API request body, the API's url, and an
// API token to get an array of email data objects from a JMAP server. The expected JMAP
// methods are "Email/query" followed by "Email/get" (two methods total).
func getEmailsList(emailsReqBody, url, token string) []any {
	emailsRes, err := makeJmapCall("POST", url, token, emailsReqBody)
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

// makeJmapCall makes a web request with content type "application/json" using the
// default http client. If the given body string is empty, nil is sent as the body.
func makeJmapCall(httpMethod, url, token, body string) (*http.Response, error) {
	var req *http.Request
	var err error
	if len(body) > 0 {
		bodyReader := bytes.NewReader([]byte(body))
		req, err = http.NewRequest(httpMethod, url, bodyReader)
	} else {
		req, err = http.NewRequest(httpMethod, url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return http.DefaultClient.Do(req)
}
