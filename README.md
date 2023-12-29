# email linter

Easily find spam and phishing emails received at single-use email addresses. This command-line app currently works with Fastmail and any other email services that have a [JMAP](https://jmap.io/index.html) API.

![demo](demo.png)

## download

[Download here](https://github.com/wheelercj/email-linter/releases). Binaries are available for several platforms. Optionally, you can add the executable file's path to your computer's PATH environment variable to make the app easier to run.

## what are single-use email addresses?

Email address protection services make it easy to use a unique email address for each online account, increasing security and reducing spam. Some examples of these services are [DuckDuckGo's Email Protection](https://duckduckgo.com/email), [1Password's Masked Email](https://1password.com/fastmail/), [Firefox Relay](https://relay.firefox.com/), and [iCloud+'s Hide My Email](https://support.apple.com/en-us/105078). Since the emails received by these addresses _should_ have predictable "from" fields, suspicious senders can be easily found. If needed, you can customize which email protection service addresses this app searches for. Use the `--help` option for more info.

## how does it work?

1. Find all emails in the inbox that went through an email protection service.
2. Find all emails outside the spam folder those single-use addresses have ever received.
3. List each single-use address and the addresses they have received from. With these, you can search your inbox for suspicious emails and decide what to do with them.

This app does not store any of your data, and only communicates with Fastmail's servers.

## OAuth or API token

Fastmail OAuth is supported. If you don't use Fastmail, you will need to get a read-only JMAP API token for this app to securely connect to your account. Provide the token by creating an environment variable named `API_TOKEN`, such as with the Bash command `API_TOKEN="your token here"`, or the PowerShell command `$env:API_TOKEN="your token here"`. If an `API_TOKEN` variable does not exist when the app runs, it will assume you are using Fastmail.

## dev resources

Here are some resources that were helpful while creating this app.

* [intro to Go](https://wheelercj.github.io/notes/pages/20221122173910.html)
* [Integrating with Fastmail](https://www.fastmail.com/for-developers/integrating-with-fastmail/)
* [Fastmail's JMAP samples](https://github.com/fastmail/JMAP-Samples/tree/main)
* [JMAP Crash Course](https://jmap.io/crash-course.html)
* [the JMAP specs and RFCs](https://jmap.io/spec.html)
* [OAuth 2 Explained In Simple Terms](https://www.youtube.com/watch?v=ZV5yTm4pT8g)
* [Fastmail OAuth](https://www.fastmail.com/for-developers/oauth/)
