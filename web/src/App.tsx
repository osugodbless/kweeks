import { useEffect, type ReactNode } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { useAuth } from "@/lib/auth";
import { LandingPage } from "@/pages/LandingPage";
import { PlayerJoin } from "@/pages/player/PlayerJoin";
import { PlayerLobby } from "@/pages/player/PlayerLobby";
import { PlayerQuestion } from "@/pages/player/PlayerQuestion";
import { PlayerStandings } from "@/pages/player/PlayerStandings";
import { PlayerPodium } from "@/pages/player/PlayerPodium";
import { ClaimPage } from "@/pages/player/ClaimPage";
import { InstructorSignup } from "@/pages/instructor/InstructorSignup";
import { InstructorLogin } from "@/pages/instructor/InstructorLogin";
import { InstructorWallet } from "@/pages/instructor/InstructorWallet";
import { InstructorFundWallet } from "@/pages/instructor/InstructorFundWallet";
import { InstructorQuizBuilder } from "@/pages/instructor/InstructorQuizBuilder";
import { InstructorLiveRoom } from "@/pages/instructor/InstructorLiveRoom";
import { InstructorHistory } from "@/pages/instructor/InstructorHistory";
import { InstructorHistoryEmpty } from "@/pages/instructor/InstructorHistoryEmpty";

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "instant" });
  }, [pathname]);
  return null;
}

function Guard({ children }: { children: ReactNode }) {
  const { instructor, loaded } = useAuth();
  if (!loaded) return null;
  if (!instructor) return <Navigate to="/instructor/login" replace />;
  return <>{children}</>;
}

export default function App() {
  const refresh = useAuth((s) => s.refresh);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <>
      <ScrollToTop />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/join" element={<PlayerJoin />} />
        <Route path="/lobby" element={<PlayerLobby />} />
        <Route path="/question" element={<PlayerQuestion />} />
        <Route path="/standings" element={<PlayerStandings />} />
        <Route path="/podium" element={<PlayerPodium />} />
        <Route path="/claim" element={<ClaimPage />} />

        <Route path="/instructor/signup" element={<InstructorSignup />} />
        <Route path="/instructor/login" element={<InstructorLogin />} />
        <Route
          path="/instructor/dashboard"
          element={
            <Guard>
              <InstructorWallet />
            </Guard>
          }
        />
        <Route
          path="/instructor/fund"
          element={
            <Guard>
              <InstructorFundWallet />
            </Guard>
          }
        />
        <Route
          path="/instructor/quiz-builder"
          element={
            <Guard>
              <InstructorQuizBuilder />
            </Guard>
          }
        />
        <Route
          path="/instructor/live-room"
          element={
            <Guard>
              <InstructorLiveRoom />
            </Guard>
          }
        />
        <Route
          path="/instructor/history"
          element={
            <Guard>
              <InstructorHistory />
            </Guard>
          }
        />
        <Route
          path="/instructor/history-empty"
          element={
            <Guard>
              <InstructorHistoryEmpty />
            </Guard>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  );
}
