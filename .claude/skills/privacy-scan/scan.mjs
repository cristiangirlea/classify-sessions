// Privacy scan. Run from the repository root: node .claude/skills/privacy-scan/scan.mjs
// Prints file:line and a rule name for every hit, never the matched prompt text.
// Exit 0 = clean, 1 = hits, 2 = scan could not run.
import { execFileSync } from 'node:child_process';
import { readFileSync, existsSync } from 'node:fs';
import { userInfo } from 'node:os';
import { firstPrompts } from './first-prompts.mjs';

const git = (...a) => execFileSync('git', a, { encoding: 'utf8' }).trim();
const esc = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

const hits = [];
const hit = (file, line, rule) => hits.push(`${file}:${line}  ${rule}`);

// 1. Nothing under .handoff/ may be tracked.
const handoff = git('ls-files', '--', '.handoff');
if (handoff) for (const f of handoff.split('\n')) hit(f, 0, 'tracked-handoff-file');

// 2. Files to scan: tracked plus untracked-not-ignored, minus fixtures and dependencies.
const files = git('ls-files', '-co', '--exclude-standard')
  .split('\n')
  .filter((f) => f && existsSync(f))
  .filter((f) => !f.startsWith('fixtures/') && !f.includes('node_modules/') && !f.startsWith('.handoff/'))
  .filter((f) => !/\.(png|jpg|ico|woff2?)$/i.test(f) && !f.endsWith('package-lock.json'));

// Identity is read at run time so this file never contains it.
const user = userInfo().username;
let email = '';
try { email = git('config', 'user.email'); } catch {}

const identity = [
  ['windows-user-path', /[A-Za-z]:\\+Users\\+/i],
  ['windows-drive-path', /(^|[^A-Za-z0-9])[A-Za-z]:\\+[A-Za-z_]/],
  ['mac-user-path', /\/Users\//],
  ['home-path', /\/home\//],
  ...(user && user.length >= 3 ? [['os-username', new RegExp(`(^|[^A-Za-z0-9])${esc(user)}([^A-Za-z0-9]|$)`, 'i')]] : []),
  ...(email ? [['owner-email', new RegExp(esc(email), 'i')]] : []),
];
// Files allowed to *name* tool-log locations because they describe or enforce the barrier.
const barrierDocs = new Set([
  'CLAUDE.md', 'AGENTS.md', 'NOTES.md', 'SPEC.md', 'PLAN.md', 'PROMPTS.md',
  '.claude/settings.json', '.claude/hooks/guard-bash.mjs', '.claude/skills/privacy-scan/SKILL.md',
  '.claude/skills/corpus-check/SKILL.md',
  '.claude/skills/privacy-scan/scan.mjs', '.claude/skills/privacy-scan/first-prompts.mjs',
  '.cursor/rules/project.mdc',
]);
const toolLogs = [
  ['claude-projects-dir', /\.claude[\\/]+projects/i],
  ['codex-dir', /\.codex\b/i],
  ['cursor-user-dir', /Cursor[\\/]+User/i],
  ['system-reminder-tag', /<system-reminder>/i],
];
// Code that implements the defensive <system-reminder> stripping may name that tag, and only that tag.
const tagCode = new Set(['backend/internal/ingest/ingest.go', 'backend/internal/model/model.go']);
// Go tests may use the invented drive path C:\data\ to test absolute --input handling; any other
// drive path, and every user path, username and email rule, still applies to them.
const inventedDrivePath = /(^|[^A-Za-z0-9])C:\\+data\\+/g;
const allowedUrl =/^https?:\/\/(127\.0\.0\.1|localhost)([:/]|$)|^https:\/\/docs\.claude\.com\/en\/docs\/about-claude\/pricing/;

// 3. First-prompt fingerprints: first 40 normalised characters of each session's first prompt.
const prompts = [];
for (const [sid, fp] of firstPrompts('fixtures/claude-code')) {
  if (!fp) continue;
  const norm = fp.text.toLowerCase().replace(/\s+/g, ' ').trim();
  if (norm.length >= 20) prompts.push([sid.slice(0, 8), norm.slice(0, 40)]);
}

for (const f of files) {
  const self = f.startsWith('.claude/skills/privacy-scan/');
  const lines = readFileSync(f, 'utf8').split(/\r?\n/);
  lines.forEach((text, i) => {
    const n = i + 1;
    if (!self)
      for (const [rule, re] of identity) {
        const checked = rule === 'windows-drive-path' && /_test\.go$/.test(f) ? text.replace(inventedDrivePath, '$1') : text;
        if (re.test(checked)) hit(f, n, rule);
      }
    if (!barrierDocs.has(f) && !/_test\.go$/.test(f) && !f.startsWith('backend/testdata/'))
      for (const [rule, re] of toolLogs) if (re.test(text) && !(rule === 'system-reminder-tag' && tagCode.has(f))) hit(f, n, rule);
    if (/^(backend\/|frontend\/src\/)/.test(f) && !/_test\.go$/.test(f))
      for (const m of text.matchAll(/https?:\/\/[^\s"'`)<>]+/g)) if (!allowedUrl.test(m[0])) hit(f, n, 'external-url');
    const low = text.toLowerCase().replace(/\s+/g, ' ');
    for (const [sid, p] of prompts) if (low.includes(p)) hit(f, n, `first-prompt-text(session ${sid})`);
  });
}

console.log(`privacy-scan: ${files.length} files scanned, ${prompts.length} first-prompt fingerprints, ${hits.length} hit(s)`);
for (const h of hits) console.log(h);
process.exit(hits.length ? 1 : 0);
