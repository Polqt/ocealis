package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/service"
)

type reReleaseRecordingSvc struct {
	last service.ReReleaseBottleInput
	got  bool
}

func (s *reReleaseRecordingSvc) CreateBottle(context.Context, service.CreateBottleInput) (*domain.Bottle, error) {
	return nil, nil
}
func (s *reReleaseRecordingSvc) GetBottle(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (s *reReleaseRecordingSvc) GetJourney(context.Context, int32) (*domain.Journey, error) {
	return nil, nil
}
func (s *reReleaseRecordingSvc) DiscoverBottle(context.Context, service.DiscoverBottleInput) (*domain.Journey, error) {
	return nil, nil
}
func (s *reReleaseRecordingSvc) ReReleaseBottle(_ context.Context, input service.ReReleaseBottleInput) (*domain.Bottle, error) {
	s.got = true
	s.last = input
	return &domain.Bottle{
		ID:          input.BottleID,
		Nickname:    input.Nickname,
		MessageText: "original Message",
		Status:      domain.BottleStatusMysteryDelay,
		VisibleAt:   time.Now().Add(20 * time.Minute),
		IsReleased:  false,
	}, nil
}

func TestReReleaseRejectsInvalidTurnstileAndMissingNickname(t *testing.T) {
	svc := &reReleaseRecordingSvc{}
	app := castApp(t, captchaStub{ok: false}, svc)

	resp, err := app.Test(httptest.NewRequest(
		http.MethodPost,
		"/api/v1/bottles/7/release",
		bytes.NewBufferString(`{"nickname":"finder","lat":39.0997,"lng":-94.5786,"turnstile_token":"bad"}`),
	))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("invalid Turnstile want 403, got %d body=%s", resp.StatusCode, body)
	}
	if svc.got {
		t.Fatal("Re-release proceeded after Turnstile failure")
	}

	svc = &reReleaseRecordingSvc{}
	app = castApp(t, captchaStub{ok: true}, svc)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/bottles/7/release",
		bytes.NewBufferString(`{"lat":39.0997,"lng":-94.5786,"turnstile_token":"ok"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("missing Nickname want 422, got %d body=%s", resp.StatusCode, body)
	}
}

func TestReReleaseAcceptsNicknameAndLocation(t *testing.T) {
	svc := &reReleaseRecordingSvc{}
	app := castApp(t, captchaStub{ok: true}, svc)
	body, _ := json.Marshal(map[string]any{
		"nickname":        "finder",
		"lat":             39.0997,
		"lng":             -94.5786,
		"turnstile_token": "ok",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles/7/release", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 200, got %d body=%s", resp.StatusCode, b)
	}
	if !svc.got || svc.last.BottleID != 7 || svc.last.Nickname != "finder" {
		t.Fatalf("Re-release input not passed to service: %+v", svc.last)
	}
	if svc.last.Lat == nil || svc.last.Lng == nil {
		t.Fatalf("finder location missing from service input: %+v", svc.last)
	}
}

func TestReReleaseIsIPRateLimited(t *testing.T) {
	svc := &reReleaseRecordingSvc{}
	app := castApp(t, middleware.AcceptTurnstile{}, svc)

	for i := 1; i <= 4; i++ {
		body := bytes.NewBufferString(`{"nickname":"finder","lat":30,"lng":-140,"turnstile_token":"ok"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles/7/release", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if i <= 3 && resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d want 200, got %d", i, resp.StatusCode)
		}
		if i == 4 && resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("fourth request want 429, got %d", resp.StatusCode)
		}
	}
}
