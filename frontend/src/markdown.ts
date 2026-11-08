// Escape source text before producing our small, inert Markdown subset.
const escape = (text: string) => text.replace(/[&<>"']/g, c => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]!));
function resolvePath(from: string, target: string) {const parts = (target.startsWith('/') ? target.slice(1) : `${from.slice(0, from.lastIndexOf('/') + 1)}${target}`).split('/'); const normalized: string[] = []; for (const part of parts) {if (part === '..') normalized.pop(); else if (part !== '.' && part) normalized.push(part);} return normalized.join('/');}
function inline(text: string, path: string): string {
  const tokens: string[] = [];
  let source = text.replace(/`([^`]+)`/g, (_, code) => `\u0000${tokens.push(`<code>${escape(code)}</code>`) - 1}\u0000`);
  source = source.replace(/\[([^\]]+)\]\(([^\s)]+)\)/g, (_, label, raw) => {
    const target = String(raw); let href = '', attributes = '';
    if (/^https?:\/\//i.test(target)) {href = target; attributes = ' target="_blank" rel="noreferrer"';}
    else if (target.startsWith('#')) href = target;
    else if (!/^[a-z][a-z\d+.-]*:/i.test(target) && !target.startsWith('//')) {const resolved = resolvePath(path, target); if (/\.md(?:#.*)?$/i.test(resolved)) {href = `/docs/${resolved}`; attributes = ` data-doc="${escape(resolved)}"`;} else {href = target.startsWith('/') ? target : `/docs/${resolved}`; if (resolved.startsWith('schemas/') || resolved.startsWith('examples/')) href = `/${resolved}`;}}
    if (!href) return escape(label);
    return `\u0000${tokens.push(`<a href="${escape(href)}"${attributes}>${escape(label)}</a>`) - 1}\u0000`;
  });
  source = escape(source).replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>').replace(/\*([^*]+)\*/g, '<em>$1</em>');
  return source.replace(/\u0000(\d+)\u0000/g, (_, n) => tokens[Number(n)] ?? '');
}
export function markdown(text: string, path: string): string {
  const lines = text.replace(/\r\n/g, '\n').split('\n'); const output: string[] = []; let i = 0;
  while (i < lines.length) {
    const line = lines[i]; if (!line.trim()) {i++; continue;}
    if (/^```/.test(line)) {const code: string[] = []; i++; while (i < lines.length && !/^```/.test(lines[i])) code.push(lines[i++]); i++; output.push(`<pre><code>${escape(code.join('\n'))}</code></pre>`); continue;}
    const heading = /^(#{1,6})\s+(.+)$/.exec(line); if (heading) {const level = heading[1].length; const id = heading[2].toLowerCase().replace(/[^a-z\d\s-]/g, '').replace(/\s+/g, '-'); output.push(`<h${level} id="${escape(id)}">${inline(heading[2], path)}</h${level}>`); i++; continue;}
    if (/^\s*[-*_]{3,}\s*$/.test(line)) {output.push('<hr/>'); i++; continue;}
    if (line.includes('|') && i + 1 < lines.length && /^\s*\|?\s*:?-{3,}/.test(lines[i + 1])) {const cells = (s: string) => s.trim().replace(/^\||\|$/g, '').split('|').map(v => v.trim()); const head = cells(line); output.push(`<div class="doc-table"><table><thead><tr>${head.map(c => `<th>${inline(c, path)}</th>`).join('')}</tr></thead><tbody>`); i += 2; while (i < lines.length && lines[i].includes('|') && lines[i].trim()) output.push(`<tr>${cells(lines[i++]).map(c => `<td>${inline(c, path)}</td>`).join('')}</tr>`); output.push('</tbody></table></div>'); continue;}
    if (/^\s*([-*+] |\d+\. )/.test(line)) {const ordered = /^\s*\d+\. /.test(line); const tag = ordered ? 'ol' : 'ul'; output.push(`<${tag}>`); while (i < lines.length && /^\s*([-*+] |\d+\. )/.test(lines[i])) output.push(`<li>${inline(lines[i++].replace(/^\s*([-*+] |\d+\. )/, ''), path)}</li>`); output.push(`</${tag}>`); continue;}
    if (/^>\s?/.test(line)) {const quoted: string[] = []; while (i < lines.length && /^>/.test(lines[i])) quoted.push(lines[i++].replace(/^>\s?/, '')); output.push(`<blockquote>${inline(quoted.join(' '), path)}</blockquote>`); continue;}
    const paragraph = [line]; i++; while (i < lines.length && lines[i].trim() && !/^(#|```|>|\s*[-*+] |\s*\d+\. )/.test(lines[i]) && !(lines[i].includes('|') && lines[i + 1]?.includes('---'))) paragraph.push(lines[i++]); output.push(`<p>${inline(paragraph.join(' '), path)}</p>`);
  }
  return output.join('\n');
}
