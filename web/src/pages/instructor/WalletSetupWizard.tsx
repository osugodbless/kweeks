import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import {
  Banknote,
  Check,
  ChevronRight,
  FileCheck,
  Landmark,
  Loader2,
  Lock,
  ScanLine,
  ShieldCheck,
  Sparkles,
  Upload,
  UserCheck,
  Wallet as WalletIcon,
  Zap,
} from "lucide-react";
import { ApiError, type WalletSetupStage } from "@/lib/api";
import { useActivateRail, useCreateBmoniUser, useCreateWallet, useLookupBVN, useSubmitKYC, useUploadKYC, useWalletSetup } from "@/lib/hooks";
import { Modal } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { cn } from "@/lib/cn";

const NIGERIAN_STATES = [
  "Abia", "Adamawa", "Akwa Ibom", "Anambra", "Bauchi", "Bayelsa", "Benue", "Borno",
  "Cross River", "Delta", "Ebonyi", "Edo", "Ekiti", "Enugu", "FCT - Abuja", "Gombe",
  "Imo", "Jigawa", "Kaduna", "Kano", "Katsina", "Kebbi", "Kogi", "Kwara", "Lagos",
  "Nasarawa", "Niger", "Ogun", "Ondo", "Osun", "Oyo", "Plateau", "Rivers", "Sokoto",
  "Taraba", "Yobe", "Zamfara",
];

// Strict BMONI lifecycle: create the smart wallet (stage 2) BEFORE verifying
// identity (stage 3), then activate the rail (stage 4).
const STEPS = [
  { id: 1, label: "Create wallet", Icon: WalletIcon },
  { id: 2, label: "Verify identity", Icon: UserCheck },
  { id: 3, label: "Activate NGN", Icon: Landmark },
] as const;

const STAGE_TO_STEP: Record<WalletSetupStage, number> = {
  unprovisioned: 1,
  wallet: 1,
  kyc: 2,
  rail: 3,
  ready: 4,
};

const BVN_RE = /^\d{11}$/;
const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;

