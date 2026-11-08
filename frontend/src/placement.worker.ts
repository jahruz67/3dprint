/// <reference lib="webworker" />
import type { Printer, Transform, Vec3 } from './api';
self.onmessage = (event: MessageEvent<{vertices: number[]; transform: Transform; printer: Printer}>) => {
  const {vertices, transform: t, printer: p} = event.data;
  const c = t.rotation.map(v => Math.cos(v * Math.PI / 180)), s = t.rotation.map(v => Math.sin(v * Math.PI / 180));
  const min: Vec3 = [Infinity, Infinity, Infinity], max: Vec3 = [-Infinity, -Infinity, -Infinity];
  const rotated = new Float64Array(vertices.length);
  for (let i = 0; i + 2 < vertices.length; i += 3) {let x = vertices[i] * t.scale[0], y = vertices[i + 1] * t.scale[1], z = vertices[i + 2] * t.scale[2]; [y, z] = [y*c[0]-z*s[0], y*s[0]+z*c[0]]; [x, z] = [x*c[1]+z*s[1], -x*s[1]+z*c[1]]; [x, y] = [x*c[2]-y*s[2], x*s[2]+y*c[2]]; const v = [x, y, z]; for (let k = 0; k < 3; k++) {rotated[i+k] = v[k]; min[k] = Math.min(min[k],v[k]); max[k] = Math.max(max[k],v[k]);}}
  const offset = [t.position[0]-(min[0]+max[0])/2, t.position[1]-(min[1]+max[1])/2, t.position[2]-min[2]];
  const b = p.buildVolume, centered = b.origin === 'center', ox = centered ? -b.width/2 : 0, oy = centered ? -b.depth/2 : 0;
  const tolerance = 1e-5;
  const lo = min.map((v,i)=>v+offset[i]), hi = max.map((v,i)=>v+offset[i]); const reasons: string[] = [];
  if (lo[0] < ox-tolerance || hi[0] > ox+b.width+tolerance) reasons.push('Model exceeds the printer’s X build limits.');
  if (lo[1] < oy-tolerance || hi[1] > oy+b.depth+tolerance) reasons.push('Model exceeds the printer’s Y build limits.');
  if (lo[2] < -tolerance || hi[2] > b.height+tolerance) reasons.push('Model exceeds the printer’s Z build limits.');
  if (b.shape === 'circular') {const cx=ox+b.width/2,cy=oy+b.depth/2; for(let i=0;i+2<rotated.length;i+=3) if(((rotated[i]+offset[0]-cx)/(b.width/2))**2+((rotated[i+1]+offset[1]-cy)/(b.depth/2))**2 > 1.00001) {reasons.push('Model extends outside the circular build plate.'); break;}}
  for(const r of b.excludedRegions??[]) if(hi[0]>r.x && lo[0]<r.x+r.width && hi[1]>r.y && lo[1]<r.y+r.depth) reasons.push('Model envelope overlaps an excluded bed region. Reposition it before slicing.');
  self.postMessage({fits: !reasons.length, reasons, dimensions: max.map((v,i)=>v-min[i])});
};
