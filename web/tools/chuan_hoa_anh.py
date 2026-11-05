#!/usr/bin/env python3
"""Chuẩn hoá bộ ảnh gốc Ma Sói -> thư mục assets/ (WebP, tên chuẩn) + manifest.json.

  python3 chuan_hoa_anh.py <thư-mục-gốc MasoiPng> <thư-mục-ra> [thư-mục khung UI]

Thư mục khung UI (tuỳ chọn) chứa khung-*.png (khung 9 lát + hoa văn góc), chép sang ui/ dạng WebP không nén mất.

Bộ gốc không bị sửa. Chạy lại bao nhiêu lần cũng được (thư mục ra bị ghi đè).
"""
import json, os, shutil, sys
from PIL import Image

SRC, OUT = sys.argv[1], sys.argv[2]
UI = sys.argv[3] if len(sys.argv) > 3 else None
H = 360  # chiều cao chuẩn của ảnh tay

# Mã vai backend -> thư mục gốc, tên thư mục mới
VAI = {
    'Villager':    ('Dan_Lang',        'dan-lang'),
    'Werewolf':    ('Soi_Thuong',      'soi'),
    'CursedWolf':  ('Soi_Nguyen',      'soi-nguyen'),
    'WolfSeer':    ('Soi_Tien_Tri',    'soi-tien-tri'),
    'Seer':        ('Tien_Tri',        'tien-tri'),
    'Witch':       ('Phu_Thuy',        'phu-thuy'),
    'Guard':       ('Bao_Ve',          'bao-ve'),
    'Visitor':     ('Nguyet_Nu',       'nguyet-nu'),
    'Detective':   ('Tham_Tu',         'tham-tu'),
    'Cupid':       ('Than_Cupid',      'cupid'),
    'Hunter':      ('Tho_San',         'tho-san'),
    'RustyKnight': ('Hiep_Si_Kiem_Gi', 'hiep-si'),
}
DANH_HIEU = {'chief': ('Thi_Truong', 'truong-lang')}

