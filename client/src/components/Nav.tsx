import { A } from "@solidjs/router";

export default function Nav() {
  return (
    <header class="ocean-nav">
      <A href="/" class="ocean-nav__brand">
        Ocealis
      </A>
      <A href="/cast" class="ocean-nav__cast">
        Cast
      </A>
    </header>
  );
}