export function WalletSetupWizard({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { data: setup, refetch } = useWalletSetup();
  const submitKYC = useSubmitKYC();
  const uploadKYC = useUploadKYC();
  const lookupBVN = useLookupBVN();
  const createUser = useCreateBmoniUser();
  const createWallet = useCreateWallet();
  const activateRail = useActivateRail();

  const [step, setStep] = useState<number | null>(null);
  const [bvn, setBvn] = useState("");

  // KYC form state
  const [kyc, setKyc] = useState({
    firstName: "", lastName: "", dateOfBirth: "", gender: "male",
    bvn: "", street: "", city: "", state: "Lagos", postalCode: "101241",
  });
  // The BVN typed into the look-up box; the holder record resolves from it.
  const [lookupInput, setLookupInput] = useState("");
  const [bvnResolved, setBvnResolved] = useState(false);
  const [idFile, setIdFile] = useState<File | null>(null);
  const [poaFile, setPoaFile] = useState<File | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [error, setError] = useState("");

  // Resume from the live setup stage when no local step is set this session.
  const activeStep = step ?? (setup ? (STAGE_TO_STEP[setup.stage] ?? 1) : 1);

  const stage = setup?.stage ?? "unprovisioned";
  const done = activeStep >= 4 || stage === "ready";

  const needsUser = Boolean(setup && !setup.bmoniUserId);
  const railOn = setup?.railConfigured === true;
  const onKycStep = activeStep === 2 && railOn && !needsUser;

  // Step 1 auto-creates the wallet (stage 2 of the lifecycle) when visible.
  useEffect(() => {
    if (open && activeStep === 1 && !needsUser && !createWallet.isPending && !createWallet.isSuccess) {
      createWallet.mutateAsync().catch(() => {});
    }
  }, [open, activeStep, needsUser, createWallet]);

  const submitting = lookupBVN.isPending || submitKYC.isPending || uploadKYC.isPending || createWallet.isPending || activateRail.isPending;

  async function handleLookup(event: FormEvent) {
    event.preventDefault();
    setError("");
    if (!BVN_RE.test(lookupInput)) {
      setError("BVN is exactly 11 digits.");
      return;
    }
    try {
      const rec = await lookupBVN.mutateAsync(lookupInput);
      // Auto-fill the identity fields from the BVN holder record for the host
      // to edit + confirm. Address stays for the host to complete.
      setKyc((k) => ({
        ...k,
        bvn: lookupInput,
        firstName: rec.firstName || k.firstName,
        lastName: rec.lastName || k.lastName,
        dateOfBirth: rec.dateOfBirth || k.dateOfBirth,
        gender: rec.gender === "female" ? "female" : rec.gender === "male" ? "male" : k.gender,
        street: rec.residentialAddress || k.street,
        state: rec.stateOfResidence && NIGERIAN_STATES.includes(rec.stateOfResidence) ? rec.stateOfResidence : k.state,
      }));
      setBvnResolved(true);
    } catch (err) {
      setBvnResolved(false);
      setError(err instanceof ApiError ? err.message : "We could not resolve that BVN — check it and try again.");
    }
  }

  function validateKYC() {
    const next: Record<string, string> = {};
    if (!kyc.firstName.trim()) next.firstName = "Enter your legal first name.";
    if (!kyc.lastName.trim()) next.lastName = "Enter your legal last name.";
    if (!DATE_RE.test(kyc.dateOfBirth)) next.dateOfBirth = "Use YYYY-MM-DD.";
    if (!BVN_RE.test(kyc.bvn)) next.bvn = "BVN is exactly 11 digits.";
    if (!kyc.street.trim()) next.street = "Enter your street address.";
    if (!kyc.city.trim()) next.city = "Enter your city.";
    if (!kyc.state) next.state = "Pick your state.";
    if (!/^\d{6}$/.test(kyc.postalCode)) next.postalCode = "6-digit postal code.";
    if (!idFile) next.idFile = "National ID / passport is required.";
    if (!poaFile) next.poaFile = "Proof of address is required.";
    return next;
  }

  async function handleKYC(event: FormEvent) {
    event.preventDefault();
    setError("");
    const next = validateKYC();
    setErrors(next);
    if (Object.keys(next).length > 0) return;

    try {
      // Strict KYC order: documents are uploaded before the profile is
      // submitted (PATCH /kyc), and both are required for NGN.
      await uploadKYC.mutateAsync({ kind: "identification", file: idFile! });
      await uploadKYC.mutateAsync({ kind: "proof-of-address", file: poaFile! });
      await submitKYC.mutateAsync(kyc);
      setBvn(kyc.bvn);
      setStep(3);
      void refetch();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not submit your identity — check the details and try again.");
    }
  }

  async function handleActivate(event: FormEvent) {
    event.preventDefault();
    setError("");
    const finalBvn = bvn || kyc.bvn;
    if (!BVN_RE.test(finalBvn)) {
      setError("Enter the 11-digit BVN used for verification.");
      return;
    }
    try {
      await activateRail.mutateAsync(finalBvn);
      setStep(4);
      void refetch();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not activate the NGN rail — try again.");
    }
  }

  async function handleCreateUser() {
    setError("");
    try {
      await createUser.mutateAsync();
      setStep(1);
      void refetch();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not create your BMONI user — try again.");
    }
  }

  function resetAndClose() {
    setStep(null);
    onClose();
  }

  const deposit = setup?.depositAccount;

  return (
    <Modal open={open} onClose={resetAndClose} hideClose={done}>
      {/* Stepper */}
      <div className="flex items-center gap-2">
        {STEPS.map((s, i) => (
          <div key={s.id} className="flex flex-1 items-center gap-2">
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  "grid size-8 shrink-0 place-items-center rounded-full text-sm font-bold transition-colors",
                  activeStep > s.id || done
                    ? "bg-mint text-white"
                    : activeStep === s.id
                      ? "bg-violet text-white"
                      : "bg-ink/5 text-soft",
                )}
              >
                {activeStep > s.id || done ? <Check className="size-4" strokeWidth={3} /> : s.id}
              </span>
              <span
                className={cn(
                  "hidden text-xs font-bold sm:block",
                  activeStep === s.id ? "text-ink" : "text-soft",
                )}
              >
                {s.label}
              </span>
            </div>
            {i < STEPS.length - 1 && <div className="h-0.5 flex-1 rounded bg-ink/10" />}
          </div>
        ))}
      </div>

      <div className="mt-6">
        {/* Step 1: Create wallet (lifecycle stage 2, precedes KYC) */}
        {!done && activeStep === 1 && railOn && !needsUser && (
          <div className="space-y-5">
            <div>
              <p className="text-xs font-bold uppercase tracking-widest text-soft">Step 1 · Create your wallet</p>
              <p className="mt-1 text-sm leading-relaxed text-soft">
                Your smart wallet is provisioned on-chain before identity verification. This takes a
                few seconds — the owner proof is signed server-side with your host key.
              </p>
            </div>

            <div className="flex items-center gap-3 rounded-2xl border-2 border-violet/20 bg-violet/5 px-5 py-4">
              {createWallet.isPending ? (
                <>
                  <Loader2 className="size-5 animate-spin text-violet" />
                  <p className="text-sm font-bold text-violet">Creating your smart wallet…</p>
                </>
              ) : createWallet.isError ? (
                <p className="text-sm font-bold text-coral">
                  {(createWallet.error as Error)?.message ?? "Wallet creation failed — try again."}
                </p>
              ) : (
                <>
                  <span className="grid size-9 place-items-center rounded-xl bg-violet text-white">
                    <WalletIcon className="size-4" />
                  </span>
                  <div className="min-w-0">
                    <p className="text-sm font-bold">Wallet ready</p>
                    <p className="truncate font-mono text-xs text-soft">{setup?.bmoniWalletAddress ?? "…"}</p>
                  </div>
                </>
              )}
            </div>

            {error && (
              <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                {error}
              </p>
            )}

            <Button
              variant="mint"
              size="lg"
              disabled={!createWallet.isSuccess}
              onClick={() => setStep(2)}
              icon={<ChevronRight className="size-5" />}
              className="w-full"
            >
              Continue to identity
            </Button>
          </div>
        )}

        {/* Step 2: Verify identity (lifecycle stage 3) */}
        {!done && onKycStep && (
          <div className="space-y-4">
            <div>
              <p className="text-xs font-bold uppercase tracking-widest text-soft">Step 2 · Verify your identity</p>
              <p className="mt-1 text-sm leading-relaxed text-soft">
                Start with your BVN. We look it up and pre-fill your details for you to confirm,
                then you upload your documents.
              </p>
            </div>

            {!bvnResolved ? (
              <form onSubmit={handleLookup} noValidate className="space-y-4">
                <div className="rounded-2xl border-2 border-violet/20 bg-violet/5 p-4">
                  <Field
                    name="bvnLookup" label="Bank Verification No. (BVN)" value={lookupInput}
                    onChange={(e) => setLookupInput(e.target.value.replace(/[^0-9]/g, "").slice(0, 11))}
                    placeholder="11 digits" inputMode="numeric" autoComplete="off"
                    icon={<ScanLine className="size-5" />}
                  />
                  {error && (
                    <p role="alert" className="mt-2 rounded-xl border-2 border-coral/30 bg-coral/10 px-4 py-2.5 text-sm font-bold text-coral">
                      {error}
                    </p>
                  )}
                  <Button type="submit" variant="violet" size="lg" loading={lookupBVN.isPending} icon={<ScanLine className="size-5" />} className="mt-4 w-full">
                    {lookupBVN.isPending ? "Looking up…" : "Look up my BVN"}
                  </Button>
                  <p className="mt-2 text-center text-xs text-soft">
                    Your BVN is only used to fetch your records for verification — it is never shown to players.
                  </p>
                </div>
              </form>
            ) : (
              <form onSubmit={handleKYC} noValidate className="space-y-4">
                <div className="flex items-start gap-2.5 rounded-2xl border-2 border-mint/30 bg-mint/10 px-4 py-3">
                  <Check className="mt-0.5 size-4 shrink-0 text-mint-dark" strokeWidth={3} />
                  <p className="text-sm font-bold text-mint-dark">
                    Records found — confirm your details below, complete your address, then upload your documents.
                  </p>
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <Field
                    name="firstName" label="Legal first name" value={kyc.firstName}
                    onChange={(e) => setKyc({ ...kyc, firstName: e.target.value })}
                    error={errors.firstName} autoComplete="given-name" spellCheck="false"
                  />
                  <Field
                    name="lastName" label="Legal last name" value={kyc.lastName}
                    onChange={(e) => setKyc({ ...kyc, lastName: e.target.value })}
                    error={errors.lastName} autoComplete="family-name" spellCheck="false"
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-3">
                  <Field
                    name="dateOfBirth" label="Date of birth" value={kyc.dateOfBirth}
                    onChange={(e) => setKyc({ ...kyc, dateOfBirth: e.target.value })}
                    placeholder="YYYY-MM-DD" error={errors.dateOfBirth} inputMode="numeric" autoComplete="bday"
                  />
                  <div>
                    <label htmlFor="gender" className="mb-2 block text-sm font-bold">Gender</label>
                    <select
                      id="gender" value={kyc.gender}
                      onChange={(e) => setKyc({ ...kyc, gender: e.target.value })}
                      className="w-full appearance-none rounded-2xl border-2 border-ink/10 bg-white px-4 py-4 font-bold text-ink outline-none transition-all focus:border-violet"
                    >
                      <option value="male">Male</option>
                      <option value="female">Female</option>
                    </select>
                  </div>
                  <Field
                    name="bvn" label="Bank Verification No." value={kyc.bvn}
                    onChange={(e) => setKyc({ ...kyc, bvn: e.target.value.replace(/[^0-9]/g, "").slice(0, 11) })}
                    error={errors.bvn} inputMode="numeric" autoComplete="off"
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <Field
                    name="street" label="Street address" value={kyc.street}
                    onChange={(e) => setKyc({ ...kyc, street: e.target.value })}
                    placeholder="15 Admiralty Way" error={errors.street} autoComplete="street-address"
                  />
                  <Field
                    name="city" label="City" value={kyc.city}
                    onChange={(e) => setKyc({ ...kyc, city: e.target.value })}
                    placeholder="Lagos" error={errors.city} autoComplete="address-level2"
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div>
                    <label htmlFor="state" className="mb-2 block text-sm font-bold">State</label>
                    <select
                      id="state" value={kyc.state}
                      onChange={(e) => setKyc({ ...kyc, state: e.target.value })}
                      className={cn(
                        "w-full appearance-none rounded-2xl border-2 bg-white px-4 py-4 font-bold text-ink outline-none transition-all focus:border-violet",
                        errors.state ? "border-coral" : "border-ink/10",
                      )}
                    >
                      {NIGERIAN_STATES.map((s) => (
                        <option key={s} value={s}>{s}</option>
                      ))}
                    </select>
                    {errors.state && <p className="mt-1.5 text-sm font-bold text-coral">{errors.state}</p>}
                  </div>
                  <Field
                    name="postalCode" label="Postal code" value={kyc.postalCode}
                    onChange={(e) => setKyc({ ...kyc, postalCode: e.target.value.replace(/[^0-9]/g, "").slice(0, 6) })}
                    placeholder="101241" error={errors.postalCode} inputMode="numeric" autoComplete="postal-code"
                  />
                </div>

                {/* Required documents, uploaded before PATCH /kyc per the strict flow. */}
                <div className="rounded-2xl border-2 border-ink/10 bg-white p-4">
                  <p className="flex items-center gap-2 text-sm font-bold">
                    <FileCheck className="size-4 text-violet" /> Documents <span className="text-coral">· required</span>
                  </p>
                  <p className="mt-0.5 text-xs text-soft">
                    Identification and proof of address are uploaded before your profile is submitted.
                  </p>
                  <div className="mt-3 grid gap-3 sm:grid-cols-2">
                    <FileInput required label="National ID / passport" file={idFile} onChange={setIdFile} error={errors.idFile} />
                    <FileInput required label="Proof of address" file={poaFile} onChange={setPoaFile} error={errors.poaFile} />
                  </div>
                </div>

                <button type="button" onClick={() => setBvnResolved(false)} className="w-full cursor-pointer text-center text-sm font-bold text-violet hover:text-violet-dark">
                  ← Use a different BVN
                </button>

                {error && (
                  <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                    {error}
                  </p>
                )}

                <Button type="submit" variant="violet" size="lg" loading={submitting} icon={<ShieldCheck className="size-5" />} className="w-full">
                  {submitting ? "Submitting…" : "Submit & continue"}
                </Button>
              </form>
            )}
          </div>
        )}

        {/* Step 3: Activate rail (lifecycle stage 4) */}
        {!done && activeStep === 3 && railOn && !needsUser && (
          <form onSubmit={handleActivate} noValidate className="space-y-4">
            <div>
              <p className="text-xs font-bold uppercase tracking-widest text-soft">Step 3 · Activate NGN rail</p>
              <p className="mt-1 text-sm leading-relaxed text-soft">
                Final step. Your BVN is verified against your profile and the naira rail is linked
                to your wallet, giving you a Nigerian bank account you can fund by transfer.
              </p>
            </div>

            <Field
              name="bvn" label="Bank Verification No." value={bvn || kyc.bvn}
              onChange={(e) => setBvn(e.target.value.replace(/[^0-9]/g, "").slice(0, 11))}
              placeholder="11 digits" inputMode="numeric" autoComplete="off"
              icon={<Lock className="size-5" />}
              hint="The BVN you verified with in step 2."
            />

            <div className="flex items-start gap-3 rounded-2xl border-2 border-gold/30 bg-gold/10 p-4">
              <Zap className="mt-0.5 size-5 shrink-0 text-gold-dark" />
              <p className="text-sm leading-relaxed text-ink">
                <span className="font-bold">After activation</span> you get a Nigerian virtual
                account ({deposit?.bankName ?? "your bank"}) you can transfer into — no code needed.
              </p>
            </div>

            {error && (
              <p role="alert" className="rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                {error}
              </p>
            )}

            <Button type="submit" variant="mint" size="lg" loading={activateRail.isPending} icon={<Landmark className="size-5" />} className="w-full">
              {activateRail.isPending ? "Activating…" : "Activate NGN rail"}
            </Button>
          </form>
        )}

        {/* Done */}
        {done && (
          <div className="space-y-5 text-center">
            <span className="mx-auto grid size-16 animate-float place-items-center rounded-full bg-mint text-white">
              <Sparkles className="size-8" />
            </span>
            <div>
              <h3 className="font-display text-2xl font-semibold tracking-tight">Your wallet is ready</h3>
              <p className="mx-auto mt-1 max-w-sm text-sm leading-relaxed text-soft">
                The NGN rail is active. Fund your wallet by transferring to your Nigerian virtual
                account below, or use instant wallet credit from the fund page.
              </p>
            </div>

            {deposit?.accountNumber && (
              <div className="rounded-2xl border-2 border-mint/30 bg-mint/10 p-5">
                <p className="text-xs font-bold uppercase tracking-widest text-mint-dark">Your NGN account</p>
                <p className="mt-2 rounded-xl bg-white px-4 py-3 font-display text-2xl font-semibold tracking-wide">
                  {deposit.accountNumber}
                </p>
                <p className="mt-1 text-sm font-bold text-mint-dark">{deposit.bankName} · NGN</p>
              </div>
            )}

            {!deposit?.accountNumber && (
              <div className="flex items-center gap-3 rounded-2xl border-2 border-gold/30 bg-gold/10 p-4">
                <Banknote className="size-5 text-gold-dark" />
                <p className="text-sm font-bold text-ink">Wallet active — you can fund it now.</p>
              </div>
            )}

            <Button variant="ink" size="lg" onClick={resetAndClose} className="w-full">
              Done
            </Button>
          </div>
        )}

        {/* No BMONI user yet but the rail is reachable: create-user failed at
            signup, so offer the idempotent retry instead of a dead-end form. */}
        {needsUser && railOn && (
          <div className="space-y-4 text-center">
            <span className="mx-auto grid size-14 place-items-center rounded-full bg-violet/10 text-violet">
              <UserCheck className="size-6" />
            </span>
            <p className="text-sm leading-relaxed text-soft">
              Your BMONI user was not created at signup. Fix it now so you can create your wallet,
              verify your identity and activate naira.
            </p>
            <Button
              variant="violet" size="lg" loading={createUser.isPending}
              icon={<UserCheck className="size-5" />}
              onClick={handleCreateUser} className="w-full"
            >
              {createUser.isPending ? "Creating…" : "Create my BMONI user"}
            </Button>
          </div>
        )}

        {/* Rail is not configured: no provisioning is possible from here. */}
        {!railOn && (
          <div className="space-y-4 text-center">
            <span className="mx-auto grid size-14 place-items-center rounded-full bg-coral/10 text-coral">
              <Lock className="size-6" />
            </span>
            <p className="text-sm leading-relaxed text-soft">
              The money rail is not reachable right now. Your account still works with wallet
              credit — retry provisioning from here once the rail is configured.
            </p>
            <Button variant="violet" size="lg" loading={createUser.isPending} onClick={() => { setStep(1); }} className="w-full">
              Retry setup
            </Button>
          </div>
        )}
      </div>
    </Modal>
  );
}

