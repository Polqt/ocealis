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

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/internal/stamp"
	"github.com/gofiber/fiber/v3"
)

type stampRecordingSvc struct {
	castRecordingSvc
	last service.StampBottleInput
	got  bool
}

func (f *stampRecordingSvc) StampBottle(_ context.Context, in service.StampBottleInput) (*domain.Journey, error) {
	prepared, err := stamp.Prepare(in.SealIcon, in.Note)
	if err != nil {
		return nil, err
	}
	f.got = true
	f.last = in
	return &domain.Journey{
		Bottle: &domain.Bottle{ID: in.BottleID, Status: domain.BottleStatusDrifting, IsReleased: true},
		Events: []domain.BottleEvent{{
			BottleID:  in.BottleID,
			EventType: domain.EventTypeStamp,
			SealIcon:  prepared.SealIcon,
			Note:      prepared.Note,
		}},
	}, nil
}

func stampRequest(t *testing.T, app *fiber.App, body map[string]any) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bottles/7/stamp", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestStampRejectsInvalidTurnstile(t *testing.T) {
	svc := &stampRecordingSvc{}
	app := castApp(t, captchaStub{ok: false}, svc)

	resp := stampRequest(t, app, map[string]any{
		"seal_icon":       "anchor",
		"turnstile_token": "bad",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 403, got %d body=%s", resp.StatusCode, body)
	}
	if svc.got {
		t.Fatal("Stamp must not proceed after Turnstile failure")
	}
}

func TestStampRejectsOverLimitNote(t *testing.T) {
	svc := &stampRecordingSvc{}
	app := castApp(t, captchaStub{ok: true}, svc)

	resp := stampRequest(t, app, map[string]any{
		"note":            strings.Repeat("潮", 81),
		"turnstile_token": "ok",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 422, got %d body=%s", resp.StatusCode, body)
	}
}

func TestStampAcceptsSealOrNoteAndIsRateLimited(t *testing.T) {
	svc := &stampRecordingSvc{}
	app := castApp(t, captchaStub{ok: true}, svc)
	body := map[string]any{
		"note":            "safe passage",
		"turnstile_token": "ok",
	}

	for attempt := 1; attempt <= 4; attempt++ {
		resp := stampRequest(t, app, body)
		if attempt <= 3 && resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("attempt %d: want 200, got %d body=%s", attempt, resp.StatusCode, data)
		}
		if attempt == 4 && resp.StatusCode != http.StatusTooManyRequests {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("attempt 4: want 429, got %d body=%s", resp.StatusCode, data)
		}
		resp.Body.Close()
	}

	if !svc.got || svc.last.BottleID != 7 || svc.last.Note != "safe passage" {
		t.Fatalf("Stamp did not reach service: %+v", svc.last)
	}
}
