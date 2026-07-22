import { createSignal, Show } from "solid-js";
import { castBottle } from "~/lib/api";

type Phase = "write" | "casting" | "out";

async function readGeo(): Promise<{ lat?: number; lng?: number }> {
  if (!navigator.geolocation) return {};
  try {
    const pos = await new Promise<GeolocationPosition>((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(resolve, reject, {
        timeout: 8000,
        maximumAge: 60_000,
      });
    });
    return { lat: pos.coords.latitude, lng: pos.coords.longitude };
  } catch {
    // Denied/missing → API BasinFallback
    return {};
  }
}

export default function Home() {
  const [nickname, setNickname] = createSignal("");
  const [message, setMessage] = createSignal("");
  const [phase, setPhase] = createSignal<Phase>("write");
  const [error, setError] = createSignal("");

  async function onCast(e: Event) {
    e.preventDefault();
    setError("");
    setPhase("casting");
    try {
      const geo = await readGeo();
      // ponytail: no Turnstile widget until site key; server accepts non-empty when secret empty
      const turnstile =
        (window as unknown as { turnstileToken?: string }).turnstileToken ?? "dev";
      await castBottle({
        nickname: nickname().trim(),
        message_text: message().trim(),
        turnstile_token: turnstile,
        start_lat: geo.lat,
        start_lng: geo.lng,
      });
      setPhase("out");
      setMessage("");
    } catch (err) {
      setPhase("write");
      setError(err instanceof Error ? err.message : "cast failed");
    }
  }

  return (
    <main class="cast-page">
      <div class="cast-sea" aria-hidden="true" />
      <section class="cast-panel">
        <p class="cast-brand">Ocealis</p>
        <Show
          when={phase() !== "out"}
          fallback={
            <div class="cast-out">
              <h1>It’s out there.</h1>
              <p>No map pin. No “your bottles.” The ocean keeps the secret.</p>
              <button type="button" class="cast-btn" onClick={() => setPhase("write")}>
                Cast another
              </button>
            </div>
          }
        >
          <h1>Cast a Bottle</h1>
          <p class="cast-lede">Write a Message. Choose a Nickname. Let go.</p>
          <form class="cast-form" onSubmit={onCast}>
            <label>
              Nickname
              <input
                name="nickname"
                maxlength={24}
                required
                value={nickname()}
                onInput={e => setNickname(e.currentTarget.value)}
                autocomplete="nickname"
              />
            </label>
            <label>
              Message
              <textarea
                name="message"
                maxlength={500}
                required
                rows={5}
                value={message()}
                onInput={e => setMessage(e.currentTarget.value)}
              />
            </label>
            <Show when={error()}>
              <p class="cast-error" role="alert">
                {error()}
              </p>
            </Show>
            <button type="submit" class="cast-btn" disabled={phase() === "casting"}>
              {phase() === "casting" ? "Casting…" : "Cast / Release"}
            </button>
          </form>
        </Show>
      </section>
    </main>
  );
}
