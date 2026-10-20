# Night

Phase `Night`: the speaker calls the roles one step at a time, in the order
of rules §3; each called player acts on their own phone while everybody
else keeps their eyes closed. Nothing is resolved until the morning
([dawn](dawn.md)). Two night steps belong to no role and have their own
guide: the [phone check](phone-check.md) and the lovers looking at each
other (end of this page).

**Read [secrecy](../secrecy.md) before building this screen.** Every step
runs its full time, real or fake (a role that is dead still gets called,
rules §3 "Gọi giả"), and every living phone must look and behave the same
whatever happens on it.

## What drives the screen

| field | who has it | use |
|---|---|---|
| `view.nightCall` | every phone | `{role, duration}` of the step running now: the role being called (`Werewolf` is the pack step, "Bầy sói"), `PhoneCheck`, or `LoversMeet`. Never says whether the call is real |
| `deadline` | every phone | when this step ends |
| `view.me.turn` | the player being called | `step`; the action screen for that step. `blocked`, `doomed`, `grief`: no action, show `notice` instead (rules §3, §6 "Hiệp sĩ kiếm gỉ", §7) |
| `view.me.turn.biteShown`, `bitten` | the Witch, while her heal potion is left | whom the pack bites now; `biteShown` with no `bitten` = nobody |
| `view.pack` | living wolf team (the cursed too) | `members`; `votes` `{wolf: target}` live during the pack step (`""` = no bite); `doomed` and `doomedNotices` (wolves who can no longer vote, rules §6 "Hiệp sĩ kiếm gỉ") |
| `view.me.statusCards`, `view.lover` | cursed players, lovers | from the phone check on |
| `view.full`, `view.secrets` | dead players | everything, live: see "Dead players" below |

A phone whose screen did not change gets **no message** (`../README.md`
"Server → client"): when a step starts every phone gets a new `view` (the
`nightCall` changed), and while the called player acts only the phones
whose own screen changed get one (the actor, the pack, the dead).

