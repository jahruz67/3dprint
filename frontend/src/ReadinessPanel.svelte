<script lang="ts">
  import { onMount } from 'svelte';
  import { api, json, number } from './api';
  import type { Model, Printer, Transform, Readiness } from './api';
  export let model: Model;
  export let printer: Printer;
  export let transform: Transform;
  export let supportAngle = 45;
  export let supportMode = 'auto';
  export let disabled = false;
  export let onreport: (report: Readiness | null) => void;
  export let onaction: (action:string) => void;
  export let onguide: () => void;
  export let showRisks = false;
  let report: Readiness | null = null;
  let loading = true, error = '', requestedKey = '', sequence = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let controller: AbortController | null = null;
  $: key = JSON.stringify({modelId:model.id,printerId:printer.id,printerRevision:printer,transform,supportAngle:Number(supportAngle)});
  $: if(key !== requestedKey) schedule(key);
  function schedule(next:string) {requestedKey=next;const token=++sequence;if(timer)clearTimeout(timer);controller?.abort();report=null;onreport(null);loading=true;error='';timer=setTimeout(()=>load(next,token),300);}
  async function load(next:string,token:number) {const snapshot=JSON.parse(next);controller=new AbortController();try {const result=await api<Readiness>(`/api/models/${encodeURIComponent(snapshot.modelId)}/analyze`,{...json({printerId:snapshot.printerId,transform:snapshot.transform,supportAngle:snapshot.supportAngle}),signal:controller.signal});if(token!==sequence)return;report=result;onreport(result);}catch(e){if(token===sequence && !(e instanceof DOMException && e.name==='AbortError'))error=e instanceof Error?e.message:String(e);}finally{if(token===sequence)loading=false;}}
  const actionLabels:Record<string,string>={supports:'Use recommended supports',orient:'Try another orientation',place_on_bed:'Place on bed',inspect:'Inspect the model',brim:'Review bed adhesion',scale:'Review model scale'};
  const actionLabel=(action:string)=>actionLabels[action]??'Review finding';
  onMount(()=>()=>{sequence++;if(timer)clearTimeout(timer);controller?.abort();});
</script>

<section class="readiness-panel" aria-label="Print readiness"><header><div><div class="eyebrow">MODEL ANALYSIS</div><h2>Print readiness</h2></div><button class="quiet small" on:click={onguide}>Guide</button></header>
{#if loading}<div class="readiness-loading" role="status"><span class="spinner"></span> Checking overhangs, bridges, and bed contact…</div>
{:else if error}<div class="readiness-error" role="alert"><p>Geometry guidance is unavailable: {error}</p><button class="quiet small" on:click={()=>schedule(key)}>Retry analysis</button></div>
{:else if report}<div class="readiness-recommendation" class:needs-supports={report.supportsRecommended}><span>{report.supportsRecommended?'△':'✓'}</span><div><strong>{report.supportsRecommended?'Supports recommended':'No supports indicated by this analysis'}</strong><small>{supportMode==='off'&&report.supportsRecommended?'Supports are currently OFF. Review this override before slicing.':`Local geometric estimate · ${report.confidence || 'sampled analysis'}`}</small></div></div><div class="readiness-metrics"><div><span>Unsupported area</span><strong>{number(report.unsupportedAreaMm2)} <small>mm²</small></strong></div><div><span>Bridge-like area</span><strong>{number(report.bridgeLikeAreaMm2)} <small>mm²</small></strong></div><div><span>Bed contact</span><strong>{number(report.bedContactAreaMm2)} <small>mm²</small></strong></div><div><span>Largest sampled gap</span><strong>{number(report.maxGapMm)} <small>mm</small></strong></div></div>
<div class="readiness-findings">{#each report.findings??[] as finding}<div class:warning={finding.severity==='warning'||finding.severity==='error'}><strong>{finding.title}</strong><p>{finding.detail}</p>{#if finding.action}<button class="quiet small" disabled={disabled} on:click={()=>onaction(finding.action)}>{actionLabel(finding.action)} →</button>{/if}</div>{/each}</div>
{#if report.supportsRecommended && supportMode==='off'}<button class="secondary full small" disabled={disabled} on:click={()=>onaction('supports')}>Use recommended supports</button>{/if}
{#if report.riskTriangles?.length}<label class="check-line readiness-risk-toggle"><input type="checkbox" bind:checked={showRisks}/> Highlight sampled risk faces</label>{/if}
<details class="readiness-assumptions"><summary>Analysis limits & assumptions</summary><p>{number(report.samples,0)} samples. Overhang threshold {supportAngle}°. Bed-contact ratio {number(report.bedContactRatio*100)}%. Overhang area {number(report.overhangAreaMm2)} mm².</p>{#each report.assumptions??[] as assumption}<p>{assumption}</p>{/each}<p>This geometric guidance cannot guarantee adhesion, bridging, stability, support removal, or a successful physical print.</p></details>{/if}
</section>
