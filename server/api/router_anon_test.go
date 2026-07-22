package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Polqt/ocealis/api"
	"github.com/Polqt/ocealis/api/handler"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/ws"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// fakeBottleSvc stubs BottleService for HTTP seam tests.
type fakeBottleSvc struct {
	bottle *domain.Bottle
}

func (f *fakeBottleSvc) CreateBottle(context.Context, service.CreateBottleInput) (*domain.Bottle, error) {
	return nil, nil
}
func (f *fakeBottleSvc) GetBottle(context.Context, int32) (*domain.Bottle, error) {
	return f.bottle, nil
}
func (f *fakeBottleSvc) GetJourney(context.Context, int32) (*domain.Journey, error) {
	return &domain.Journey{Bottle: f.bottle, Events: nil}, nil
}
func (f *fakeBottleSvc) DiscoverBottle(context.Context, service.DiscoverBottleInput) (*domain.Journey, error) {
	return nil, nil
}
func (f *fakeBottleSvc) ReleaseBottle(context.Context, int32, int32, float64, float64) (*domain.Bottle, error) {
	return nil, nil
}

func TestAnonymousVisitorCanGetBottleWithoutAuth(t *testing.T) {
	log := zap.NewNop()
	hub := ws.NewHub()
	bottle := &domain.Bottle{
		ID:         1,
		MessageText: "hello ocean",
		Status:     domain.BottleStatusDrifting,
		CreatedAt:  time.Now(),
	}

	app := fiber.New()
	api.RegisterRoutes(app, api.Handlers{
		Health:    &handler.HealthHandler{},
		Bottle:    handler.NewBottleHandler(&fakeBottleSvc{bottle: bottle}, nil),
		Event:     handler.NewEventHandler(nil),
		Discovery: handler.NewDiscoveryHandler(nil),
	}, hub, log)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bottles/1", nil)
	// no Authorization header — Visitor is anonymous
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("anonymous bottle read must not require JWT; got 401 body=%s", body)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("want 200, got %d body=%s", resp.StatusCode, body)
	}
}

func TestUserJWTCreateLoginIsNotProductRoute(t *testing.T) {
	log := zap.NewNop()
	hub := ws.NewHub()
	app := fiber.New()
	api.RegisterRoutes(app, api.Handlers{
		Health:    &handler.HealthHandler{},
		Bottle:    handler.NewBottleHandler(&fakeBottleSvc{}, nil),
		Event:     handler.NewEventHandler(nil),
		Discovery: handler.NewDiscoveryHandler(nil),
	}, hub, log)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("JWT user create must not be product route; got %d", resp.StatusCode)
	}
}
