/// <reference lib="webworker" />
export {};
// Group and compact large previews away from the UI thread.
self.onmessage = (event: MessageEvent<number[]>) => {
  const data = event.data;
  const counts = new Uint32Array(5);
  const kindAt = (i: number) => Math.max(0, Math.min(4, Math.round(data[i + 6])));
  for (let i = 0; i + 8 < data.length; i += 9) counts[kindAt(i)]++;
  const groups = Array.from({length: 5}, (_, kind) => ({kind, positions: new Float32Array(counts[kind] * 6), layers: new Uint32Array(counts[kind]), ordinals: new Uint32Array(counts[kind])}));
  const offsets = new Uint32Array(5);
  for (let i = 0; i + 8 < data.length; i += 9) {const kind = kindAt(i), offset = offsets[kind]++, group = groups[kind]; for (let j = 0; j < 6; j++) group.positions[offset * 6 + j] = data[i + j]; group.layers[offset] = data[i + 7]; group.ordinals[offset] = i / 9;}
  const transfers = groups.flatMap(g => [g.positions.buffer, g.layers.buffer, g.ordinals.buffer]);
  self.postMessage(groups, {transfer: transfers});
};
