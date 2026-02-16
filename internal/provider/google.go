package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tnm/email-cli/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Google struct {
	from    string
	config  *config.GoogleConfig
	service *gmail.Service
}

func NewGoogle(from string, cfg *config.GoogleConfig) (*Google, error) {
	g := &Google{
		from:   from,
		config: cfg,
	}

	if err := g.initService(); err != nil {
		return nil, err
	}

	return g, nil
}

func googleOAuthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailSendScope},
		RedirectURL:  googleRedirectURL,
	}
}

func (g *Google) initService() error {
	oauth2Config := googleOAuthConfig(g.config.ClientID, g.config.ClientSecret)

	token := &oauth2.Token{
		AccessToken:  g.config.AccessToken,
		RefreshToken: g.config.RefreshToken,
	}

	if g.config.TokenExpiry != "" {
		if expiry, err := time.Parse(time.RFC3339, g.config.TokenExpiry); err == nil {
			token.Expiry = expiry
		}
	}

	client := oauth2Config.Client(context.Background(), token)

	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	g.service = service
	return nil
}

func (g *Google) Name() string {
	return "google"
}

func (g *Google) Send(email *Email) error {
	if len(email.Bcc) == 0 {
		return g.sendSingle(email)
	}

	visible := &Email{
		To:          append([]string(nil), email.To...),
		Cc:          append([]string(nil), email.Cc...),
		Subject:     email.Subject,
		Body:        email.Body,
		HTML:        email.HTML,
		Attachments: email.Attachments,
	}

	if len(visible.To)+len(visible.Cc) > 0 {
		if err := g.sendSingle(visible); err != nil {
			return err
		}
	}

	for _, bccRecipient := range sanitizeAddressList(email.Bcc) {
		private := &Email{
			To:          []string{bccRecipient},
			Subject:     email.Subject,
			Body:        email.Body,
			HTML:        email.HTML,
			Attachments: email.Attachments,
		}
		if err := g.sendSingle(private); err != nil {
			return err
		}
	}

	return nil
}

func (g *Google) sendSingle(email *Email) error {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", sanitizeHeaderValue(g.from)))

	to := sanitizeAddressList(email.To)
	if len(to) > 0 {
		msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	}

	cc := sanitizeAddressList(email.Cc)
	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}

	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", sanitizeHeaderValue(email.Subject)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(email.Attachments) > 0 {
		return g.sendWithAttachments(email, &msg)
	}

	contentType := "text/plain"
	if email.HTML {
		contentType = "text/html"
	}
	msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\r\n", contentType))
	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	return g.sendRaw(msg.String())
}

func (g *Google) sendWithAttachments(email *Email, headerBuilder *strings.Builder) error {
	var buf strings.Builder
	writer := multipart.NewWriter(&buf)
	boundary := writer.Boundary()

	headerBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	headerBuilder.WriteString("\r\n")

	buf.WriteString(headerBuilder.String())

	contentType := "text/plain"
	if email.HTML {
		contentType = "text/html"
	}

	bodyHeader := make(textproto.MIMEHeader)
	bodyHeader.Set("Content-Type", fmt.Sprintf("%s; charset=\"UTF-8\"", contentType))
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return err
	}
	if _, err := bodyPart.Write([]byte(email.Body)); err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	for _, att := range email.Attachments {
		content := att.Content
		if content == nil && att.Path != "" {
			data, err := os.ReadFile(att.Path)
			if err != nil {
				return fmt.Errorf("failed to read attachment %s: %w", att.Path, err)
			}
			content = data
		}

		filename := att.Filename
		if filename == "" && att.Path != "" {
			filename = filepath.Base(att.Path)
		}
		filename = sanitizeFilename(filename)

		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		attHeader := make(textproto.MIMEHeader)
		attHeader.Set("Content-Type", mimeType)
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		attPart, err := writer.CreatePart(attHeader)
		if err != nil {
			return err
		}

		encoded := base64.StdEncoding.EncodeToString(content)
		if _, err := attPart.Write([]byte(encoded)); err != nil {
			return fmt.Errorf("failed to write attachment %s: %w", filename, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize mime message: %w", err)
	}

	return g.sendRaw(buf.String())
}

func (g *Google) sendRaw(raw string) error {
	message := &gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(raw)),
	}

	_, err := g.service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

const googleRedirectURL = "http://127.0.0.1:8089/callback"

func GenerateGoogleOAuthState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate oauth state: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func GetGoogleAuthURL(clientID, clientSecret, state string) string {
	oauth2Config := googleOAuthConfig(clientID, clientSecret)
	return oauth2Config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func ExchangeGoogleCode(clientID, clientSecret, code string) (*oauth2.Token, error) {
	oauth2Config := googleOAuthConfig(clientID, clientSecret)

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

func GetGoogleDeviceAuth(clientID, clientSecret string) (*oauth2.DeviceAuthResponse, error) {
	oauth2Config := googleOAuthConfig(clientID, clientSecret)
	auth, err := oauth2Config.DeviceAuth(
		context.Background(),
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start device auth flow: %w", err)
	}
	return auth, nil
}

func ExchangeGoogleDeviceAuth(clientID, clientSecret string, auth *oauth2.DeviceAuthResponse) (*oauth2.Token, error) {
	oauth2Config := googleOAuthConfig(clientID, clientSecret)
	token, err := oauth2Config.DeviceAccessToken(
		context.Background(),
		auth,
		oauth2.AccessTypeOffline,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to complete device auth flow: %w", err)
	}
	return token, nil
}

func RunGoogleAuthServer(expectedState string) (string, error) {
	if expectedState == "" {
		return "", fmt.Errorf("expected oauth state is required")
	}

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	sendErr := func(err error) {
		select {
		case errChan <- err:
		default:
		}
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    "127.0.0.1:8089",
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			sendErr(fmt.Errorf("invalid oauth state"))
			http.Error(w, "Error: invalid state", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			sendErr(fmt.Errorf("no code in callback"))
			http.Error(w, "Error: no code received", http.StatusBadRequest)
			return
		}

		select {
		case codeChan <- code:
		default:
		}

		_, _ = fmt.Fprint(w, "Authorization successful! You can close this window.")
		go func() {
			_ = server.Shutdown(context.Background())
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sendErr(err)
		}
	}()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	select {
	case code := <-codeChan:
		return code, nil
	case err := <-errChan:
		_ = server.Shutdown(context.Background())
		return "", err
	case <-timeout.C:
		_ = server.Shutdown(context.Background())
		return "", fmt.Errorf("timed out waiting for oauth callback")
	}
}

