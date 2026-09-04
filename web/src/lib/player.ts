import { create } from "zustand";

// Player session: which room + who I am, shared across the join→podium flow.
interface PlayerState {
  roomId: string | null;
  code: string | null;
  participantId: string | null;
  email: string | null;
  nickname: string | null;
  avatar: string | null;
  setRoom: (v: { roomId: string; code: string }) => void;
  setParticipant: (p: {
    id: string;
    email: string;
    nickname: string;
    avatar: string;
  }) => void;
  clear: () => void;
}

export const usePlayer = create<PlayerState>((set) => ({
  roomId: null,
  code: null,
  participantId: null,
  email: null,
  nickname: null,
  avatar: null,
  setRoom: (v) => set({ roomId: v.roomId, code: v.code }),
  setParticipant: (p) =>
    set({ participantId: p.id, email: p.email, nickname: p.nickname, avatar: p.avatar }),
  clear: () =>
    set({ roomId: null, code: null, participantId: null, email: null, nickname: null, avatar: null }),
}));

export const AVATARS = ["😎", "🤠", "🦁", "🐸", "👾", "🥷", "🦊", "🐼", "🐙", "🦄", "🐯", "🤖"];

export function fmtNaira(naira: string | number): string {
  const n = typeof naira === "number" ? naira : parseInt(naira || "0", 10);
  return new Intl.NumberFormat("en-NG", { maximumFractionDigits: 0 }).format(n);
}

export function naira(naira: string | number): string {
  return `₦${fmtNaira(naira)}`;
}
