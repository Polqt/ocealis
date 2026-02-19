package util

import "math"

const earthRadiusKm = 6371.0

func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func ApplyDrift(lat, lng, speedKmH, bearingDeg, durationHours float64) (newLat, newLng float64) {
	distKm := speedKmH * durationHours
	distRad := distKm / earthRadiusKm
	bearingRad := toRad(bearingDeg)
	latRad := toRad(lat)
	lngRad := toRad(lng)

	newLatRad := math.Asin(
		math.Sin(latRad)*math.Cos(distRad) +
			math.Cos(latRad)*math.Sin(distRad)*math.Cos(bearingRad),
	)

	newLngRad := lngRad + math.Atan2(
		math.Sin(bearingRad)*math.Sin(distRad)*math.Cos(latRad),
		math.Cos(distRad)-math.Sin(latRad)*math.Sin(newLatRad),
	)

	newLng = math.Mod(toDeg(newLngRad)+540, 360) - 180 // Normalize to [-180, 180]
	newLat = toDeg(newLatRad)
	return
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }
func toDeg(rad float64) float64 { return rad * 180 / math.Pi }