Every phone, called or not, shows the same night screen: the role being
called and the countdown. The called player's phone adds the action screen
**on the same layout**, without any sound, vibration, light or animation
that the other phones do not have ([secrecy](../secrecy.md#night)).

<!-- example: server/view.night.asleep.json type deadline view.phase view.night view.nightCall view.me -->
```json
{
  "type": "view",
  "deadline": 1792094520000,
  "view": {
    "me": {
      "alive": true,
      "canShoot": false,
      "cursed": false,
      "id": "p3",
      "role": "Guard",
      "roleCard": {
        "description": "Mỗi đêm che một người khỏi bị Sói cắn. Được che chính mình, không che một người hai đêm liền.",
        "goal": "Dân thắng khi không còn con Sói nào.",
        "name": "Bảo vệ",
        "team": "Dân"
      },
      "team": "Village"
    },
    "night": 1,
    "nightCall": {
      "duration": 30000,
      "role": "Cupid"
    },
    "phase": "Night"
  }
}
```
<!-- /example -->

## What the phone may send

| step (`nightCall.role`) | send | who | notes |
|---|---|---|---|
| `Cupid` (night 1) | `cupidPair` `{a, b}` | Cupid | two different living players, Cupid allowed; if none chosen, the server picks |
| `Visitor` | `visitorVisit` `{target}` | Nguyệt Nữ | a living player, not herself, not last night's; if none chosen, the server picks |
| `Guard` | `guardProtect` `{target}` | Bảo vệ | a living player, oneself allowed, not last night's |
| `CursedWolf` (nights 1–3) | `cursedWolfCurse` `{target}` | Sói nguyền | a living non-wolf; once a game |
| `WolfSeer` | `wolfSeerCheck` `{target}` | Sói tiên tri | any living player |
| `Werewolf` | `wolfVote` `{target}` | every living wolf-team player | `""` = "Không cắn ai"; may change until the step ends |
| `Witch` | `witchHeal`, `witchPoison` `{target}` | Phù thủy | each potion once a game; heal only when someone is bitten |
| `Seer` | `seerCheck` `{target}` | Tiên tri | a living player |
| `Detective` | `detectiveCheck` `{target}` | Thám tử | once a game; still called every night after (rules §3) |

Every choice is final except the pack's vote. The rules of each role:
rules §5, §6, §7; the role's own "how to play" text: [guide](guide.md).
No step ends early when the choice is made: keep the step's screen until
`deadline`.

Results come as events to the actor only: `SeerChecked` (`isWolf`),
`WolfSeerChecked` (to the whole pack), `DetectiveChecked` (`hasWolf`),
`PlayerProtected`, `PlayerVisited`, `PlayerCursed`, `LoversPaired`,
`WitchHealed`, `WitchPoisoned`; `WolfVoted` to the pack.

<!-- example: server/view.night.witch.json type deadline view.nightCall view.me.turn -->
```json
{
  "type": "view",
  "deadline": 1792094720000,
  "view": {
    "me": {
      "turn": {
        "biteShown": true,
        "bitten": "p1",
        "blocked": false,
        "doomed": false,
        "grief": false,
        "step": "Witch"
      }
    },
    "nightCall": {
      "duration": 30000,
      "role": "Witch"
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.night.pack.json view.nightCall view.me.turn view.pack -->
```json
{
  "view": {
    "me": {
      "turn": {
        "biteShown": false,
        "blocked": false,
        "doomed": false,
        "grief": false,
        "step": "Werewolf"
      }
    },
    "nightCall": {
      "duration": 30000,
      "role": "Werewolf"
    },
    "pack": {
      "members": [
        "p9",
        "p11",
        "p12"
      ],
      "votes": {
        "p11": "p1",
        "p12": "p1",
        "p9": "p1"
      }
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.night.seer.blocked.json view.nightCall view.me.turn -->
```json
{
  "view": {
    "me": {
      "turn": {
        "biteShown": false,
        "blocked": true,
        "doomed": false,
        "grief": false,
        "notice": "Bạn không thể sử dụng chức năng đêm nay.",
        "step": "Seer"
      }
    },
    "nightCall": {
      "duration": 30000,
      "role": "Seer"
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.night.pack.doomed.json view.me.turn view.pack -->
```json
{
  "view": {
    "me": {
      "turn": {
        "biteShown": false,
        "blocked": false,
        "doomed": true,
        "grief": false,
        "notice": "Bạn đã trúng độc và chết từ lúc này, bạn không được vote.",
        "step": "Werewolf"
      }
    },
    "pack": {
      "doomed": [
        "p9"
      ],
      "doomedNotices": [
        "Minh không thể vote đêm nay."
      ],
      "members": [
        "p9",
        "p11",
        "p12"
      ]
    }
  }
}
```
<!-- /example -->

<!-- example: events/SeerChecked.json -->
```json
{
  "type": "SeerChecked",
  "data": {
    "isWolf": true,
    "seer": "p7",
    "target": "p12"
  }
}
```
<!-- /example -->

## Errors

`not_your_turn`: anybody who does not hold the step being called, whatever
their role and state, gets exactly this, and so does a tap after the step
ended; show it like any other refusal. `blocked` (the same text as
`me.turn.notice`), `already_acted`, `ability_used`, `potion_used`,
`nothing_to_heal`, `same_guard_target`, `same_visitor_target`,
`invalid_target`, `wrong_phase`, `rate_limited`. A correct app never gets
`not_alive`.

<!-- example: server/error.not_your_turn.json -->
```json
{
  "type": "error",
  "code": "not_your_turn",
  "message": "not your turn"
}
```
<!-- /example -->

## Events and announcements

`NightStarted` (`say`), then per step `NightStepCalled` `{role, duration}`
with `say` naming the role, never a player ("Bầy sói hãy mở mắt...").
`NightEnded` closes the night. Steps without a role read their own line
(`PhoneCheckAnnounced`, `LoversMeetAnnounced`) instead.

<!-- example: events/NightStepCalled.json -->
```json
{
  "type": "NightStepCalled",
  "data": {
    "duration": 30000,
    "role": "Cupid"
  },
  "say": "Cupid hãy mở mắt, thực hiện xong thì nhắm mắt lại."
}
```
<!-- /example -->

## Lovers look at each other

Night 1 with Cupid in the role list: after the phone check, `nightCall.role`
is `LoversMeet` (10 s by default). The two lovers' phones have `me.turn.step`
`LoversMeet`: they open their eyes and look at each other, with nothing to
send. Every other phone shows the same night screen as for any step.

<!-- example: server/view.night.loversMeet.json view.nightCall view.me.turn view.lover -->
```json
{
  "view": {
    "lover": {
      "colour": "white",
      "partner": "p5",
      "partnerRole": "Villager"
    },
    "me": {
      "turn": {
        "biteShown": false,
        "blocked": false,
        "doomed": false,
        "grief": false,
        "step": "LoversMeet"
      }
    },
    "nightCall": {
      "duration": 10000,
      "role": "LoversMeet"
    }
  }
}
```
<!-- /example -->

## Dead players

A dead player sees everything, live (rules §4), so their phone keeps getting
messages while others act. It MUST show it only while its owner presses
and holds, day and night; the rest of the time a plain cover that never
changes when a message arrives, dark at night
([secrecy](../secrecy.md#dead-players)).

<!-- example: server/view.night.dead.json view.phase view.nightCall view.me view.full view.pack -->
```json
{
  "view": {
    "full": true,
    "me": {
      "alive": false,
      "canShoot": false,
      "cursed": false,
      "id": "p8",
      "role": "Hunter",
      "roleCard": {
        "description": "Khi chết vì bất kỳ lý do gì, bạn lộ vai và bắt buộc bắn một người còn sống.",
        "goal": "Dân thắng khi không còn con Sói nào.",
        "name": "Thợ săn",
        "team": "Dân"
      },
      "team": "Village"
    },
    "nightCall": {
      "duration": 30000,
      "role": "Visitor"
    },
    "pack": {
      "members": [
        "p9",
        "p11",
        "p12"
      ]
    },
    "phase": "Night"
  }
}
```
<!-- /example -->
