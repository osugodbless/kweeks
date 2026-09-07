import { create } from "zustand";
import { persist } from "zustand/middleware";

// Player session: which room + who I am, shared across the join → lobby →
// question → standings → podium flow. Persisted to localStorage so a refresh
// mid-game keeps the player in the room.

export interface PlayerIdentity {
  participantId: string;
  roomId: string;
  email: string;
  nickname: string;
  avatar: string;
}

interface PlayerState {
  roomId: string | null;
  code: string | null;
  participantId: string | null;
  email: string | null;
  nickname: string | null;
  avatar: string | null;
  /** Index of the last question this player answered — drives the
   *  standings → next-question auto-advance. */
  lastAnsweredIndex: number | null;
  setRoom: (roomId: string, code: string) => void;
  setParticipant: (p: PlayerIdentity) => void;
  markAnswered: (index: number) => void;
  clear: () => void;
}

export const usePlayer = create<PlayerState>()(
  persist(
    (set) => ({
      roomId: null,
      code: null,
      participantId: null,
      email: null,
      nickname: null,
      avatar: null,
      lastAnsweredIndex: null,
      setRoom: (roomId, code) => set({ roomId, code }),
      setParticipant: (p) =>
        set({
          roomId: p.roomId,
          participantId: p.participantId,
          email: p.email,
          nickname: p.nickname,
          avatar: p.avatar,
        }),
      markAnswered: (index) => set({ lastAnsweredIndex: index }),
      clear: () =>
        set({
          roomId: null,
          code: null,
          participantId: null,
          email: null,
          nickname: null,
          avatar: null,
          lastAnsweredIndex: null,
        }),
    }),
    { name: "kweeks.player" },
  ),
);

export function fmtNaira(naira: string | number): string {
  const n = typeof naira === "number" ? naira : parseInt(naira || "0", 10);
  if (Number.isNaN(n)) return "0";
  return new Intl.NumberFormat("en-NG", { maximumFractionDigits: 0 }).format(n);
}

export function naira(naira: string | number): string {
  return `₦${fmtNaira(naira)}`;
}

/**
 * Mirrors the backend `SplitPodium` (internal/domain/money.go): descending
 * weighted shares, rounding remainder to first place so the shares sum exactly
 * to the pool. Used for prize-preview copy on lobby/podium screens.
 */
export function splitPodium(poolNaira: string | number, n: number): number[] {
  const pool = typeof poolNaira === "number" ? poolNaira : parseInt(poolNaira || "0", 10);
  if (n <= 0) return [];
  if (n === 1) return [pool];
  const sumWeights = (n * (n + 1)) / 2;
  const shares: number[] = [];
  let assigned = 0;
  for (let i = 0; i < n; i++) {
    const w = n - i;
    const share = Math.floor((pool * w) / sumWeights);
    shares.push(share);
    assigned += share;
  }
  shares[0] += pool - assigned;
  return shares;
}
