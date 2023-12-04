# email linter

Automatically detect spam and phishing emails received at single-use email addresses.

Email address protection services make it easy to use a unique email address for each online account, increasing security and reducing spam. Some examples of these services are [DuckDuckGo's Email Protection](https://duckduckgo.com/email), [Firefox Relay](https://relay.firefox.com/), and [iCloud+'s Hide My Email](https://support.apple.com/en-us/105078). Since the emails received by these addresses should have predictable "from" fields, suspicious senders can be automatically detected.

This command-line tool currently only works with Fastmail, and automatically carries out these steps:

1. Find all emails in the inbox that went through an email protection service.
2. Compare their "from" fields to those of all other emails outside the spam folder with matching single-use addresses.
3. If any single-use address has emails with different senders, the relevant email addresses are listed. With these, you can search your inbox for the suspicious emails and decide what to do with them.

## dev resources

Here are some resources that were helpful while creating this application.

* [Integrating with Fastmail](https://www.fastmail.com/for-developers/integrating-with-fastmail/)
* [Fastmail's JMAP samples](https://github.com/fastmail/JMAP-Samples/tree/main)
* [JMAP Crash Course](https://jmap.io/crash-course.html)
* [the JMAP specs and RFCs](https://jmap.io/spec.html)
