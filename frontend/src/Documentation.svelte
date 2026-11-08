<script lang="ts">
  import { onMount } from 'svelte';
  import { api, downloadJSON } from './api';
  import { markdown } from './markdown';
  export let path = 'configuration/getting-started.md';
  export let onclose: () => void;
  export let oncreate: () => void;
  let title = 'Printer profile guide', content = '', error = '', loading = true, loadedPath = '', sequence = 0;
  let article: HTMLDivElement;
  const chapters = [{path: 'configuration/getting-started.md', label: 'Start with your printer'}, {path: 'configuration/overview.md', label: 'How profiles work'}, {path: 'configuration/printer-schema.md', label: 'Printer field reference'}, {path: 'configuration/material-catalogue.md', label: 'Materials & compatibility'}, {path: 'configuration/supports.md', label: 'Supports & print readiness'}, {path: 'configuration/validation-rules.md', label: 'Validation & uncertainty'}, {path: 'configuration/firmware-compatibility.md', label: 'Firmware boundaries'}, {path: 'configuration/external-ai-guide.md', label: 'Use an external assistant'}, {path: 'configuration/troubleshooting.md', label: 'Troubleshooting'}];
  async function load(next: string) {const token = ++sequence; loading = true; error = ''; loadedPath = next; try {const data = await api<{title: string; content: string}>(`/api/docs?path=${encodeURIComponent(next.split('#')[0])}`); if (token !== sequence) return; title = data.title; content = data.content; article?.scrollTo(0, 0);} catch (e) {if (token === sequence) error = e instanceof Error ? e.message : String(e);} finally {if (token === sequence) loading = false;}}
  function navigate(next: string) {path = next;}
  function follow(event: MouseEvent) {const link = (event.target as HTMLElement).closest('a[data-doc]'); if (link) {event.preventDefault(); navigate(link.getAttribute('data-doc')!);}}
  async function template() {try {downloadJSON(await api('/api/printer-template'), 'my-printer.draft.json');} catch (e) {error = e instanceof Error ? e.message : String(e);}}
  $: if (path !== loadedPath) void load(path);
  $: html = markdown(content, path);
  onMount(() => {article?.focus(); article?.addEventListener('click',follow); return()=>article?.removeEventListener('click',follow);});
</script>

<div class="modal-backdrop documentation-backdrop" role="presentation"><div class="documentation-reader" role="dialog" aria-modal="true" aria-labelledby="guide-title" tabindex="-1">
  <header class="doc-header"><div><div class="eyebrow">Included documentation · available offline</div><h2 id="guide-title">Printer profile guide</h2></div><button class="icon-button" aria-label="Close profile guide" on:click={onclose}>×</button></header>
  <div class="doc-layout"><nav class="doc-navigation" aria-label="Documentation chapters">{#each chapters as chapter}<button class:active={path.split('#')[0] === chapter.path} on:click={() => navigate(chapter.path)}>{chapter.label}</button>{/each}<div class="doc-downloads"><span>FILES TO KEEP</span><a href="/api/profile-kit" download>Complete profile kit ZIP ↓</a><a href="/schemas/printer.schema.json" download>Printer JSON Schema ↓</a><button on:click={template}>Unfilled printer draft ↓</button><button on:click={oncreate}>Create a printer profile →</button></div></nav><div class="doc-content" bind:this={article} role="document" tabindex="-1">{#if loading}<div class="doc-loading"><span class="spinner"></span> Loading local documentation…</div>{:else if error}<div class="alert error" role="alert">{error}<button class="quiet" on:click={() => load(path)}>Try again</button></div>{:else}<div class="doc-breadcrumb">CONFIGURATION / {title}</div><article class="markdown-document">{@html html}</article>{/if}</div></div>
</div></div>
