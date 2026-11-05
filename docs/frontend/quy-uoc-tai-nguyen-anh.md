# Quy ước tài nguyên ảnh (assets) — Ma Sói

Ảnh **gốc** (PNG/JPG, tên tuỳ ý) để riêng, không đưa vào app. Ảnh **app dùng** do script `web/tools/chuan_hoa_anh.py` sinh ra:

```
python3 web/tools/chuan_hoa_anh.py <thư-mục-gốc MasoiPng> <thư-mục-ra>/assets art/ui
```

## Cấu trúc
```
web/assets/
  manifest.json          bảng mã vai backend -> đường dẫn ảnh
  nen/ngay.webp, dem.webp            nền 1678×937 (cùng bố cục -> chuyển dần ngày/đêm)
  the/mat-sau.webp                   mặt sau lá bài
  ui/khung-{lang|soi|xam}-9.webp     khung 9 lát (border-image 36 / 9px), ui/khung-{lang|soi}-goc.webp hoa văn góc 38px
  chung/tay-ngay-{trai|phai}[-da-vote].webp   tay ban ngày, mọi người như nhau
  vai/<thư-mục>/the.webp             thẻ vai 765×1024
  vai/<thư-mục>/tay-{trai|phai}[-da-dung].webp
  vai/tho-san/hai-tay.webp           ngoại lệ: hai tay cầm chung một khẩu súng
```

## Quy tắc
- Tên: chữ thường, không dấu, gạch nối, không dấu cách. Đuôi duy nhất: `.webp`.
- Tay: không hậu tố = còn dùng được; `-da-dung` = đã dùng. Hai trạng thái cùng bên có cùng kích thước (chung khung cắt) để đổi trạng thái tay không nhảy.
- Tay cao 360px. Thẻ, nền nén có mất mức 90/85; tay mức 92, kênh trong suốt giữ nguyên.
- Ảnh xem thử (`*-xem-thu.png`) chỉ để kiểm tra, không vào app.

## Mã vai -> thư mục
| Mã backend | Thư mục | Ảnh tay |
|---|---|---|
| Villager | dan-lang | tay-trai / tay-phai (bộ cầm đèn, ban đêm) |
| Werewolf | soi | tay-trai / tay-phai |
| CursedWolf | soi-nguyen | tay-trai (+ -da-dung) / tay-phai |
| WolfSeer | soi-tien-tri | tay-trai / tay-phai |
| Seer | tien-tri | tay-trai / tay-phai |
| Witch | phu-thuy | trái = bình cứu, phải = bình độc, mỗi bên + -da-dung |
| Guard | bao-ve | tay-trai / tay-phai |
| Visitor | nguyet-nu | tay-trai / tay-phai |
| Detective | tham-tu | trái = kính lúp, phải = sổ, mỗi bên + -da-dung |
| Cupid | cupid | trái = cung, phải = tên, mỗi bên + -da-dung |
| Hunter | tho-san | hai-tay |
| RustyKnight | hiep-si | tay-trai / tay-phai |
| chief (danh hiệu) | truong-lang | tay-trai / tay-phai |

## Dùng trong code
```js
import m from './assets/manifest.json'
const v = m.vai[view.me.role]                          // vd 'Witch'
const tayTrai = v.tay[daDungCuu ? 'tay-trai-da-dung' : 'tay-trai']
```

## Còn tồn
- Thám tử, tay phải: khung rộng 470px vì sổ mở và sổ gập nằm hai chỗ khác nhau -> vùng chạm to, tay nhảy khi đổi trạng thái. Nên tạo lại ảnh "đã dùng" giữ đúng vị trí sổ.
- Dân làng ban ngày: màu da tay trái lệch giữa chưa vote / đã vote.
