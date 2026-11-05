// Detailed 32x48 villager painter, built from parts (light from upper-left).
const SKIN = ['#f3cfa6', '#e2b183', '#c98f5f', '#9c6842', '#7a4e30'];
const HAIR = ['#2b1d14', '#4a2e18', '#7a4a24', '#b07a3a', '#d8b46a', '#8c8c8c', '#e6e0d4'];
const CLOTH = ['#5b6e3a', '#7a3b2e', '#3b4f6b', '#6b5a3a', '#5a3b5e', '#8a6a2a', '#4a584d', '#6e2f3f', '#2f5a5a'];
const SHIRT = ['#e8dcc0', '#d6c8a6', '#c2b9a5', '#e0d2b8'];
const HAIRSTYLE = ['short', 'short', 'long', 'bun', 'bald', 'curly', 'braid'];
const HATS = ['none', 'none', 'none', 'hood', 'straw', 'cap', 'scarf'];
const GARB = ['tunic', 'vest', 'apron', 'shawl', 'tunic'];

function hexRgb(h) { const n = parseInt(h.slice(1), 16); return [n >> 16 & 255, n >> 8 & 255, n & 255]; }
function rgbHex(r, g, b) { return '#' + [r, g, b].map((v) => Math.max(0, Math.min(255, Math.round(v))).toString(16).padStart(2, '0')).join(''); }
// darker shades drift toward blue/purple, lighter toward warm cream: classic pixel-art ramp
function shade(h, f) {
  const [r, g, b] = hexRgb(h);
  if (f < 1) { const k = 1 - f; return rgbHex(r * f + 6 * k, g * f + 2 * k, b * f + 26 * k); }
  const k = f - 1; return rgbHex(r + (255 - r) * k * 0.9, g + (242 - g) * k * 0.85, b + (200 - b) * k * 0.6);
}
function ramp(h) { return { hi: shade(h, 1.28), mid: h, lo: shade(h, 0.72), dk: shade(h, 0.5) }; }

