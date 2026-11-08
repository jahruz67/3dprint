export type Vec3 = [number, number, number];
export interface Printer {
  schemaVersion: string; id: string; name: string; manufacturer: string; model: string;
  firmware: {family: string; version: string; gcodeFlavor: string; extrusionMode: string; supportedCommands: string[]; unsupportedCommands: string[]; startGcode: string[]; endGcode: string[]};
  buildVolume: {width: number; depth: number; height: number; shape: string; origin: string; excludedRegions?: {x: number; y: number; width: number; depth: number}[]};
  motion: {maxSpeed: number; maxTravelSpeed: number; maxAcceleration: number};
  extruders: {nozzleDiameter: number; filamentDiameter: number; maxVolumetricFlow: number; flowVerified?: boolean; retractionLength: number; retractionSpeed: number; type: string}[];
  thermal: {minNozzle: number; maxNozzle: number; minBed: number; maxBed: number; hasHeatedBed: boolean; partCooling: boolean};
  capabilities: Record<string, boolean>; provenance: Record<string, string>;
}
export interface Filament {schemaVersion: string; id: string; name: string; material: string; density: number; nozzleTemperature: number; bedTemperature: number; maxVolumetricFlow: number; minLayerHeight: number; maxLayerHeight: number; fanPercent: number;notes?:string[]}
export interface Model {id: string; name: string; triangles: number; bounds: {min: Vec3; max: Vec3}; volumeMm3: number; areaMm2: number; warnings: string[]}
export interface Transform {rotation: Vec3; scale: Vec3; position: Vec3}
export interface Settings {layerHeight: number; lineWidth: number; walls: number; infill: number; infillPattern: string; speed: number; travelSpeed: number; acceleration: number; supports: boolean; supportStyle: string; rotation: Vec3}
export interface Estimate {seconds: number; massG: number; volumeMm3: number; qualityError: number; strengthProxy: number}
export interface Actual {seconds: number; massG: number; filamentMm: number; pathMm: number; supportFilamentMm: number; sliceSeconds: number}
export interface Report {mode: string; selected: Settings; baseline: Settings; predicted: Estimate; baselinePredicted: Estimate; evaluated: number; paretoCount: number; constraints: string[]; assumptions: string[]; changes: string[]; actual?: Actual; baselineActual?: Actual; timeDifferencePercent?: number; massDifferencePercent?: number; warnings?: string[];readiness?:Readiness;supportCandidates?:SupportCandidate[];supportSelectionReason?:string}
export interface Job {id: string; status: 'queued'|'optimizing'|'slicing'|'validating'|'complete'|'cancelled'|'failed'; progress: number; message: string; error?: string; report?: Report; previewReady: boolean}
export interface Preview {segments: number[]; layers: number[]; estimatedSeconds: number; filamentMm: number; massG: number; warnings: string[]}
export interface Engine {available: boolean; path: string; version: string}
export interface MaterialCompatibility {profile: Filament; compatible: boolean; reasons: string[];notes?:string[]}
export type SupportMode = 'auto'|'optimized'|'easy_remove'|'stable'|'off';
export type SupportStyle = 'auto'|'organic'|'snug'|'grid';
export type SupportPlacement = 'everywhere'|'buildplate';
export interface Preferences {schemaVersion?:string;mode:string;budget:number;minWalls:number;minInfill:number;supports:boolean;loadDirection?:Vec3;supportMode?:SupportMode;supportStyle?:SupportStyle;supportPlacement?:SupportPlacement;supportAngle?:number}
export interface ReadinessFinding {code:string;severity:string;title:string;detail:string;action:string}
export interface Readiness {supportsRecommended:boolean;overhangAreaMm2:number;unsupportedAreaMm2:number;bridgeLikeAreaMm2:number;bedContactAreaMm2:number;bedContactRatio:number;maxGapMm:number;samples:number;confidence:string;findings:ReadinessFinding[];assumptions:string[];riskTriangles?:number[]}
export interface SupportCandidate {style:string;seconds:number;massG:number;supportFilamentMm:number;selected:boolean}
export interface AppState {printers: Printer[]; filaments: Filament[]; engine: Engine; version: string; nativeDialogs?: boolean; warnings?: string[]; activePrinterId?: string; preferences?: Preferences}
export const supportModes: {id:SupportMode;name:string;description:string}[] = [
  {id:'auto',name:'Auto · recommended',description:'Use geometry to decide whether supports are needed, then compare available native organic and snug strategies by calculated time and material.'},
  {id:'optimized',name:'Optimized',description:'Explicitly enable supports and compare available native strategies using calculated time and material. Extra slicing computation may be needed.'},
  {id:'easy_remove',name:'Easy removal',description:'Favor a support strategy intended to reduce contact and removal effort. Removal quality still depends on calibration and geometry.'},
  {id:'stable',name:'Stable',description:'Favor stable support coverage and contact. This can use more material and require more cleanup.'},
  {id:'off',name:'Off',description:'Generate no supports. Review every unsupported feature and bridge before using this override.'}
];
export const modes = [
  {id: 'fast', name: 'Fast draft', short: 'Time', icon: '↗', description: 'Minimize calculated print time within machine and printability constraints.'},
  {id: 'balanced', name: 'Balanced', short: 'Balance', icon: '◈', description: 'A weighted trade-off between time, material, quality, and reliability.'},
  {id: 'quality', name: 'High quality', short: 'Quality', icon: '✧', description: 'Prioritize dimensional fidelity and the analytical surface error metric.'},
  {id: 'strength', name: 'Structural strength', short: 'Strength', icon: '▥', description: 'Prioritize wall and infill strength proxies. Mechanical performance requires physical validation.'},
  {id: 'material', name: 'Material saver', short: 'Material', icon: '♧', description: 'Minimize estimated extrusion mass subject to functional shell constraints.'}
] as const;
export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) { const data = await response.json().catch(() => null); throw new Error(data?.error || `Request failed (${response.status})`); }
  if (response.status === 204) return undefined as T;
  return response.json();
}
export const json = (body: unknown): RequestInit => ({method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)});
export function duration(seconds: number) {if (!Number.isFinite(seconds)) return '—'; const minutes = Math.max(1, Math.round(seconds / 60)); return minutes >= 60 ? `${Math.floor(minutes / 60)}h ${minutes % 60}m` : `${minutes}m`;}
export function number(value: number, digits = 1) {return Number.isFinite(value) ? value.toLocaleString(undefined, {maximumFractionDigits: digits}) : '—';}
export function downloadJSON(profile: unknown, name: string) {const url = URL.createObjectURL(new Blob([JSON.stringify(profile, null, 2)], {type: 'application/json'})); const link = document.createElement('a'); link.href = url; link.download = name; link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000);}
