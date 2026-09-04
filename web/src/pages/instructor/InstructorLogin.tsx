import { useState, type FormEvent } from "react";
import { useNavigate, Link } from "react-router-dom";
import { AuthShell } from "@/components/ui/auth";
import { useAuth } from "@/lib/auth";

export function InstructorLogin() {
  const nav = useNavigate();
  const login = useAuth((s) => s.login);
  const [email, setEmail] = useState("");
  const [pass, setPass] = useState("");
  const [pending, setPending] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const ready = email.trim() !== "" && pass.length >= 6;

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (!ready || pending) return;
    setPending(true);
    setErr(null);
    try {
      await login(email.trim(), pass);
      nav("/instructor/dashboard");
    } catch (x) {
      setErr(x instanceof Error ? x.message : "Login failed — check your details and try again.");
      setPending(false);
    }
  };

  return (
    <AuthShell
      lead="Welcome back to the arena."
      sub="Your wallet, your quizzes, your live rooms — all waiting behind this door."
      chipLabel="₦150,000 ready to host"
      foot="Naira only · bank-grade escrow · instant payouts"
    >
      <h2 className="font-display text-[26px] font-extrabold text-paper">
        Log in to your account
      </h2>
      <p className="mt-1 font-body text-[13.5px] text-text-2">Pick up where you left off.</p>

      <form onSubmit={submit} noValidate>
        <div className="mt-5 flex flex-col gap-[18px]">
          <label className="block">
            <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
              EMAIL
            </span>
            <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface px-4">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="host@kweeks.ng"
                className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
              />
            </div>
          </label>
          <label className="block">
            <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
              PASSWORD
            </span>
            <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface px-4">
              <input
                type="password"
                value={pass}
                onChange={(e) => setPass(e.target.value)}
                placeholder="••••••••••"
                className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
              />
            </div>
          </label>
        </div>

        <button
          type="submit"
          disabled={!ready || pending}
          className="mt-6 flex h-14 w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90 disabled:opacity-40"
        >
          {pending ? "LOGGING IN…" : "LOG IN"}
        </button>

        {err && (
          <p className="mt-4 rounded-2xl border border-stroke bg-surface px-4 py-3 text-center font-body text-[13px] font-semibold text-red">
            {err}
          </p>
        )}
      </form>

      <div className="mt-5 flex items-center justify-center gap-1.5">
        <span className="font-body text-[13px] text-text-3">New to Kweeks?</span>
        <Link to="/instructor/signup" className="font-body text-[13px] font-bold text-gold">
          Create an account
        </Link>
      </div>
    </AuthShell>
  );
}
