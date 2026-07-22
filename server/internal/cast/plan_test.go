package cast_test

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/Polqt/ocealis/internal/cast"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/geo"
)

func TestCastRejectsOverLimitNicknameAndMessage(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(1))
	lat, lng := 30.0, -140.0

	_, err := cast.Prepare("x"+strings.Repeat("y", 24), "hi", &lat, &lng, now, rng)
	if err == nil {
		t.Fatal("want reject nickname >24")
	}

	_, err = cast.Prepare("ok", strings.Repeat("m", 501), &lat, &lng, now, rng)
	if err == nil {
		t.Fatal("want reject message >500")
	}
}

func TestCastAppliesMysteryDelayAndOceanDrop(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(42))
	// inland Kansas
	lat, lng := 39.0997, -94.5786

	plan, err := cast.Prepare("sailor", "hello ocean", &lat, &lng, now, rng)
	if err != nil {
		t.Fatal(err)
	}

	if plan.Status != domain.BottleStatusMysteryDelay {
		t.Fatalf("status=%s want Mystery Delay", plan.Status)
	}
	if plan.IsReleased {
		t.Fatal("Mystery Delay bottle must not be released/visible yet")
	}

	minAt := now.Add(15 * time.Minute)
	maxAt := now.Add(30 * time.Minute)
	if plan.VisibleAt.Before(minAt) || plan.VisibleAt.After(maxAt) {
		t.Fatalf("visible_at=%v want in [%v, %v]", plan.VisibleAt, minAt, maxAt)
	}

	if geo.IsLand(plan.Lat, plan.Lng) {
		t.Fatalf("drop must be Ocean, got %v,%v", plan.Lat, plan.Lng)
	}
	if plan.Nickname != "sailor" {
		t.Fatalf("nickname=%q", plan.Nickname)
	}
}

func TestCastMissingGeoUsesBasinFallback(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(1))

	plan, err := cast.Prepare("anon", "from desktop", nil, nil, now, rng)
	if err != nil {
		t.Fatal(err)
	}
	fb := geo.BasinFallback()
	if plan.Lat != fb.Lat || plan.Lng != fb.Lng {
		t.Fatalf("want basin fallback %v got %v,%v", fb, plan.Lat, plan.Lng)
	}
}

func TestCastSanitizesMessageHTML(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	rng := rand.New(rand.NewSource(1))
	lat, lng := 30.0, -140.0

	plan, err := cast.Prepare("anon", "<b>hi</b><script>x</script>", &lat, &lng, now, rng)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(plan.MessageText, "<") || strings.Contains(plan.MessageText, "script") {
		t.Fatalf("message still has HTML: %q", plan.MessageText)
	}
	if plan.MessageText == "" {
		t.Fatal("sanitized message empty")
	}
}
