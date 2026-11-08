<script lang="ts">
  import { onMount } from 'svelte';
  import * as THREE from 'three';
  import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
  import type { Model, Printer, Transform, Preview, Vec3 } from './api';
  export let model: Model | null = null;
  export let vertices: number[] = [];
  export let printer: Printer | undefined;
  export let transform: Transform;
  export let preview: Preview | null = null;
  export let previewMode = false;
  export let layer = 0;
  export let singleLayer = false;
  export let travel = false;
  export let printOrder = 100;
  export let light = false;
  export let riskTriangles:number[]=[];
  export let showRisks=true;
  export let onorient: (rotation: Vec3, note: string) => void = () => {};
  export let onerror: (message: string) => void = () => {};
  let host: HTMLDivElement;
  let ready = false;
  let renderer: THREE.WebGLRenderer;
  let scene: THREE.Scene;
  let camera: THREE.PerspectiveCamera;
  let controls: OrbitControls;
  let bed = new THREE.Group();
  let modelGroup = new THREE.Group();
  let paths = new THREE.Group();
  let mesh: THREE.Mesh | null = null;
  let riskMesh:THREE.Mesh|null=null;
  let currentRisk:number[]|null=null;
  let originOffset:Vec3=[0,0,0];
  let webglError = '';
  let orientationBusy = false;
  let orientWorker: Worker | null = null;
  let previewWorker: Worker | null = null;
  let previewBusy = false;
  let pathEntries: {line: THREE.LineSegments; layers: Uint32Array; ordinals: Uint32Array; kind: number}[] = [];
  const colors = [0x667b90, 0x44d8bd, 0x6996f1, 0xe9b65f, 0xbb8fe8];
  let currentVertices: number[] | null = null;
  let currentPreview: Preview | null = null;
  let currentPrinter: Printer | undefined;
  let frame = 0;
  function render() {if (!ready || frame) return; frame = requestAnimationFrame(() => {frame = 0; renderer.render(scene, camera);});}
  function disposeGroup(group: THREE.Group) {for (const child of [...group.children]) {group.remove(child); if ('geometry' in child) (child.geometry as THREE.BufferGeometry).dispose(); if ('material' in child) {const m = child.material as THREE.Material | THREE.Material[]; if (Array.isArray(m)) m.forEach(x => x.dispose()); else m.dispose();}}}
  function drawBed() {
    disposeGroup(bed);
    const w = printer?.buildVolume.width ?? 220, d = printer?.buildVolume.depth ?? 220;
    const centered = printer?.buildVolume.origin === 'center';
    const ox = centered ? -w / 2 : 0, oy = centered ? -d / 2 : 0;
    const circle = ['circular', 'circle'].includes(printer?.buildVolume.shape ?? '');
    const geometry = circle ? new THREE.CircleGeometry(w / 2, 96) : new THREE.PlaneGeometry(w, d);
    const plate = new THREE.Mesh(geometry, new THREE.MeshStandardMaterial({color: light ? 0xdce7e8 : 0x1a272d, roughness: 1, metalness: 0}));
    plate.position.set(ox + w / 2, oy + d / 2, -0.15); bed.add(plate);
    const grid: number[] = [];
    for (let x = 0; x <= w; x += 10) {const half = circle ? Math.sqrt(Math.max(0, (w / 2) ** 2 - (x - w / 2) ** 2)) : d / 2; grid.push(ox + x, oy + d / 2 - half, 0, ox + x, oy + d / 2 + half, 0);}
    for (let y = 0; y <= d; y += 10) {const half = circle ? Math.sqrt(Math.max(0, (w / 2) ** 2 - (y - d / 2) ** 2)) : w / 2; grid.push(ox + w / 2 - half, oy + y, 0, ox + w / 2 + half, oy + y, 0);}
    const gridGeo = new THREE.BufferGeometry(); gridGeo.setAttribute('position', new THREE.Float32BufferAttribute(grid, 3));
    bed.add(new THREE.LineSegments(gridGeo, new THREE.LineBasicMaterial({color: light ? 0xa7bdc1 : 0x31434b, transparent: true, opacity: 0.65})));
    const box = new THREE.Box3(new THREE.Vector3(ox, oy, 0), new THREE.Vector3(ox + w, oy + d, printer?.buildVolume.height ?? 250));
    const helper = new THREE.Box3Helper(box, light ? 0xb6c9cb : 0x34454d); bed.add(helper);
    for (const r of printer?.buildVolume.excludedRegions ?? []) {const excluded = new THREE.Mesh(new THREE.PlaneGeometry(r.width, r.depth), new THREE.MeshBasicMaterial({color: 0xe87873, opacity: .35, transparent: true})); excluded.position.set(r.x + r.width / 2, r.y + r.depth / 2, .06); bed.add(excluded);}
  }
  function drawModel() {
    clearRisk();currentRisk=null;
    disposeGroup(modelGroup); mesh = null;
    if (!vertices.length) return;
    const geo = new THREE.BufferGeometry(); geo.setAttribute('position', new THREE.Float32BufferAttribute(vertices, 3)); geo.computeBoundingBox();
    const b = geo.boundingBox!;originOffset=[-(b.min.x+b.max.x)/2,-(b.min.y+b.max.y)/2,-b.min.z]; geo.translate(...originOffset); geo.computeVertexNormals();
    mesh = new THREE.Mesh(geo, new THREE.MeshStandardMaterial({color: 0x70cfbe, roughness: .62, metalness: .12, side: THREE.DoubleSide})); modelGroup.add(mesh);
  }
  function clearRisk(){if(!riskMesh)return;riskMesh.removeFromParent();riskMesh.geometry.dispose();(riskMesh.material as THREE.Material).dispose();riskMesh=null;}
  function updateRisk(indices:number[],visible:boolean){if(currentRisk!==indices){clearRisk();currentRisk=indices;if(mesh && indices.length){const valid=indices.filter(i=>Number.isInteger(i)&&i>=0&&(i+1)*9<=vertices.length);const positions=new Float32Array(valid.length*9);valid.forEach((triangle,index)=>{for(let i=0;i<9;i++)positions[index*9+i]=vertices[triangle*9+i]+originOffset[i%3];});const geo=new THREE.BufferGeometry();geo.setAttribute('position',new THREE.BufferAttribute(positions,3));riskMesh=new THREE.Mesh(geo,new THREE.MeshBasicMaterial({color:0xf2b064,side:THREE.DoubleSide,transparent:true,opacity:.7,depthWrite:false,polygonOffset:true,polygonOffsetFactor:-1,polygonOffsetUnits:-1}));mesh.add(riskMesh);}}if(riskMesh)riskMesh.visible=visible;}
  function moveModel() {
    if (!mesh) return;
    // Backend applies the X rotation, then Y, then Z (matrix Rz * Ry * Rx).
    mesh.rotation.set(transform.rotation[0] * Math.PI / 180, transform.rotation[1] * Math.PI / 180, transform.rotation[2] * Math.PI / 180, 'ZYX');
    mesh.scale.set(...transform.scale); mesh.position.set(0, 0, 0); mesh.updateMatrixWorld(true);
    const box = new THREE.Box3().setFromObject(mesh);
    mesh.position.set(transform.position[0] - (box.min.x + box.max.x) / 2, transform.position[1] - (box.min.y + box.max.y) / 2, transform.position[2] - box.min.z);
    modelGroup.visible = !previewMode;
  }
  function drawPaths() {
    previewWorker?.terminate(); previewWorker = null; previewBusy = false;
    disposeGroup(paths); pathEntries = [];
    if (!preview) return;
    previewBusy = true; previewWorker = new Worker(new URL('./preview.worker.ts', import.meta.url), {type: 'module'});
    previewWorker.onmessage = event => {
      if (!ready) return;
      const groups = event.data as {kind: number; positions: Float32Array; layers: Uint32Array; ordinals: Uint32Array}[];
      // One GPU buffer per feature, never one UI object per segment.
      groups.forEach(g => {if (!g.layers.length) return; const kind = g.kind; const geo = new THREE.BufferGeometry(); geo.setAttribute('position', new THREE.BufferAttribute(g.positions, 3)); const line = new THREE.LineSegments(geo, new THREE.LineBasicMaterial({color: colors[kind], transparent: kind === 0, opacity: kind === 0 ? .3 : 1})); paths.add(line); pathEntries.push({line, layers: g.layers, ordinals: g.ordinals, kind});});
      previewBusy = false; previewWorker?.terminate(); previewWorker = null; showLayers(); render();
    };
    previewWorker.onerror = () => {previewBusy = false; previewWorker?.terminate(); previewWorker = null; onerror('The toolpath preview could not be prepared. Try a lower-complexity model.');};
    previewWorker.postMessage(preview.segments);
  }
  function lower(values: ArrayLike<number>, target: number) {let lo = 0, hi = values.length; while (lo < hi) {const mid = (lo + hi) >>> 1; if (values[mid] < target) lo = mid + 1; else hi = mid;} return lo;}
  function showLayers() {
    paths.visible = previewMode;
    for (const entry of pathEntries) {entry.line.visible = entry.kind !== 0 || travel; const from = singleLayer ? lower(entry.layers, layer) : 0; const to = Math.min(lower(entry.layers, layer + 1), lower(entry.ordinals, (preview?.segments.length ?? 0) / 9 * printOrder / 100)); entry.line.geometry.setDrawRange(from * 2, Math.max(0, to - from) * 2);}
  }
  export function fit() {
    if (!ready) return;
    const w = printer?.buildVolume.width ?? 220, d = printer?.buildVolume.depth ?? 220;
    const centered = printer?.buildVolume.origin === 'center';
    const target = model && mesh ? new THREE.Box3().setFromObject(mesh).getCenter(new THREE.Vector3()) : new THREE.Vector3(centered ? 0 : w / 2, centered ? 0 : d / 2, 0);
    const distance = Math.max(w, d, model ? model.bounds.max[2] - model.bounds.min[2] : 0) * 1.25;
    camera.position.copy(target).add(new THREE.Vector3(distance, -distance * 1.3, distance * 1.15)); controls.target.copy(target); controls.update(); render();
  }
  export function view(direction: 'top'|'front'|'iso') {if (!ready) return; if (direction === 'iso') {fit(); return;} const dist = camera.position.distanceTo(controls.target); camera.position.copy(controls.target).add(direction === 'top' ? new THREE.Vector3(0, -.001, dist) : new THREE.Vector3(0, -dist, .001)); controls.update(); render();}
  export function orient() {
    if (!vertices.length || !printer || orientationBusy) return;
    orientationBusy = true;
    orientWorker = new Worker(new URL('./orientation.worker.ts', import.meta.url), {type: 'module'});
    orientWorker.onmessage = event => {orientationBusy = false; orientWorker?.terminate(); orientWorker = null; if (event.data.error) onerror(event.data.error); else onorient(event.data.rotation, event.data.note);};
    orientWorker.onerror = () => {orientationBusy = false; orientWorker?.terminate(); onerror('Orientation analysis failed. You can rotate the model manually.');};
    orientWorker.postMessage({vertices, scale: transform.scale, width: printer.buildVolume.width, depth: printer.buildVolume.depth, height: printer.buildVolume.height});
  }
  $: if (ready) {if (currentPrinter !== printer) {currentPrinter = printer; drawBed(); fit();} if (currentVertices !== vertices) {currentVertices = vertices; drawModel(); moveModel(); fit();} if (currentPreview !== preview) {currentPreview = preview; drawPaths();} render();}
  $: if (ready && transform) {moveModel(); render();}
  $: if(ready){vertices;updateRisk(riskTriangles,showRisks);render();}
  $: if (ready) {previewMode; layer; singleLayer; travel; printOrder; showLayers(); if (mesh) modelGroup.visible = !previewMode; render();}
  $: if (ready) {scene.background = new THREE.Color(light ? 0xecf2f1 : 0x10191f); light; drawBed(); render();}
  onMount(() => {
    try {
      scene = new THREE.Scene(); scene.background = new THREE.Color(0x10191f);
      camera = new THREE.PerspectiveCamera(38, 1, .1, 20000); camera.up.set(0, 0, 1);
      renderer = new THREE.WebGLRenderer({antialias: true, alpha: false}); renderer.setPixelRatio(Math.min(devicePixelRatio, 2)); renderer.outputColorSpace = THREE.SRGBColorSpace; host.appendChild(renderer.domElement);
      controls = new OrbitControls(camera, renderer.domElement); controls.enableDamping = false; controls.screenSpacePanning = true; controls.minDistance = 2; controls.maxDistance = 5000; controls.addEventListener('change', render);
      scene.add(new THREE.HemisphereLight(0xe1fff8, 0x344354, 2.2)); const sun = new THREE.DirectionalLight(0xffffff, 2.5); sun.position.set(120, -180, 400); scene.add(sun); scene.add(bed, modelGroup, paths);
      ready = true; drawBed(); drawModel(); moveModel(); drawPaths(); showLayers(); fit();
      const resize = new ResizeObserver(entries => {const {width, height} = entries[0].contentRect; if (!width || !height) return; camera.aspect = width / height; camera.updateProjectionMatrix(); renderer.setSize(width, height); render();}); resize.observe(host);
      return () => {ready = false; cancelAnimationFrame(frame); resize.disconnect(); orientWorker?.terminate(); previewWorker?.terminate(); controls.dispose();clearRisk(); disposeGroup(bed); disposeGroup(modelGroup); disposeGroup(paths); renderer.dispose(); renderer.domElement.remove();};
    } catch (error) {webglError = 'The 3D viewport needs a working WebGL2 graphics driver. Slicing and export remain available.'; onerror(error instanceof Error ? error.message : String(error));}
  });
