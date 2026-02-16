package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tnm/email-cli/internal/config"
	"github.com/tnm/email-cli/internal/provider"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

func configAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a new provider configuration",
		ArgsUsage: "<name>",
		Description: "Add a new provider configuration.\n\n" +
			"Interactive mode (default):\n" +
			"  email-cli config add mymail\n\n" +
			"Non-interactive mode (for scripts/agents):\n" +
			"  # SMTP\n" +
			"  email-cli config add mymail \\\n" +
			"    --type smtp \\\n" +
			"    --from me@example.com \\\n" +
			"    --host smtp.example.com \\\n" +
			"    --port 587 \\\n" +
			"    --username me@example.com \\\n" +
			"    --password \"secret\" \\\n" +
			"    --tls\n\n" +
			"  # Proton Mail\n" +
			"  email-cli config add proton \\\n" +
			"    --type proton \\\n" +
			"    --from me@proton.me \\\n" +
			"    --username me@proton.me \\\n" +
			"    --password \"bridge-password\"\n\n" +
			"  # Google (device auth by default)\n" +
			"  email-cli config add google \\\n" +
			"    --type google \\\n" +
			"    --from me@gmail.com \\\n" +
			"    --client-id \"xxx.apps.googleusercontent.com\" \\\n" +
			"    --client-secret \"xxx\"\n\n" +
			"  # Google with local callback flow\n" +
			"  email-cli config add google \\\n" +
			"    --type google \\\n" +
			"    --from me@gmail.com \\\n" +
			"    --client-id \"xxx.apps.googleusercontent.com\" \\\n" +
			"    --client-secret \"xxx\" \\\n" +
			"    --oauth-method local",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "type", Usage: "Provider type: smtp, proton, google"},
			&cli.StringFlag{Name: "from", Usage: "From email address"},
			&cli.StringFlag{Name: "host", Usage: "SMTP host / Bridge host"},
			&cli.IntFlag{Name: "port", Usage: "SMTP port / Bridge port"},
			&cli.StringFlag{Name: "username", Usage: "Username"},
			&cli.StringFlag{Name: "password", Usage: "Password"},
			&cli.BoolFlag{Name: "tls", Value: true, Usage: "Use TLS (SMTP)"},
			&cli.StringFlag{Name: "client-id", Usage: "Google OAuth client ID"},
			&cli.StringFlag{Name: "client-secret", Usage: "Google OAuth client secret"},
			&cli.StringFlag{Name: "access-token", Usage: "Google OAuth access token"},
			&cli.StringFlag{Name: "refresh-token", Usage: "Google OAuth refresh token"},
			&cli.StringFlag{Name: "oauth-method", Value: "device", Usage: "Google OAuth method when tokens are not provided: device or local"},
			&cli.BoolFlag{Name: "default", Usage: "Set as default provider"},
		},
		Action: runConfigAdd,
	}
}

func runConfigAdd(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("usage: email-cli config add <name>")
	}
	name := c.Args().First()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Providers[name]; exists {
		return fmt.Errorf("provider %q already exists", name)
	}

	var providerCfg config.ProviderConfig
	providerCfg.Name = name

	if isNonInteractive(c) {
		if err := buildProviderConfigFromFlags(c, &providerCfg); err != nil {
			return err
		}
	} else {
		if err := buildProviderConfigInteractive(&providerCfg); err != nil {
			return err
		}
	}

	cfg.Providers[name] = providerCfg

	// Set as default if first provider or --default flag.
	if cfg.DefaultProvider == "" || c.Bool("default") {
		cfg.DefaultProvider = name
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Provider %q added successfully!\n", name)
	if cfg.DefaultProvider == name {
		fmt.Println("(Set as default provider)")
	}
	return nil
}

func buildProviderConfigFromFlags(c *cli.Context, providerCfg *config.ProviderConfig) error {
	cfgFrom := c.String("from")
	if cfgFrom == "" {
		return fmt.Errorf("--from is required")
	}
	providerCfg.From = cfgFrom

	cfgType := c.String("type")
	switch cfgType {
	case "smtp":
		cfgHost := c.String("host")
		if cfgHost == "" {
			return fmt.Errorf("--host is required for SMTP")
		}
		port := c.Int("port")
		if port == 0 {
			port = 587
		}
		providerCfg.Type = config.ProviderSMTP
		providerCfg.SMTP = &config.SMTPConfig{
			Host:     cfgHost,
			Port:     port,
			Username: c.String("username"),
			Password: c.String("password"),
			UseTLS:   c.Bool("tls"),
		}

	case "proton":
		host := c.String("host")
		if host == "" {
			host = "127.0.0.1"
		}
		port := c.Int("port")
		if port == 0 {
			port = 1025
		}
		providerCfg.Type = config.ProviderProton
		providerCfg.Proton = &config.ProtonConfig{
			Host:     host,
			Port:     port,
			Username: c.String("username"),
			Password: c.String("password"),
		}

	case "google":
		clientID := c.String("client-id")
		clientSecret := c.String("client-secret")
		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("--client-id and --client-secret are required for Google")
		}
		accessToken := c.String("access-token")
		refreshToken := c.String("refresh-token")
		tokenExpiry := ""
		if accessToken == "" || refreshToken == "" {
			var err error
			accessToken, refreshToken, tokenExpiry, err = obtainGoogleTokens(clientID, clientSecret, c.String("oauth-method"))
			if err != nil {
				return err
			}
		}

		providerCfg.Type = config.ProviderGoogle
		providerCfg.Google = &config.GoogleConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenExpiry:  tokenExpiry,
		}

	default:
		return fmt.Errorf("invalid --type: must be smtp, proton, or google")
	}

	return nil
}

