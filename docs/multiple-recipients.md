# Multiple recipients

Emails with multiple recipients usually don't say which address is yours. Sometimes there are patterns in the recipient addresses that hint at the answer, and Email Linter looks for some of those, but sometimes there are not. This means there's a chance Email Linter could say someone else's address is yours. If this happens to you but you see a pattern in the recipient addresses that could be used to improve the output, please let me know by [creating a new issue](https://github.com/wheelercj/email-linter/issues/new)!

I've considered letting users enter their email addresses, but I doubt anyone who really puts email protection services to good use would want to be constantly updating the list.

**Example:**

When an email is forwarded to a duck address, all the recipient addresses are changed to include the duck address. For example, let's say these are the email's original recipient addresses, and that one of them is yours:

```
alex@hotmail.com
sue@gmail.com
bob@icloud.com
```

If your duck address the email is forwarded to is `a1b2c3@duck.com`, then DuckDuckGo's email protection service will change them to:

```
alex_at_hotmail.com_a1b2c3@duck.com
sue_at_gmail.com_a1b2c3@duck.com
bob_at_icloud.com_a1b2c3@duck.com
```

Since Email Linter can't tell which is yours, its output of recipient addresses includes only the last part, your duck address `a1b2c3@duck.com`.
