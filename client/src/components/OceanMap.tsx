import { For, Show, createSignal, onCleanup, onMount } from "solid-js";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { browseMap, getJourney, openBottle, stampBottle, type MapBrowseResult } from "~/lib/api";
import type { Bottle, BottleEvent } from "~/lib/types";

const HEAT_SRC = "ocean-heat";
const HEAT_LAYER = "ocean-heat-circles";
const CORK_SRC = "ocean-corks";
const CORK_LAYER = "ocean-corks-circles";

type Opened = { bottle: Bottle; events: BottleEvent[] };

export default function OceanMap() {
  let el!: HTMLDivElement;
  const [opened, setOpened] = createSignal<Opened | null>(null);
  const [openErr, setOpenErr] = createSignal("");
  const [sealIcon, setSealIcon] = createSignal("");
  const [stampNote, setStampNote] = createSignal("");
  const [stamping, setStamping] = createSignal(false);
  const [stampErr, setStampErr] = createSignal("");

  onMount(() => {
    const map = new maplibregl.Map({
      container: el,
      style: "https://demotiles.maplibre.org/style.json",
      center: [-140, 30],
      zoom: 2.2,
      minZoom: 1,
      maxZoom: 10,
      attributionControl: { compact: true },
    });
    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "top-right");

    let timer: ReturnType<typeof setTimeout> | undefined;
    const schedule = () => {
      clearTimeout(timer);
      timer = setTimeout(() => void refresh(map), 280);
    };

    map.on("load", () => {
      map.addSource(HEAT_SRC, {
        type: "geojson",
        data: emptyFC(),
      });
      map.addLayer({
        id: HEAT_LAYER,
        type: "circle",
        source: HEAT_SRC,
        paint: {
          "circle-radius": ["interpolate", ["linear"], ["get", "count"], 1, 8, 20, 28],
          "circle-color": "#f0c27b",
          "circle-opacity": 0.45,
          "circle-blur": 0.4,
        },
      });
      map.addSource(CORK_SRC, {
        type: "geojson",
        data: emptyFC(),
      });
      map.addLayer({
        id: CORK_LAYER,
        type: "circle",
        source: CORK_SRC,
        paint: {
          "circle-radius": 5,
          "circle-color": ["case", ["get", "is_seed"], "#c8e7f0", "#f0c27b"],
          "circle-stroke-width": 1.5,
          "circle-stroke-color": "#061820",
        },
      });
      void refresh(map);
    });

    map.on("mouseenter", CORK_LAYER, () => {
      map.getCanvas().style.cursor = "pointer";
    });
    map.on("mouseleave", CORK_LAYER, () => {
      map.getCanvas().style.cursor = "";
    });
    map.on("click", CORK_LAYER, e => {
      const id = e.features?.[0]?.properties?.id;
      if (id == null) return;
      void openCork(Number(id));
    });

    map.on("moveend", schedule);
    map.on("zoomend", schedule);

    onCleanup(() => {
      clearTimeout(timer);
      map.remove();
    });
  });

  async function openCork(id: number) {
    setOpenErr("");
    setSealIcon("");
    setStampNote("");
    setStampErr("");
    try {
      const [bottle, journey] = await Promise.all([openBottle(id), getJourney(id)]);
      setOpened({ bottle, events: journey.events ?? [] });
    } catch (err) {
      setOpened(null);
      setOpenErr(err instanceof Error ? err.message : "could not Open");
    }
  }

  async function addStamp(event: SubmitEvent) {
    event.preventDefault();
    const current = opened();
    if (!current || (!sealIcon() && !stampNote().trim())) return;

    setStamping(true);
    setStampErr("");
    try {
      const journey = await stampBottle(current.bottle.id, {
        seal_icon: sealIcon() || undefined,
        note: stampNote().trim() || undefined,
        turnstile_token: "dev",
      });
      setOpened({ bottle: journey.bottle, events: journey.events ?? [] });
      setSealIcon("");
      setStampNote("");
    } catch (err) {
      setStampErr(err instanceof Error ? err.message : "could not Stamp");
    } finally {
      setStamping(false);
    }
  }

  return (
    <>
      <div class="ocean-map" ref={el} role="application" aria-label="Ocean map" />
      <Show when={openErr()}>
        <p class="cork-open__err" role="alert">
          {openErr()}
        </p>
      </Show>
      <Show when={opened()}>
        {o => (
          <aside class="cork-open" aria-label="Opened Bottle">
            <button type="button" class="cork-open__close" onClick={() => setOpened(null)}>
              Close
            </button>
            <p class="cork-open__nick">{o().bottle.nickname}</p>
            <p class="cork-open__msg">{o().bottle.message_text}</p>
            <h2 class="cork-open__journey-title">Journey</h2>
            <Show
              when={o().events.length > 0}
              fallback={<p class="cork-open__empty">No Journey events yet.</p>}
            >
              <ol class="cork-open__events">
                <For each={o().events}>{ev => <li>{journeyLabel(ev)}</li>}</For>
              </ol>
            </Show>
            <form class="cork-open__stamp" onSubmit={addStamp}>
              <h2>Stamp this Bottle</h2>
              <label>
                Seal
                <select value={sealIcon()} onInput={e => setSealIcon(e.currentTarget.value)}>
                  <option value="">No seal</option>
                  <option value="⚓">⚓ Anchor</option>
                  <option value="🐚">🐚 Shell</option>
                  <option value="🌊">🌊 Wave</option>
                  <option value="⭐">⭐ Star</option>
                </select>
              </label>
              <label>
                Note <span>{stampNote().length}/80</span>
                <textarea
                  rows={2}
                  maxLength={80}
                  value={stampNote()}
                  onInput={e => setStampNote(e.currentTarget.value)}
                  placeholder="Optional passport note"
                />
              </label>
              <Show when={stampErr()}>
                <p class="cork-open__stamp-error" role="alert">
                  {stampErr()}
                </p>
              </Show>
              <button
                type="submit"
                disabled={stamping() || (!sealIcon() && !stampNote().trim())}
              >
                {stamping() ? "Stamping…" : "Add Stamp"}
              </button>
            </form>
          </aside>
        )}
      </Show>
    </>
  );
}