func buildProviderConfigInteractive(providerCfg *config.ProviderConfig) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select provider type:")
	fmt.Println("  1. Google Workspace (Gmail API with OAuth2)")
	fmt.Println("  2. Proton Mail (via Bridge)")
	fmt.Println("  3. Generic SMTP")
	fmt.Print("\nChoice [1-3]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		providerCfg.Type = config.ProviderGoogle
		providerCfg.Google = &config.GoogleConfig{}

		providerCfg.From = prompt(reader, "From email address")
		providerCfg.Google.ClientID = prompt(reader, "Client ID")
		providerCfg.Google.ClientSecret = prompt(reader, "Client Secret")

		oauthMethod := promptDefault(reader, "OAuth method (device/local)", "device")
		accessToken, refreshToken, tokenExpiry, err := obtainGoogleTokens(providerCfg.Google.ClientID, providerCfg.Google.ClientSecret, oauthMethod)
		if err != nil {
			return err
		}
		providerCfg.Google.AccessToken = accessToken
		providerCfg.Google.RefreshToken = refreshToken
		providerCfg.Google.TokenExpiry = tokenExpiry

	case "2":
		providerCfg.Type = config.ProviderProton
		providerCfg.Proton = &config.ProtonConfig{}

		providerCfg.From = prompt(reader, "From email address")
		providerCfg.Proton.Host = promptDefault(reader, "Bridge host", "127.0.0.1")
		portStr := promptDefault(reader, "Bridge port", "1025")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		providerCfg.Proton.Port = port
		providerCfg.Proton.Username = prompt(reader, "Username (email)")
		providerCfg.Proton.Password = prompt(reader, "Bridge password")

	case "3":
		providerCfg.Type = config.ProviderSMTP
		providerCfg.SMTP = &config.SMTPConfig{}

		providerCfg.From = prompt(reader, "From email address")
		providerCfg.SMTP.Host = prompt(reader, "SMTP host")
		portStr := promptDefault(reader, "SMTP port", "587")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		providerCfg.SMTP.Port = port
		providerCfg.SMTP.Username = prompt(reader, "Username")
		providerCfg.SMTP.Password = prompt(reader, "Password")

		useTLS := promptDefault(reader, "Use TLS? (y/n)", "y")
		providerCfg.SMTP.UseTLS = strings.ToLower(useTLS) == "y"

	default:
		return fmt.Errorf("invalid choice")
	}

	return nil
}

func obtainGoogleTokens(clientID, clientSecret, oauthMethod string) (string, string, string, error) {
	method := strings.ToLower(strings.TrimSpace(oauthMethod))
	if method == "" {
		method = "device"
	}

	var (
		token *oauth2.Token
	)

	switch method {
	case "device":
		auth, err := provider.GetGoogleDeviceAuth(clientID, clientSecret)
		if err != nil {
			return "", "", "", err
		}

		fmt.Println("\nAuthorize this CLI with Google.")
		if auth.VerificationURIComplete != "" {
			fmt.Printf("Open this URL:\n%s\n", auth.VerificationURIComplete)
		} else {
			fmt.Printf("Open this URL:\n%s\n", auth.VerificationURI)
			fmt.Printf("Enter this code: %s\n", auth.UserCode)
		}
		fmt.Println("Waiting for authorization...")

		token, err = provider.ExchangeGoogleDeviceAuth(clientID, clientSecret, auth)
		if err != nil {
			return "", "", "", err
		}

	case "local":
		state, err := provider.GenerateGoogleOAuthState()
		if err != nil {
			return "", "", "", err
		}

		fmt.Println("\nStarting local server for OAuth callback...")
		fmt.Println("Open this URL in your browser to authorize:")
		fmt.Println(provider.GetGoogleAuthURL(clientID, clientSecret, state))
		fmt.Println("\nWaiting for authorization...")

		code, err := provider.RunGoogleAuthServer(state)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get authorization: %w", err)
		}

		token, err = provider.ExchangeGoogleCode(clientID, clientSecret, code)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get token: %w", err)
		}

	default:
		return "", "", "", fmt.Errorf("invalid oauth method %q: must be device or local", oauthMethod)
	}

	if token.AccessToken == "" {
		return "", "", "", fmt.Errorf("google oauth returned empty access token")
	}

	tokenExpiry := ""
	if !token.Expiry.IsZero() {
		tokenExpiry = token.Expiry.Format(time.RFC3339)
	}

	return token.AccessToken, token.RefreshToken, tokenExpiry, nil
}
