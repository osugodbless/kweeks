import type { ReactNode } from "react";
import { AlertTriangle, LogOut } from "lucide-react";
import { Modal } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  body?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Styles the confirm action as destructive (coral). */
  danger?: boolean;
  /** Uses the sign-out icon instead of the warning triangle. */
  signOut?: boolean;
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

/** Confirmation modal for destructive operations (logout, deletes). Nothing
 *  happens until the host confirms — a cancel or backdrop click aborts. */
export function ConfirmDialog({
  open,
  title,
  body,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  danger,
  signOut,
  loading,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal open={open} onClose={onCancel}>
      <div className="flex items-start gap-4">
        <span
          className={
            "grid size-11 shrink-0 place-items-center rounded-2xl " +
            (danger ? "bg-coral/10 text-coral" : signOut ? "bg-violet/10 text-violet" : "bg-gold/10 text-gold-dark")
          }
        >
          {signOut ? <LogOut className="size-5" /> : <AlertTriangle className="size-5" />}
        </span>
        <div className="min-w-0 flex-1">
          <h3 className="font-display text-xl font-semibold tracking-tight">{title}</h3>
          {body && <p className="mt-1.5 text-sm leading-relaxed text-soft">{body}</p>}
        </div>
      </div>

      <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <Button variant="ghost" size="md" onClick={onCancel} disabled={loading} className="sm:min-w-28">
          {cancelLabel}
        </Button>
        <Button
          variant={danger ? "coral" : "ink"}
          size="md"
          loading={loading}
          onClick={onConfirm}
          className="sm:min-w-28"
        >
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  );
}

export default ConfirmDialog;
