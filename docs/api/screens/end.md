# End of the game

Phase `Ended`: a side won (rules §8), all died (a draw), or the table voted
to call the game off (rules §8b). Everybody now sees everything: every
role, every night's choices, the lovers (rules §4).

## What drives the screen

| field | use |
|---|---|
| `view.winner` | the winning team (`Village`, `Wolves`, `Lovers`); absent on a draw or an abort |
| `view.draw` | `true` when nobody won (all dead, or called off) |
| `view.winners` | the ids of the players who win, dead or alive, in seat order |
| `view.summary` | the game's story, one Vietnamese line per night and day plus the outcome: show as given |
| `view.players[]` | every seat with `role`, `team` and `card` |
| `view.full`, `view.secrets` | the hidden facts, for a detailed view |
| `people` (next to `view`) | everybody in the room, who is host, who is connected (for "Ván mới") |

`GameEnded` comes with `say` (the outcome) and `outcome.aborted` tells a
called-off game from a draw. The events list arrives again with
`reset: true`: hidden events are back in their places
(`../README.md` "Server → client").

<!-- example: server/view.ended.json type people[:2] view.phase view.winner view.winners view.draw view.summary -->
```json
{
  "type": "view",
  "people": [
    {
      "connected": true,
      "id": "p1",
      "isHost": true,
      "nickname": "An"
    },
    {
      "connected": true,
      "id": "p2",
      "isHost": false,
      "nickname": "Bình"
    }
  ],
  "view": {
    "draw": false,
    "phase": "Ended",
    "summary": [
      "Lan được bầu làm Trưởng làng.",
      "Đêm 1: Cupid ghép đôi Bình và Giang. Nguyệt Nữ chặn Khoa, Sói tiên tri soi Khoa: Tiên tri. Bảo vệ che chở Hà, Thám tử dò Hà: không có Sói. Sói nguyền nguyền Quân. Sói cắn An, Phù thủy cứu An. Hiệp sĩ kiếm gỉ đầu độc Minh. → Không ai chết.",
      "Ngày 1: Làng tha cho Phúc. Làng tha cho Chi. Không ai bị treo.",
      "Đêm 2: Nguyệt Nữ chặn Phúc. Bảo vệ che chở Khoa. Sói cắn Lan. Tiên tri soi Quân: Sói. Lan là Thợ săn. Lan bắn Phúc. Trưởng làng mới là Chi. → Lan, Minh chết.",
      "Ngày 2: Bình đã bị treo.",
      "Đêm 3: Nguyệt Nữ chặn Nga. Phù thủy đầu độc Quân. → Giang, Quân chết.",
      "Phe Dân chiến thắng, tất cả bầy sói đã bị diệt."
    ],
    "winner": "Village",
    "winners": [
      "p1",
      "p2",
      "p3",
      "p4",
      "p5",
      "p6",
      "p7",
      "p8",
      "p10"
    ]
  }
}
```
<!-- /example -->

<!-- example: server/view.ended.aborted.json view.draw view.votes.abortVotes view.summary -->
```json
{
  "view": {
    "draw": true,
    "summary": [
      "Hà được bầu làm Trưởng làng.",
      "Đêm 1: Cupid ghép đôi Quân và Nga. Nguyệt Nữ chặn Phúc. → Không ai chết.",
      "Ngày 1: Hà đã bị treo. Hà là Thợ săn. Trưởng làng không kịp chọn người kế nhiệm, danh hiệu đã mất.",
      "Bàn bỏ phiếu huỷ ván. Không có phe thắng."
    ],
    "votes": {
      "abortVotes": {
        "p1": true,
        "p2": true,
        "p3": true,
        "p4": true,
        "p5": true,
        "p7": true,
        "p8": true,
        "p9": true
      }
    }
  }
}
```
<!-- /example -->

<!-- example: events/GameEnded.json -->
```json
{
  "type": "GameEnded",
  "data": {
    "outcome": {
      "aborted": false,
      "draw": false,
      "winner": "Village",
      "winners": [
        "p1",
        "p2",
        "p3",
        "p4",
        "p5",
        "p6",
        "p7",
        "p8",
        "p10"
      ]
    }
  },
  "say": "Phe Dân chiến thắng, tất cả bầy sói đã bị diệt."
}
```
<!-- /example -->

## What the phone may send

| send | who | notes |
|---|---|---|
| `newGame` | host | back to the [lobby](lobby.md) with the same players, seats and role list; the host deals again (rules §8c) |
| `transferHost` `{target}` | host | e.g. when the host has to leave |

Errors: `not_host`, `wrong_phase`. While the game ended, a host who stays
disconnected loses the title after 60 seconds (rules §2 "Chủ phòng"), so
somebody can always start the next game.

Events: `NewGameStarted` (no data), then the lobby view.
