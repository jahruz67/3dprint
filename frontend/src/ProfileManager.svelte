<script lang="ts">
  import { onMount } from 'svelte';
  import { api, json, downloadJSON } from './api';
  import type { Printer, Filament, Preferences } from './api';
  export let printer: Printer | undefined;
  export let filament: Filament | undefined;
  export let initialKind: 'printer'|'filament'|'optimization' = 'printer';
  export let initialText = '';
  export let initialErrors: string[] = [];
  export let currentPreferences: Preferences | undefined = undefined;
  export let onclose: () => void;
  export let onguide: (path?: string) => void;
  export let onsaved: (kind: 'printer'|'filament'|'optimization', profile: Record<string, unknown>) => void;
  let kind = initialKind;
  let text = initialText;
  let errors: string[] = initialErrors;
  let message = '', busy = false, loading = true;
  let fileInput: HTMLInputElement;
  const preferences = currentPreferences ?? {schemaVersion: '1.0', mode: 'balanced', budget: 64, minWalls: 2, minInfill: .1, supports: true,supportMode:'auto',supportStyle:'auto',supportPlacement:'everywhere',supportAngle:45};
  async function selectKind(value: typeof kind) {
    kind = value; errors = []; message = ''; loading = true;
    try {if (kind === 'printer') text = JSON.stringify(printer ?? await api('/api/printer-template'), null, 2); else if (kind === 'filament') text = JSON.stringify(filament ?? {schemaVersion: '1.0', id: 'my-filament', name: 'My verified filament', material: null, density: null, nozzleTemperature: null, bedTemperature: null, maxVolumetricFlow: null, minLayerHeight: null, maxLayerHeight: null, fanPercent: null}, null, 2); else text = JSON.stringify(preferences, null, 2);}
    catch (e) {errors = [messageOf(e)]; text = '';}
    finally {loading = false;}
  }
  function parse(): Record<string, unknown> {const profile = JSON.parse(text); if (!profile || typeof profile !== 'object' || Array.isArray(profile)) throw new Error('A profile must be a JSON object.'); return profile;}
  async function validate(save = false) {
    busy = true; errors = []; message = '';
    try {const profile = parse(); const result = await api<{valid: boolean; errors: string[]}>('/api/profiles/validate', json({kind, profile})); errors = result.errors ?? []; if (result.valid) {if (save) {await api('/api/profiles', json({kind, profile})); message = 'Validated profile saved and selected.'; onsaved(kind, profile);} else message = 'Schema, machine limits, firmware boundaries, and required evidence passed validation.';}}
    catch (e) {errors = [messageOf(e)];} finally {busy = false;}
  }
  async function createDraft() {loading = true; errors = []; message = ''; try {text = JSON.stringify(kind === 'printer' ? await api('/api/printer-template') : kind === 'filament' ? {schemaVersion: '1.0', id: 'my-filament', name: 'My verified filament', material: null, density: null, nozzleTemperature: null, bedTemperature: null, maxVolumetricFlow: null, minLayerHeight: null, maxLayerHeight: null, fanPercent: null} : preferences, null, 2); message = 'Draft created. Replace every unknown value using your exact machine documentation or calibration before validation.';} catch (e) {errors = [messageOf(e)];} finally {loading = false;}}
  function duplicate() {try {const p = parse(); if (kind !== 'optimization') {p.id = `${p.id ?? 'profile'}-copy-${Date.now().toString(36)}`; p.name = `${p.name ?? 'Profile'} copy`;} text = JSON.stringify(p, null, 2); message = 'Duplicate ready. Check every specification for the new hardware before saving.'; errors = [];} catch (e) {errors = [messageOf(e)];}}
  async function load(file: File | undefined) {if (!file) return; errors = []; message = ''; try {if (file.size > 2 * 1024 * 1024) throw new Error('Profile files must be smaller than 2 MiB.'); text = JSON.stringify(JSON.parse(await file.text()), null, 2); await validate(true);} catch (e) {errors = [messageOf(e)];} finally {fileInput.value = '';}}
  function exportProfile() {try {const p = parse(); downloadJSON(p, `${p.id || kind}${text.includes(': null') ? '.draft' : ''}.json`);} catch (e) {errors = [messageOf(e)];}}
  function messageOf(e: unknown) {return e instanceof Error ? e.message : String(e);}
  onMount(() => {if (initialText) loading = false; else void selectKind(initialKind);});
