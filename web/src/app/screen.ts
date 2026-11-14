// Which screen the phone shows, from the store alone plus the form the
// player opened before joining (docs/frontend/giao-dien-3d.md table "pha →
// cảnh": before the room the 2D screens; Lobby; then the game, drawn in
// later steps). Built from the latest view only (web/CLAUDE.md).
import type { Server } from '../api/types.ts';
import type { State } from '../state/store.ts';

/** The pre-room form the player is on. */
export type Route = { name: 'home' } | { name: 'create' } | { name: 'join'; code?: string };

export type ScreenName = 'home' | 'create' | 'join' | 'pin' | 'lobby' | 'game';

/** Screens drawn by later steps (3b onwards); Bước 3 shows a placeholder for them. */
export function screenOfPhase(phase: Server.Phase): 'lobby' | 'game' {
  switch (phase) {
    case 'Lobby':
      return 'lobby';
    case 'RoleView':
    case 'ChiefElection':
    case 'Night':
    case 'Dawn':
    case 'Day':
    case 'NominationLock':
    case 'Trial':
    case 'Verdict':
    case 'Dusk':
    case 'HunterShot':
    case 'ChiefSuccession':
    case 'Ended':
      return 'game';
    default: {
      const unknown: never = phase;
      void unknown;
      return 'game'; // a phase from a newer server
    }
  }
}

/** True when the phone is in a room: joined, or dropped and rejoining with a screen to keep. */
export function inRoom(s: State): boolean {
  const c = s.connection.state;
  return s.view !== undefined && (c === 'joined' || c === 'waiting' || c === 'connecting');
}

export function pickScreen(s: State, route: Route): ScreenName {
  if (!inRoom(s)) return route.name;
  const screen = screenOfPhase(s.view!.phase);
  // The PIN comes with a join, in the lobby; once the deal starts the game it
  // must not hide the role view.
  return s.pin !== undefined && screen === 'lobby' ? 'pin' : screen;
}

/** The host's controls show for the host only (docs/api/screens/lobby.md "What drives the lobby"). */
export function isHost(s: State): boolean {
  const me = s.view?.me?.id;
  return me !== undefined && (s.people ?? []).some((p) => p.id === me && p.isHost);
}
