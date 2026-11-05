# Day: discussion, nomination, trial, verdict

Every day vote is public (rules §4). The day runs through these phases:

| phase | what happens (rules §4 "Thảo luận và xử trảm") |
|---|---|
| `Day` | discussion and nominations, on the day clock; "skip" votes |
| `NominationLock` | every living player chose and someone is nominated: a short countdown; any change restarts it (back to `Day`) |
| `Trial` | the defendant defends themselves; the day clock stops |
| `Verdict` | everybody but the defendant votes kill or spare |
| `Dusk` | the day is decided (hanged, or nobody); the speaker moves on to the night |

After a hanging: a dead [Hunter](hunter.md) shoots, the win check runs
(rules §8), a dead Chief names a [successor](chief.md), then `Dusk`. After
a spare, the nominations and skips are cleared and the day goes on if the
day clock has time left. When the day clock ran out with someone
nominated, the top nominee is still tried and the night follows whatever
the verdict (rules §4 step 6).

## What drives the screen

| field | use |
|---|---|
| `deadline` | end of the running phase (the day clock in `Day`; the countdown in `NominationLock`) |
| `view.votes.nominations` | `{voter: nominee}`, `""` = "no one", live |
| `view.votes.skips` | who voted to skip the rest of the day |
| `view.votes.spared` | spared today: cannot be nominated again until the night |
| `view.votes.defendant` | the player on trial (`Trial`, `Verdict`) |
| `view.votes.verdicts` | `{voter: kill}`, live |
| `view.votes.abortVotes` | `{voter: yes}`: the vote to call the game off (rules §8b) |
| `view.chief` | the Chief: their nomination and verdict count double |
| `view.players[].alive` | who may vote |

