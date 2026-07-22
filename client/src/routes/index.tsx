import { clientOnly } from "@solidjs/start";

const OceanMap = clientOnly(() => import("~/components/OceanMap"));

export default function Home() {
  return (
    <main class="ocean-page">
      <OceanMap
        fallback={
          <div class="ocean-fallback" role="status">
            Loading Ocean…
          </div>
        }
      />
    </main>
  );
}
