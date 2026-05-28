import { forwardRef, useId } from "react";
import type { SelectHTMLAttributes } from "react";
import { ChevronDownIcon } from "@/components/icons";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, error, className, children, id: externalId, ...props }, ref) => {
    const generatedId = useId();
    const id = externalId || generatedId;
    const errorId = `${id}-error`;

    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label htmlFor={id} className="text-sm font-medium text-[var(--text-secondary)]">
            {label}
          </label>
        )}
        <div className="relative">
          <select
            ref={ref}
            id={id}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
            className={[
              "w-full appearance-none rounded-[var(--radius-md)] border bg-[var(--bg-surface)]",
              "px-3 py-2 pr-9 text-sm text-[var(--text-primary)]",
              "transition-colors focus:outline-none focus:ring-2 focus:ring-offset-1",
              "focus:ring-offset-[var(--bg-base)]",
              error
                ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
                : "border-[var(--border-default)] focus:ring-[var(--accent-default)]",
              "disabled:opacity-50 disabled:cursor-not-allowed",
              className,
            ]
              .filter(Boolean)
              .join(" ")}
            {...props}
          >
            {children}
          </select>
          <ChevronDownIcon
            className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2
              text-[var(--text-tertiary)]"
            width={16}
            height={16}
          />
        </div>
        {error && (
          <p id={errorId} className="text-xs text-[var(--color-error)]" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  },
);

Select.displayName = "Select";