# Ảnh tay: (thư mục gốc, file gốc) -> (thư mục ra, tên ra không đuôi)
# Trạng thái "còn dùng được" không hậu tố; đã dùng = "-da-dung".
TAY = {
    ('Dan_Lang', 'dan-lang-tay-trai-dem.png'):        ('vai/dan-lang', 'tay-trai'),
    ('Dan_Lang', 'dan-lang-tay-phai-dem.png'):        ('vai/dan-lang', 'tay-phai'),
    ('Dan_Lang', 'dan-lang-tay-trai-chua-vote.png'):  ('chung', 'tay-ngay-trai'),
    ('Dan_Lang', 'dan-lang-tay-trai-da-vote.png'):    ('chung', 'tay-ngay-trai-da-vote'),
    ('Dan_Lang', 'dan-lang-tay-phai-chua-vote.png'):  ('chung', 'tay-ngay-phai'),
    ('Dan_Lang', 'dan-lang-tay-phai-da-vote.png'):    ('chung', 'tay-ngay-phai-da-vote'),
    ('Soi_Thuong', 'Tay_Trai_Soi_Thuong .jpg'):       ('vai/soi', 'tay-trai'),
    ('Soi_Thuong', 'Tay_Phai_Soi_Thuong.png'):        ('vai/soi', 'tay-phai'),
    ('Soi_Nguyen', 'Anh_Tay_Trai_On.png'):            ('vai/soi-nguyen', 'tay-trai'),
    ('Soi_Nguyen', 'Anh_Tay_Trai_Off.png'):           ('vai/soi-nguyen', 'tay-trai-da-dung'),
    ('Soi_Nguyen', 'Anh_Tay_Phai.png'):               ('vai/soi-nguyen', 'tay-phai'),
    ('Soi_Tien_Tri', 'Tay_Trai_SoiTT.png'):           ('vai/soi-tien-tri', 'tay-trai'),
    ('Soi_Tien_Tri', 'Tay_Phai_SoiTT.png'):           ('vai/soi-tien-tri', 'tay-phai'),
    ('Tien_Tri', 'tay-trai-tien-tri.png'):            ('vai/tien-tri', 'tay-trai'),
    ('Tien_Tri', 'tay-phai-tien-tri.png'):            ('vai/tien-tri', 'tay-phai'),
    ('Phu_Thuy', 'Left_Hand_On.png'):                 ('vai/phu-thuy', 'tay-trai'),
    ('Phu_Thuy', 'Left_Hand_Off.png'):                ('vai/phu-thuy', 'tay-trai-da-dung'),
    ('Phu_Thuy', 'Right_Hand_On.png'):                ('vai/phu-thuy', 'tay-phai'),
    ('Phu_Thuy', 'Right_Hand_Off.png'):               ('vai/phu-thuy', 'tay-phai-da-dung'),
    ('Bao_Ve', 'bao-ve-tay-trai.png'):                ('vai/bao-ve', 'tay-trai'),
    ('Bao_Ve', 'bao-ve-tay-phai.png'):                ('vai/bao-ve', 'tay-phai'),
    ('Nguyet_Nu', 'nguyet-nu-tay-trai.png'):          ('vai/nguyet-nu', 'tay-trai'),
    ('Nguyet_Nu', 'nguyet-nu-tay-phai.png'):          ('vai/nguyet-nu', 'tay-phai'),
    ('Tham_Tu', 'tham-tu-tay-trai-chua-dung.png'):    ('vai/tham-tu', 'tay-trai'),
    ('Tham_Tu', 'tham-tu-tay-trai-da-dung.png'):      ('vai/tham-tu', 'tay-trai-da-dung'),
    ('Tham_Tu', 'tham-tu-tay-phai-chua-dung.png'):    ('vai/tham-tu', 'tay-phai'),
    ('Tham_Tu', 'tham-tu-tay-phai-da-dung.png'):      ('vai/tham-tu', 'tay-phai-da-dung'),
    ('Than_Cupid', 'cupid-tay-trai-chua-dung.png'):   ('vai/cupid', 'tay-trai'),
    ('Than_Cupid', 'cupid-tay-trai-da-dung.png'):     ('vai/cupid', 'tay-trai-da-dung'),
    ('Than_Cupid', 'cupid-tay-phai-chua-dung.png'):   ('vai/cupid', 'tay-phai'),
    ('Than_Cupid', 'cupid-tay-phai-da-dung.png'):     ('vai/cupid', 'tay-phai-da-dung'),
    ('Tho_San', 'tay-tho-san.png'):                   ('vai/tho-san', 'hai-tay'),
    ('Hiep_Si_Kiem_Gi', 'hiep-si-tay-trai.png'):      ('vai/hiep-si', 'tay-trai'),
    ('Hiep_Si_Kiem_Gi', 'hiep-si-tay-phai.png'):      ('vai/hiep-si', 'tay-phai'),
    ('Thi_Truong', 'truong-lang-tay-trai.png'):       ('vai/truong-lang', 'tay-trai'),
    ('Thi_Truong', 'truong-lang-tay-phai.png'):       ('vai/truong-lang', 'tay-phai'),
}

def out(rel):
    p = os.path.join(OUT, rel); os.makedirs(os.path.dirname(p), exist_ok=True); return p

def save_lossy(im, rel, q):
    im.convert('RGB').save(out(rel), 'WEBP', quality=q, method=6)

def save_hand(im, rel):
    # nén có mất ở mức 92 (sai lệch màu ~4/255, mắt không thấy), kênh trong suốt giữ nguyên
    im.save(out(rel), 'WEBP', quality=92, alpha_quality=100, method=6)

if os.path.exists(OUT): shutil.rmtree(OUT)
log = []

