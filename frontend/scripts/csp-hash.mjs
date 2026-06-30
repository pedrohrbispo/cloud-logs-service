// Recompute the CSP sha256 for the inline no-flash theme script and inject it
// into nginx.conf. Runs as a postbuild step so the served HTML and the CSP
// header never drift (a mismatch would block the script → theme flash/breakage).
import { readFileSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const root = join(here, '..');

// Prefer the built shell (what is actually served); fall back to source.
let html;
try {
  html = readFileSync(join(root, 'dist', 'index.html'), 'utf8');
} catch {
  html = readFileSync(join(root, 'index.html'), 'utf8');
}

const match = html.match(/<script>([\s\S]*?)<\/script>/);
if (!match) {
  console.error('csp-hash: no inline <script> found in index.html');
  process.exit(1);
}

const hash = createHash('sha256').update(match[1], 'utf8').digest('base64');
const token = `sha256-${hash}`;

const nginxPath = join(root, 'nginx.conf');
const conf = readFileSync(nginxPath, 'utf8');
// Replace the placeholder or a previously-injected hash, always re-emitting the
// full `sha256-<hash>` token (keep the prefix).
const updated = conf.replace(/sha256-[A-Za-z0-9+/=_]+/, token);
writeFileSync(nginxPath, updated);

console.log(`csp-hash: inline script CSP token = '${token}' (written to nginx.conf)`);
