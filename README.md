# Email Linter

[![Go Reference](https://pkg.go.dev/badge/github.com/wheelercj/email-linter.svg)](https://pkg.go.dev/github.com/wheelercj/email-linter)

Easily find spam and phishing emails received at [masked email addresses](#what-are-masked-email-addresses). This command line app currently works with [Fastmail](https://www.fastmail.com/features/), [Topicbox](https://www.topicbox.com/), and any other email services that have a [JMAP](https://jmap.io/index.html) API. See more examples of these services [here](https://jmap.io/software.html).

```
$ email-linter
Your inbox's 2 masked addresses and those they received from:
abc123@duck.com
        donotreply_at_email.schwab.com_abc123@duck.com
        donotreply_at_mail.schwab.com_abc123@duck.com
        id_at_proxyvote.com_abc123@duck.com
def456@duck.com
        venmo_at_venmo.com_def456@duck.com
```

Email Linter lists each of your masked addresses and all the addresses they have received from so you can quickly spot suspicious senders.

## Download

Either:

* [download a zipped executable file](https://github.com/wheelercj/email-linter/releases), unzip it, and run the app with `./email-linter --help`
* or run `go install github.com/wheelercj/email-linter@latest` and then `email-linter --help`

## Privacy

Email Linter communicates with your email service and optionally stores your API token in your device's keyring. No other communication nor storage takes place. There is a chance future versions of Email Linter will store email addresses locally to work more efficiently and offer more features.

## What are masked email addresses?

They are email addresses created to be used for only one account each. Whenever one of these addresses starts receiving spam or phishing emails, you know exactly which account was compromised and can disconnect the address from your inbox. This way, you immediately stop receiving spam and never have to give your main email address to anyone you don't trust. Some examples of these email protection services are [DuckDuckGo's Email Protection](https://duckduckgo.com/email), [Fastmail's Masked Email](https://www.fastmail.help/hc/en-us/articles/4406536368911-Masked-Email), [Proton's hide-my-email aliases](https://proton.me/pass/aliases), [Firefox Relay](https://relay.firefox.com/), and [iCloud+'s Hide My Email](https://support.apple.com/en-us/105078). Since the emails received by these addresses _should_ have predictable "from" fields, suspicious senders can be easily found with Email Linter. If needed, you can customize which email protection service addresses to search for. Use the `--help` option for more info.

## Why

I got phished. Fortunately, it was a fake phishing email for training against phishing, but I learned to not look at emails while half-asleep and, more importantly, the sender's address was different from normal for the masked address I used. Email services don't seem to consider that suspicious (at least not yet), and checking the sender's address manually for every email is tedious if you don't remember the correct sender address. Email Linter automates checking sender addresses for you. I hope email services will make it obsolete.

## How does it work?

1. First, Email Linter finds all emails in your inbox that went through an email protection service.
2. Next, it finds all emails outside your spam folder those masked addresses have ever received.
3. Then it lists each masked address and the addresses they have received from. This makes it simple to spot suspicious senders so you can easily search your inbox for malicious emails and decide what to do with them.

## API token

Email Linter needs a read-only JMAP API token to securely connect to your account. If you're using Fastmail, you can [create an API token here](https://www.fastmail.com/settings/security/tokens).

When you run the app, you will be asked to enter the token and whether you want to save it in your device's keyring.

## Caveats

* Email Linter does not protect against [email spoofing](https://til.chriswheeler.dev/email-spoofing/).
* There's a chance Email Linter will think someone else's email address is yours. Why and when is explained in [./docs/multiple-recipients.md](./docs/multiple-recipients.md).

## Dev resources

Here are some resources that were helpful while creating this app.

* [intro to Go](https://til.chriswheeler.dev/intro-to-go/)
* [Integrating with Fastmail](https://www.fastmail.com/for-developers/integrating-with-fastmail/)
* [Fastmail's JMAP samples](https://github.com/fastmail/JMAP-Samples/tree/main)
* [JMAP Crash Course](https://jmap.io/crash-course.html)
* [the JMAP specs and RFCs](https://jmap.io/spec.html)
* [How and why we built Masked Email with JMAP](https://blog.1password.com/making-masked-email-with-jmap/) by Madeline Hanley at 1Password
* [spf13/cobra](https://github.com/spf13/cobra)
* [GoReleaser](https://goreleaser.com/)
* [GoReleaser Action](https://github.com/marketplace/actions/goreleaser-action)