# Nền + mặt sau thẻ
save_lossy(Image.open(f'{SRC}/BackGround/Anh_Ngay.png'), 'nen/ngay.webp', 85)
save_lossy(Image.open(f'{SRC}/BackGround/Anh_dem.png'), 'nen/dem.webp', 85)
save_lossy(Image.open(f'{SRC}/Mat_Sau_La_Bai.jpg'), 'the/mat-sau.webp', 90)

# Thẻ vai
for code, (d, slug) in {**VAI, **DANH_HIEU}.items():
    save_lossy(Image.open(f'{SRC}/{d}/{d}.jpg'), f'vai/{slug}/the.webp', 90)

# Ảnh tay: gom theo (thư mục ra, bên tay) -> dùng chung một khung cắt rồi đưa về cao 360
groups = {}
for (d, f), (od, name) in TAY.items():
    im = Image.open(f'{SRC}/{d}/{f}').convert('RGBA')
    side = 'hai' if name == 'hai-tay' else ('trai' if 'trai' in name else 'phai')
    key = (od, side + ('-ngay' if 'ngay' in name else ''))
    groups.setdefault(key, []).append((name, im, f'{d}/{f}'))

for (od, side), items in groups.items():
    sizes = {im.size for _, im, _ in items}
    if len(sizes) > 1:
        sys.exit(f'LỖI: các trạng thái của {od} {side} khác kích thước {sizes} — không thể dùng chung khung')
    box = None
    for _, im, _ in items:
        b = im.getchannel('A').point(lambda a: 255 if a > 8 else 0).getbbox()
        box = b if box is None else (min(box[0], b[0]), min(box[1], b[1]), max(box[2], b[2]), max(box[3], b[3]))
    for name, im, src in items:
        c = im.crop(box)
        if c.size[1] != H:
            c = c.resize((round(c.size[0] * H / c.size[1]), H), Image.LANCZOS)
        if c.size != im.size:
            log.append(f'{src}: {im.size[0]}x{im.size[1]} -> {c.size[0]}x{c.size[1]}')
        save_hand(c, f'{od}/{name}.webp')

# Khung giao diện (pixel art -> không nén mất)
ui = {}
if UI:
    for f in sorted(os.listdir(UI)):
        if f.startswith('khung-') and f.endswith('.png'):
            Image.open(os.path.join(UI, f)).save(out(f'ui/{f[:-4]}.webp'), 'WEBP', lossless=True, quality=100, method=6)
            ui[f[:-4]] = f'ui/{f[:-4]}.webp'

# Manifest cho code
def tay(slug):
    d = os.path.join(OUT, 'vai', slug); r = {}
    for f in sorted(os.listdir(d)):
        if f.startswith('tay-') or f.startswith('hai-tay'):
            r[f[:-5]] = f'vai/{slug}/{f}'
    return r

man = {
    'ghiChu': 'Đường dẫn tính từ thư mục assets/. Tay: không hậu tố = còn dùng được, "-da-dung" = đã dùng.',
    'nen': {'ngay': 'nen/ngay.webp', 'dem': 'nen/dem.webp'},
    'matSau': 'the/mat-sau.webp',
    'ui': ui,
    'tayNgay': {f[:-5]: f'chung/{f}' for f in sorted(os.listdir(os.path.join(OUT, 'chung')))},
    'vai': {code: {'thuMuc': slug, 'the': f'vai/{slug}/the.webp', 'tay': tay(slug)} for code, (_, slug) in VAI.items()},
    'danhHieu': {code: {'thuMuc': slug, 'the': f'vai/{slug}/the.webp', 'tay': tay(slug)} for code, (_, slug) in DANH_HIEU.items()},
}
with open(out('manifest.json'), 'w', encoding='utf-8') as fp:
    json.dump(man, fp, ensure_ascii=False, indent=2)

print('Đã cắt/đổi cỡ:'); print('\n'.join('  ' + l for l in log) or '  (không)')
