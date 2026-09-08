import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { Banknote, Check, ChevronDown, Hash, Landmark, Loader2, Mail, PartyPopper, ShieldCheck, Sparkles } from "lucide-react";
import { ApiError } from "@/lib/api";
import { useResolveClaim, useSubmitPayout } from "@/lib/hooks";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Footer } from "@/components/ui/footer";
import { Money } from "@/components/ui/money";
import { PlayerTopBar } from "@/components/ui/player-topbar";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;
const NUBAN_RE = /^\d{10}$/;

const PAYOUT_STATES: Record<string, { title: string; copy: string }> = {
  bank_submitted: { title: "Account locked in", copy: "Your bank account is verified. We are moving your prize now." },
  paying: { title: "Payout in flight", copy: "The money is on its way from the host wallet to your bank account." },
  paid: { title: "Paid out", copy: "The transfer settled. Check your bank — it should be there shortly." },
  failed: { title: "Payout needs attention", copy: "The transfer could not complete. Please try again or contact the host." },
};

export function ClaimPage() {
  const [params] = useSearchParams();
  const resolveClaim = useResolveClaim();
  const submitPayout = useSubmitPayout();

  // Step 1: identify the claim.
  const [claimCode, setClaimCode] = useState(params.get("code") ?? "");
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [lookupError, setLookupError] = useState("");

  // Step 2: bank details.
  const [accountNumber, setAccountNumber] = useState("");
  const [bankCode, setBankCode] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [payoutError, setPayoutError] = useState("");

  const claim = resolveClaim.data?.claim ?? null;
  const banks = resolveClaim.data?.banks ?? [];
  const paid = submitPayout.data?.claim ?? null;

  useEffect(() => {
    const code = params.get("code");
    const mail = params.get("email");
    if (code && mail && !resolveClaim.isSuccess && !resolveClaim.isPending) {
      void resolveClaim.mutateAsync({ claimCode: code, email: mail }).catch(() => {});
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  async function handleLookup(event: FormEvent) {
    event.preventDefault();
    setLookupError("");
    if (!claimCode.trim()) {
      setLookupError("Enter the claim code from your email or podium screen.");
      return;
    }
    if (!EMAIL_RE.test(email.trim())) {
      setLookupError("Enter the email you joined with.");
      return;
    }
    try {
      await resolveClaim.mutateAsync({ claimCode: claimCode.trim(), email: email.trim().toLowerCase() });
    } catch (err) {
      setLookupError(err instanceof ApiError && err.message ? err.message : "We could not find that claim.");
    }
  }

  function validateBank() {
    const next: Record<string, string> = {};
    if (!NUBAN_RE.test(accountNumber.trim())) next.accountNumber = "Enter the 10-digit account number.";
    if (!bankCode) next.bank = "Pick your bank.";
    return next;
  }

  async function handlePayout(event: FormEvent) {
    event.preventDefault();
    if (!claim) return;
    setPayoutError("");
    const next = validateBank();
    setErrors(next);
    if (Object.keys(next).length > 0) return;

    const bank = banks.find((b) => b.code === bankCode);
    try {
      await submitPayout.mutateAsync({
        claimCode: claimCode.trim(),
        email: email.trim().toLowerCase(),
        accountNumber: accountNumber.trim(),
        bankCode,
        bankName: bank?.name ?? "",
      });
    } catch (err) {
      setPayoutError(err instanceof ApiError ? err.message : "Payout failed — check the account details and try again.");
    }
  }

  const finalState = paid?.state ?? claim?.state ?? null;
  const status = finalState ? PAYOUT_STATES[finalState] : null;
  const selecting = !claim;

  const leftCard = useMemo(
    () => (
      <div className="rounded-[2rem] border-2 border-gold/30 bg-gold/10 p-6">
        <p className="text-xs font-bold uppercase tracking-[0.25em] text-gold-dark">Your prize</p>
        <Money value={claim?.amountNaira ?? "0"} className="mt-1 block text-5xl" />
        <p className="mt-2 text-sm leading-relaxed text-ink">
          {claim
            ? `Claim code ${claim.claimCode?.slice(0, 8)}… · sent to ${email}`
            : "Enter the code from your email or podium screen to lock in your winnings."}
        </p>
        <div className="mt-4 flex items-start gap-3 rounded-2xl border-2 border-mint/30 bg-mint/10 p-4">
          <ShieldCheck className="mt-0.5 size-5 shrink-0 text-mint-dark" />
          <p className="text-sm leading-relaxed text-ink">
            <span className="font-bold">Your claim code is private.</span> Anyone with it can claim your
            prize — keep it to yourself.
          </p>
        </div>
      </div>
    ),
    [claim, email],
  );

  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-24 size-72 rounded-full bg-gold/15 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-24 size-80 rounded-full bg-mint/15 blur-3xl" />

      <PlayerTopBar status="claim" />

      <div className="relative mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 lg:py-16">
        <div className="grid items-start gap-12 lg:grid-cols-[0.85fr_1.15fr]">
          <div className="lg:sticky lg:top-24">
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
              <Sparkles className="size-4 text-gold" /> Claim your prize
            </span>
            <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
              Your winnings, <span className="text-pop">sorted</span>
            </h1>
            <p className="mt-3 max-w-md text-lg text-soft">
              Enter your claim code, pick the Nigerian bank account you want paid into, and the host
              wallet transfers your prize in seconds.
            </p>
            <div className="mt-8">{leftCard}</div>
          </div>

          <section className="relative rounded-[2rem] border-2 border-ink/5 bg-cream p-6 card-3d sm:p-9">
            {selecting ? (
              <>
                <span className="inline-flex items-center gap-2 rounded-full bg-coral/10 px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-coral">
                  <Hash className="size-4" /> Step 1 · Identify
                </span>
                <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight sm:text-4xl">
                  Find your claim
                </h2>
                <p className="mt-1.5 text-[15px] text-soft">
                  Use the claim code we emailed you (or showed on your podium screen) and the email
                  you joined with.
                </p>
                <form onSubmit={handleLookup} noValidate className="mt-8 space-y-5">
                  <Field
                    name="claimCode"
                    label="Claim code"
                    icon={<Hash className="size-5" />}
                    value={claimCode}
                    onChange={(e) => setClaimCode(e.target.value.trim().slice(0, 64))}
                    placeholder="Paste your claim code"
                    error={lookupError}
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
                    autoComplete="email"
                    spellCheck="false"
                  />
                  <Button type="submit" size="lg" loading={resolveClaim.isPending} className="w-full" icon={<Check className="size-5" />}>
                    {resolveClaim.isPending ? "Looking up…" : "Find my claim"}
                  </Button>
                  {lookupError && (
                    <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                      {lookupError}
                    </p>
                  )}
                </form>
              </>
            ) : status ? (
              <>
                <span
                  className={`inline-flex items-center gap-2 rounded-full px-4 py-1.5 text-sm font-bold uppercase tracking-widest ${
                    finalState === "paid"
                      ? "bg-mint/10 text-mint-dark"
                      : finalState === "failed"
                        ? "bg-coral/10 text-coral"
                        : "bg-gold/10 text-gold-dark"
                  }`}
                >
                  {finalState === "paid" ? <PartyPopper className="size-4" /> : <Landmark className="size-4" />}
                  Step 2 · Payout
                </span>
                <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight sm:text-4xl">{status.title}</h2>
                <p className="mt-2 text-[15px] leading-relaxed text-soft">{status.copy}</p>

                <div className="mt-6 space-y-3">
                  <div className="flex items-center justify-between rounded-2xl border-2 border-ink/5 bg-white px-5 py-4">
                    <span className="text-sm font-bold text-soft">Prize</span>
                    <Money value={claim?.amountNaira ?? "0"} className="text-base" />
                  </div>
                  <div className="flex items-center justify-between rounded-2xl border-2 border-ink/5 bg-white px-5 py-4">
                    <span className="text-sm font-bold text-soft">Account</span>
                    <span className="text-sm font-bold">
                      {paid?.payoutRef ? `•• ${accountNumber.slice(-4)} · ${banks.find((b) => b.code === bankCode)?.name ?? "bank"}` : "Confirming…"}
                    </span>
                  </div>
                  {paid?.payoutRef && (
                    <div className="flex items-center justify-between rounded-2xl border-2 border-mint/30 bg-mint/10 px-5 py-4">
                      <span className="text-sm font-bold text-mint-dark">Settlement ref</span>
                      <span className="text-sm font-bold text-mint-dark">{paid.payoutRef}</span>
                    </div>
                  )}
                </div>
              </>
            ) : (
              <>
                <span className="inline-flex items-center gap-2 rounded-full bg-violet/10 px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-violet">
                  <Banknote className="size-4" /> Step 2 · Bank details
                </span>
                <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight sm:text-4xl">
                  Where should we pay?
                </h2>
                <p className="mt-1.5 text-[15px] text-soft">
                  Pick the bank account your <Money value={claim?.amountNaira ?? "0"} className="text-base" /> prize
                  lands in. The name on the account is verified automatically.
                </p>

                <form onSubmit={handlePayout} noValidate className="mt-8 space-y-5">
                  <div>
                    <label htmlFor="bank" className="mb-2 block text-sm font-bold">
                      Bank
                    </label>
                    <div className="relative">
                      <select
                        id="bank"
                        value={bankCode}
                        onChange={(e) => setBankCode(e.target.value)}
                        className={`w-full appearance-none rounded-2xl border-2 bg-white px-4 py-3.5 text-[15px] font-bold outline-none transition-colors focus:border-violet ${
                          errors.bank ? "border-coral" : "border-ink/10"
                        }`}
                      >
                        <option value="">Select your bank…</option>
                        {banks.map((b) => (
                          <option key={b.code} value={b.code}>
                            {b.name}
                          </option>
                        ))}
                      </select>
                      <ChevronDown className="pointer-events-none absolute right-4 top-1/2 size-5 -translate-y-1/2 text-soft" />
                    </div>
                    {errors.bank && <p className="mt-1.5 text-xs font-bold text-coral">{errors.bank}</p>}
                  </div>

                  <Field
                    name="accountNumber"
                    label="Account number"
                    icon={<Landmark className="size-5" />}
                    type="text"
                    inputMode="numeric"
                    value={accountNumber}
                    onChange={(e) => setAccountNumber(e.target.value.replace(/[^0-9]/g, "").slice(0, 10))}
                    placeholder="0123456789"
                    error={errors.accountNumber}
                    autoComplete="off"
                  />

                  {payoutError && (
                    <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                      {payoutError}
                    </p>
                  )}

                  <Button
                    type="submit"
                    variant="mint"
                    size="lg"
                    loading={submitPayout.isPending}
                    icon={submitPayout.isPending ? <Loader2 className="size-5 animate-spin" /> : <Check className="size-5" />}
                    className="w-full"
                  >
                    {submitPayout.isPending ? "Transferring…" : `Pay me ${"· "} ${claim?.amountNaira ?? ""}`}
                  </Button>
                  <p className="text-center text-xs leading-relaxed text-soft">
                    The host wallet sends the exact prize amount to this account via BMONI. Nothing is
                    deducted on your side.
                  </p>
                </form>
              </>
            )}
          </section>
        </div>
      </div>

      <Footer variant="player" />
    </main>
  );
}

export default ClaimPage;
