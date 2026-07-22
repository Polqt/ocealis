import { createEffect, onCleanup, onMount } from "solid-js";
import * as THREE from "three";
import gsap from "gsap";
import type { Bottle } from "~/lib/types";
import { latLngToPlane } from "~/lib/coords";

type Props = {
  bottles: () => Bottle[];
  selectedId: () => number | null;
  journeyPoints: () => { lat: number; lng: number }[];
  onSelect: (id: number) => void;
  castPulse: () => number;
};

const STYLE_COLORS = [
  0xd4a574, 0xe8c4a0, 0xc9845a, 0xb08968, 0x8d6e63, 0xa5b4a3, 0x7ea8be, 0xcbd5e1, 0xf0e6d2, 0x9aa6b2
];

type OceanApi = {
  syncBottles: (bottles: Bottle[], selectedId: number | null) => void;
  syncJourney: (points: { lat: number; lng: number }[]) => void;
  pulseCast: () => void;
};

function makeMessageBottle(styleColor: number, scale = 1) {
  const group = new THREE.Group();

  const glass = new THREE.Mesh(
    new THREE.CylinderGeometry(0.22 * scale, 0.28 * scale, 0.9 * scale, 12),
    new THREE.MeshStandardMaterial({
      color: styleColor,
      roughness: 0.25,
      metalness: 0.35,
      transparent: true,
      opacity: 0.92,
      emissive: 0x1a3040,
      emissiveIntensity: 0.25
    })
  );
  glass.position.y = 0.45 * scale;
  group.add(glass);

  const neck = new THREE.Mesh(
    new THREE.CylinderGeometry(0.1 * scale, 0.14 * scale, 0.28 * scale, 10),
    new THREE.MeshStandardMaterial({
      color: styleColor,
      roughness: 0.3,
      metalness: 0.3,
      transparent: true,
      opacity: 0.95
    })
  );
  neck.position.y = 1.05 * scale;
  group.add(neck);

  const cork = new THREE.Mesh(
    new THREE.CylinderGeometry(0.11 * scale, 0.11 * scale, 0.12 * scale, 10),
    new THREE.MeshStandardMaterial({ color: 0xc4a574, roughness: 0.8, metalness: 0.05 })
  );
  cork.position.y = 1.25 * scale;
  group.add(cork);

  const paper = new THREE.Mesh(
    new THREE.PlaneGeometry(0.22 * scale, 0.35 * scale),
    new THREE.MeshStandardMaterial({ color: 0xf2e6c8, roughness: 0.9, side: THREE.DoubleSide })
  );
  paper.position.set(0.02 * scale, 0.45 * scale, 0.2 * scale);
  paper.rotation.y = 0.2;
  group.add(paper);

  group.rotation.z = Math.PI / 12;
  return group;
}

