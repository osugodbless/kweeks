// Thin typed client for the kweeks REST API. All money fields are whole-naira
// display strings (e.g. "150000"), matching the backend contract.
export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

const BASE = import.meta.env.VITE_API_BASE ?? "/api";

function token(): string | null {
  return localStorage.getItem("kweeks.token");
}

export function setToken(t: string | null) {
  if (t) localStorage.setItem("kweeks.token", t);
  else localStorage.removeItem("kweeks.token");
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  const tok = token();
  if (tok) headers.Authorization = `Bearer ${tok}`;
  let payload: string | undefined;
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }
  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, { method, headers, body: payload });
  } catch {
    throw new ApiError(0, "Network error — is the backend running on :8080?");
  }
  if (!res.ok) {
    let message = res.statusText;
    try {
      const j = await res.json();
      if (j?.error) message = j.error;
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  put: <T>(path: string, body?: unknown) => request<T>("PUT", path, body),
};

// ---- Types (mirror of web/docs/API_CONTRACT.md) ----

export interface Instructor {
  id: string;
  name: string;
  email: string;
  avatar: string;
}

export interface Wallet {
  id: string;
  balanceNaira: string;
  bmoniUserId?: string;
  bmoniWalletId?: string;
  bmoniWalletAddress?: string;
}

export interface AuthResult {
  token: string;
  instructor: Instructor;
  wallet: Wallet;
}

export interface WalletTransaction {
  id: string;
  walletId: string;
  kind: "fund" | "pool" | "payout" | "credit";
  amountNaira: string;
  note: string;
  createdAt: string;
}

export interface WalletView {
  wallet: Wallet;
  transactions: WalletTransaction[];
}

export interface ParticipantBrief {
  id: string;
  nickname: string;
  avatar: string;
}

export interface PublicQuestion {
  id: string;
  index: number;
  prompt: string;
  options: string[];
  startedAt: string;
  durationMs: number;
  remainingMs: number;
}

export interface PublicWinner {
  participantId: string;
  nickname: string;
  avatar: string;
  correctCount: number;
  totalLatencyMs: number;
}

export type RoomState = "lobby" | "live" | "podium" | "ended";

export interface PublicRoom {
  id: string;
  code: string;
  quizId: string;
  title: string;
  poolNaira: string;
  winnerCount: number;
  pacing: "auto" | "manual";
  state: RoomState;
  questionCount: number;
  currentIndex: number;
  participantCount: number;
  stateAt: string;
  host: { name?: string };
  participants: ParticipantBrief[];
  currentQuestion: PublicQuestion | null;
  winners: PublicWinner[] | null;
}

export interface Standing {
  participantId: string;
  nickname: string;
  avatar: string;
  correctCount: number;
  totalLatencyMs: number;
  joinedAt: string;
}

export interface Participant {
  id: string;
  roomId: string;
  email: string;
  nickname: string;
  avatar: string;
  joinedAt: string;
}

export interface AnswerReceipt {
  id: string;
  correct: boolean;
  score: number;
  latencyMs: number;
}

export interface QuizListItem {
  id: string;
  title: string;
  poolNaira: string;
  winnerCount: number;
  pacing: "auto" | "manual";
  questionCount: number;
  createdAt?: string;
  roomCode?: string;
  roomId?: string;
  state?: RoomState;
}

export interface QuizQuestion {
  id: string;
  prompt: string;
  options: string[];
  correctIndex: number;
  durationMs: number;
}

export interface QuizDetail {
  id: string;
  title: string;
  poolNaira: string;
  winnerCount: number;
  pacing: "auto" | "manual";
  defaultDurationMs: number;
  questions: QuizQuestion[];
}

export interface DashboardStat {
  quizzesHosted: number;
  playersHosted: number;
  winnersPaid: number;
  availableNaira: string;
  quizzes: QuizListItem[];
}

export interface HistoryItem {
  id: string;
  at: string;
  type: "fund" | "quiz" | "payout" | "room";
  title: string;
  amountNaira?: string;
  meta?: string;
  state?: string;
}

export interface ClaimResult {
  id: string;
  amountNaira: string;
  state: string;
  claimCode?: string;
}
