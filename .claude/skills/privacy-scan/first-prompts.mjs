// First genuine prompt per session, read from the fixture JSONL.
// Used by scan.mjs as fingerprints; this module never prints prompt text. Rule: earliest
// dated `user` record, not sidechain, not isMeta, whose content is a string or text blocks
// with no tool_result block, after defensive removal of <system-reminder> and <command-*> tags.
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, basename } from 'node:path';

const TAGS = /<(system-reminder|command-[a-z-]+)>[\s\S]*?<\/\1>/g;

function walk(dir) {
  const out = [];
  for (const name of readdirSync(dir).sort()) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) out.push(...walk(p));
    else if (name.endsWith('.jsonl')) out.push(p);
  }
  return out;
}

function promptText(rec) {
  if (rec.type !== 'user' || rec.isSidechain || rec.isMeta) return null;
  const c = rec.message && rec.message.content;
  let text = null;
  if (typeof c === 'string') text = c;
  else if (Array.isArray(c)) {
    if (c.some((b) => b && b.type === 'tool_result')) return null;
    const parts = c.filter((b) => b && b.type === 'text' && typeof b.text === 'string');
    if (parts.length) text = parts.map((b) => b.text).join('\n');
  }
  if (text === null) return null;
  text = text.replace(TAGS, '').trim();
  return text || null;
}

// Returns Map<sessionId, {at, text}> keyed by the file's session id (file name).
export function firstPrompts(root) {
  const res = new Map();
  for (const file of walk(root)) {
    const sid = basename(file, '.jsonl');
    let best = null;
    for (const line of readFileSync(file, 'utf8').split('\n')) {
      if (!line.trim()) continue;
      let rec;
      try { rec = JSON.parse(line); } catch { continue; }
      const at = Date.parse(rec.timestamp);
      if (Number.isNaN(at)) continue;
      const text = promptText(rec);
      if (text && (!best || at < best.at)) best = { at, ts: rec.timestamp, text };
    }
    res.set(sid, best);
  }
  return res;
}
