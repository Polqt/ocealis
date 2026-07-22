import { createSignal, onCleanup, onMount, Show } from "solid-js";
import OceanCanvas from "~/components/OceanCanvas";
import CastPanel from "~/components/CastPanel";
import BottlePanel from "~/components/BottlePanel";
import { ensureAnonSession, getJourney, listOceanBottles } from "~/lib/api";
import { connectOceanWs } from "~/lib/ocean-ws";
import type { Bottle } from "~/lib/types";

export default function OceanHome() {
  const [ready, setReady] = createSignal(false);
  const [error, setError] = createSignal("");
  const [live, setLive] = createSignal(false);
  const [bottles, setBottles] = createSignal<Bottle[]>([]);
  const [selectedId, setSelectedId] = createSignal<number | null>(null);
  const [castOpen, setCastOpen] = createSignal(false);
  const [castPulse, setCastPulse] = createSignal(0);
  const [journeyPoints, setJourneyPoints] = createSignal<{ lat: number; lng: number }[]>([]);

  const upsertBottle = (bottle: Bottle) => {
    setBottles(prev => {
      const idx = prev.findIndex(b => b.id === bottle.id);
      if (idx === -1) return [bottle, ...prev];
      const next = prev.slice();
      next[idx] = { ...next[idx], ...bottle };
      return next;
    });
  };

  onMount(() => {
    let stopWs = () => {};
    void (async () => {
      try {
        await ensureAnonSession();
        const ocean = await listOceanBottles(120);
        setBottles(ocean.data ?? []);
        setReady(true);
        stopWs = connectOceanWs({
          onStatus: setLive,
          onDrift: payload => {
            setBottles(prev =>
              prev.map(b =>
                b.id === payload.bottle_id
                  ? {
                      ...b,
                      current_lat: payload.lat,
                      current_lng: payload.lng,
                      hops: payload.hops,
                      bottle_style: payload.bottle_style
                    }
                  : b
              )
            );
          },
          onDiscovered: id => {
            setBottles(prev => prev.filter(b => b.id !== id));
            if (selectedId() === id) {
              /* keep panel; journey refresh handles status */
            }
          },
          onReleased: id => {
            void getJourney(id)
              .then(j => upsertBottle(j.bottle))
              .catch(() => {});
          }
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : "Could not open the ocean");
      }
    })();

    onCleanup(() => stopWs());
  });

  return (
    <main class="ocean-shell">
      <div class="ocean-stage">
        <Show when={ready()} fallback={<div class="ocean-loading">Opening the sea…</div>}>
          <OceanCanvas
            bottles={bottles}
            selectedId={selectedId}
            journeyPoints={journeyPoints}
            castPulse={castPulse}
            onSelect={id => setSelectedId(id)}
          />
        </Show>

        <header class="brand-hero">
          <p class="brand">Ocealis</p>
          <h1>A living ocean of anonymous bottles</h1>
          <p class="tagline">Cast a message. Watch it drift. Let strangers find it.</p>
          <div class="cta-row">
            <button type="button" class="primary" onClick={() => setCastOpen(true)}>
              Cast a bottle
            </button>
            <span classList={{ pulse: true, on: live() }}>{live() ? "Live drift" : "Connecting…"}</span>
          </div>
        </header>

        <Show when={error()}>
          <p class="banner-error">{error()}</p>
        </Show>
      </div>

      <CastPanel
        open={castOpen}
        onClose={() => setCastOpen(false)}
        onCast={id => {
          setCastPulse(v => v + 1);
          void getJourney(id)
            .then(j => {
              upsertBottle(j.bottle);
              setSelectedId(id);
            })
            .catch(() => {});
        }}
      />

      <BottlePanel
        bottleId={selectedId}
        onClose={() => {
          setSelectedId(null);
          setJourneyPoints([]);
        }}
        onJourney={setJourneyPoints}
        onReleased={id => {
          void getJourney(id)
            .then(j => upsertBottle(j.bottle))
            .catch(() => {});
        }}
        onDiscovered={id => {
          setBottles(prev => prev.filter(b => b.id !== id));
        }}
      />
    </main>
  );
}
