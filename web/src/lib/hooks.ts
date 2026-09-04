import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import {
  api,
  AnswerReceipt,
  ClaimResult,
  DashboardStat,
  HistoryItem,
  Participant,
  PublicRoom,
  QuizDetail,
  QuizListItem,
  Standing,
  Wallet,
  WalletView,
} from "@/lib/api";
import { useAuth } from "@/lib/auth";

export const qk = {
  me: ["me"] as const,
  wallet: ["wallet"] as const,
  dashboard: ["dashboard"] as const,
  history: ["history"] as const,
  quizzes: ["quizzes"] as const,
  quiz: (id: string) => ["quiz", id] as const,
  room: (id: string) => ["room", id] as const,
  lookup: (code: string) => ["room", "code", code] as const,
  standings: (roomId: string) => ["standings", roomId] as const,
};

const authed = () => Boolean(localStorage.getItem("kweeks.token"));

export function useMe() {
  return useQuery({
    queryKey: qk.me,
    queryFn: () => api.get<{ instructor: import("@/lib/api").Instructor; wallet: import("@/lib/api").Wallet }>("/auth/me"),
    enabled: authed(),
  });
}

export function useWallet() {
  const setWallet = useAuth((s) => s.setWallet);
  const q = useQuery({
    queryKey: qk.wallet,
    queryFn: () => api.get<WalletView>("/wallet"),
    enabled: authed(),
  });
  useEffect(() => {
    if (q.data?.wallet) setWallet(q.data.wallet);
  }, [q.data, setWallet]);
  return q;
}

export function useFundWallet() {
  const qc = useQueryClient();
  const setWallet = useAuth((s) => s.setWallet);
  return useMutation({
    mutationFn: (v: { amountNaira: string; method: string }) =>
      api.post<{ wallet: Wallet }>("/wallet/fund", v),
    onSuccess: (res) => {
      if (res?.wallet) setWallet(res.wallet);
      void qc.invalidateQueries({ queryKey: qk.wallet });
      void qc.invalidateQueries({ queryKey: qk.dashboard });
    },
  });
}

export function useDashboard() {
  return useQuery({ queryKey: qk.dashboard, queryFn: () => api.get<DashboardStat>("/instructor/dashboard"), enabled: authed() });
}

export function useHistory() {
  return useQuery({ queryKey: qk.history, queryFn: () => api.get<HistoryItem[]>("/instructor/history"), enabled: authed() });
}

export function useQuizzes() {
  return useQuery({ queryKey: qk.quizzes, queryFn: () => api.get<QuizListItem[]>("/quizzes"), enabled: authed() });
}

export function useQuiz(id: string | undefined) {
  return useQuery({
    queryKey: qk.quiz(id ?? ""),
    queryFn: () => api.get<QuizDetail>(`/quizzes/${id}`),
    enabled: authed() && Boolean(id),
  });
}

export function useCreateQuiz() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (q: QuizDetail) => api.post<{ id: string }>("/quizzes", q),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.quizzes });
      void qc.invalidateQueries({ queryKey: qk.dashboard });
    },
  });
}

export function useOpenRoom() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (quizId: string) => api.post<{ id: string; code: string }>("/rooms", { quizId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.dashboard });
      void qc.invalidateQueries({ queryKey: qk.quizzes });
    },
  });
}

export function useRoom(roomId: string | undefined) {
  return useQuery({
    queryKey: qk.room(roomId ?? ""),
    queryFn: () => api.get<PublicRoom>(`/rooms/${roomId}`),
    enabled: Boolean(roomId),
    refetchInterval: 1500,
  });
}

export function useRoomByCode(code: string | undefined) {
  return useQuery({
    queryKey: qk.lookup(code ?? ""),
    queryFn: () => api.get<PublicRoom>(`/lookup/${code}`),
    enabled: Boolean(code && code.trim().length >= 3),
    refetchInterval: 1500,
  });
}

export function useStandings(roomId: string | undefined) {
  return useQuery({
    queryKey: qk.standings(roomId ?? ""),
    queryFn: () => api.get<Standing[]>(`/rooms/${roomId}/standings`),
    enabled: Boolean(roomId),
    refetchInterval: 2000,
  });
}

export function useJoinRoom() {
  return useMutation({
    mutationFn: (v: { roomId: string; email: string; nickname: string; avatar: string }) =>
      api.post<Participant>(`/rooms/${v.roomId}/join`, {
        email: v.email,
        nickname: v.nickname,
        avatar: v.avatar,
      }),
  });
}

export function useSubmitAnswer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { roomId: string; participantId: string; questionId: string; optionIndex: number }) =>
      api.post<AnswerReceipt>(`/rooms/${v.roomId}/answer`, {
        participantId: v.participantId,
        questionId: v.questionId,
        optionIndex: v.optionIndex,
      }),
    onSuccess: (_d, v) => {
      void qc.invalidateQueries({ queryKey: qk.standings(v.roomId) });
    },
  });
}

export function useRoomControl() {
  const qc = useQueryClient();
  const invalidate = (roomId: string) => {
    void qc.invalidateQueries({ queryKey: qk.room(roomId) });
    void qc.invalidateQueries({ queryKey: qk.standings(roomId) });
    void qc.invalidateQueries({ queryKey: qk.dashboard });
  };
  return {
    start: useMutation({
      mutationFn: (roomId: string) => api.post(`/rooms/${roomId}/start`),
      onSuccess: (_d, roomId) => invalidate(roomId),
    }),
    next: useMutation({
      mutationFn: (roomId: string) => api.post(`/rooms/${roomId}/next`),
      onSuccess: (_d, roomId) => invalidate(roomId),
    }),
    podium: useMutation({
      mutationFn: (roomId: string) => api.post<Standing[]>(`/rooms/${roomId}/podium`),
      onSuccess: (_d, roomId) => invalidate(roomId),
    }),
  };
}

export function useRedeem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { roomId: string; email: string }) =>
      api.post<ClaimResult>(`/rooms/${v.roomId}/redeem`, { email: v.email }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: qk.history });
    },
  });
}

// ---- Realtime ----

export interface WsEvent<T = unknown> {
  type: string;
  data: T;
}

const WS_BASE = () => {
  const base = import.meta.env.VITE_API_BASE ?? "/api";
  const wsBase = base.startsWith("http")
    ? base.replace(/^http/, "ws").replace(/\/api$/, "")
    : `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}`;
  return wsBase;
};

/** Subscribes to a room's websocket and returns the latest event + connection state. */
export function useRoomSocket(roomId: string | undefined) {
  const [event, setEvent] = useState<WsEvent | null>(null);
  const [connected, setConnected] = useState(false);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    if (!roomId) return;
    let ws: WebSocket | null = null;
    let closed = false;
    let timer: ReturnType<typeof setTimeout>;

    const open = () => {
      if (closed) return;
      ws = new WebSocket(`${WS_BASE()}/api/rooms/${roomId}/ws`);
      ws.onopen = () => setConnected(true);
      ws.onmessage = (m) => {
        try {
          const j = JSON.parse(m.data as string);
          setEvent(j);
        } catch {
          /* ignore malformed frame */
        }
      };
      ws.onclose = () => {
        setConnected(false);
        if (!closed) timer = setTimeout(() => setAttempt((a) => a + 1), 2000);
      };
      ws.onerror = () => ws?.close();
    };
    open();
    return () => {
      closed = true;
      clearTimeout(timer);
      ws?.close();
    };
  }, [roomId, attempt]);

  return { event, connected };
}
