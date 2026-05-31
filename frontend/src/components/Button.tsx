import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { SpinnerIcon } from "@/components/icons";

type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";
type ButtonSize = "sm" | "md" | "lg";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  icon?: ReactNode;
  children: ReactNode;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "bg-[var(--button-primary-bg)] text-[var(--button-primary-text)] hover:bg-[var(--accent-bright)] " +
    "hover:shadow-md hover:shadow-[var(--accent-default)]/20 " +
    "focus-visible:ring-[var(--accent-default)]",
  secondary:
    "bg-[var(--bg-elevated)] text-[var(--text-primary)] border border-[var(--border-default)] " +
    "hover:bg-[var(--bg-hover)] focus-visible:ring-[var(--border-strong)]",
  danger:
    "bg-[var(--color-error)] text-white hover:brightness-110 " +
    "focus-visible:ring-[var(--color-error)]",
  ghost:
    "bg-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)] " +
    "hover:bg-[var(--bg-hover)] focus-visible:ring-[var(--border-default)]",
};

const sizeStyles: Record<ButtonSize, string> = {
  sm: "px-3 py-1.5 text-sm gap-1.5",
  md: "px-4 py-2 text-sm gap-2",
  lg: "px-5 py-2.5 text-base gap-2",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", loading = false, icon, children, className, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={[
          "inline-flex items-center justify-center font-medium transition-all duration-150",
          `rounded-[var(--radius-md)]`,
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
          "focus-visible:ring-offset-[var(--bg-base)]",
          "disabled:opacity-60 disabled:cursor-not-allowed",
          "active:scale-[0.98]",
          variantStyles[variant],
          sizeStyles[size],
          className,
        ]
          .filter(Boolean)
          .join(" ")}
        disabled={loading || props.disabled}
        {...props}
      >
        {loading ? <SpinnerIcon width={16} height={16} /> : icon}
        {children}
      </button>
    );
  },
);

Button.displayName = "Button";
