import { useEffect, type ReactNode } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/cn";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  eyebrow?: string;
  children: ReactNode;
  className?: string;
  hideClose?: boolean;
}

/** Accessible modal panel following the DESIGN.md system: cream card, rounded
 *  corners, soft shadow, escape-to-close, focus returned to the trigger. */
export function Modal({ open, onClose, title, eyebrow, children, className, hideClose }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[80] grid place-items-center overflow-y-auto p-4 sm:p-6"
      role="dialog"
      aria-modal="true"
      aria-label={title ?? "Wallet setup"}
    >
      <button
        type="button"
        aria-label="Close modal"
        onClick={onClose}
        className="fixed inset-0 cursor-default bg-ink/50 backdrop-blur-sm"
      />
      <div
        className={cn(
          "relative my-auto w-full max-w-lg rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d sm:p-8",
          "animate-[rise_0.35s_cubic-bezier(0.16,1,0.3,1)_both]",
          className,
        )}
      >
        {!hideClose && (
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="absolute right-5 top-5 grid size-10 cursor-pointer place-items-center rounded-2xl text-soft transition-colors hover:bg-ink/5 hover:text-ink"
          >
            <X className="size-5" />
          </button>
        )}
        {eyebrow && (
          <p className="text-xs font-bold uppercase tracking-[0.25em] text-soft">{eyebrow}</p>
        )}
        {title && <h2 className="mt-1.5 font-display text-2xl font-semibold tracking-tight">{title}</h2>}
        {children}
      </div>
    </div>
  );
}