</script>

<div class="viewport-canvas" bind:this={host} aria-label="Interactive 3D build plate"></div>
{#if webglError}<div class="viewport-notice">{webglError}</div>{/if}
{#if orientationBusy}<div class="viewport-notice"><span class="spinner"></span> Comparing stable orientations…</div>{/if}
{#if previewBusy}<div class="viewport-notice"><span class="spinner"></span> Preparing buffered toolpaths…</div>{/if}
{#if riskTriangles.length && showRisks && !previewMode}<div class="risk-face-legend"><i></i> Sampled support-risk faces</div>{/if}

<style>
  .viewport-canvas {position: absolute; inset: 0; overflow: hidden} .viewport-canvas :global(canvas) {display: block; width: 100%; height: 100%; outline: none}
  .viewport-notice {position: absolute; top: 80px; left: 50%; transform: translateX(-50%); padding: 14px 20px; max-width: 420px; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); color: var(--text); font-size: 12px; display: flex; align-items: center; gap: 10px}
  .risk-face-legend {position:absolute;top:72px;left:22px;display:flex;align-items:center;gap:7px;font-size:9px;color:var(--muted);background:var(--panel);border:1px solid var(--border);padding:7px 9px;border-radius:4px;pointer-events:none}.risk-face-legend i{background:#f2b064;width:7px;height:7px;border-radius:2px}
</style>