function FileInput({
  label, file, onChange, error, required,
}: {
  label: string;
  file: File | null;
  onChange: (f: File | null) => void;
  error?: string;
  required?: boolean;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  return (
    <div>
      <button
        type="button"
        onClick={() => inputRef.current?.click()}
        className={cn(
          "flex w-full cursor-pointer items-center gap-3 rounded-2xl border-2 bg-white px-4 py-3 text-left transition-all hover:border-violet",
          error ? "border-coral" : file ? "border-mint" : "border-ink/10",
        )}
      >
        <span className={cn("grid size-10 shrink-0 place-items-center rounded-xl", file ? "bg-mint/10 text-mint-dark" : "bg-violet/10 text-violet")}>
          {file ? <FileCheck className="size-5" /> : <Upload className="size-5" />}
        </span>
        <span className="min-w-0 flex-1">
          <span className="block truncate text-sm font-bold">
            {file ? file.name : label}
            {required && !file && <span className="ml-1.5 text-xs font-bold uppercase text-coral">Required</span>}
          </span>
          <span className="block text-xs text-soft">{file ? "Ready to submit" : "JPEG or PNG"}</span>
        </span>
        <input
          ref={inputRef} type="file" accept="image/jpeg,image/png" className="hidden"
          onChange={(e) => { onChange(e.target.files?.[0] ?? null); }}
        />
      </button>
      {error && <p className="mt-1.5 text-sm font-bold text-coral">{error}</p>}
    </div>
  );
}

export default WalletSetupWizard;
