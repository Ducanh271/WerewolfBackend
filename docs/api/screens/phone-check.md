# Phone check

A night step (`view.nightCall.role` `PhoneCheck`, 5 s by default) when
**everybody** opens their eyes and reads their phone (rules §3 "Kiểm tra
điện thoại"). It runs on a fixed schedule set by the public role list:
night 1 after Cupid when the list has Cupid, nights 1–3 after the Cursed
wolf when it has a Cursed wolf (both on night 1 when it has both). It runs
whether or not anybody was cursed or paired, so it tells nothing.

## What drives the screen

| field | use |
|---|---|
| `view.me.phoneCheck` | the notice, on every living phone while the step runs |
| `.cursedText` | the curse line ("Bạn không bị nguyền." / "Bạn đã bị nguyền."), present only when the role list has a Cursed wolf |
| `.loveText` | the love line ("Bạn không ở trong cặp đôi." / "Bạn yêu {nickname}, vai {role}."), present only when it has Cupid |
| `.cursed`, `.love` | the same as data; `love.colour` is the lovers' link colour (rules §7) |
| `deadline` | end of the step |

Show the lines as given, never rewritten. Every notice of a game has the
same keys whatever the answer, and the screen must be the same too: same
size, same layout, same colours for every answer, and the answer itself
behind press-and-hold like the role card, because at this step **every
eye at the table is open** ([secrecy](../secrecy.md#phone-check)).

From this step on, a cursed player's `me.team`, `me.roleCard` side and
`me.statusCards` (`cursed`) change, and lovers get `view.lover` and a
`lovers` or `lovers_cross` card (rules §9b "Thẻ phụ").

<!-- example: server/view.night.phoneCheck.notInLove.json type deadline view.nightCall view.me.phoneCheck -->
```json
{
  "type": "view",
  "deadline": 1792094525000,
  "view": {
    "me": {
      "phoneCheck": {
        "cursed": false,
        "cursedText": "Bạn không bị nguyền.",
        "love": {
          "colour": "",
          "inLove": false,
          "partner": "",
          "partnerRole": ""
        },
        "loveText": "Bạn không ở trong cặp đôi."
      }
    },
    "nightCall": {
      "duration": 5000,
      "role": "PhoneCheck"
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.night.phoneCheck.cursed.json view.me.phoneCheck -->
```json
{
  "view": {
    "me": {
      "phoneCheck": {
        "cursed": true,
        "cursedText": "Bạn đã bị nguyền.",
        "love": {
          "colour": "",
          "inLove": false,
          "partner": "",
          "partnerRole": ""
        },
        "loveText": "Bạn không ở trong cặp đôi."
      }
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.night.phoneCheck.inLove.json view.me.phoneCheck view.lover -->
```json
{
  "view": {
    "lover": {
      "colour": "white",
      "partner": "p5",
      "partnerRole": "Villager"
    },
    "me": {
      "phoneCheck": {
        "cursed": false,
        "cursedText": "Bạn không bị nguyền.",
        "love": {
          "colour": "white",
          "inLove": true,
          "partner": "p5",
          "partnerRole": "Villager"
        },
        "loveText": "Bạn yêu Giang, vai Dân làng."
      }
    }
  }
}
```
<!-- /example -->

## What the phone may send

Nothing. Errors: any night command is `not_your_turn`.

## Events

`PhoneCheckAnnounced` with `say` ("Mọi người mở mắt xem điện thoại, rồi
nhắm mắt lại."), read by the speaker; `PhoneChecked` to each living player,
their own notice as data.
