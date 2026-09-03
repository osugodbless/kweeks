import { Routes, Route } from "react-router-dom";
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

export default function App() {
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
      <Route path="/instructor/dashboard" element={<InstructorWallet />} />
      <Route path="/instructor/fund" element={<InstructorFundWallet />} />
      <Route path="/instructor/quiz-builder" element={<InstructorQuizBuilder />} />
      <Route path="/instructor/live-room" element={<InstructorLiveRoom />} />
      <Route path="/instructor/history" element={<InstructorHistory />} />
      <Route path="/instructor/history-empty" element={<InstructorHistoryEmpty />} />
    </Routes>
  );
}
