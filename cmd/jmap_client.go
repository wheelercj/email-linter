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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// getAccountIdAndApiUrl makes a web request to get the user's account ID and the API
// session URL to use for all further web requests.
func getAccountIdAndApiUrl(token string) (string, string) {
	session := getApiSession(token)
	primaryAccounts := session["primaryAccounts"].(map[string]any)
	accountId := primaryAccounts["urn:ietf:params:jmap:mail"].(string)
	if Verbose {
		fmt.Printf("account ID: %s\n", accountId)
	}
	url := session["apiUrl"].(string)
	return accountId, url
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

// getInboxAndSpamIds makes a web request for the IDs of the inbox and the spam folders.
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

	if Verbose {
		fmt.Printf("inbox ID: %s\n", inboxId)
		fmt.Printf("spam folder ID: %s\n", spamId)
	}

	return inboxId, spamId
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

	return getEmailsList(emailsReqBody, url, token)
}

// getEmailsList makes a web request to get an array of email data objects from a JMAP
// server. The expected JMAP methods are "Email/query" followed by "Email/get" (two
// methods total).
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
