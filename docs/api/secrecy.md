# Secrecy rules for the app

The server never sends a phone what its player may not know
(`README.md` "Server → client"; rules [§3](../rules.md#3-ban-đêm)). But
the game is played around one table: a phone that lights up in the dark,
buzzes, plays a sound, or shows a bigger card than its neighbour's tells
the table what the server kept secret. These rules are for the app. Each
one names the rule of [`rules.md`](../rules.md) it protects; breaking one
is a bug of the same weight as a server leak.

## Everywhere

- **Same behaviour on every phone.** Nothing a phone does that others can
  notice (sound, vibration, flash, screen turning on, brightness,
  animation, a screen change, a notification) may depend on the player's
  role, team, state, or choices, or on whether a night call is real or
  fake (rules [§3](../rules.md#3-ban-đêm) "Gọi giả", "Mỗi lượt gọi chạy
  đủ thời gian").
- **No vibration, no sound, no notification for game messages.** The only
  sound is the speaker phone reading `say` texts, which are the same
  whoever is alive (rules [§1](../rules.md#1-tổng-quan) "Máy làm loa").
- **Nothing secret outside the screen.** The page title, the app's
  notifications, the lock screen and share previews never show a role, a
  team or a night result.
- **Time comes from the server.** A step ends at `deadline` on every phone,
  never earlier on the phone of a player who already chose; a real call and
  a fake call last the same.
- **Errors look alike.** Every refusal is shown the same way, with its text
  from [errors.md](errors.md). `not_your_turn` is sent to every player who
  does not hold the step being called, blocked, dead, or any other role
  alike; the app adds nothing to it.

## Role cards

Rules [§9b](../rules.md#9b-thẻ-vai-mô-tả-cho-người-chơi): every card has a
**fixed size, is covered by default, and shows only while pressed and
held**; letting go hides it again.

- The same size and layout for every role and every status card
  (`me.roleCard`, `me.statusCards`, rules §9b "Thẻ phụ"), whatever the
  length of the text, so the shape of a card says nothing.
- No team colour, icon or label outside the held card: the covered card
  looks the same for a Wolf and a Villager.
- The player's own seat in `view.players` carries their `role`, `team` and
  `card`; a lover sees the partner's role, a wolf the pack. Show these only
  behind the same press-and-hold, never on the table of seats.
- The card is reachable during the whole game, behind the same gesture.

## Night

Rules [§3](../rules.md#3-ban-đêm): everybody's eyes are closed except the
called role's.

- Every living phone shows the same night screen: the role being called
  (`nightCall.role`) and the countdown. Nothing else differs from one phone
  to another that a neighbour could see in the dark.
- The called player's phone shows its action on that same screen, with no
  sound, vibration, flash or brightness change; a blocked or doomed
  player's notice (`me.turn.notice`, never naming who) takes the same
  place.
- Nothing shows that someone has chosen ("done", a tick, a progress bar)
  on any phone but the actor's own (and the pack's, for the wolves' votes).
- The step's screen stays until `deadline` whether the choice was made or
  not.

## Phone check

Rules [§3](../rules.md#3-ban-đêm) "Kiểm tra điện thoại", §7: at this step
**every eye is open**, so neighbours can see each other's phones.

- Every living phone shows one notice of the same size, layout, colours
  and timing, whatever it says: cursed or not, in love or not.
- The answer is covered and shown only while pressed and held, like a role
  card. The lovers' link colour (`love.colour`, rules §7) is shown only
  inside it.
- Show `cursedText` and `loveText` as given; a line that is absent is
  absent for everybody (it depends only on the public role list).

## Dead players

Rules [§4](../rules.md#4-ban-ngày): a dead player sees every role and
everything that happens, live, while the living still sit next to them.

- **MUST: a dead player's phone shows game information only while its
  owner presses and holds**, for the whole game, by day and by night;
  letting go covers it again, like a role card. This is a requirement, not
  advice: a glance at the phone of a dead neighbour must show nothing.
  - Game information is everything on the screen that comes from the
    game: every seat's `role`, `team` and `card` (`view.full`),
    `view.secrets`, `view.pack` and its live votes, the votes, the night
    calls, the events list and the announcements' text.
  - The cover is one plain screen that stays the same whatever arrives:
    the phone keeps getting messages while others act (the pack's live
    votes, every choice), so a cover that changes, lights up or animates
    beside a sleeping neighbour would show that someone is acting, and
    when. At night it stays dark.
  - The speaker phone follows the same rule when its player is dead: it
    still reads the `say` texts aloud, with the screen covered.
  - It ends with the game: once the phase is `Ended` everybody sees
    everything ([end](screens/end.md)), so the cover goes away.
- **Two pickers are shown openly**, without holding: the dead Hunter's shot
  (`view.me.canShoot`, [hunter](screens/hunter.md)) and the dead Chief
  naming a successor (`view.succession` is this player,
  [Chief](screens/chief.md)). The whole table waits on them, so they must be
  usable without a held finger. The picker shows only the names (and
  seats) of the players to choose from: no role, team, card or anything
  else from the full view. The rest of the screen stays behind the hold.
- **The picker's list MUST be the public "alive" list**: the seats with
  `alive: true` in `view.players`, the same seats every living phone shows
  as alive. A player already dead by the rules but not yet announced (for
  example a wolf the Rusty knight's poison killed, rules §6 "Hiệp sĩ kiếm
  gỉ", listed in `view.pack.doomed`) shows as alive and stays in the list.
  Never filter or mark it with what only the dead phone's full view holds
  (`view.secrets`, `view.pack`, other seats' `role`, hidden events): the
  whole table is looking at this screen, and a missing or greyed name
  would tell them who is dead or what role someone holds. The server
  checks the target itself (`invalid_target`).

## Announcements

Rules [§4](../rules.md#4-ban-ngày) "Thông báo buổi sáng": deaths are
named in seat order, never with cause or role. Show and read the server's
`say` text as given; never build an announcement from events (their order
inside the server is not the order the table may hear).

## Guide

The rules guide (`GET /guide`) is public, but a phone showing the "how to
play" page of one role tells whoever sees it which role the player holds.
Open the guide on the list of every role in the game, never straight on
the player's own role.
