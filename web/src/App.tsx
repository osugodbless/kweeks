import { useEffect } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { LandingPage } from "@/pages/LandingPage";
import { PlayerJoin } from "@/pages/player/PlayerJoin";
import { PlayerLobby } from "@/pages/player/PlayerLobby";
import { PlayerQuestion } from "@/pages/player/PlayerQuestion";
import { PlayerStandings } from "@/pages/player/PlayerStandings";
import { PlayerPodium } from "@/pages/player/PlayerPodium";
import { InstructorSignup } from "@/pages/instructor/InstructorSignup";
import { InstructorLogin } from "@/pages/instructor/InstructorLogin";
import { InstructorWallet } from "@/pages/instructor/InstructorWallet";
import { InstructorFundWallet } from "@/pages/instructor/InstructorFundWallet";
import { InstructorQuizBuilder } from "@/pages/instructor/InstructorQuizBuilder";
import { InstructorLiveRoom } from "@/pages/instructor/InstructorLiveRoom";
import { InstructorHistory } from "@/pages/instructor/InstructorHistory";
import { InstructorHistoryEmpty } from "@/pages/instructor/InstructorHistoryEmpty";
import { useAuth } from "@/lib/auth";

function Guard({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem("kweeks.token");
  if (!token) return <Navigate to="/instructor/login" replace />;
  return <>{children}</>;
}

export default function App() {
  const refresh = useAuth((s) => s.refresh);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/join" element={<PlayerJoin />} />
      <Route path="/lobby" element={<PlayerLobby />} />
      <Route path="/question" element={<PlayerQuestion />} />
      <Route path="/standings" element={<PlayerStandings />} />
      <Route path="/podium" element={<PlayerPodium />} />
      <Route path="/instructor/signup" element={<InstructorSignup />} />
      <Route path="/instructor/login" element={<InstructorLogin />} />
      <Route
        path="/instructor/dashboard"
        element={<Guard><InstructorWallet /></Guard>}
      />
      <Route
        path="/instructor/fund"
        element={<Guard><InstructorFundWallet /></Guard>}
      />
      <Route
        path="/instructor/quiz-builder"
        element={<Guard><InstructorQuizBuilder /></Guard>}
      />
      <Route
        path="/instructor/live-room"
        element={<Guard><InstructorLiveRoom /></Guard>}
      />
      <Route
        path="/instructor/history"
        element={<Guard><InstructorHistory /></Guard>}
      />
      <Route
        path="/instructor/history-empty"
        element={<Guard><InstructorHistoryEmpty /></Guard>}
      />
    </Routes>
  );
}