<!-- example: server/view.day.json type deadline view.phase view.votes view.chief -->
```json
{
  "type": "view",
  "deadline": 1792095120000,
  "view": {
    "chief": "p8",
    "phase": "Day",
    "votes": {
      "abortVotes": {
        "p2": false
      },
      "nominations": {
        "p2": "p3"
      },
      "spared": [
        "p11"
      ]
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.nominationLock.json type deadline view.phase view.votes -->
```json
{
  "type": "view",
  "deadline": 1792094800000,
  "view": {
    "phase": "NominationLock",
    "votes": {
      "abortVotes": {
        "p2": false
      },
      "nominations": {
        "p1": "p11",
        "p10": "p11",
        "p11": "p3",
        "p12": "p11",
        "p2": "p11",
        "p3": "p11",
        "p4": "",
        "p5": "p11",
        "p6": "p11",
        "p7": "",
        "p8": "p11",
        "p9": "p11"
      }
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.verdict.json type deadline view.phase view.votes.defendant view.votes.verdicts -->
```json
{
  "type": "view",
  "deadline": 1792094860000,
  "view": {
    "phase": "Verdict",
    "votes": {
      "defendant": "p11",
      "verdicts": {
        "p1": true,
        "p10": true,
        "p2": false,
        "p3": false,
        "p4": true,
        "p5": false,
        "p6": false,
        "p7": true,
        "p8": false,
        "p9": false
      }
    }
  }
}
```
<!-- /example -->

<!-- example: server/view.dusk.json type deadline view.phase view.votes view.chief -->
```json
{
  "type": "view",
  "deadline": 1792095495000,
  "view": {
    "chief": "p3",
    "phase": "Dusk",
    "votes": {
      "nominations": {
        "p1": "p2",
        "p10": "p2",
        "p12": "p2",
        "p2": "p12",
        "p3": "p2",
        "p4": "p2",
        "p5": "p2",
        "p6": "p2",
        "p7": "p2"
      },
      "verdicts": {
        "p1": true,
        "p10": true,
        "p12": true,
        "p3": true,
        "p4": true,
        "p5": true,
        "p6": true,
        "p7": true
      }
    }
  }
}
```
<!-- /example -->

## What the phone may send

| send | phases | who | notes |
|---|---|---|---|
| `nominate` `{target}` | `Day`, `NominationLock` | living players | not oneself, not a spared player; `""` = no one; may change |
| `skipDay` `{on}` | `Day`, `NominationLock` | living players | toggle; passes over 66% of the living (rules §4) |
| `verdict` `{kill}` | `Verdict` | living players but the defendant | `true` kill, `false` spare |
| `abortVote` `{on}` | `Day`, `NominationLock`, `Trial`, `Verdict`, `Dusk` | living players | toggle; passes at 2/3 of the living (rules §8b) |
| `sendChat` `{text}` | `Day`, `NominationLock` | living players | see [Chat](#chat) |
| `sendChat` `{text}` | `Trial` | the defendant only | see [Chat](#chat) |

Hide the controls a player may not use (dead, the defendant, oneself and
spared players in the nominee list), so the errors below stay for races.

## Errors

`wrong_phase` (a tap after the phase moved on), `rate_limited`. A correct
app never gets `self_nomination`, `already_spared`, `on_trial`, `not_alive`.
Chat has its own: see [Chat](#chat).

## Events and announcements

With `say`: `DayStarted`, `TrialStarted` (names the defendant),
`VerdictStarted`, `Hanged`, `Spared`, `DayEnded` ("Không ai bị treo.").
Without: `NominationCast`, `SkipSet`, `NominationsLocking`,
`NominationsReopened`, `VerdictCast`, `DayClockExpired`, `AbortVoteCast`,
`ChatSent`.
A passed abort vote ends the game: [end](end.md).

<!-- example: events/TrialStarted.json -->
```json
{
  "type": "TrialStarted",
  "data": {
    "defendant": "p11",
    "duration": 30000
  },
  "say": "Phúc lên giàn bào chữa."
}
```
<!-- /example -->

## Chat

One public channel for the whole room, typed, next to what the table says
aloud (rules §4 "Nhắn tin"). There is no private channel for the wolves or
the lovers. A message moves no timer, vote or countdown, and the speaker never
reads it.

**Who may write.** Show the input only then, and hide it otherwise:

| phase | who |
|---|---|
| `ChiefElection`, `Day`, `NominationLock` | every living player |
| `Trial` | the player in `view.votes.defendant` |
| any other | nobody, the dead included |

**Reading.** Every phone reads the chat at any time, the dead too. Build it
from the `events` messages: each `ChatSent` is one line, `data.by` is the
player and `data.text` the message, in the order of the list. An `events`
message with `reset: true` replaces the whole list, so rebuild the chat from
it. That list keeps the earlier games of the room: **show only what follows the
latest `NewGameStarted`**, because a new game's chat starts empty (rules §8c).

Show `data.text`, not what the player typed: the server normalizes it (a line
break becomes a space, invisible characters go). It is what a player typed all
the same, so render it as plain text: [secrecy](../secrecy.md#everywhere).

<!-- example: events/ChatSent.json -->
```json
{
  "type": "ChatSent",
  "data": {
    "by": "p2",
    "text": "Tôi nghi Phúc."
  }
}
```
<!-- /example -->

**Limits.** 200 characters a message; 1 message per 2 seconds for each player,
up to 5 at once after a silence; 2000 messages a game. Coming back on a new
socket does not reset them.

| error | what to do |
|---|---|
| `chat_closed` | a message sent just as the phase changed: show the text, hide the input. At night every player gets exactly this |
| `chat_empty` | keep the send button off until there is something to send |
| `chat_too_long` | count the characters in the input, and keep what was typed so it can be cut |
| `chat_too_fast` | show the text, and keep what was typed so it can be sent again |
| `chat_full` | show the text and hide the input until the next game |

A correct app never gets `not_alive` here. A message over the 4 KB limit gets
`message_too_large` and the socket stays open; 200 characters always fit.