export default function OceanCanvas(props: Props) {
  let host!: HTMLDivElement;
  let api: OceanApi | null = null;

  createEffect(() => {
    api?.syncBottles(props.bottles(), props.selectedId());
  });

  createEffect(() => {
    api?.syncJourney(props.journeyPoints());
  });

  createEffect(() => {
    const pulse = props.castPulse();
    if (pulse) api?.pulseCast();
  });

  onMount(() => {
    const width = () => host.clientWidth || window.innerWidth;
    const height = () => host.clientHeight || window.innerHeight;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x061a28);
    scene.fog = new THREE.FogExp2(0x072033, 0.012);

    const camera = new THREE.PerspectiveCamera(42, width() / height(), 0.1, 400);
    // Low angle across a vast sea — horizon reads as infinite.
    camera.position.set(0, 14, 42);
    camera.lookAt(0, 1.2, 0);

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setSize(width(), height());
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.15;
    host.appendChild(renderer.domElement);

    scene.add(new THREE.AmbientLight(0x6ea8b8, 0.55));
    const sun = new THREE.DirectionalLight(0xffe2c4, 1.35);
    sun.position.set(30, 50, 18);
    scene.add(sun);
    const fill = new THREE.DirectionalLight(0x3d7a8c, 0.45);
    fill.position.set(-25, 12, -10);
    scene.add(fill);

    // Sky dome
    const sky = new THREE.Mesh(
      new THREE.SphereGeometry(180, 32, 16),
      new THREE.ShaderMaterial({
        side: THREE.BackSide,
        depthWrite: false,
        uniforms: {
          topColor: { value: new THREE.Color(0x0b3044) },
          bottomColor: { value: new THREE.Color(0x163d4f) }
        },
        vertexShader: `
          varying vec3 vWorld;
          void main() {
            vec4 w = modelMatrix * vec4(position, 1.0);
            vWorld = w.xyz;
            gl_Position = projectionMatrix * viewMatrix * w;
          }
        `,
        fragmentShader: `
          uniform vec3 topColor;
          uniform vec3 bottomColor;
          varying vec3 vWorld;
          void main() {
            float h = normalize(vWorld).y;
            gl_FragColor = vec4(mix(bottomColor, topColor, max(h, 0.0)), 1.0);
          }
        `
      })
    );
    scene.add(sky);

    // Vast ocean surface
    const oceanGeo = new THREE.PlaneGeometry(220, 220, 140, 140);
    const oceanMat = new THREE.MeshStandardMaterial({
      color: 0x0a5568,
      roughness: 0.28,
      metalness: 0.22,
      flatShading: true
    });
    const ocean = new THREE.Mesh(oceanGeo, oceanMat);
    ocean.rotation.x = -Math.PI / 2;
    scene.add(ocean);

    // Soft sun path on water
    const glitter = new THREE.Mesh(
      new THREE.CircleGeometry(18, 48),
      new THREE.MeshBasicMaterial({
        color: 0xd9c39a,
        transparent: true,
        opacity: 0.12,
        depthWrite: false
      })
    );
    glitter.rotation.x = -Math.PI / 2;
    glitter.position.set(8, 0.05, -6);
    scene.add(glitter);

    // Center hero bottle — brand focal point / cast ritual anchor
    const hero = makeMessageBottle(0xe8c4a0, 2.4);
    hero.position.set(0, 0.2, 0);
    hero.userData.isHero = true;
    scene.add(hero);
    gsap.to(hero.rotation, { y: Math.PI * 2, duration: 28, repeat: -1, ease: "none" });
    gsap.to(hero.position, { y: 0.55, duration: 2.8, yoyo: true, repeat: -1, ease: "sine.inOut" });

    const bottleGroup = new THREE.Group();
    scene.add(bottleGroup);

    const meshes = new Map<number, THREE.Object3D>();
    const raycaster = new THREE.Raycaster();
    const pointer = new THREE.Vector2();

    const journeyLine = new THREE.Line(
      new THREE.BufferGeometry(),
      new THREE.LineBasicMaterial({ color: 0xf2d7a1, transparent: true, opacity: 0.85 })
    );
    scene.add(journeyLine);

    const makeDriftBottle = (style: number) => {
      const bottle = makeMessageBottle(STYLE_COLORS[style % STYLE_COLORS.length], 1);
      bottle.scale.setScalar(1);
      return bottle;
    };

    api = {
      syncBottles(bottles, selectedId) {
        const seen = new Set<number>();
        for (const b of bottles) {
          seen.add(b.id);
          const { x, z } = latLngToPlane(b.current_lat, b.current_lng, 90);
          let mesh = meshes.get(b.id);
          if (!mesh) {
            mesh = makeDriftBottle(b.bottle_style);
            mesh.userData.bottleId = b.id;
            meshes.set(b.id, mesh);
            bottleGroup.add(mesh);
            mesh.position.set(x, 0.4, z);
            gsap.from(mesh.scale, { x: 0.01, y: 0.01, z: 0.01, duration: 0.7, ease: "back.out(1.6)" });
          } else {
            gsap.to(mesh.position, { x, z, duration: 1.2, ease: "sine.inOut" });
          }
          const selected = selectedId === b.id;
          const s = selected ? 1.45 : 1;
          mesh.scale.setScalar(s);
        }
        for (const [id, mesh] of meshes) {
          if (!seen.has(id)) {
            bottleGroup.remove(mesh);
            mesh.traverse(obj => {
              const m = obj as THREE.Mesh;
              if (m.geometry) m.geometry.dispose();
              if (m.material) {
                const mat = m.material;
                if (Array.isArray(mat)) mat.forEach(x => x.dispose());
                else mat.dispose();
              }
            });
            meshes.delete(id);
          }
        }
      },
      syncJourney(pts) {
        if (pts.length < 2) {
          journeyLine.visible = false;
          return;
        }
        const positions = new Float32Array(pts.length * 3);
        pts.forEach((p, i) => {
          const { x, z } = latLngToPlane(p.lat, p.lng, 90);
          positions[i * 3] = x;
          positions[i * 3 + 1] = 0.9;
          positions[i * 3 + 2] = z;
        });
        journeyLine.geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
        journeyLine.geometry.computeBoundingSphere();
        journeyLine.visible = true;
      },
      pulseCast() {
        gsap.fromTo(hero.scale, { x: 1, y: 1, z: 1 }, { x: 1.2, y: 1.2, z: 1.2, duration: 0.45, yoyo: true, repeat: 1 });
        gsap.fromTo(
          camera.position,
          { y: 14, z: 42 },
          { y: 11, z: 34, duration: 1.5, ease: "power2.out", yoyo: true, repeat: 1 }
        );
      }
    };

    api.syncBottles(props.bottles(), props.selectedId());
    api.syncJourney(props.journeyPoints());

    let frame = 0;
    const tick = () => {
      frame = requestAnimationFrame(tick);
      const t = performance.now() * 0.001;
      const pos = oceanGeo.attributes.position as THREE.BufferAttribute;
      for (let i = 0; i < pos.count; i++) {
        const x = pos.getX(i);
        const y = pos.getY(i);
        const wave =
          Math.sin(x * 0.12 + t * 0.9) * 0.55 +
          Math.cos(y * 0.1 + t * 0.7) * 0.4 +
          Math.sin((x + y) * 0.08 + t * 1.3) * 0.2;
        pos.setZ(i, wave);
      }
      pos.needsUpdate = true;
      oceanGeo.computeVertexNormals();

      for (const mesh of meshes.values()) {
        const id = mesh.userData.bottleId as number;
        mesh.position.y = 0.35 + Math.sin(t * 1.8 + id) * 0.18;
        mesh.rotation.y = Math.sin(t * 0.6 + id) * 0.35;
        mesh.rotation.z = Math.PI / 14 + Math.sin(t + id) * 0.05;
      }

      (glitter.material as THREE.MeshBasicMaterial).opacity = 0.08 + Math.sin(t * 0.7) * 0.04;
      renderer.render(scene, camera);
    };
    tick();

    const onResize = () => {
      camera.aspect = width() / height();
      camera.updateProjectionMatrix();
      renderer.setSize(width(), height());
    };
    window.addEventListener("resize", onResize);

    const onClick = (ev: MouseEvent) => {
      const rect = renderer.domElement.getBoundingClientRect();
      pointer.x = ((ev.clientX - rect.left) / rect.width) * 2 - 1;
      pointer.y = -((ev.clientY - rect.top) / rect.height) * 2 + 1;
      raycaster.setFromCamera(pointer, camera);
      const targets: THREE.Object3D[] = [];
      for (const mesh of meshes.values()) {
        mesh.traverse(o => {
          if ((o as THREE.Mesh).isMesh) targets.push(o);
        });
      }
      const hits = raycaster.intersectObjects(targets, false);
      if (hits[0]) {
        let obj: THREE.Object3D | null = hits[0].object;
        while (obj && obj.userData.bottleId == null) obj = obj.parent;
        if (obj?.userData.bottleId != null) {
          props.onSelect(obj.userData.bottleId as number);
        }
      }
    };
    renderer.domElement.addEventListener("click", onClick);

    onCleanup(() => {
      api = null;
      cancelAnimationFrame(frame);
      window.removeEventListener("resize", onResize);
      renderer.domElement.removeEventListener("click", onClick);
      renderer.dispose();
      oceanGeo.dispose();
      oceanMat.dispose();
      host.replaceChildren();
    });
  });

  return <div class="ocean-canvas" ref={host} aria-label="Living ocean map" />;
}