function journeyLabel(event: BottleEvent): string {
  switch (event.event_type) {
    case "released":
      return "Cast";
    case "drift":
      return "Drift";
    case "stamp":
      return `Stamp${event.seal_icon ? ` ${event.seal_icon}` : ""}${event.note ? ` — ${event.note}` : ""}`;
    case "re_released":
      return "Re-release";
    case "sink":
      return "Sink";
    default:
      return event.event_type;
  }
}

function emptyFC(): GeoJSON.FeatureCollection {
  return { type: "FeatureCollection", features: [] };
}

async function refresh(map: maplibregl.Map) {
  const b = map.getBounds();
  let result: MapBrowseResult;
  try {
    result = await browseMap({
      min_lat: b.getSouth(),
      max_lat: b.getNorth(),
      min_lng: b.getWest(),
      max_lng: b.getEast(),
      zoom: map.getZoom(),
    });
  } catch {
    return;
  }

  const heat = map.getSource(HEAT_SRC) as maplibregl.GeoJSONSource | undefined;
  const corks = map.getSource(CORK_SRC) as maplibregl.GeoJSONSource | undefined;
  if (!heat || !corks) return;

  if (result.mode === "heat") {
    heat.setData({
      type: "FeatureCollection",
      features: (result.heat ?? []).map(c => ({
        type: "Feature",
        properties: { count: c.count },
        geometry: { type: "Point", coordinates: [c.lng, c.lat] },
      })),
    });
    corks.setData(emptyFC());
    map.setLayoutProperty(HEAT_LAYER, "visibility", "visible");
    map.setLayoutProperty(CORK_LAYER, "visibility", "none");
  } else {
    corks.setData({
      type: "FeatureCollection",
      features: (result.corks ?? []).map(c => ({
        type: "Feature",
        properties: { id: c.id, is_seed: !!c.is_seed },
        geometry: { type: "Point", coordinates: [c.lng, c.lat] },
      })),
    });
    heat.setData(emptyFC());
    map.setLayoutProperty(HEAT_LAYER, "visibility", "none");
    map.setLayoutProperty(CORK_LAYER, "visibility", "visible");
  }
}
