import { useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Banknote, Check, Landmark, Sparkles, Zap } from "lucide-react";
import { ApiError } from "@/lib/api";
import { useDepositAccount, useFundWallet } from "@/lib/hooks";
import { useAuth } from "@/lib/auth";
import { naira } from "@/lib/player";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { Money } from "@/components/ui/money";
import { cn } from "@/lib/cn";

const QUICK_PICKS = ["1000", "5000", "50000", "100000"];

const METHODS = [
  { id: "credit", label: "Wallet credit", Icon: Zap, copy: "Instant platform credit" },
  { id: "transfer", label: "Bank transfer", Icon: Landmark, copy: "Transfer to your NGN account" },
];

export function InstructorFundWallet() {
  const navigate = useNavigate();
  const fund = useFundWallet();
  const { wallet } = useAuth();
  const { data: deposit, isError: depositError } = useDepositAccount();

  const [amount, setAmount] = useState("50000");
  const [method, setMethod] = useState("credit");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const numeric = useMemo(() => parseInt(amount || "0", 10), [amount]);
  const valid = numeric > 0;

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setSuccess("");
    if (!valid) {
      setError("Enter an amount to fund.");
      return;
    }
    if (method !== "credit") return; // bank transfer settles via the deposit account, not here
    try {
      await fund.mutateAsync({ amountNaira: String(numeric), method });
      setSuccess(`Wallet credited with ${naira(numeric)} — it is spendable right away.`);
      setAmount("");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Funding failed — try again.");
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden pb-24">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-24 size-72 rounded-full bg-mint/15 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-24 size-80 rounded-full bg-gold/15 blur-3xl" />

      <InstructorNav active="/instructor/dashboard" />

      <div className="relative mx-auto w-full max-w-2xl px-4 pt-12 sm:px-6 lg:pt-16">
        <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
          <Banknote className="size-4 text-mint-dark" /> Fund wallet
        </span>
        <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
          Add naira to your <span className="text-pop">wallet</span>
        </h1>
        <p className="mt-3 max-w-md text-lg text-soft">
          Funds land instantly and are spendable on quiz pools. Available balance:{" "}
          <Money value={wallet?.balanceNaira ?? "0"} className="text-base" />
        </p>

        <form onSubmit={handleSubmit} noValidate className="mt-10 space-y-8">
          <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d">
            <Field
              name="amount"
              label="Amount"
              type="number"
              min={1}
              inputMode="numeric"
              value={amount}
              onChange={(e) => setAmount(e.target.value.replace(/[^0-9]/g, "").slice(0, 9))}
              placeholder="0"
              error={error}
              right={<span className="pr-3 font-display text-lg font-semibold text-soft">₦</span>}
            />

            <div className="mt-5 flex flex-wrap gap-2.5">
              {QUICK_PICKS.map((p) => {
                const active = amount === p;
                return (
                  <button
                    key={p}
                    type="button"
                    onClick={() => setAmount(p)}
                    className={cn(
                      "cursor-pointer rounded-full border-2 px-4 py-2 text-sm font-bold transition-all",
                      active
                        ? "border-mint bg-mint text-white"
                        : "border-ink/10 bg-white text-soft hover:border-mint hover:text-mint-dark",
                    )}
                  >
                    {naira(p)}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d">
            <p className="text-sm font-bold">Funding method</p>
            <div className="mt-4 grid gap-3">
              {METHODS.map((m) => {
                const active = method === m.id;
                return (
                  <button
                    key={m.id}
                    type="button"
                    role="radio"
                    aria-checked={active}
                    onClick={() => setMethod(m.id)}
                    className={cn(
                      "flex cursor-pointer items-center gap-4 rounded-2xl border-2 px-5 py-4 text-left transition-all",
                      active
                        ? "border-violet bg-violet/5 shadow-[0_0_0_4px_rgba(108,76,241,0.12)]"
                        : "border-ink/10 bg-white hover:border-ink/25",
                    )}
                  >
                    <span
                      className={cn(
                        "grid size-11 shrink-0 place-items-center rounded-2xl",
                        active ? "bg-violet text-white" : "bg-ink/5 text-soft",
                      )}
                    >
                      <m.Icon className="size-5" strokeWidth={2.2} />
                    </span>
                    <span className="flex-1">
                      <span className="block font-bold">{m.label}</span>
                      <span className="block text-xs text-soft">{m.copy}</span>
                    </span>
                    {active && (
                      <span className="grid size-5 place-items-center rounded-full bg-violet text-white">
                        <Check className="size-3" strokeWidth={3.5} />
                      </span>
                    )}
                  </button>
                );
              })}
            </div>

            {method === "transfer" && (
              <div className="mt-5 rounded-2xl border-2 border-mint/30 bg-mint/10 p-5">
                <p className="text-sm font-bold text-mint-dark">Send a bank transfer to this account</p>
                {deposit ? (
                  <>
                    <p className="mt-3 rounded-2xl bg-white px-5 py-4 font-display text-2xl font-semibold tracking-wide text-ink">
                      {deposit.accountNumber}
                    </p>
                    <p className="mt-2 text-sm font-bold text-mint-dark">{deposit.bankName} · NGN</p>
                    <p className="mt-2 text-xs leading-relaxed text-soft">
                      Your wallet is credited automatically when the transfer lands. No code needed.
                    </p>
                  </>
                ) : (
                  <p className="mt-2 text-sm leading-relaxed text-soft">
                    {depositError
                      ? "Your wallet is not set up on the money rail yet — complete the setup wizard on your dashboard, or use wallet credit."
                      : "Fetching your NGN account…"}
                  </p>
                )}
              </div>
            )}
          </div>

          {success && (
            <p className="flex items-start gap-2 rounded-2xl border-2 border-mint/30 bg-mint/10 px-4 py-3 text-sm font-bold text-mint-dark" role="status">
              <Check className="mt-0.5 size-4 shrink-0" strokeWidth={3} />
              {success}
            </p>
          )}

          {method === "credit" && (
            <>
              <Button type="submit" variant="mint" size="lg" loading={fund.isPending} className="w-full" icon={<Sparkles className="size-5" />}>
                {fund.isPending ? "Funding…" : `Fund ${valid ? naira(numeric) : "wallet"}`}
              </Button>
              <p className="text-center text-xs leading-relaxed text-soft">
                Wallet credits post instantly to your available balance. Bank transfers settle through
                your NGN account instead.
              </p>
            </>
          )}

          <div className="text-center">
            <button type="button" onClick={() => navigate("/instructor/dashboard")} className="cursor-pointer text-sm font-bold text-violet hover:text-violet-dark">
              ← Back to dashboard
            </button>
          </div>
        </form>
      </div>

      <Footer className="mt-16" />
    </main>
  );
}
