import { FileQuestion, History, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";

export function HistoryEmpty({ onCreate }: { onCreate?: () => void }) {
  return (
    <div className="mx-auto max-w-md text-center">
      <span className="mx-auto grid size-16 place-items-center rounded-[1.5rem] bg-cream text-soft card-3d">
        <History className="size-8" strokeWidth={2} />
      </span>
      <h2 className="mt-6 font-display text-3xl font-semibold tracking-tight">No history yet</h2>
      <p className="mt-2 text-[15px] leading-relaxed text-soft">
        Every quiz you host, pool you fund and winner you pay will show up right here.
      </p>
      <Button variant="coral" size="lg" className="mt-6" onClick={onCreate}>
        <Plus className="size-5" /> Create a quiz
      </Button>
    </div>
  );
}

export function InstructorHistoryEmpty() {
  return (
    <main className="relative flex min-h-screen flex-col overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-violet/20 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-gold/15 blur-3xl" />

      <InstructorNav active="/instructor/history" />

      <div className="relative mx-auto flex w-full max-w-6xl flex-1 flex-col px-4 pt-16 sm:px-6">
        <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
          <FileQuestion className="size-4 text-violet" /> History
        </span>
        <h1 className="mt-4 font-display text-4xl font-semibold tracking-tight sm:text-5xl">Your activity</h1>
        <div className="mt-10 flex-1">
          <HistoryEmpty onCreate={() => undefined} />
        </div>
      </div>

      <Footer className="mt-16" />
    </main>
  );
}

export default InstructorHistoryEmpty;
