package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Polqt/ocealis/api"
	"github.com/Polqt/ocealis/api/handler"
	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/ws"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type captchaStub struct {
	ok bool
}

func (c captchaStub) Verify(ctx context.Context, token, ip string) error {
	if !c.ok || token == "" || token == "bad" {
		return middleware.ErrTurnstileFailed
	}
	return nil
}

type castRecordingSvc struct {
	last service.CreateBottleInput
	got  bool
}

func (f *castRecordingSvc) CreateBottle(ctx context.Context, in service.CreateBottleInput) (*domain.Bottle, error) {
	f.got = true
	f.last = in
	return &domain.Bottle{
		ID:          1,
		Nickname:    in.Nickname,
		MessageText: in.MessageText,
		Status:      domain.BottleStatusMysteryDelay,
		VisibleAt:   time.Now().Add(20 * time.Minute),
		CreatedAt:   time.Now(),
	}, nil
}
func (f *castRecordingSvc) GetBottle(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (f *castRecordingSvc) GetJourney(context.Context, int32) (*domain.Journey, error) {
	return nil, nil
}
func (f *castRecordingSvc) DiscoverBottle(context.Context, service.DiscoverBottleInput) (*domain.Journey, error) {
	return nil, nil
}
func (f *castRecordingSvc) ReleaseBottle(context.Context, int32, int32, float64, float64) (*domain.Bottle, error) {
	return nil, nil
}

func castApp(t *testing.T, verify middleware.TurnstileVerifier, svc service.BottleService) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "Internal Server Error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})
	api.RegisterRoutes(app, api.Handlers{
		Health:    &handler.HealthHandler{},
		Bottle:    handler.NewBottleHandler(svc, verify),
		Event:     handler.NewEventHandler(nil),
		Discovery: handler.NewDiscoveryHandler(nil),
	}, ws.NewHub(), zap.NewNop())
	return app
}

func TestCastRejectsInvalidTurnstile(t *testing.T) {
	svc := &castRecordingSvc{}
	app := castApp(t, captchaStub{ok: false}, svc)

	body, _ := json.Marshal(map[string]any{
		"nickname":         "sailor",
		"message_text":     "hello",
		"turnstile_token":  "bad",
		"start_lat":        30.0,
		"start_lng":        -140.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 403, got %d body=%s", resp.StatusCode, b)
	}
	if svc.got {
		t.Fatal("Cast must not proceed after Turnstile failure")
	}
}

func TestCastAcceptsValidTurnstileAndReturnsMysteryDelay(t *testing.T) {
	svc := &castRecordingSvc{}
	app := castApp(t, captchaStub{ok: true}, svc)

	body, _ := json.Marshal(map[string]any{
		"nickname":        "sailor",
		"message_text":    "hello ocean",
		"turnstile_token": "ok",
		"start_lat":       39.0997,
		"start_lng":       -94.5786,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 201, got %d body=%s", resp.StatusCode, b)
	}
	if !svc.got || svc.last.Nickname != "sailor" {
		t.Fatalf("Cast did not reach service with nickname: %+v", svc.last)
	}
}

func TestCastRejectsOverLimitViaAPI(t *testing.T) {
	svc := &castRecordingSvc{}
	app := castApp(t, captchaStub{ok: true}, svc)

	body, _ := json.Marshal(map[string]any{
		"nickname":        strings.Repeat("n", 25),
		"message_text":    "hello",
		"turnstile_token": "ok",
		"start_lat":       30.0,
		"start_lng":       -140.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 422, got %d body=%s", resp.StatusCode, b)
	}
}