</script>

<div class="modal-backdrop" role="presentation"><div class="modal profile-modal" role="dialog" aria-modal="true" aria-labelledby="profile-heading" tabindex="-1">
  <header class="modal-header"><div><div class="eyebrow">Your hardware · portable configuration</div><h2 id="profile-heading">Profile workshop</h2></div><button class="icon-button" aria-label="Close profile workshop" on:click={onclose}>×</button></header>
  <div class="profile-tabs"><button class:active={kind === 'printer'} disabled={busy} on:click={() => selectKind('printer')}>Printer hardware</button><button class:active={kind === 'filament'} disabled={busy} on:click={() => selectKind('filament')}>Filament</button><button class:active={kind === 'optimization'} disabled={busy} on:click={() => selectKind('optimization')}>Preferences</button></div>
  <div class="workshop-layout"><div class="workshop-editor"><p class="muted">{kind === 'printer' ? 'Import a profile for your exact printer, or complete the unfilled draft below. A valid profile defines the build plate, firmware, temperatures, motion, and verified extrusion limits.' : kind === 'filament' ? 'Use values from your filament manufacturer and your calibrated process. The active printer determines which material recipes are compatible.' : 'Define your functional constraints separately from the machine and material profiles.'}</p>
    <div class="editor-toolbar"><span class="mono">{kind}.json</span><div><button class="quiet small" disabled={busy} on:click={() => fileInput.click()}>Import JSON</button><button class="quiet small" disabled={busy || loading} on:click={createDraft}>New draft</button><button class="quiet small" disabled={busy || !text} on:click={duplicate}>Duplicate</button><button class="quiet small" disabled={!text} on:click={exportProfile}>Export</button></div></div>
    <input class="hidden" type="file" accept=".json,application/json" bind:this={fileInput} on:change={() => load(fileInput.files?.[0])}/>
    {#if loading}<div class="editor-loading"><span class="spinner"></span> Loading an unfilled profile template…</div>{:else}<textarea class="json-editor" spellcheck="false" aria-label={`${kind} profile JSON`} bind:value={text} on:input={() => {errors = []; message = '';}}></textarea>{/if}
    {#if errors.length}<div class="alert error profile-validation" role="alert"><strong>Profile needs correction</strong><p>It has not been activated. Fix the fields below using documented or calibrated values.</p><ul>{#each errors as error}<li>{error}</li>{/each}</ul></div>{/if}
    {#if message}<div class="alert success" role="status">{message}</div>{/if}
  </div><aside class="workshop-guide"><div class="eyebrow">Complete it with confidence</div><h3>Start with the guide</h3><p>Every field has a unit, accepted values, validation rules, examples, and a source requirement.</p><button class="secondary full" on:click={() => onguide(kind === 'printer' ? 'configuration/getting-started.md' : kind === 'filament' ? 'configuration/filament-schema.md' : 'configuration/optimization-schema.md')}>Read {kind} guide →</button><h3>Unknown means unknown</h3><p>Drafts keep missing hardware values as <code>null</code>. They cannot be used to slice until those values are verified.</p><h3>External assistant workflow</h3><p>You may give the included schema and documentation to an assistant outside this app. Check its sources, import its JSON here, and validate.</p><button class="quiet guide-inline" on:click={() => onguide('configuration/external-ai-guide.md')}>Open the external assistant guide →</button><a href="/api/profile-kit" download>Download complete profile kit ↓</a><a href={`/schemas/${kind === 'optimization' ? 'optimization' : kind}.schema.json`} download>Download JSON Schema ↓</a></aside></div>
  <footer class="modal-footer"><span class="muted">Validated imports are saved and selected automatically. Files stay on this computer.</span><div><button class="secondary" disabled={busy || loading || !text} on:click={() => validate(false)}>Validate</button><button class="primary" disabled={busy || loading || !text} on:click={() => validate(true)}>{busy ? 'Checking…' : kind === 'printer' ? 'Validate & use printer' : 'Validate & save'}</button></div></footer>
</div></div>
