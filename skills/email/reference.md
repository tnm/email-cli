# email-cli Reference

## Commands

### send

Send an email.

```bash
email-cli send [flags]
```

**Required flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--to` | `-t` | Recipient email (repeatable) |
| `--subject` | `-s` | Email subject |

**Optional flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--body` | `-m` | Message body (or pipe via stdin) |
| `--cc` | `-c` | CC recipient (repeatable) |
| `--bcc` | `-b` | BCC recipient (repeatable) |
| `--attach` | `-a` | File attachment (repeatable) |
| `--html` | | Treat body as HTML |
| `--provider` | `-p` | Use specific provider |

**Examples:**
```bash
# Basic
email-cli send -t user@example.com -s "Hello" -m "Body"

# Multiple recipients with CC
email-cli send -t a@example.com -t b@example.com -c cc@example.com -s "Team" -m "Hi"

# Attachment
email-cli send -t user@example.com -s "Report" -m "Attached" -a file.pdf

# Pipe body
echo "Content" | email-cli send -t user@example.com -s "Subject"

# HTML
email-cli send -t user@example.com -s "News" -m "<h1>Title</h1>" --html
```

---

### config add

Add a new provider.

```bash
email-cli config add --name <name> [flags]
```

**Interactive mode** (no flags): Prompts for all values.

**Non-interactive mode** (with `--type`):

| Flag | Description |
|------|-------------|
| `--name` | Provider name |
| `--type` | Provider type: `agentmail`, `smtp`, `proton`, `google` |
| `--api-key` | AgentMail API key |
| `--inbox-id` | AgentMail inbox ID (email address) |
| `--from` | From email address (SMTP/Proton/Google) |
| `--host` | SMTP host |
| `--port` | SMTP port (default: 587) |
| `--username` | Auth username |
| `--password` | Auth password |
| `--tls` | Use TLS (default: true) |
| `--client-id` | Google OAuth client ID |
| `--client-secret` | Google OAuth client secret |
| `--access-token` | Google OAuth access token |
| `--refresh-token` | Google OAuth refresh token |
| `--oauth-method` | Google OAuth method: `device` (default) or `local` |
| `--default` | Set as default provider |
| `--use-keychain` | Store secrets in macOS Keychain (macOS only) |

**Examples:**
```bash
# AgentMail (easiest)
email-cli config add --name agent \
  --type agentmail \
  --api-key "am_..." \
  --inbox-id "myagent@agentmail.to" \
  --default

# Gmail (recommended for existing accounts)
email-cli config add --name gmail-smtp \
  --type smtp \
  --from me@gmail.com \
  --host smtp.gmail.com \
  --port 587 \
  --username me@gmail.com \
  --password "$GMAIL_APP_PASSWORD" \
  --tls \
  --default

# SMTP
email-cli config add --name work \
  --type smtp \
  --from me@company.com \
  --host smtp.company.com \
  --port 587 \
  --username me \
  --password "secret" \
  --default

# Proton
email-cli config add --name proton \
  --type proton \
  --from me@proton.me \
  --username me@proton.me \
  --password "bridge-pass"

# Google
email-cli config add --name gmail \
  --type google \
  --from me@gmail.com \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "xxx"

# With Keychain (macOS only)
email-cli config add --name agent \
  --type agentmail \
  --api-key "am_..." \
  --inbox-id "myagent@agentmail.to" \
  --use-keychain
```

---

### config list

List configured providers.

```bash
email-cli config list
```

---

### config show

Show config as JSON.

```bash
email-cli config show                     # Full config (redacted)
email-cli config show <name>              # Specific provider (redacted)
email-cli config show --show-secrets      # Full config with secrets
email-cli config show --show-secrets <name>  # Specific provider with secrets
```

---

### config set

Update a config value.

```bash
email-cli config set <name> <key> <value>
```

**Keys by provider type:**

| Provider | Keys |
|----------|------|
| AgentMail | `api-key`, `inbox-id` |
| SMTP | `from`, `host`, `port`, `username`, `password`, `tls` |
| Proton | `from`, `host`, `port`, `username`, `password` |
| Google | `from`, `client-id`, `client-secret`, `access-token`, `refresh-token` |

**Flags:**
| Flag | Description |
|------|-------------|
| `--use-keychain` | Store secret in macOS Keychain (macOS only) |

**Examples:**
```bash
email-cli config set work password "new-pass"
email-cli config set work host smtp.newserver.com
email-cli config set agent api-key "am_newkey..."

# Store in Keychain (macOS)
email-cli config set --use-keychain work password "new-pass"
```

---

### config default

Set default provider.

```bash
email-cli config default <name>
```

---

### config remove

Remove a provider.

```bash
email-cli config remove <name>
```

---

### config path

Show config file path.

```bash
email-cli config path
# Output: ~/.config/email-cli/config.json
```

---

## Config File Format

Located at `~/.config/email-cli/config.json`:

```json
{
  "default_provider": "agent",
  "providers": {
    "agent": {
      "type": "agentmail",
      "name": "agent",
      "agentmail": {
        "api_key": "am_...",
        "inbox_id": "myagent@agentmail.to"
      }
    },
    "work": {
      "type": "smtp",
      "name": "work",
      "from": "me@company.com",
      "smtp": {
        "host": "smtp.company.com",
        "port": 587,
        "username": "me",
        "password": "secret",
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
| 1 | Error |

---

## Provider Defaults

| Provider | Default Host | Default Port |
|----------|--------------|--------------|
| AgentMail | api.agentmail.to | N/A (REST API) |
| Proton | 127.0.0.1 | 1025 |
| SMTP | (required) | 587 |
| Google | Gmail API | N/A |
