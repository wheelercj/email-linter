# email-linter

[![Go Reference](https://pkg.go.dev/badge/github.com/wheelercj/email-linter.svg)](https://pkg.go.dev/github.com/wheelercj/email-linter)

Easily find spam and phishing emails received at [single-use email addresses](#what-are-single-use-email-addresses). This command-line app currently works with Fastmail and any other email services that have a [JMAP](https://jmap.io/index.html) API.

![demo](demo.png)

## download

Either:

* run `go install github.com/wheelercj/email-linter@latest` and then `email-linter --help`
* or [download a zipped executable file](https://github.com/wheelercj/email-linter/releases), unzip it, and run the app with `./email-linter --help`

## what are single-use email addresses?

They are email addresses created to be used for only one account each. Whenever one of these email address starts receiving spam or phishing emails, you know exactly which account was compromised and can disconnect the address from your inbox. This way, you immediately stop receiving spam and never have to give your main email address to anyone you don't trust. Some examples of these email protection services are [DuckDuckGo's Email Protection](https://duckduckgo.com/email), [1Password's Masked Email](https://1password.com/fastmail/), [Firefox Relay](https://relay.firefox.com/), and [iCloud+'s Hide My Email](https://support.apple.com/en-us/105078). Since the emails received by these addresses _should_ have predictable "from" fields, suspicious senders can be easily found. If needed, you can customize which email protection service addresses this app searches for. Use the `--help` option for more info.

## how does it work?

1. First, email-linter finds all emails in your inbox that went through an email protection service.
2. Next, it finds all emails outside your spam folder those single-use addresses have ever received.
3. Then it lists each single-use address and the addresses they have received from. This makes it simple to spot suspicious senders so you can easily search your inbox for malicious emails and decide what to do with them.

This app does not store any of your data anywhere and only communicates with your email service.

## API token

email-linter needs a read-only JMAP API token to securely connect to your account. If you're using Fastmail, you can [create an API token here](https://www.fastmail.com/settings/security/tokens).

**Choose one.** The token can be entered in any one of three ways:

* **When you run the app**, you can enter the token interactively if you haven't chosen any of the other options.
* **Create a file** for the token with the location and name `~/.config/email-linter/jmap_token` (`~` is the user folder, such as `C:/Users/chris`).
* **Create an environment variable** named `JMAP_TOKEN`. This option is generally not recommended because any process can read the environment variable.

If both a token file and environment variable are provided, the file is used.

## dev resources

Here are some resources that were helpful while creating this app.

* [intro to Go](https://wheelercj.github.io/notes/pages/20221122173910.html)
* [Integrating with Fastmail](https://www.fastmail.com/for-developers/integrating-with-fastmail/)
* [Fastmail's JMAP samples](https://github.com/fastmail/JMAP-Samples/tree/main)
* [JMAP Crash Course](https://jmap.io/crash-course.html)
* [the JMAP specs and RFCs](https://jmap.io/spec.html)
