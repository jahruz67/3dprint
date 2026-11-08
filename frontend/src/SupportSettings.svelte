<script lang="ts">
  import { supportModes } from './api';
  import type { SupportMode, SupportStyle, SupportPlacement } from './api';
  export let mode:SupportMode='auto';
  export let style:SupportStyle='auto';
  export let placement:SupportPlacement='everywhere';
  export let angle=45;
  export let disabled=false;
  export let onchange:()=>void;
  export let onguide:()=>void;
  export let onsave:()=>void;
  $: selectedMode=supportModes.find(item=>item.id===mode)!;
</script>

<section class="settings-section support-settings"><div class="section-number">03</div><div class="support-heading"><h2>Support strategy</h2><button class="quiet small" on:click={onguide}>Guide</button></div><label class="field-label" for="support-mode">Support objective</label><select id="support-mode" bind:value={mode} disabled={disabled} on:change={onchange}>{#each supportModes as option}<option value={option.id}>{option.name}</option>{/each}</select><p class="support-description">{selectedMode.description}</p>{#if style!=='auto'&&mode!=='off'}<p class="support-override-note">The explicit {style} structure overrides automatic structure selection.</p>{/if}
<div class="support-option-grid"><label>Structure<select aria-label="Support structure" bind:value={style} disabled={disabled||mode==='off'} on:change={onchange}><option value="auto">Automatic</option><option value="organic">Organic / tree</option><option value="snug">Snug</option><option value="grid">Grid</option></select></label><label>Placement<select aria-label="Support placement" bind:value={placement} disabled={disabled||mode==='off'} on:change={onchange}><option value="everywhere">Everywhere</option><option value="buildplate">Build plate only</option></select></label></div><label class="field-label support-angle-label" for="support-angle">Overhang threshold <span>{angle}°</span></label><input id="support-angle" type="range" min="20" max="80" step="1" bind:value={angle} disabled={disabled} on:input={onchange}/><p class="hint">Measured from the horizontal. Raising this threshold can mark more surfaces as needing support. Native bridge detection avoids supporting recognized bridge spans.</p><p class="support-caveat">Material use, stability, and ease of removal conflict. Check support contact and access in the toolpath preview; geometry alone cannot establish a safe print.</p><button class="quiet support-save" disabled={disabled} on:click={onsave}>Save current print preferences</button></section>
