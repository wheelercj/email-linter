# email linter

Easily find spam and phishing emails received at single-use email addresses. This command-line tool currently only works with Fastmail and the email protection services run by DuckDuckGo, Firefox, and Apple.

![demo](demo.png)

The Fastmail API token needed can be entered when you run the app or by creating an environment variable named `JMAP_TOKEN`, such as with the PowerShell command `$env:JMAP_TOKEN="your token here"`.

## what are single-use email addresses?

Email address protection services make it easy to use a unique email address for each online account, increasing security and reducing spam. Some examples of these services are [DuckDuckGo's Email Protection](https://duckduckgo.com/email), [Firefox Relay](https://relay.firefox.com/), and [iCloud+'s Hide My Email](https://support.apple.com/en-us/105078). Since the emails received by these addresses _should_ have predictable "from" fields, suspicious senders can be easily found.

## how does it work?

1. Find all emails in the inbox that went through an email protection service.
2. Find all emails outside the spam folder those single-use addresses have ever received.
3. List each single-use address and the addresses they have received from. With these, you can search your inbox for suspicious emails and decide what to do with them.

## dev resources

Here are some resources that were helpful while creating this application.

* [intro to Go](https://wheelercj.github.io/notes/pages/20221122173910.html)
* [Integrating with Fastmail](https://www.fastmail.com/for-developers/integrating-with-fastmail/)
* [Fastmail's JMAP samples](https://github.com/fastmail/JMAP-Samples/tree/main)
* [JMAP Crash Course](https://jmap.io/crash-course.html)
* [the JMAP specs and RFCs](https://jmap.io/spec.html)
