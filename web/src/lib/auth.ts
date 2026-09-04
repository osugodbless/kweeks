import { create } from "zustand";
import { api, AuthResult, Instructor, setToken, Wallet } from "@/lib/api";

interface AuthState {
  instructor: Instructor | null;
  wallet: Wallet | null;
  loaded: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (name: string, email: string, password: string) => Promise<void>;
  refresh: () => Promise<void>;
  setWallet: (w: Wallet) => void;
  logout: () => void;
}

export const useAuth = create<AuthState>((set) => ({
  instructor: null,
  wallet: null,
  loaded: false,

  login: async (email, password) => {
    const res = await api.post<AuthResult>("/auth/login", { email, password });
    setToken(res.token);
    set({ instructor: res.instructor, wallet: res.wallet, loaded: true });
  },

  signup: async (name, email, password) => {
    const res = await api.post<AuthResult>("/auth/signup", { name, email, password });
    setToken(res.token);
    set({ instructor: res.instructor, wallet: res.wallet, loaded: true });
  },

  refresh: async () => {
    if (!localStorage.getItem("kweeks.token")) {
      set({ loaded: true });
      return;
    }
    try {
      const res = await api.get<{ instructor: Instructor; wallet: Wallet }>("/auth/me");
      set({ instructor: res.instructor, wallet: res.wallet, loaded: true });
    } catch {
      setToken(null);
      set({ instructor: null, wallet: null, loaded: true });
    }
  },

  setWallet: (w) => set((s) => ({ wallet: w ?? s.wallet })),

  logout: () => {
    setToken(null);
    set({ instructor: null, wallet: null, loaded: true });
  },
}));
