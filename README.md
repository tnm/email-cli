# email-cli

A simple CLI for agents to send emails. Supports Google Workspace, Proton Mail, and generic SMTP.

## Install

### Install Script (requires Go)

```bash
curl -fsSL https://raw.githubusercontent.com/tnm/email-cli/main/email/install.sh | bash
```

Installs `email-cli` into `~/.local/bin` by default. If that directory is not on your `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Or install from a local checkout:

```bash
git clone https://github.com/tnm/email-cli
cd email-cli/email
./install.sh
```

### Go Install

```bash
go install github.com/tnm/email-cli@latest
```

Or build from source:

```bash
git clone https://github.com/tnm/email-cli
cd email-cli/email
go build -o email-cli .
```

### Claude Code Plugin

Install as a Claude Code plugin for AI agent access:

```bash
# Add the marketplace
/plugin marketplace add tnm/email-cli

# Install the plugin
/plugin install email-cli@email-cli

# Or load locally during development
claude --plugin-dir /path/to/email-cli
```

Once installed, Claude automatically uses email-cli when you ask it to send emails.

## Quick Start

```bash
# Add a provider (interactive)
email-cli config add mymail

# Send an email
email-cli send -t recipient@example.com -s "Hello" -m "Message body"
```

---

## Configuration

Config is stored at `~/.config/email-cli/config.json`

## Security

`email-cli` follows a pragmatic local-file security model:

- Secrets are stored in plaintext in `~/.config/email-cli/config.json`.
- File permissions are restricted to owner-only (`0600`) when writing config.
- `email-cli config show` redacts secrets by default.
- `email-cli config show --show-secrets` prints raw secrets and should be treated as sensitive.

Operational guidance:

- Avoid passing secrets directly on command lines when possible (`--password`, `--access-token`, `--refresh-token`) because shell history and process inspection may expose them.
- Prefer app-specific passwords for SMTP providers (for Gmail, use an App Password).
- Rotate credentials immediately if a machine, shell history, or agent transcript is exposed.
- For install safety, prefer reviewing `install.sh` from a pinned version tag before running it.

### Interactive Mode (for humans)

```bash
email-cli config add mymail
```

Walks you through provider selection and credential entry.

### Non-Interactive Mode (for scripts/agents)

#### AgentMail (easiest - just API key)

```bash
email-cli config add --name agent \
  --type agentmail \
  --api-key "$AGENTMAIL_API_KEY" \
  --inbox-id "$AGENTMAIL_INBOX_ID" \
  --default
```

No OAuth, no app passwords, no local servers. Just an API key from [agentmail.to](https://agentmail.to). Free tier: 3 inboxes, 3k emails/month.

#### Recommended for Gmail (lowest friction): SMTP + App Password

```bash
email-cli config add --name gmail-smtp \
  --type smtp \
  --from me@gmail.com \
  --host smtp.gmail.com \
  --port 587 \
  --username me@gmail.com \
  --password "$GMAIL_APP_PASSWORD" \
  --tls \
  --default
```

Use this when you want the fastest setup for humans and agents. It avoids Google Cloud OAuth client setup.

#### SMTP

```bash
email-cli config add --name mymail \
  --type smtp \
  --from me@example.com \
  --host smtp.example.com \
  --port 587 \
  --username me@example.com \
  --password "secret" \
  --tls
```

#### Proton Mail

```bash
email-cli config add --name proton \
  --type proton \
  --from me@proton.me \
  --username me@proton.me \
  --password "bridge-password"
```

#### Google Workspace (Gmail API)

```bash
email-cli config add --name google \
  --type google \
  --from me@gmail.com \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "xxx"
```

By default this starts OAuth device flow and prints a verification URL/code in your terminal.
To force localhost callback flow instead:

```bash
email-cli config add --name google \
  --type google \
  --from me@gmail.com \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "xxx" \
  --oauth-method local
```

### Config Commands

```bash
# List providers
email-cli config list

# Show config as JSON (useful for agents)
email-cli config show                     # full config (secrets redacted)
email-cli config show mymail              # specific provider (secrets redacted)
email-cli config show --show-secrets      # include secrets/tokens
email-cli config show mymail --show-secrets

# Set individual values
email-cli config set mymail password "new-password"
email-cli config set mymail host smtp.newserver.com

# Set default provider
email-cli config default mymail

# Remove provider
email-cli config remove mymail

