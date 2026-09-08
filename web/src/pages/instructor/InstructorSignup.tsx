import { useState } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Eye, EyeOff, Lock, Mail, Phone, ShieldCheck, UserPlus, UserRound } from "lucide-react";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { AuthShell } from "@/components/ui/auth";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;
const PHONE_RE = /^\+?\d{10,15}$/;

export function InstructorSignup() {
  const navigate = useNavigate();
  const { signup } = useAuth();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    if (!name.trim()) {
      setError("Enter your full name.");
      return;
    }
    if (!EMAIL_RE.test(email.trim())) {
      setError("Enter a valid email address.");
      return;
    }
    if (!PHONE_RE.test(phone.trim())) {
      setError("Enter your phone number (E.164, e.g. +2348012345678).");
      return;
    }
    if (password.length < 6) {
      setError("Password must be at least 6 characters.");
      return;
    }

    setSubmitting(true);
    try {
      await signup(name.trim(), email.trim().toLowerCase(), phone.trim(), password);
      navigate("/instructor/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Sign up failed — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell tagline="Your own wallet is issued at signup">
      <div
        className="rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d sm:p-9"
        style={{ animation: "rise 0.6s cubic-bezier(0.16,1,0.3,1) both" }}
      >
        <h1 className="font-display text-3xl font-semibold tracking-tight">Create your host account</h1>
        <p className="mt-1.5 text-[15px] text-soft">
          One account to build quizzes, fund pools and pay out winners.
        </p>

        <form onSubmit={handleSubmit} noValidate className="mt-7 space-y-5">
          <Field
            name="name"
            label="Full name"
            icon={<UserRound className="size-5" />}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Adeola Peters"
            autoComplete="name"
            spellCheck="false"
          />
          <Field
            name="email"
            label="Email"
            type="email"
            icon={<Mail className="size-5" />}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            autoComplete="email"
            spellCheck="false"
          />
          <Field
            name="phone"
            label="Phone number"
            type="tel"
            icon={<Phone className="size-5" />}
            value={phone}
            onChange={(e) => setPhone(e.target.value.replace(/[^\d+]/g, "").slice(0, 16))}
            placeholder="+2348012345678"
            autoComplete="tel"
            spellCheck="false"
          />
          <Field
            name="password"
            label="Password"
            type={showPassword ? "text" : "password"}
            icon={<Lock className="size-5" />}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 6 characters"
            autoComplete="new-password"
            right={
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? "Hide password" : "Show password"}
                className="grid size-9 cursor-pointer place-items-center rounded-xl text-soft transition-colors hover:bg-ink/5 hover:text-ink"
              >
                {showPassword ? <EyeOff className="size-5" /> : <Eye className="size-5" />}
              </button>
            }
          />

          {error && (
            <p
              className="flex items-start gap-2 rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral"
              role="alert"
            >
              <ShieldCheck className="mt-0.5 size-4 shrink-0" />
              {error}
            </p>
          )}

          <Button type="submit" variant="coral" size="lg" loading={submitting} icon={<UserPlus className="size-5" />} className="w-full">
            {submitting ? "Creating…" : "Create account"}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-soft">
          Already have an account?{" "}
          <Link to="/instructor/login" className="font-bold text-violet underline-offset-4 transition-colors hover:text-violet-dark hover:underline">
            Sign in instead
          </Link>
        </p>
      </div>

      <p className="mt-6 text-center text-xs leading-relaxed text-soft">
        Host accounts are free. Players always play free. Your phone number is your private BMONI
        wallet identity — it is never shared with players.
      </p>
    </AuthShell>
  );
}

export default InstructorSignup;
