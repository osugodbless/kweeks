import { cn } from "@/lib/cn";

export function Field({
  label,
  placeholder,
  value,
  onChange,
  type = "text",
  monoValue,
  className,
}: {
  label: string;
  placeholder?: string;
  value?: string;
  onChange?: (v: string) => void;
  type?: string;
  monoValue?: string;
  className?: string;
}) {
  const showValue = value !== undefined;
  return (
    <label className={cn("block", className)}>
      <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
        {label}
      </span>
      <div className="flex h-[50px] w-full items-center rounded-[14px] border border-stroke bg-surface px-4">
        <input
          type={type}
          value={showValue ? value : undefined}
          placeholder={placeholder}
          onChange={(e) => onChange?.(e.target.value)}
          className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
        />
        {monoValue && (
          <span className="font-display text-[26px] font-extrabold text-paper">{monoValue}</span>
        )}
      </div>
    </label>
  );
}

export function MonoField({
  label,
  prefix,
  value,
  className,
}: {
  label: string;
  prefix?: string;
  value: string;
  className?: string;
}) {
  return (
    <label className={cn("block", className)}>
      <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
        {label}
      </span>
      <div className="flex h-[58px] w-full items-center gap-2 rounded-[14px] border border-stroke bg-surface px-[18px]">
        {prefix && (
          <span className="font-display text-[24px] font-extrabold text-naira">{prefix}</span>
        )}
        <span className="font-display text-[26px] font-extrabold text-paper">{value}</span>
      </div>
    </label>
  );
}
