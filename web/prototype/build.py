"""Ghép bản chạy thử: python3 prototype/build.py  (chạy từ thư mục gốc repo)."""
import base64, json, os
D = os.path.dirname(os.path.abspath(__file__)) + '/'
A = os.path.normpath(D + '../assets') + '/'
s = open(D + 'ban-3d.template.html', encoding='utf-8').read()
b64 = lambda p, mt='image/webp': 'data:%s;base64,' % mt + base64.b64encode(open(p, 'rb').read()).decode()
m = {'__BV_L__': 'vai/bao-ve/tay-trai.webp', '__BV_R__': 'vai/bao-ve/tay-phai.webp', '__SOI_L__': 'vai/soi/tay-trai.webp', '__SOI_R__': 'vai/soi/tay-phai.webp',
     '__DAY_L__': 'chung/tay-ngay-trai.webp', '__DAY_LV__': 'chung/tay-ngay-trai-da-vote.webp', '__DAY_R__': 'chung/tay-ngay-phai.webp',
     '__BG_DAY__': 'nen/ngay.webp', '__BG_NIGHT__': 'nen/dem.webp'}
out = s.replace('__VILLAGER__', open(D + 'villager.js', encoding='utf-8').read())
for k, p in m.items(): out = out.replace(k, b64(A + p))
out = out.replace('__WOLF__', b64(D + 'wolf.png', 'image/png'))
hands = {}
for r in sorted(os.listdir(A + 'vai')):
    for f in sorted(os.listdir(A + 'vai/' + r)):
        if f.startswith('tay-'): hands[r + '/' + f[:-5]] = b64(A + 'vai/' + r + '/' + f)
out = out.replace('__HANDS__', json.dumps(hands))
out = out.replace('__CARDS__', json.dumps({f[:-5]: b64(D + 'cards/' + f) for f in sorted(os.listdir(D + 'cards'))}))
open(D + 'ban-3d.html', 'w', encoding='utf-8').write(out)
print('ok', len(out) // 1024, 'KB')
