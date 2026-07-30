package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Polqt/ocealis/api/middleware"
	"github.com/gofiber/fiber/v3"
)

func stampRequest(t *testing.T, app *fiber.App, body map[string]any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles/7/stamp", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestStampAcceptsSealOrNoteAndAppendsJourneyEvent(t *testing.T) {
	tests := []struct {
		name string
		seal string
		note string
	}{
		{name: "seal only", seal: "⚓"},
		{name: "note only", note: "Fair winds"},
		{name: "seal and note", seal: "🌊", note: "Still drifting"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &castRecordingSvc{}
			app := castApp(t, captchaStub{ok: true}, svc)
			resp := stampRequest(t, app, map[string]any{
				"seal_icon":       tt.seal,
				"note":            tt.note,
				"turnstile_token": "ok",
			})
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("want 201, got %d body=%s", resp.StatusCode, body)
			}
			if !svc.stampGot || svc.stampLast.BottleID != 7 ||
				svc.stampLast.SealIcon != tt.seal || svc.stampLast.Note != tt.note {
				t.Fatalf("Stamp did not reach service: %+v", svc.stampLast)
			}
		})
	}
}

func TestStampRejectsMissingOrInvalidTurnstile(t *testing.T) {
	for _, token := range []string{"", "bad"} {
		t.Run(token, func(t *testing.T) {
			svc := &castRecordingSvc{}
			app := castApp(t, captchaStub{ok: false}, svc)
			resp := stampRequest(t, app, map[string]any{
				"seal_icon":       "⚓",
				"turnstile_token": token,
			})
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("want 403, got %d body=%s", resp.StatusCode, body)
			}
			if svc.stampGot {
				t.Fatal("Stamp must not proceed after Turnstile failure")
			}
		})
	}
}

func TestStampRejectsEmptyAndOverLimitNote(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
	}{
		{name: "empty", body: map[string]any{"turnstile_token": "ok"}},
		{name: "over limit", body: map[string]any{
			"note":            strings.Repeat("n", 81),
			"turnstile_token": "ok",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &castRecordingSvc{}
			app := castApp(t, captchaStub{ok: true}, svc)
			resp := stampRequest(t, app, tt.body)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnprocessableEntity {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("want 422, got %d body=%s", resp.StatusCode, body)
			}
			if svc.stampGot {
				t.Fatal("invalid Stamp must not reach service")
			}
		})
	}
}

func TestStampRateLimitIsPerIP(t *testing.T) {
	svc := &castRecordingSvc{}
	app := castApp(t, middleware.AcceptTurnstile{}, svc)

	for i := 1; i <= 4; i++ {
		resp := stampRequest(t, app, map[string]any{
			"seal_icon":       "⚓",
			"turnstile_token": "ok",
		})
		resp.Body.Close()
		want := http.StatusCreated
		if i == 4 {
			want = http.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Fatalf("request %d: want %d, got %d", i, want, resp.StatusCode)
		}
	}
}
