import { For, Show, createEffect, createSignal, onCleanup } from "solid-js";
import { discoverBottle, getJourney, releaseBottle, getStoredUser } from "~/lib/api";
import type { BottleEvent, Journey } from "~/lib/types";

type Props = {
  bottleId: () => number | null;
  reloadToken: () => number;
  onClose: () => void;
  onJourney: (points: { lat: number; lng: number }[]) => void;
  onReleased: (id: number) => void;
  onDiscovered: (id: number) => void;
};

const CHAPTER: Record<string, { title: string; line: string }> = {
  released: { title: "Cast", line: "The bottle first touched water." },
  drift: { title: "Drift", line: "Currents carried it farther from shore." },
  discovered: { title: "Found", line: "Someone lifted it from the sea." },
  re_released: { title: "Returned", line: "It was set free again." }
};

function chapterCopy(event: BottleEvent) {
  return CHAPTER[event.event_type] ?? { title: event.event_type, line: "A quiet moment on the journey." };
}

function journeyPointsFrom(j: Journey) {
  const points = [...j.events].reverse().map(e => ({ lat: e.lat, lng: e.lng }));
  points.push({ lat: j.bottle.current_lat, lng: j.bottle.current_lng });
  return points;
}

export default function BottlePanel(props: Props) {
  const [journey, setJourney] = createSignal<Journey | null>(null);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal("");

  createEffect(() => {
    const id = props.bottleId();
    props.reloadToken(); // remote discover/refresh
    setJourney(null);
    setError("");
    if (!id) {
      props.onJourney([]);
      return;
    }

    let cancelled = false;
    void (async () => {
      try {
        const j = await getJourney(id);
        if (cancelled || props.bottleId() !== id) return;
        setJourney(j);
        props.onJourney(journeyPointsFrom(j));
      } catch (err) {
        if (cancelled || props.bottleId() !== id) return;
        setError(err instanceof Error ? err.message : "Could not load journey");
      }
    })();

    onCleanup(() => {
      cancelled = true;
    });
  });

  const isOwn = () => {
    const user = getStoredUser();
    const j = journey();
    return !!(user && j && user.id === j.bottle.sender_id);
  };

  const onDiscover = async () => {
    const j = journey();
    if (!j || busy()) return;
    setBusy(true);
    setError("");
    try {
      const next = await discoverBottle(j.bottle.id, j.bottle.current_lat, j.bottle.current_lng);
      setJourney(next);
      props.onDiscovered(j.bottle.id);
      props.onJourney(journeyPointsFrom(next));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not discover");
    } finally {
      setBusy(false);
    }
  };

  const onRelease = async () => {
    const j = journey();
    if (!j || busy()) return;
    setBusy(true);
    setError("");
    try {
      const bottle = await releaseBottle(j.bottle.id, j.bottle.current_lat, j.bottle.current_lng);
      props.onReleased(bottle.id);
      props.onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not re-release");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Show when={props.bottleId()}>
      <aside class="panel bottle-panel" aria-label="Bottle journey">
        <header class="panel-header">
          <h2>Bottle #{props.bottleId()}</h2>
          <button type="button" class="ghost" onClick={() => props.onClose()}>
            Close
          </button>
        </header>

        <Show when={journey()} fallback={<p class="panel-lead">Reading the tide…</p>}>
          {j => (
            <>
              <p class="message">{j().bottle.message_text}</p>
              <div class="meta">
                <span>{j().bottle.hops} hops</span>
                <span>{j().bottle.status}</span>
              </div>

              <section class="journey-chapters" aria-label="Journey chapters">
                <h3>Journey</h3>
                <ol>
                  <For each={[...j().events].reverse()}>
                    {event => {
                      const copy = chapterCopy(event);
                      return (
                        <li>
                          <strong>{copy.title}</strong>
                          <span>{copy.line}</span>
                          <em>
                            {event.lat.toFixed(1)}°, {event.lng.toFixed(1)}°
                          </em>
                        </li>
                      );
                    }}
                  </For>
                </ol>
              </section>

              <Show when={error()}>
                <p class="error">{error()}</p>
              </Show>

              <div class="actions">
                <Show when={j().bottle.status === "drifting" && !isOwn()}>
                  <button type="button" class="primary" disabled={busy()} onClick={() => void onDiscover()}>
                    Discover
                  </button>
                </Show>
                <Show when={j().bottle.status === "discovered"}>
                  <button type="button" class="primary" disabled={busy()} onClick={() => void onRelease()}>
                    Re-release
                  </button>
                </Show>
              </div>
            </>
          )}
        </Show>
      </aside>
    </Show>
  );
}
