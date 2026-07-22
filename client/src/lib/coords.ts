/** Map lat/lng to a flat ocean plane (x,z). */
export function latLngToPlane(lat: number, lng: number, span = 48): { x: number; z: number } {
  const x = (lng / 180) * (span / 2);
  const z = (-lat / 90) * (span / 2);
  return { x, z };
}

export function planeToLatLng(x: number, z: number, span = 48): { lat: number; lng: number } {
  const lng = (x / (span / 2)) * 180;
  const lat = (-z / (span / 2)) * 90;
  return {
    lat: Math.max(-85, Math.min(85, lat)),
    lng: Math.max(-180, Math.min(180, lng))
  };
}

export function randomCastPoint(): { lat: number; lng: number } {
  // Prefer mid-ocean bands so bottles start in water-ish zones.
  const lat = (Math.random() * 50 - 25) | 0;
  const lng = Math.random() * 300 - 150;
  return { lat: lat + Math.random(), lng };
}
