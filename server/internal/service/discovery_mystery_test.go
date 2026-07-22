package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/service"
)

// fakeBottles returns whatever FindNearby is given — discovery must still hide Mystery Delay.
type fakeBottles struct {
	rows []domain.Bottle
}

func (f *fakeBottles) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (f *fakeBottles) GetByID(context.Context, int32) (*domain.Bottle, error) { return nil, nil }
func (f *fakeBottles) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (f *fakeBottles) UpdatePosition(context.Context, int32, float64, float64, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (f *fakeBottles) ListActive(context.Context) ([]domain.Bottle, error) { return nil, nil }
func (f *fakeBottles) ReleaseScheduled(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (f *fakeBottles) FindNearby(context.Context, repository.FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	return &domain.CursorResult[domain.Bottle]{Data: f.rows}, nil
}
func (f *fakeBottles) WithTx(*ocealis.Queries) repository.BottleRepository { return f }

func TestMysteryDelayBottleInvisibleToNearby(t *testing.T) {
	now := time.Now()
	repo := &fakeBottles{rows: []domain.Bottle{
		{
			ID:          1,
			MessageText: "secret",
			Status:      domain.BottleStatusMysteryDelay,
			IsReleased:  false,
			VisibleAt:   now.Add(20 * time.Minute),
			CurrentLat:  30,
			CurrentLng:  -140,
		},
		{
			ID:          2,
			MessageText: "visible cork",
			Status:      domain.BottleStatusDrifting,
			IsReleased:  true,
			VisibleAt:   now.Add(-time.Hour),
			CurrentLat:  30.1,
			CurrentLng:  -140.1,
		},
	}}
	svc := service.NewDiscoveryService(repo)
	out, err := svc.FindNearby(context.Background(), service.FindNearbyInput{
		Lat: 30, Lng: -140, RadiusKm: 500, Limit: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Data) != 1 || out.Data[0].ID != 2 {
		t.Fatalf("Mystery Delay must be invisible; got %+v", out.Data)
	}
}
