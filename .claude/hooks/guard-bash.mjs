// PreToolUse hook for Bash. Reads the hook JSON from stdin and blocks (exit 2) any command
// whose text names a location outside this repository that must never be touched: AI tool
// logs, other assistants' data, or the earlier attempt. Only tool_input.command is tested;
// the hook payload itself always carries a transcript path under ~/.claude/projects.
const BLOCK = /(\.claude[\\/]+projects|\.codex\b|[\\/]\.cursor\b|AppData[\\/]+Roaming[\\/]+Cursor|Cursor[\\/]+User|[\\/]\.hrn\b|session-insights-verified-final|harness-session-analytics)/i;

let raw = '';
process.stdin.on('data', (d) => (raw += d));
process.stdin.on('end', () => {
  let cmd = '';
  try { cmd = String(JSON.parse(raw).tool_input?.command ?? ''); } catch { process.exit(0); }
  const m = cmd.match(BLOCK);
  if (m) {
    process.stderr.write(`blocked by .claude/hooks/guard-bash.mjs: the command references "${m[0]}", which is outside this repository's allowed inputs\n`);
    process.exit(2);
  }
  process.exit(0);
});
