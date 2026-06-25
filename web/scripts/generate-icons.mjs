// SPDX-License-Identifier: Apache-2.0
//
// Generates the PWA PNG icons (no external deps) so the app is installable
// without committing binary blobs that are hard to review. Re-run with:
//   node scripts/generate-icons.mjs
//
// Draws the Konatsu mark: a brand-yellow rounded square with a dark envelope.

import { deflateSync } from 'node:zlib';
import { writeFileSync, mkdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const BRAND = [0xff, 0xd2, 0x0a]; // #ffd20a
const INK = [0x1f, 0x23, 0x29]; // #1f2329
const SS = 4; // supersampling factor for anti-aliasing

const outDir = resolve(dirname(fileURLToPath(import.meta.url)), '../public/icons');
mkdirSync(outDir, { recursive: true });

function lerp(a, b, t) {
  return [
    Math.round(a[0] + (b[0] - a[0]) * t),
    Math.round(a[1] + (b[1] - a[1]) * t),
    Math.round(a[2] + (b[2] - a[2]) * t),
  ];
}

/** Signed coverage [0,1] of a point against a rounded rect (1=inside). */
function insideRoundedRect(x, y, x0, y0, x1, y1, r) {
  const cx = Math.min(Math.max(x, x0 + r), x1 - r);
  const cy = Math.min(Math.max(y, y0 + r), y1 - r);
  if (x >= x0 && x <= x1 && y >= y0 && y <= y1) {
    const dx = x < x0 + r ? x0 + r - x : x > x1 - r ? x - (x1 - r) : 0;
    const dy = y < y0 + r ? y0 + r - y : y > y1 - r ? y - (y1 - r) : 0;
    return dx * dx + dy * dy <= r * r;
  }
  void cx;
  void cy;
  return false;
}

function pointInTriangle(px, py, ax, ay, bx, by, cx, cy) {
  const d = (by - cy) * (ax - cx) + (cx - bx) * (ay - cy);
  const a = ((by - cy) * (px - cx) + (cx - bx) * (py - cy)) / d;
  const b = ((cy - ay) * (px - cx) + (ax - cx) * (py - cy)) / d;
  const c = 1 - a - b;
  return a >= 0 && b >= 0 && c >= 0;
}

/** Color a single sample point. `maskable` fills the whole square (safe zone). */
function sample(x, y, S, maskable) {
  const u = x / S;
  const v = y / S;

  // Background.
  let bgInside;
  if (maskable) {
    bgInside = true; // full-bleed; the launcher applies its own mask
  } else {
    bgInside = insideRoundedRect(u, v, 0.02, 0.02, 0.98, 0.98, 0.22);
  }
  if (!bgInside) return null; // transparent

  // Scale the envelope down a bit for maskable so it stays in the safe zone.
  const k = maskable ? 0.72 : 1;
  const cxp = 0.5;
  const cyp = 0.5;
  const ex0 = cxp + (0.21 - cxp) * k;
  const ex1 = cxp + (0.79 - cxp) * k;
  const ey0 = cyp + (0.31 - cyp) * k;
  const ey1 = cyp + (0.71 - cyp) * k;

  const inBody = insideRoundedRect(u, v, ex0, ey0, ex1, ey1, 0.04 * k);
  if (!inBody) return BRAND;

  // Flap: dark triangle gives way to a brand chevron crease near the top.
  const flapApexY = ey0 + (ey1 - ey0) * 0.55;
  const inFlap = pointInTriangle(u, v, ex0, ey0, cxp, flapApexY, ex1, ey0);
  return inFlap ? BRAND : INK;
}

function renderRGBA(S, maskable) {
  const data = Buffer.alloc(S * S * 4);
  for (let y = 0; y < S; y++) {
    for (let x = 0; x < S; x++) {
      let r = 0;
      let g = 0;
      let b = 0;
      let a = 0;
      for (let sy = 0; sy < SS; sy++) {
        for (let sx = 0; sx < SS; sx++) {
          const c = sample(x + (sx + 0.5) / SS, y + (sy + 0.5) / SS, S, maskable);
          if (c) {
            r += c[0];
            g += c[1];
            b += c[2];
            a += 255;
          }
        }
      }
      const n = SS * SS;
      const i = (y * S + x) * 4;
      // Premultiply-free: average color over covered subsamples.
      const cov = a / n / 255;
      data[i] = cov > 0 ? Math.round(r / (a / 255)) : 0;
      data[i + 1] = cov > 0 ? Math.round(g / (a / 255)) : 0;
      data[i + 2] = cov > 0 ? Math.round(b / (a / 255)) : 0;
      data[i + 3] = Math.round(a / n);
    }
  }
  return data;
}

// --- Minimal PNG encoder ---

const CRC_TABLE = (() => {
  const t = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[n] = c >>> 0;
  }
  return t;
})();

