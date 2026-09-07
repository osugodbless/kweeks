import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "@/lib/cn";

export interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: ReactNode;
  error?: string;
  icon?: ReactNode;
  right?: ReactNode;
  /** Applies the error styling without needing a label wrapper. */
  invalid?: boolean;
}

export const Field = forwardRef<HTMLInputElement, FieldProps>(function Field(
  { className, label, hint, error, icon, right, invalid, id, ...props },
  ref,
) {
  const fieldId = id ?? props.name;
  const hasError = invalid || Boolean(error);

  return (
    <div className="w-full">
      {label && (
        <label htmlFor={fieldId} className="mb-2 block text-sm font-bold">
          {label}
        </label>
      )}
      <div className="relative">
        {icon && (
          <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-soft">
            {icon}
          </span>
        )}
        <input
          ref={ref}
          id={fieldId}
          className={cn(
            "w-full rounded-2xl border-2 bg-white py-4 font-bold text-ink outline-none transition-all",
            "placeholder:font-normal placeholder:text-soft/60",
            icon && "pl-12",
            right && "pr-12",
            hasError
              ? "border-coral focus:border-coral"
              : "border-ink/10 hover:border-ink/20 focus:border-violet focus:shadow-[0_0_0_4px_rgba(108,76,241,0.15)]",
            className,
          )}
          {...props}
        />
        {right && <span className="absolute right-2 top-1/2 -translate-y-1/2">{right}</span>}
      </div>
      {hint && !error && <div className="mt-1.5 text-xs text-soft">{hint}</div>}
      {error && <div className="mt-1.5 text-sm font-bold text-coral">{error}</div>}
    </div>
  );
});
