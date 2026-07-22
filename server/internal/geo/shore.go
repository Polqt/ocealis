package geo

import "math"

// Point is a lat/lng pair on the world.
type Point struct {
	Lat float64
	Lng float64
}

// shoreDrop pairs a coastline sample with a just-offshore Ocean drop.
type shoreDrop struct {
	shore Point
	drop  Point
}

// Coarse land rectangles — good enough for Cast snap tests, not a gazetteer.
// ponytail: axis-aligned continents; finer coastline when Drift needs it.
var landRects = []struct{ minLat, maxLat, minLng, maxLng float64 }{
	{24, 49, -125, -66},  // contiguous US / southern Canada strip
	{14, 32, -118, -86},  // Mexico / Central America (partial)
	{-56, 12, -82, -34},  // South America
	{36, 71, -10, 40},    // Europe
	{0, 37, -18, 52},     // Africa north/west
	{-35, 0, 8, 52},      // Africa south
	{8, 55, 60, 145},     // Asia bulk
	{-40, -10, 112, 154}, // Australia
}

// Sample shoreline drops (shore ≈ coast, drop ≈ just offshore).
var shores = []shoreDrop{
	{shore: Point{29.1, -94.8}, drop: Point{28.5, -94.5}},   // Gulf of Mexico near Houston
	{shore: Point{32.7, -117.2}, drop: Point{32.5, -117.5}}, // San Diego offshore
	{shore: Point{40.6, -74.0}, drop: Point{40.2, -73.5}},   // NY / Atlantic
	{shore: Point{25.8, -80.1}, drop: Point{25.5, -79.8}},   // Miami offshore
	{shore: Point{47.6, -122.3}, drop: Point{47.5, -124.5}}, // Seattle → Pacific
	{shore: Point{51.5, -0.1}, drop: Point{50.5, -1.5}},     // UK south coast
	{shore: Point{35.7, 139.7}, drop: Point{34.5, 140.5}},   // Tokyo → Pacific
	{shore: Point{-33.9, 151.2}, drop: Point{-34.2, 151.5}}, // Sydney offshore
}

// BasinFallback is used when Visitor denies/missing geolocation.
// Mid North Pacific — open Ocean, no country picker.
func BasinFallback() Point {
	return Point{Lat: 30.0, Lng: -140.0}
}

// IsLand reports whether the point sits inside a coarse land rectangle.
// Known offshore Cast drops and BasinFallback are Ocean even if a rect overlaps them
// (coarse US box covers Gulf water — drops stay authoritative).
func IsLand(lat, lng float64) bool {
	if isKnownOcean(lat, lng) {
		return false
	}
	for _, r := range landRects {
		if lat >= r.minLat && lat <= r.maxLat && lng >= r.minLng && lng <= r.maxLng {
			return true
		}
	}
	return false
}

func isKnownOcean(lat, lng float64) bool {
	fb := BasinFallback()
	if near(lat, lng, fb.Lat, fb.Lng) {
		return true
	}
	for _, s := range shores {
		if near(lat, lng, s.drop.Lat, s.drop.Lng) {
			return true
		}
	}
	return false
}

func near(lat1, lng1, lat2, lng2 float64) bool {
	const eps = 0.05 // ~5km
	return math.Abs(lat1-lat2) < eps && math.Abs(lng1-lng2) < eps
}

// ResolveDrop returns an Ocean-only Cast drop.
// Inland → nearest Shoreline just offshore. Already Ocean → keep.
func ResolveDrop(lat, lng float64) (float64, float64) {
	if !IsLand(lat, lng) {
		return lat, lng
	}
	best := shores[0]
	bestD := haversineKm(lat, lng, best.shore.Lat, best.shore.Lng)
	for _, s := range shores[1:] {
		d := haversineKm(lat, lng, s.shore.Lat, s.shore.Lng)
		if d < bestD {
			bestD = d
			best = s
		}
	}
	return best.drop.Lat, best.drop.Lng
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
