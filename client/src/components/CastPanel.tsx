import { createSignal, Show } from "solid-js";
import { createBottle } from "~/lib/api";
import { randomCastPoint } from "~/lib/coords";

type Props = {
  open: () => boolean;
  onClose: () => void;
  onCast: (bottleId: number) => void;
};

const STYLES = [
  { id: 0, label: "Amber" },
  { id: 1, label: "Sand" },
  { id: 2, label: "Clay" },
  { id: 3, label: "Driftwood" },
  { id: 4, label: "Bark" },
  { id: 5, label: "Kelp" },
  { id: 6, label: "Tide" },
  { id: 7, label: "Mist" },
  { id: 8, label: "Foam" },
  { id: 9, label: "Stone" }
];

export default function CastPanel(props: Props) {
  const [message, setMessage] = createSignal("");
  const [style, setStyle] = createSignal(6);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal("");

  const submit = async (e: Event) => {
    e.preventDefault();
    if (!message().trim() || busy()) return;
    setBusy(true);
    setError("");
    try {
      const point = randomCastPoint();
      const bottle = await createBottle({
        message_text: message().trim(),
        bottle_style: style(),
        start_lat: point.lat,
        start_lng: point.lng
      });
      setMessage("");
      props.onCast(bottle.id);
      props.onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not cast bottle");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Show when={props.open()}>
      <div class="panel cast-panel" role="dialog" aria-label="Cast a bottle">
        <header class="panel-header">
          <h2>Cast into the sea</h2>
          <button type="button" class="ghost" onClick={() => props.onClose()} aria-label="Close">
            Close
          </button>
        </header>
        <p class="panel-lead">Write something short. Anyone on the ocean may find it.</p>
        <form onSubmit={submit}>
          <label class="field">
            <span>Message</span>
            <textarea
              maxlength={1000}
              rows={4}
              value={message()}
              onInput={e => setMessage(e.currentTarget.value)}
              placeholder="A note for whoever finds this…"
              required
            />
          </label>
          <label class="field">
            <span>Bottle style</span>
            <div class="style-row">
              {STYLES.map(s => (
                <button
                  type="button"
                  classList={{ "style-chip": true, active: style() === s.id }}
                  onClick={() => setStyle(s.id)}
                >
                  {s.label}
                </button>
              ))}
            </div>
          </label>
          <Show when={error()}>
            <p class="error">{error()}</p>
          </Show>
          <button type="submit" class="primary" disabled={busy()}>
            {busy() ? "Casting…" : "Release bottle"}
          </button>
        </form>
      </div>
    </Show>
  );
}
