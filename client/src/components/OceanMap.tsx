import { For, Show, createSignal, onCleanup, onMount } from "solid-js";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import {
  browseMap,
  getJourney,
  openBottle,
  reReleaseBottle,
  stampBottle,
  type MapBrowseResult,
} from "~/lib/api";
import type { Bottle, BottleEvent } from "~/lib/types";

const HEAT_SRC = "ocean-heat";
const HEAT_LAYER = "ocean-heat-circles";
const CORK_SRC = "ocean-corks";
const CORK_LAYER = "ocean-corks-circles";

type Opened = { bottle: Bottle; events: BottleEvent[] };

export default function OceanMap() {
  let el!: HTMLDivElement;
  let oceanMap: maplibregl.Map | undefined;
  const [opened, setOpened] = createSignal<Opened | null>(null);
  const [openErr, setOpenErr] = createSignal("");
  const [sealIcon, setSealIcon] = createSignal("");
  const [stampNote, setStampNote] = createSignal("");
  const [stampErr, setStampErr] = createSignal("");
  const [stampDone, setStampDone] = createSignal(false);
  const [stampBusy, setStampBusy] = createSignal(false);
  const [reReleaseNickname, setReReleaseNickname] = createSignal("");
  const [reReleaseErr, setReReleaseErr] = createSignal("");
  const [reReleaseDone, setReReleaseDone] = createSignal("");
  const [reReleaseBusy, setReReleaseBusy] = createSignal(false);

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
    oceanMap = map;
    map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "top-right");

    let timer: ReturnType<typeof setTimeout> | undefined;
    let poll: ReturnType<typeof setInterval> | undefined;
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
      poll = setInterval(() => void refresh(map), 60_000);
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
      clearInterval(poll);
      oceanMap = undefined;
      map.remove();
    });
  });

  async function openCork(id: number) {
    setOpenErr("");
    resetStamp();
    resetReRelease();
    try {
      const [bottle, journey] = await Promise.all([openBottle(id), getJourney(id)]);
      setOpened({ bottle, events: journey.events ?? [] });
    } catch (err) {
      setOpened(null);
      setOpenErr(err instanceof Error ? err.message : "could not Open");
    }
  }

  function resetStamp() {
    setSealIcon("");
    setStampNote("");
    setStampErr("");
    setStampDone(false);
    setStampBusy(false);
  }

  function closeCork() {
    setOpened(null);
    resetStamp();
    resetReRelease();
  }

  async function submitStamp(event: SubmitEvent) {
    event.preventDefault();
    const current = opened();
    if (!current || stampBusy()) return;

    const note = stampNote().trim();
    const seal = sealIcon();
    if (!seal && !note) {
      setStampErr("Choose a seal or leave a note.");
      return;
    }

    setStampBusy(true);
    setStampErr("");
    setStampDone(false);
    try {
      const turnstile =
        (window as unknown as { turnstileToken?: string }).turnstileToken ?? "dev";
      const stamped = await stampBottle(current.bottle.id, {
        seal_icon: seal || undefined,
        note: note || undefined,
        turnstile_token: turnstile,
      });
      setOpened(previous =>
        previous?.bottle.id === current.bottle.id
          ? { ...previous, events: [...previous.events, stamped] }
          : previous,
      );
      setSealIcon("");
      setStampNote("");
      setStampDone(true);
    } catch (err) {
      setStampErr(err instanceof Error ? err.message : "could not Stamp");
    } finally {
      setStampBusy(false);
    }
  }

  function resetReRelease() {
    setReReleaseNickname("");
    setReReleaseErr("");
    setReReleaseDone("");
    setReReleaseBusy(false);
  }

  async function submitReRelease(event: SubmitEvent) {
    event.preventDefault();
    const current = opened();
    if (!current || reReleaseBusy()) return;

    const nickname = reReleaseNickname().trim();
    if (!nickname) {
      setReReleaseErr("Nickname is required.");
      return;
    }

    setReReleaseBusy(true);
    setReReleaseErr("");
    setReReleaseDone("");
    try {
      const location = await readGeo();
      const turnstile =
        (window as unknown as { turnstileToken?: string }).turnstileToken ?? "dev";
      await reReleaseBottle(current.bottle.id, {
        nickname,
        turnstile_token: turnstile,
        lat: location.lat,
        lng: location.lng,
      });
      setOpened(null);
      setReReleaseNickname("");
      setReReleaseDone("The Bottle is hidden by Mystery Delay, then it will Drift elsewhere.");
      if (oceanMap) await refresh(oceanMap);
    } catch (err) {
      setReReleaseErr(err instanceof Error ? err.message : "could not Re-release");
    } finally {
      setReReleaseBusy(false);
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
      <Show when={reReleaseDone()}>
        <p class="cork-open__notice" role="status">
          {reReleaseDone()}
        </p>
      </Show>
      <Show when={opened()}>
        {o => (
          <aside class="cork-open" aria-label="Opened Bottle">
            <button type="button" class="cork-open__close" onClick={closeCork}>
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
            <form class="stamp-form" onSubmit={submitStamp}>
              <h2>Stamp this Journey</h2>
              <label>
                Seal
                <select value={sealIcon()} onInput={event => setSealIcon(event.currentTarget.value)}>
                  <option value="">No seal</option>
                  <option value="⚓">⚓ Anchor</option>
                  <option value="🌊">🌊 Wave</option>
                  <option value="🐚">🐚 Shell</option>
                  <option value="⭐">⭐ Star</option>
                </select>
              </label>
              <label>
                Note
                <textarea
                  value={stampNote()}
                  onInput={event => setStampNote(event.currentTarget.value)}
                  maxLength={80}
                  rows={2}
                  placeholder="A few words for the next Visitor"
                />
              </label>
              <button
                type="submit"
                disabled={stampBusy() || (!sealIcon() && !stampNote().trim())}
              >
                {stampBusy() ? "Stamping…" : "Stamp"}
              </button>
              <Show when={stampErr()}>
                <p class="stamp-form__error" role="alert">
                  {stampErr()}
                </p>
              </Show>
              <Show when={stampDone()}>
                <p class="stamp-form__done" role="status">
                  Journey stamped.
                </p>
              </Show>
            </form>
            <form class="stamp-form" onSubmit={submitReRelease}>
              <h2>Re-release this Bottle</h2>
              <label>
                Your Nickname
                <input
                  name="re-release-nickname"
                  value={reReleaseNickname()}
                  onInput={event => setReReleaseNickname(event.currentTarget.value)}
                  maxLength={24}
                  required
                  autocomplete="nickname"
                />
              </label>
              <button type="submit" disabled={reReleaseBusy() || !reReleaseNickname().trim()}>
                {reReleaseBusy() ? "Re-releasing…" : "Re-release"}
              </button>
              <Show when={reReleaseErr()}>
                <p class="stamp-form__error" role="alert">
                  {reReleaseErr()}
                </p>
              </Show>
            </form>
          </aside>
        )}
      </Show>
    </>
  );
}

async function readGeo(): Promise<{ lat?: number; lng?: number }> {
  if (!navigator.geolocation) return {};
  try {
    const position = await new Promise<GeolocationPosition>((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(resolve, reject, {
        timeout: 8000,
        maximumAge: 60_000,
      });
    });
    return { lat: position.coords.latitude, lng: position.coords.longitude };
  } catch {
    return {};
  }
}

function journeyLabel(event: BottleEvent): string {
  switch (event.event_type) {
    case "released":
      return "Cast";
    case "drift":
      return "Drift";
    case "stamp":
      return ["Stamp", event.seal_icon, event.note ? `— ${event.note}` : ""]
        .filter(Boolean)
        .join(" ");
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