function crc32(buf) {
  let c = 0xffffffff;
  for (let i = 0; i < buf.length; i++) c = CRC_TABLE[(c ^ buf[i]) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}

function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const typeBuf = Buffer.from(type, 'ascii');
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(Buffer.concat([typeBuf, data])), 0);
  return Buffer.concat([len, typeBuf, data, crc]);
}

function encodePNG(rgba, S) {
  const sig = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(S, 0);
  ihdr.writeUInt32BE(S, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // color type RGBA
  // 10,11,12 = compression, filter, interlace = 0

  // Filter type 0 per scanline.
  const stride = S * 4;
  const raw = Buffer.alloc((stride + 1) * S);
  for (let y = 0; y < S; y++) {
    raw[y * (stride + 1)] = 0;
    rgba.copy(raw, y * (stride + 1) + 1, y * stride, y * stride + stride);
  }
  const idat = deflateSync(raw, { level: 9 });

  return Buffer.concat([
    sig,
    chunk('IHDR', ihdr),
    chunk('IDAT', idat),
    chunk('IEND', Buffer.alloc(0)),
  ]);
}

function write(name, S, maskable) {
  const png = encodePNG(renderRGBA(S, maskable), S);
  writeFileSync(resolve(outDir, name), png);
  console.log(`wrote ${name} (${S}x${S}, ${png.length} bytes)`);
}

// --- Favicon: a real .ico (PNG-encoded entries) + PNG favicons in public root ---

const publicRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../public');

/** Build a Vista+ ICO whose entries are PNG-encoded (widely supported). */
function buildICO(sizes) {
  const images = sizes.map((s) => encodePNG(renderRGBA(s, false), s));
  const count = images.length;

  const header = Buffer.alloc(6);
  header.writeUInt16LE(0, 0); // reserved
  header.writeUInt16LE(1, 2); // type: icon
  header.writeUInt16LE(count, 4);

  const dir = Buffer.alloc(16 * count);
  let offset = 6 + 16 * count;
  images.forEach((img, i) => {
    const s = sizes[i];
    const e = 16 * i;
    dir[e] = s >= 256 ? 0 : s; // width (0 means 256)
    dir[e + 1] = s >= 256 ? 0 : s; // height
    dir[e + 2] = 0; // palette colors
    dir[e + 3] = 0; // reserved
    dir.writeUInt16LE(1, e + 4); // color planes
    dir.writeUInt16LE(32, e + 6); // bits per pixel
    dir.writeUInt32LE(img.length, e + 8);
    dir.writeUInt32LE(offset, e + 12);
    offset += img.length;
  });

  return Buffer.concat([header, dir, ...images]);
}

function writePublic(name, buf, label) {
  writeFileSync(resolve(publicRoot, name), buf);
  console.log(`wrote ${name} (${label}, ${buf.length} bytes)`);
}

// Maskable + standard PWA icons.
write('icon-192.png', 192, false);
write('icon-512.png', 512, false);
write('icon-maskable-512.png', 512, true);

// Favicons (public root).
writePublic('favicon-16.png', encodePNG(renderRGBA(16, false), 16), '16x16');
writePublic('favicon-32.png', encodePNG(renderRGBA(32, false), 32), '32x32');
writePublic('favicon.ico', buildICO([16, 32, 48]), 'ico 16/32/48');