# Show config file path
email-cli config path
```

### Config Set Keys

| Provider | Available Keys |
|----------|---------------|
| AgentMail | `api-key`, `inbox-id` |
| SMTP | `from`, `host`, `port`, `username`, `password`, `tls` |
| Proton | `from`, `host`, `port`, `username`, `password` |
| Google | `from`, `client-id`, `client-secret`, `access-token`, `refresh-token` |

---

## Sending Email

### Basic Usage

```bash
email-cli send -t user@example.com -s "Subject" -m "Body"
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--to` | `-t` | Recipient(s) - required, repeatable |
| `--subject` | `-s` | Subject line - required |
| `--body` | `-m` | Message body |
| `--cc` | `-c` | CC recipient(s) |
| `--bcc` | `-b` | BCC recipient(s) |
| `--attach` | `-a` | File attachment(s) |
| `--html` | | Treat body as HTML |
| `--provider` | `-p` | Use specific provider |

### Examples

```bash
# Multiple recipients
email-cli send -t a@x.com -t b@x.com -s "Team Update" -m "Hello team"

# With CC/BCC
email-cli send -t user@x.com -c cc@x.com -b bcc@x.com -s "Subject" -m "Body"

# HTML email
email-cli send -t user@x.com -s "Newsletter" -m "<h1>Hello</h1><p>World</p>" --html

# Attachments
email-cli send -t user@x.com -s "Report" -m "See attached" -a report.pdf -a data.csv

# Read body from stdin
cat message.txt | email-cli send -t user@x.com -s "From file"
echo "Quick message" | email-cli send -t user@x.com -s "Piped"

# Use specific provider
email-cli send -p work -t user@x.com -s "Subject" -m "Body"
```

---

## Provider Setup

### AgentMail (Easiest)

[AgentMail](https://agentmail.to) is email infrastructure designed for AI agents. No OAuth, no app passwords, no local servers — just an API key.

1. Sign up at [agentmail.to](https://agentmail.to)
2. Create an API key in the console
3. Create an inbox (or use the API to create one)
4. Configure:

```bash
email-cli config add --name agent \
  --type agentmail \
  --api-key "am_..." \
  --inbox-id "inbox_..." \
  --default
```

**Free tier:** 3 inboxes, 3,000 emails/month, no credit card required.

**Note:** You'll send from `@agentmail.to` addresses unless you add a custom domain (paid plans).

### Gmail (Recommended for Lowest Friction)

Use Gmail over SMTP with an App Password.

1. Turn on Google 2-Step Verification.
2. Create an App Password at [Google App Passwords](https://myaccount.google.com/apppasswords).
3. Configure `email-cli`:

```bash
email-cli config add --name gmail-smtp \
  --type smtp \
  --from me@gmail.com \
  --host smtp.gmail.com \
  --port 587 \
  --username me@gmail.com \
  --password "YOUR_APP_PASSWORD" \
  --tls \
  --default
```

Notes:
- This is the simplest path for agent-driven setup.
- Agents cannot create App Passwords for users; the user must do that account-security step once.

### Google Workspace (Gmail API)

Uses Gmail API with OAuth2.

`email-cli` defaults to OAuth device flow for Google setup, so you do not need a localhost callback server in the common case.
This path requires Google Cloud OAuth project setup.

1. Create a project in [Google Cloud Console](https://console.cloud.google.com/)
2. Enable the Gmail API
3. Create OAuth2 credentials (Desktop app type)
4. Run `email-cli config add google` and complete device auth in terminal
5. Optional for local callback mode: add `http://127.0.0.1:8089/callback` to authorized redirect URIs and run with `--oauth-method local`

If you already have tokens, you can set them directly:

```bash
email-cli config add --name google \
  --type google \
  --from me@gmail.com \
  --client-id "YOUR_CLIENT_ID" \
  --client-secret "YOUR_CLIENT_SECRET" \
  --access-token "YOUR_ACCESS_TOKEN" \
  --refresh-token "YOUR_REFRESH_TOKEN"
```

### Proton Mail