function paintVillager(g, rnd, opts) {
  opts = opts || {};
  const pick = (a) => a[Math.floor(rnd() * a.length)];
  const W = 32, H = 64, TB = 48, grid = Array.from({ length: H }, () => Array(W).fill(null));
  const set = (c, x, y) => { if (x >= 0 && y >= 0 && x < W && y < H) grid[y][x] = c; };
  const rect = (c, x, y, w, h) => { for (let j = 0; j < h; j++) for (let i = 0; i < w; i++) set(c, x + i, y + j); };
  const get = (x, y) => (x >= 0 && y >= 0 && x < W && y < H) ? grid[y][x] : null;

  const sk = ramp(pick(SKIN)), hr = ramp(pick(HAIR)), cl = ramp(pick(CLOTH)), sh = ramp(pick(SHIRT));
  let cl2 = ramp(pick(CLOTH)); if (cl2.mid === cl.mid) cl2 = ramp('#3a2a1c');
  const hairStyle = pick(HAIRSTYLE), hat = pick(HATS), garb = pick(GARB);
  const beard = rnd() < 0.28, older = rnd() < 0.3, eyeCol = pick(['#1a1410', '#2b3a5a', '#3a2a14']);

  // ---- body ----
  // shoulders taper into a seated torso; left side lit, right side in shadow
  for (let y = 23; y < TB; y++) {
    const half = y < 25 ? 9 + (y - 23) * 2 : 12 + Math.min(1, (y - 25) >> 3);
    const x0 = 16 - half, x1 = 15 + half;
    for (let x = x0; x <= x1; x++) set(x - x0 < 3 ? cl.hi : (x1 - x < 4 ? cl.lo : cl.mid), x, y);
  }
  // cloth folds
  for (let k = 0; k < 4; k++) { const fx = 8 + Math.floor(rnd() * 16), fy = 30 + Math.floor(rnd() * 8); for (let y = fy; y < Math.min(H, fy + 6 + Math.floor(rnd() * 6)); y++) set(cl.lo, fx, y); }
  // sleeves and resting hands
  rect(cl.lo, 3, 26, 3, 15); rect(cl.mid, 4, 26, 1, 14); rect(cl.lo, 26, 26, 3, 15); rect(cl.dk, 28, 27, 1, 13);
  rect(cl.dk, 3, 40, 3, 1); rect(cl.dk, 26, 40, 3, 1);
  rect(sk.mid, 3, 41, 4, 3); rect(sk.hi, 3, 41, 2, 1); rect(sk.lo, 6, 42, 1, 2);
  rect(sk.mid, 25, 41, 4, 3); rect(sk.lo, 28, 41, 1, 3); rect(sk.hi, 25, 41, 1, 1);
  // shirt neckline
  for (let y = 23; y < 29; y++) { const w = Math.max(1, 4 - (y - 23) * 0.6); rect(y < 25 ? sh.hi : sh.mid, Math.round(16 - w), y, Math.round(w * 2), 1); }
  set(sh.lo, 16, 27); set(sh.lo, 16, 28);
  if (garb === 'vest') {
    rect(sh.mid, 12, 23, 8, TB - 23); rect(sh.hi, 12, 23, 2, TB - 23); rect(sh.lo, 18, 23, 2, TB - 23);
    for (let y = 23; y < TB; y++) { set(cl2.mid, 11, y); set(cl2.mid, 20, y); set(cl2.hi, 10, y); set(cl2.lo, 21, y); }
    [30, 34, 38].forEach((y) => { set('#c59c47', 12, y); set('#c59c47', 19, y); });
  }
  if (garb === 'apron') {
    rect(sh.mid, 9, 33, 14, TB - 33 + 6); rect(sh.hi, 9, 33, 3, TB - 33 + 6); rect(sh.lo, 20, 33, 3, TB - 33 + 6);
    rect(sh.dk, 9, 33, 14, 1); rect(sh.lo, 13, 41, 5, 3); rect(sh.dk, 13, 41, 5, 1);
    rect(sh.lo, 8, 32, 16, 1);
  }
  if (garb === 'shawl') {
    for (let y = 23; y < 34; y++) { const w = 14 - Math.floor((y - 23) * 1.1); rect(cl2.mid, 16 - w, y, w, 1); rect(cl2.mid, 16, y, w, 1); set(cl2.lo, 16 + w - 1, y); set(cl2.hi, 16 - w, y); }
    for (let x = 4; x < 28; x += 2) set(cl2.dk, x, 33 - Math.abs(16 - x) * 0.4 | 0);
    set('#c59c47', 16, 26); set('#c59c47', 15, 26);
  }
  if (garb === 'tunic' || garb === 'shawl') {
    rect('#3a2614', 5, 37, 22, 2); rect('#5a3c22', 5, 37, 22, 1); rect('#c59c47', 14, 37, 3, 2); set('#8a6a2a', 15, 38);
  }


  // ---- seated legs: lap seen from the front, shins, shoes ----
  const tr = ramp(pick(['#3a2a1c', '#4a3a2a', '#2f3440', '#5a4632', '#3b3a2e'])), skirt = rnd() < 0.3;
  if (skirt) {
    for (let y = TB; y < 61; y++) { const half = 11 + Math.min(3, (y - TB) >> 2); rect(cl.mid, 16 - half, y, half * 2, 1); set(cl.hi, 16 - half, y); set(cl.lo, 15 + half, y); }
    for (let x = 6; x < 27; x += 4) rect(cl.lo, x, 52, 1, 9);
    rect(cl.dk, 3, 60, 26, 1);
  } else {
    rect(tr.mid, 5, TB, 22, 6); rect(tr.hi, 5, TB, 22, 1); rect(tr.lo, 15, TB + 1, 2, 5); rect(tr.lo, 23, TB + 1, 4, 5);
    rect(tr.mid, 8, TB + 6, 6, 8); rect(tr.hi, 8, TB + 6, 2, 8); rect(tr.lo, 12, TB + 6, 2, 8);
    rect(tr.mid, 18, TB + 6, 6, 8); rect(tr.hi, 18, TB + 6, 2, 8); rect(tr.lo, 22, TB + 6, 2, 8);
  }
  const shoe = ramp(pick(['#2b1d14', '#4a2e18', '#3a3a3a']));
  rect(shoe.mid, 7, 61, 8, 3); rect(shoe.hi, 7, 61, 3, 1); rect(shoe.mid, 17, 61, 8, 3); rect(shoe.hi, 17, 61, 3, 1);

  // ---- neck & head ----
  rect(sk.lo, 14, 20, 4, 3); rect(sk.dk, 14, 22, 4, 1);
  const hx = 11, hy = 7, hw = 10, hh = 13;
  rect(sk.mid, hx, hy, hw, hh);
  rect(sk.hi, hx, hy, 2, hh - 2); rect(sk.hi, hx + 2, hy, 3, 2);
  rect(sk.lo, hx + hw - 2, hy + 1, 2, hh - 1); rect(sk.lo, hx + 1, hy + hh - 1, hw - 2, 1);
  [[hx, hy], [hx + hw - 1, hy], [hx, hy + hh - 1], [hx + hw - 1, hy + hh - 1]].forEach(([x, y]) => set(null, x, y));
  set(sk.dk, hx + 2, hy + hh - 1); set(sk.dk, hx + hw - 3, hy + hh - 1);
  // ears
  rect(sk.mid, hx - 1, hy + 5, 1, 3); set(sk.lo, hx - 1, hy + 7); rect(sk.lo, hx + hw, hy + 5, 1, 3);
  // brows, eyes, nose, mouth
  const ey = hy + 6;
  rect(hr.dk, hx + 2, ey - 2, 2, 1); rect(hr.dk, hx + 6, ey - 2, 2, 1);
  rect('#f4efe4', hx + 2, ey, 2, 1); rect('#f4efe4', hx + 6, ey, 2, 1);
  set(eyeCol, hx + 3, ey); set(eyeCol, hx + 6, ey); set(eyeCol, hx + 3, ey - 1); set(eyeCol, hx + 6, ey - 1);
  set(sk.lo, hx + 2, ey + 1); set(sk.lo, hx + 7, ey + 1);
  set(sk.lo, hx + 5, ey + 2); set(sk.dk, hx + 5, ey + 3); set(sk.hi, hx + 4, ey + 2);
  rect(shade(sk.mid, 0.62), hx + 3, ey + 5, 4, 1); set(sk.lo, hx + 4, ey + 6); set(sk.lo, hx + 5, ey + 6);
  if (rnd() < 0.5) { set(shade('#d06a5a', 1), hx + 1, ey + 3); set(shade('#d06a5a', 1), hx + 8, ey + 3); }
  if (older) { set(sk.lo, hx + 1, ey + 2); set(sk.lo, hx + 8, ey + 2); rect(sk.lo, hx + 3, ey - 4, 4, 1); }

  // ---- hair ----
  const top = (y0) => { rect(hr.mid, hx - 1, y0, hw + 2, hy + 3 - y0); rect(hr.hi, hx, y0, 4, 1); rect(hr.hi, hx - 1, y0 + 1, 2, 2); rect(hr.lo, hx + hw - 1, y0 + 1, 2, hy + 3 - y0 - 1); set(null, hx - 1, y0); set(null, hx + hw, y0); };
  if (hairStyle === 'bald') { set(sk.hi, hx + 3, hy + 1); set(sk.hi, hx + 4, hy + 1); rect(hr.mid, hx - 1, hy + 4, 1, 3); rect(hr.lo, hx + hw, hy + 4, 1, 3); }
  else {
    top(hy - 2);
    if (hairStyle === 'curly') for (let x = hx - 1; x <= hx + hw; x += 2) set(hr.mid, x, hy - 3);
    rect(hr.mid, hx - 1, hy + 3, 1, 3); rect(hr.lo, hx + hw, hy + 3, 1, 3);
    if (hairStyle === 'long' || hairStyle === 'braid') { rect(hr.mid, hx - 2, hy + 2, 2, 15); rect(hr.lo, hx + hw, hy + 2, 2, 15); rect(hr.hi, hx - 2, hy + 3, 1, 10); }
    if (hairStyle === 'braid') { for (let y = hy + 16; y < hy + 26; y += 2) { set(hr.mid, hx + hw + 1, y); set(hr.lo, hx + hw + 1, y + 1); } set('#be3424', hx + hw + 1, hy + 26); }
    if (hairStyle === 'bun') { rect(hr.mid, 14, hy - 6, 4, 4); rect(hr.hi, 14, hy - 6, 2, 1); rect(hr.lo, 17, hy - 5, 1, 3); }
    rect(hr.mid, hx + 1, hy, 3, 1); rect(hr.lo, hx + 6, hy, 3, 1);
  }
  if (beard) {
    rect(hr.mid, hx, ey + 3, hw, 5); rect(hr.lo, hx + hw - 2, ey + 3, 2, 5); rect(hr.hi, hx + 1, ey + 3, 2, 3);
    rect(hr.mid, hx + 1, ey + 8, hw - 2, 1); rect(shade(sk.mid, 0.55), hx + 3, ey + 5, 4, 1); rect(hr.lo, hx + 3, ey + 4, 4, 1);
  }

  // ---- headwear ----
  if (hat === 'hood') {
    for (let y = hy - 3; y < hy + 16; y++) { const o = y < hy ? hy - y : 0; rect(cl2.mid, hx - 3 + o, y, 3, 1); rect(cl2.lo, hx + hw - o, y, 3, 1); }
    rect(cl2.mid, hx - 1, hy - 4, hw + 2, 3); rect(cl2.hi, hx, hy - 4, 4, 1); rect(cl2.dk, hx, hy - 1, hw, 1);
    rect(cl2.mid, 6, 21, 20, 3); rect(cl2.hi, 6, 21, 4, 1); rect(cl2.lo, 20, 22, 6, 2);
  }
  if (hat === 'straw') {
    const s = ramp('#d4aa5a');
    rect(s.mid, hx - 6, hy, hw + 12, 2); rect(s.hi, hx - 6, hy, 8, 1); rect(s.lo, hx + hw, hy + 1, 6, 1);
    rect(s.mid, hx, hy - 5, hw, 5); rect(s.hi, hx, hy - 5, 3, 4); rect(s.lo, hx + hw - 2, hy - 4, 2, 4);
    rect('#7a3b2e', hx, hy - 1, hw, 1);
    for (let x = hx - 5; x < hx + hw + 5; x += 3) set(s.lo, x, hy + 1);
  }
  if (hat === 'cap') {
    rect(cl2.mid, hx - 1, hy - 3, hw + 2, 4); rect(cl2.hi, hx, hy - 3, 4, 1); rect(cl2.lo, hx + hw - 2, hy - 2, 3, 3);
    rect(cl2.dk, hx - 1, hy + 1, hw + 3, 1); rect(cl2.lo, hx + hw - 1, hy + 1, 4, 1);
  }
  if (hat === 'scarf') {
    rect(cl2.mid, hx - 1, hy - 3, hw + 2, 5); rect(cl2.hi, hx, hy - 3, 5, 1); rect(cl2.lo, hx + hw - 2, hy - 2, 3, 4);
    for (let x = hx; x < hx + hw; x += 3) set(cl2.hi, x, hy);
    rect(cl2.mid, hx + hw, hy + 1, 2, 3); set(cl2.lo, hx + hw + 2, hy + 4);
  }


  // ---- village chief: teal cloak with gold trim, gold chain and key, black half-mask (from the card) ----
  if (opts.chief) {
    const ck = ramp('#1f6b63'), gd = ramp('#c59c47');
    for (let y = 21; y < TB; y++) {
      const w = y < 24 ? 10 + (y - 21) * 2 : 15;
      rect(ck.mid, 16 - w, y, 5, 1); rect(ck.lo, 11 + w, y, 5, 1); set(ck.hi, 16 - w, y); set(ck.dk, 15 + w, y);
      set(gd.mid, 21 - w, y); set(gd.lo, 10 + w, y);
    }
    rect(ck.mid, 6, 21, 20, 2); rect(ck.hi, 6, 21, 6, 1); rect(gd.mid, 9, 22, 14, 1); rect(gd.hi, 10, 22, 4, 1);
    for (let x = 11; x <= 20; x++) set(gd.mid, x, 24 + Math.round(Math.abs(x - 15.5) * -0.6 + 3));
    rect(gd.mid, 15, 28, 2, 4); rect(gd.hi, 15, 28, 1, 2); rect(gd.lo, 15, 32, 2, 1); set(gd.mid, 17, 31); set(gd.mid, 17, 30);
    rect('#1a1410', hx, ey - 1, hw, 3); rect('#2b2420', hx + 1, ey - 1, hw - 2, 1);
    rect('#f4efe4', hx + 2, ey, 2, 1); rect('#f4efe4', hx + 6, ey, 2, 1); set(eyeCol, hx + 3, ey); set(eyeCol, hx + 6, ey);
    rect(gd.mid, hx - 1, ey - 1, 1, 1); rect(gd.mid, hx + hw, ey - 1, 1, 1);
  }

  // ---- paint + 1px outline (selective: darker outline at bottom/right) ----
  for (let y = 0; y < H; y++) for (let x = 0; x < W; x++) { const c = grid[y][x]; if (c) { g.fillStyle = c; g.fillRect(x, y, 1, 1); } }
  for (let y = 0; y < H; y++) for (let x = 0; x < W; x++) {
    if (get(x, y)) continue;
    const n = get(x - 1, y) || get(x + 1, y) || get(x, y - 1) || get(x, y + 1);
    if (n) { g.fillStyle = '#140f0b'; g.fillRect(x, y, 1, 1); }
  }
}
