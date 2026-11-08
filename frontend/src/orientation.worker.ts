/// <reference lib="webworker" />
// Bounded face-normal search; a geometric support-area proxy, not a global optimum.
import { Euler, Quaternion, Vector3 } from 'three';
self.onmessage = (event: MessageEvent<{vertices: number[]; scale: [number, number, number]; width: number; depth: number; height: number}>) => {
  try {
    const {vertices, scale, width, depth, height} = event.data;
    const step = Math.max(1, Math.ceil(vertices.length / 9 / 24000));
    const faces: {n: Vector3; area: number; v: Vector3[]}[] = [];
    for (let i = 0; i + 8 < vertices.length; i += 9 * step) {
      const v = [0, 3, 6].map(o => new Vector3(vertices[i + o] * scale[0], vertices[i + o + 1] * scale[1], vertices[i + o + 2] * scale[2]));
      const cross = new Vector3().subVectors(v[1], v[0]).cross(new Vector3().subVectors(v[2], v[0]));
      if (cross.lengthSq() > 1e-14) faces.push({n: cross.clone().normalize(), area: cross.length() / 2, v});
    }
    const candidates = [new Quaternion()];
    for (const face of [...faces].sort((a, b) => b.area - a.area)) {
      const q = new Quaternion().setFromUnitVectors(face.n, new Vector3(0, 0, -1));
      if (candidates.every(c => Math.abs(c.dot(q)) < 0.995)) candidates.push(q);
      if (candidates.length >= 24) break;
    }
    let best = candidates[0]; let bestScore = Infinity;
    for (const q of candidates) {
      const min = new Vector3(Infinity, Infinity, Infinity), max = new Vector3(-Infinity, -Infinity, -Infinity);
      let support = 0;
      for (const face of faces) {
        const n = face.n.clone().applyQuaternion(q);
        if (n.z < -Math.SQRT1_2) support += face.area * -n.z;
        for (const point of face.v) {const p = point.clone().applyQuaternion(q); min.min(p); max.max(p);}
      }
      const size = max.clone().sub(min);
      if (size.x > width || size.y > depth || size.z > height) continue;
      // Reward a large contact face; sampled support proxy omits the grounded face.
      let contact = 0;
      for (const face of faces) if (face.v.every(p => p.clone().applyQuaternion(q).z < min.z + 0.05)) contact += face.area;
      const score = Math.max(0, support - contact) + size.z * 0.1;
      if (score < bestScore) {best = q; bestScore = score;}
    }
    if (!Number.isFinite(bestScore)) throw new Error('No candidate orientation fits the selected build volume. Scale the model or choose another printer.');
    const e = new Euler().setFromQuaternion(best, 'ZYX');
    self.postMessage({rotation: [e.x, e.y, e.z].map(v => v * 180 / Math.PI), note: `Compared ${candidates.length} face-based orientations using a sampled overhang/contact-area proxy. Inspect stability and support needs before printing.`});
  } catch (error) {self.postMessage({error: error instanceof Error ? error.message : String(error)});}
};
