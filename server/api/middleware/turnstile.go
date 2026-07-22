package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrTurnstileFailed = errors.New("turnstile verification failed")

// TurnstileVerifier checks a Cloudflare Turnstile token at the abuse seam.
type TurnstileVerifier interface {
	Verify(ctx context.Context, token, ip string) error
}

// Turnstile is the production Cloudflare siteverify client.
type Turnstile struct {
	Secret string
	Client *http.Client
}

type siteverifyResponse struct {
	Success bool `json:"success"`
}

func (t *Turnstile) Verify(ctx context.Context, token, ip string) error {
	if strings.TrimSpace(token) == "" {
		return ErrTurnstileFailed
	}
	if t.Secret == "" {
		// ponytail: empty secret = local/dev accept any non-empty token
		return nil
	}
	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	form := url.Values{}
	form.Set("secret", t.Secret)
	form.Set("response", token)
	if ip != "" {
		form.Set("remoteip", ip)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()))
	if err != nil {
		return ErrTurnstileFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return ErrTurnstileFailed
	}
	defer resp.Body.Close()
	var out siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || !out.Success {
		return ErrTurnstileFailed
	}
	return nil
}

// AcceptTurnstile accepts any non-empty token — tests / explicit local stub.
type AcceptTurnstile struct{}

func (AcceptTurnstile) Verify(_ context.Context, token, _ string) error {
	if strings.TrimSpace(token) == "" {
		return ErrTurnstileFailed
	}
	return nil
}
