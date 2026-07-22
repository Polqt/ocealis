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
    scene.fog = new THREE.FogExp2(0x0a2a3a, 0.018);

    const camera = new THREE.PerspectiveCamera(48, width() / height(), 0.1, 200);
    camera.position.set(0, 28, 36);
    camera.lookAt(0, 0, 0);

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setSize(width(), height());
    renderer.setClearColor(0x041520, 1);
    host.appendChild(renderer.domElement);

    scene.add(new THREE.AmbientLight(0x7eb6c9, 0.7));
    const sun = new THREE.DirectionalLight(0xffe6c8, 1.1);
    sun.position.set(20, 40, 10);
    scene.add(sun);

    const oceanGeo = new THREE.PlaneGeometry(80, 80, 96, 96);
    const oceanMat = new THREE.MeshStandardMaterial({
      color: 0x0d4a5c,
      roughness: 0.35,
      metalness: 0.15,
      flatShading: true
    });
    const ocean = new THREE.Mesh(oceanGeo, oceanMat);
    ocean.rotation.x = -Math.PI / 2;
    scene.add(ocean);

    const horizon = new THREE.Mesh(
      new THREE.PlaneGeometry(120, 40),
      new THREE.MeshBasicMaterial({ color: 0x1a6b7c, transparent: true, opacity: 0.35 })
    );
    horizon.position.set(0, 8, -40);
    scene.add(horizon);

    const bottleGroup = new THREE.Group();
    scene.add(bottleGroup);

    const meshes = new Map<number, THREE.Mesh>();
    const raycaster = new THREE.Raycaster();
    const pointer = new THREE.Vector2();

    const journeyLine = new THREE.Line(
      new THREE.BufferGeometry(),
      new THREE.LineBasicMaterial({ color: 0xf2d7a1, transparent: true, opacity: 0.85 })
    );
    scene.add(journeyLine);

    const makeBottle = (style: number) => {
      const geo = new THREE.CapsuleGeometry(0.28, 0.55, 4, 8);
      const mat = new THREE.MeshStandardMaterial({
        color: STYLE_COLORS[style % STYLE_COLORS.length],
        roughness: 0.4,
        metalness: 0.2,
        emissive: 0x112222,
        emissiveIntensity: 0.2
      });
      const mesh = new THREE.Mesh(geo, mat);
      mesh.rotation.z = Math.PI / 10;
      return mesh;
    };

    api = {
      syncBottles(bottles, selectedId) {
        const seen = new Set<number>();
        for (const b of bottles) {
          seen.add(b.id);
          const { x, z } = latLngToPlane(b.current_lat, b.current_lng);
          let mesh = meshes.get(b.id);
          if (!mesh) {
            mesh = makeBottle(b.bottle_style);
            mesh.userData.bottleId = b.id;
            meshes.set(b.id, mesh);
            bottleGroup.add(mesh);
            mesh.position.set(x, 0.6, z);
            gsap.from(mesh.scale, { x: 0.01, y: 0.01, z: 0.01, duration: 0.6, ease: "back.out(1.6)" });
          } else {
            gsap.to(mesh.position, { x, z, duration: 1.2, ease: "sine.inOut" });
          }
          const selected = selectedId === b.id;
          mesh.scale.setScalar(selected ? 1.35 : 1);
          (mesh.material as THREE.MeshStandardMaterial).emissiveIntensity = selected ? 0.55 : 0.2;
        }
        for (const [id, mesh] of meshes) {
          if (!seen.has(id)) {
            bottleGroup.remove(mesh);
            mesh.geometry.dispose();
            (mesh.material as THREE.Material).dispose();
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
          const { x, z } = latLngToPlane(p.lat, p.lng);
          positions[i * 3] = x;
          positions[i * 3 + 1] = 0.9;
          positions[i * 3 + 2] = z;
        });
        journeyLine.geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
        journeyLine.geometry.computeBoundingSphere();
        journeyLine.visible = true;
      },
      pulseCast() {
        gsap.fromTo(
          camera.position,
          { y: 30, z: 38 },
          { y: 26, z: 34, duration: 1.4, ease: "power2.out", yoyo: true, repeat: 1 }
        );
      }
    };

    // Initial sync
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
        const wave = Math.sin(x * 0.35 + t * 1.2) * 0.18 + Math.cos(y * 0.28 + t * 0.9) * 0.12;
        pos.setZ(i, wave);
      }
      pos.needsUpdate = true;
      oceanGeo.computeVertexNormals();

      for (const mesh of meshes.values()) {
        mesh.position.y = 0.55 + Math.sin(t * 2 + mesh.userData.bottleId) * 0.12;
        mesh.rotation.y = Math.sin(t + mesh.userData.bottleId) * 0.2;
      }
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
      const hits = raycaster.intersectObjects([...meshes.values()]);
      if (hits[0]) {
        props.onSelect(hits[0].object.userData.bottleId as number);
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
