---
name: email
description: Send emails via CLI. Use when the user asks to send an email, compose a message, or needs email functionality. Supports Google Workspace, Proton Mail, and generic SMTP.
---

# Email CLI

Send emails using the `email-cli` command.

## Quick Start

```bash
# Send a simple email
email-cli send -t recipient@example.com -s "Subject" -m "Message body"

# Send with attachment
email-cli send -t user@example.com -s "Report" -m "See attached" -a report.pdf

# Pipe content from stdin
echo "Generated content" | email-cli send -t user@example.com -s "Report"
```

## Common Patterns

### Send to multiple recipients
```bash
email-cli send -t a@x.com -t b@x.com -s "Team Update" -m "Hello team"
```

### HTML email
```bash
email-cli send -t user@x.com -s "Newsletter" -m "<h1>Hello</h1>" --html
```

### Use specific provider
```bash
email-cli send -p work -t user@x.com -s "Subject" -m "Body"
```

## Configuration

Before sending, a provider must be configured. Check existing config:

```bash
email-cli config list
```

### Recommended decision order (for agents)

1. Prefer SMTP setup first.
2. For Gmail, prefer `smtp.gmail.com` with an App Password.
3. Only use Google API OAuth if the user explicitly asks for Gmail API/OAuth.
4. Use Proton only for Proton Mail Bridge users.

### Add provider (non-interactive, for agents)

**Gmail (lowest friction):**
```bash
email-cli config add gmail-smtp \
  --type smtp \
  --from me@gmail.com \
  --host smtp.gmail.com \
  --port 587 \
  --username me@gmail.com \
  --password "$GMAIL_APP_PASSWORD" \
  --tls \
  --default
```

If user does not have an app password yet:
- Ask them to enable 2-Step Verification.
- Ask them to generate an app password at `https://myaccount.google.com/apppasswords`.
- Then run the command above.

**SMTP:**
```bash
email-cli config add mymail \
  --type smtp \
  --from me@example.com \
  --host smtp.example.com \
  --port 587 \
  --username me@example.com \
  --password "password" \
  --tls
```

**Proton Mail:**
```bash
email-cli config add proton \
  --type proton \
  --from me@proton.me \
  --username me@proton.me \
  --password "bridge-password"
```

**Google API (only when explicitly requested):**
```bash
email-cli config add google \
  --type google \
  --from me@gmail.com \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "xxx"
```

### Update config
```bash
email-cli config set mymail password "new-password"
```

### View config as JSON
```bash
email-cli config show                     # redacted by default
email-cli config show mymail              # redacted by default
email-cli config show --show-secrets
email-cli config show mymail --show-secrets
```

## Exit Codes

- `0` = Success
- `1` = Error (check stderr for details)

## Reference

See `reference.md` for complete API documentation.