Uses [Proton Mail Bridge](https://proton.me/mail/bridge) which runs a local SMTP server.

1. Install and run Proton Mail Bridge
2. Get the bridge password from Bridge settings (not your account password)
3. Configure:

```bash
# Interactive
email-cli config add proton

# Non-interactive
email-cli config add --name proton \
  --type proton \
  --from me@proton.me \
  --username me@proton.me \
  --password "BRIDGE_PASSWORD"
```

Default bridge address: `127.0.0.1:1025`

### Generic SMTP

Works with any SMTP server: SendGrid, Mailgun, Fastmail, AWS SES, etc.

```bash
# Interactive
email-cli config add smtp

# Non-interactive (SendGrid example)
email-cli config add --name sendgrid \
  --type smtp \
  --from me@example.com \
  --host smtp.sendgrid.net \
  --port 587 \
  --username apikey \
  --password "SG.xxxx" \
  --tls
```

---

## For AI Agents

This CLI is designed to be easily used by AI agents and automation.

### Recommended Setup Order

Use this order when helping users configure email:
1. **AgentMail first** — easiest, just needs API key + inbox ID
2. For existing email accounts, try SMTP
3. For Gmail, prefer `smtp.gmail.com` + App Password
4. Use Google API only if the user explicitly wants Gmail API/OAuth features
5. Use Proton Bridge only when the user is on Proton Mail

### Agent Setup Playbook

1. Check for an existing provider:
```bash
email-cli config list
```
2. If no provider exists, recommend AgentMail for simplest setup:
```bash
email-cli config add --name agent \
  --type agentmail \
  --api-key "$AGENTMAIL_API_KEY" \
  --inbox-id "$AGENTMAIL_INBOX_ID" \
  --default
```
3. Or for existing email, collect SMTP details and configure:
```bash
email-cli config add --name agent-mail \
  --type smtp \
  --from "$EMAIL_FROM" \
  --host "$SMTP_HOST" \
  --port "$SMTP_PORT" \
  --username "$SMTP_USER" \
  --password "$SMTP_PASS" \
  --default
```
4. Send:
```bash
email-cli send -t "$TO" -s "$SUBJECT" -m "$BODY"
```

### Claude Code Plugin

The easiest way to use email-cli with Claude Code:

```bash
/plugin marketplace add tnm/email-cli
/plugin install email-cli@email-cli
```

Claude will automatically invoke the email skill when you ask it to send emails. The skill includes full documentation so Claude knows all available options.

### Simple Interface

```bash
# Predictable, scriptable
email-cli send -t "$TO" -s "$SUBJECT" -m "$BODY"

# Exit codes: 0 = success, non-zero = failure
email-cli send -t user@x.com -s "Test" -m "Hello" && echo "Sent!"
```

### Pipe Content

```bash
# Pipe generated content
echo "$GENERATED_REPORT" | email-cli send -t user@x.com -s "Daily Report"

# From files
cat analysis.txt | email-cli send -t user@x.com -s "Analysis Results"
```

### Non-Interactive Config

```bash
# Set up without prompts
email-cli config add --name agent-mail \
  --type smtp \
  --from agent@example.com \
  --host smtp.example.com \
  --port 587 \
  --username agent@example.com \
  --password "$SMTP_PASSWORD" \
  --default

# Update credentials programmatically
email-cli config set agent-mail password "$NEW_PASSWORD"
```

### JSON Output

```bash
# Get config as JSON for parsing (secrets are redacted by default)
email-cli config show
email-cli config show agent-mail

# Include secrets/tokens only when needed
email-cli config show agent-mail --show-secrets
```

### Environment Variables

You can use environment variables in scripts:

```bash
email-cli config add --name mymail \
  --type smtp \
  --from "$EMAIL_FROM" \
  --host "$SMTP_HOST" \
  --port "$SMTP_PORT" \
  --username "$SMTP_USER" \
  --password "$SMTP_PASS"
```

---

## Config File Format

`~/.config/email-cli/config.json`:

```json
{
  "default_provider": "work",
  "providers": {
    "work": {
      "type": "google",
      "name": "work",
      "from": "me@company.com",
      "google": {
        "client_id": "...",
        "client_secret": "...",
        "access_token": "...",
        "refresh_token": "...",
        "token_expiry": "2024-01-01T00:00:00Z"
      }
    },
    "personal": {
      "type": "smtp",
      "name": "personal",
      "from": "me@fastmail.com",
      "smtp": {
        "host": "smtp.fastmail.com",
        "port": 587,
        "username": "me@fastmail.com",
        "password": "app-password",
        "use_tls": true
      }
    }
  }
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (config, auth, send failure, etc.) |

---

## License

MIT
