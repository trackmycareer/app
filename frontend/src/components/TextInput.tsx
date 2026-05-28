import { forwardRef, useId } from "react";
import type { InputHTMLAttributes } from "react";

interface TextInputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(
  ({ label, error, className, id: externalId, ...props }, ref) => {
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
        <input
          ref={ref}
          id={id}
          aria-invalid={!!error}
          aria-describedby={error ? errorId : undefined}
          className={[
            "w-full rounded-[var(--radius-md)] border bg-[var(--bg-surface)] px-3 py-2",
            "text-sm text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]",
            "transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-1",
            "focus:ring-offset-[var(--bg-base)]",
            "focus:shadow-[0_0_0_3px] focus:shadow-[var(--accent-default)]/15",
            error
              ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
              : "border-[var(--border-default)] focus:ring-[var(--accent-default)]",
            "disabled:opacity-50 disabled:cursor-not-allowed",
            className,
          ]
            .filter(Boolean)
            .join(" ")}
          {...props}
        />
        {error && (
          <p id={errorId} className="text-xs text-[var(--color-error)]" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  },
);

TextInput.displayName = "TextInput";
