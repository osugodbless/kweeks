import { useState } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ArrowRight, Eye, EyeOff, Lock, Mail, ShieldCheck } from "lucide-react";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { AuthShell } from "@/components/ui/auth";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

export function InstructorLogin() {
  const navigate = useNavigate();
  const { login } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    if (!EMAIL_RE.test(email.trim())) {
      setError("Enter a valid email address.");
      return;
    }
    if (!password) {
      setError("Enter your password.");
      return;
    }

    setSubmitting(true);
    try {
      await login(email.trim().toLowerCase(), password);
      navigate("/instructor/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Sign in failed — try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell tagline="Ready to host" chipAmount="0">
      <div
        className="rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d sm:p-9"
        style={{ animation: "rise 0.6s cubic-bezier(0.16,1,0.3,1) both" }}
      >
        <h1 className="font-display text-3xl font-semibold tracking-tight">Welcome back, host</h1>
        <p className="mt-1.5 text-[15px] text-soft">
          Sign in to build quizzes, fund pools and manage payouts.
        </p>

        <form onSubmit={handleSubmit} noValidate className="mt-7 space-y-5">
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
            name="password"
            label="Password"
            type={showPassword ? "text" : "password"}
            icon={<Lock className="size-5" />}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Your password"
            autoComplete="current-password"
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

          <Button type="submit" variant="violet" size="lg" loading={submitting} icon={<ArrowRight className="size-5" />} className="w-full">
            {submitting ? "Signing in…" : "Sign in"}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-soft">
          New to Kweeks?{" "}
          <Link to="/instructor/signup" className="font-bold text-violet underline-offset-4 transition-colors hover:text-violet-dark hover:underline">
            Create an account
          </Link>
        </p>
      </div>

      <p className="mt-6 text-center text-xs leading-relaxed text-soft">
        Host accounts are free. Players always play free. By continuing you agree to the platform rules.
      </p>
    </AuthShell>
  );
}

export default InstructorLogin;
