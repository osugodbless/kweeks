import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Check, Coins, Hash, Mail, ShieldCheck, Sparkles, Timer, Trophy, UserRound, Users } from "lucide-react";
import { api, ApiError, type PublicRoom } from "@/lib/api";
import { useJoinRoom } from "@/lib/hooks";
import { usePlayer } from "@/lib/player";
import { AvatarPicker } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Field } from "@/components/ui/field";
import { Footer } from "@/components/ui/footer";
import { Money } from "@/components/ui/money";
import { PlayerTopBar } from "@/components/ui/player-topbar";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

const HOW_PLAYS = [
  { Icon: Hash, text: "Your host shares a 4-letter room code on screen." },
  { Icon: Users, text: "Pick an avatar and a nickname — that is your leaderboard face." },
  { Icon: Trophy, text: "Answer fast and correct; the podium splits the pool." },
];

export function PlayerJoin() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const { setRoom, setParticipant } = usePlayer();
  const joinRoom = useJoinRoom();

  const [code, setCode] = useState(params.get("code") ?? "");
  const [room, setRoomData] = useState<PublicRoom | null>(null);
  const [lookupError, setLookupError] = useState("");
  const [lookingUp, setLookingUp] = useState(false);

  const [nickname, setNickname] = useState("");
  const [email, setEmail] = useState("");
  const [avatarId, setAvatarId] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [joinError, setJoinError] = useState("");
  const [joining, setJoining] = useState(false);

  async function lookup(event?: FormEvent) {
    event?.preventDefault();
    const value = code.trim().toUpperCase();
    if (value.length < 4) {
      setLookupError("Room codes are 4 letters.");
      return;
    }
    setLookupError("");
    setLookingUp(true);
    try {
      const found = await api.get<PublicRoom>(`/lookup/${value}`);
      setRoomData(found);
    } catch (err) {
      setRoomData(null);
      setLookupError(err instanceof ApiError && err.message ? err.message : "We could not find that room.");
    } finally {
      setLookingUp(false);
    }
  }

  // When the host shares a deep link (/join?code=AB12), prefill + resolve the
  // room automatically so the player lands straight on the join form. One-shot,
  // guarded by the room/lookingUp flags.
  useEffect(() => {
    const fromLink = params.get("code");
    if (fromLink && !room && !lookingUp) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      void lookup();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  function validate() {
    const next: Record<string, string> = {};
    if (!nickname.trim()) next.nickname = "Pick a nickname for the leaderboard.";
    else if (nickname.trim().length > 16) next.nickname = "Keep it under 16 characters.";
    if (!email.trim()) next.email = "We need this to send your winnings.";
    else if (!EMAIL_RE.test(email.trim())) next.email = "That email does not look right.";
    if (!avatarId) next.avatar = "Choose your player avatar.";
    return next;
  }

  async function handleJoin(event: FormEvent) {
    event.preventDefault();
    if (!room) return;
    const next = validate();
    setErrors(next);
    if (Object.keys(next).length > 0) return;

    setJoinError("");
    setJoining(true);
    try {
      const participant = await joinRoom.mutateAsync({
        roomId: room.id,
        email: email.trim().toLowerCase(),
        nickname: nickname.trim(),
        avatar: avatarId!,
      });
      setRoom(room.id, room.code);
      setParticipant({
        participantId: participant.id,
        roomId: room.id,
        email: participant.email,
        nickname: participant.nickname,
        avatar: participant.avatar,
      });
      navigate("/lobby");
    } catch (err) {
      setJoinError(err instanceof ApiError ? err.message : "Could not join — try again.");
    } finally {
      setJoining(false);
    }
  }

  return (
    <main className="relative overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-24 size-72 rounded-full bg-sky/20 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 top-1/2 size-80 rounded-full bg-coral/20 blur-3xl" />
      <span className="pointer-events-none absolute bottom-0 left-1/3 size-72 rounded-full bg-mint/15 blur-3xl" />

      <PlayerTopBar status="join" code={room?.code} />

      <div className="relative mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 lg:py-16">
        <div className="grid items-start gap-12 lg:grid-cols-[0.92fr_1.08fr]">
          {/* Left: pool promo */}
          <div className="lg:sticky lg:top-24">
            <h1 className="font-display text-5xl font-semibold leading-[1.02] tracking-tight sm:text-6xl">
              Ready to <span className="text-pop">play?</span>
            </h1>
            <p className="mt-4 max-w-md text-lg leading-relaxed text-soft">
              Grab the code your host is showing, claim your spot on the leaderboard, and
              answer fast enough to take a cut of the pool.
            </p>

            <div className="mt-8 rounded-[2rem] border-2 border-mint/30 bg-mint/10 p-6">
              <p className="text-xs font-bold uppercase tracking-[0.25em] text-mint-dark">Prize pool</p>
              <Money value={room?.poolNaira ?? "50000"} className="mt-1 block text-5xl" />
              <p className="mt-2 text-sm leading-relaxed text-ink">
                {room
                  ? `Live in "${room.title}" · ${room.participantCount} player${room.participantCount === 1 ? "" : "s"} in the room.`
                  : "Enter the code on your host's screen to see tonight's pool."}
              </p>
            </div>

            <ol className="mt-8 space-y-4">
              {HOW_PLAYS.map((step, i) => (
                <li key={step.text} className="flex items-start gap-4">
                  <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-cream text-violet shadow-[0_4px_0_rgba(108,76,241,0.15)]">
                    <step.Icon className="size-5" strokeWidth={2.4} />
                  </span>
                  <div>
                    <p className="font-display text-lg font-semibold">
                      <span className="mr-2 text-soft/70">{String(i + 1).padStart(2, "0")}</span>
                      {i === 0 ? "Enter the room code" : i === 1 ? "Pick an avatar & nickname" : "Answer fast, cash out"}
                    </p>
                    <p className="text-[15px] text-soft">{step.text}</p>
                  </div>
                </li>
              ))}
            </ol>

            <div className="mt-9 flex items-start gap-4 rounded-3xl border-2 border-gold/30 bg-gold/10 p-5">
              <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-gold text-ink">
                <Coins className="size-5" strokeWidth={2.4} />
              </span>
              <p className="text-sm leading-relaxed text-ink">
                <span className="font-bold">Your email is your payout address.</span> Winnings
                from every room are sent to the address you join with — keep it accurate.
              </p>
            </div>
          </div>

          {/* Right: join card */}
          <section className="relative rounded-[2rem] border-2 border-ink/5 bg-cream p-6 card-3d sm:p-9">
            <span className="inline-flex items-center gap-2 rounded-full bg-coral/10 px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-coral">
              <Sparkles className="size-4" /> Join a game
            </span>
            <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight sm:text-4xl">
              {room ? `Join "${room.title}"` : "Enter the room code"}
            </h2>
            <p className="mt-1.5 text-[15px] text-soft">
              {room
                ? `Hosted by ${room.host?.name || "your host"} · ${room.questionCount} questions · top ${room.winnerCount} split the pool`
                : "All fields needed before the clock starts."}
            </p>

            {!room ? (
              <form onSubmit={lookup} noValidate className="mt-8 space-y-5">
                <Field
                  name="code"
                  label="Room code"
                  icon={<Hash className="size-5" />}
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 4))}
                  placeholder="e.g. AB12"
                  error={lookupError}
                  autoComplete="off"
                  spellCheck="false"
                  maxLength={4}
                />
                <Button type="submit" size="lg" loading={lookingUp} className="w-full">
                  Find room
                </Button>
              </form>
            ) : (
              <form onSubmit={handleJoin} noValidate className="mt-8 space-y-6">
                <div className="flex items-center gap-2 rounded-2xl border-2 border-mint/30 bg-mint/10 px-4 py-3 text-sm font-bold text-mint-dark">
                  <Check className="size-4" strokeWidth={3} />
                  Room found — pool locked at <Money value={room.poolNaira} className="text-sm" />
                </div>

                <div className="grid gap-6 sm:grid-cols-2">
                  <Field
                    name="nickname"
                    label="Nickname"
                    icon={<UserRound className="size-5" />}
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value.slice(0, 16))}
                    placeholder="e.g. QuizWizard"
                    error={errors.nickname}
                    maxLength={16}
                    autoComplete="off"
                    spellCheck="false"
                  />
                  <Field
                    name="email"
                    label="Email"
                    icon={<Mail className="size-5" />}
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                    error={errors.email}
                    autoComplete="email"
                    spellCheck="false"
                  />
                </div>

                <div>
                  <div className="mb-2 flex items-center justify-between">
                    <label htmlFor="avatar-grid" className="block text-sm font-bold">
                      Pick your player
                    </label>
                    {avatarId && (
                      <span className="flex items-center gap-1.5 text-xs font-bold text-violet-dark">
                        <Check className="size-3.5" strokeWidth={3} />
                        Avatar selected
                      </span>
                    )}
                  </div>
                  <AvatarPicker value={avatarId} onChange={setAvatarId} error={errors.avatar} />
                </div>

                {joinError && (
                  <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                    {joinError}
                  </p>
                )}

                <Button type="submit" size="lg" loading={joining} icon={<Trophy className="size-5" />} className="w-full">
                  {joining ? "Joining…" : "Join the game"}
                </Button>

                <p className="flex items-start gap-2 text-xs leading-relaxed text-soft">
                  <ShieldCheck className="mt-0.5 size-4 shrink-0 text-mint-dark" />
                  By joining you accept the official rules. Winnings are paid to the payout
                  email above — keep it accurate.
                </p>
              </form>
            )}

            <div className="mt-6 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm font-bold text-soft">
              <Chip tone="gold">
                <Trophy className="size-3.5" /> Top {room?.winnerCount ?? 3} paid
              </Chip>
            </div>
          </section>
        </div>

        <p className="mt-12 flex items-center justify-center gap-2 text-sm font-bold text-soft">
          <Timer className="size-4 text-coral" />
          First question drops seconds after you join — no app download needed.
        </p>
      </div>

      <Footer variant="player" />
    </main>
  );
}
