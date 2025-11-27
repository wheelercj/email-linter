# Mailbox roles

Each mailbox (inbox, drafts, spam, archive, etc.) has a role and a name (and other attributes).

> Note, you should always use the `role` attribute, as names may be localised (or even different between different servers with the same language)!
>
> — [JMAP Client Guide](https://jmap.io/client.html#:~:text=note%2C%20you%20should%20always%20use%20the%20role%20attribute%2C%20as%20names%20may%20be%20localised%20%28or%20even%20different%20between%20different%20servers%20with%20the%20same%20language%29!)

Many roles are possible. Here are some of the most common roles used by JMAP: `all`, `archive`, `drafts`, `flagged`, `important`, `inbox`, `junk`, `scheduled`, `sent`, `snoozed`, `subscribed`, `trash`. Custom mailboxes seem to never have a role; their `role` attribute has a value of `null`. I created the list of roles by combining [IMAP Mailbox Name Attributes](https://www.iana.org/assignments/imap-mailbox-name-attributes/imap-mailbox-name-attributes.xhtml) with the output of the code below.

```go
func getMailboxRoles(accountId, url, token string) {
	reqBody := fmt.Sprintf(`
		{
			"using": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"],
			"methodCalls": [
				[
					"Mailbox/get",
					{
						"accountId": "%s",
						"properties": ["id", "role", "name"]
					},
					"0"
				]
			]
		}
	`, accountId)
	res, err := makeJmapCall("POST", url, token, reqBody)
	if err != nil {
		panic(err)
	}
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(resBytes))
}
```
