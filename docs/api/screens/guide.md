# Guide: rules and "how to play your role"

`GET /guide` returns the in-app rules guide in Vietnamese, built by the
server from the real settings, so its numbers are the game's own. It needs
no room and is the same for everybody; fetch it once and keep it.

| field | use |
|---|---|
| `quickRules` | the rules in a few lines, for a first-time player |
| `firstTime` | advice for a first game (covering the screen, ...) |
| `roles[]` | per role: `role` (id), `title`, `whenCalled`, `howTo`, `onTimeout`, `gotcha`, optional `note` |
| `statuses[]` | the same for the Chief title, the lovers and the cursed (`role`: `chief`, `lovers`, `cursed`) |
| `tips` | short tips per role |

Show the texts as given. The public role list of a game (`view.roleCards`)
names the roles in play; link each to its `roles[]` entry.

The guide is public, but opening the guide of one role during the game
tells whoever sees the screen which role the player holds or suspects. Open
it on a list of every role in the game, never straight on the player's own
role ([secrecy](../secrecy.md#guide)).

<!-- example: http/guide.reply.json quickRules[:2] firstTime[:2] roles[:1] statuses[:1] tips[:1] -->
```json
{
  "firstTime": [
    "Che màn hình (tay, hoặc ngồi quay màn hình vào trong) mỗi khi điện thoại hiện vai của bạn (có 30 giây để xem) hoặc lượt đêm — người ngồi cạnh không được thấy.",
    "Thẻ vai che mặc định; nhấn giữ để xem, thả tay là ẩn lại ngay — đừng mở thẻ vai lâu để người khác liếc thấy."
  ],
  "quickRules": [
    "Ma Sói chơi bằng điện thoại của chính mỗi người, không cần ai làm quản trò riêng.",
    "Có hai phe chính: Dân làng và Sói; có thể có thêm phe thứ ba là một cặp đôi khác phe."
  ],
  "roles": [
    {
      "title": "Sói",
      "gotcha": "Đổi phiếu phút cuối rồi quên đổi lại — phiếu tính đúng lúc hết giờ, không phải phiếu lúc đầu.",
      "howTo": "Chọn một người sống để cắn, hoặc chọn \"Không cắn ai\"; thấy phiếu của cả bầy theo thời gian thực, đổi phiếu tuỳ ý cho tới hết giờ.",
      "onTimeout": "Tính phiếu lúc đó — người được quá nửa phiếu bị cắn; nếu \"Không cắn ai\" được quá nửa, hoặc không ai quá nửa, đêm đó không ai bị cắn.",
      "role": "Werewolf",
      "whenCalled": "Được gọi mỗi đêm cùng cả bầy, kể cả đêm đầu tiên."
    }
  ],
  "statuses": [
    {
      "title": "Trưởng làng (danh hiệu công khai, không phải vai riêng)",
      "gotcha": "Nghĩ vai gốc của mình đổi khi làm Trưởng làng — vai và phe gốc không đổi, chỉ có phiếu tính gấp đôi.",
      "howTo": "Lúc bầu, chọn một người (được tự bầu mình); khi đang giữ danh hiệu, phiếu đề cử và phiếu giết/tha của bạn tự tính gấp đôi, không cần làm gì thêm; khi Trưởng làng chết, trong 15 giây chọn một người sống kế nhiệm.",
      "note": "Đang bị đề cử lên giàn thì mất quyền bỏ phiếu như mọi người khác, kể cả khi đang là Trưởng làng.",
      "onTimeout": "Lúc bầu, không ai bầu cho ai thì server chọn ngẫu nhiên một người. Lúc chọn kế nhiệm, hết giờ thì danh hiệu mất, không ai thay.",
      "role": "chief",
      "whenCalled": "Bầu ngay sau khi xem vai, trước đêm 1; giữ cho tới khi chết."
    }
  ],
  "tips": [
    "Tiên tri: cân nhắc trước khi nói huỵch toẹt kết quả soi ra ban ngày — nói sớm quá dễ lộ mình là Tiên tri."
  ]
}
```
<!-- /example -->
