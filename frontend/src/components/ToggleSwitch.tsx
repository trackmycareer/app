export function ToggleSwitch({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (val: boolean) => void;
}) {
  return (
    <label className="flex cursor-pointer items-center justify-between py-2">
      <span className="text-sm text-[var(--text-secondary)]">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={[
          "relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent",
          "transition-colors focus-visible:outline-none focus-visible:ring-2",
          "focus-visible:ring-[var(--accent-default)] focus-visible:ring-offset-2",
          "focus-visible:ring-offset-[var(--bg-base)]",
          checked ? "bg-[var(--accent-default)]" : "bg-[var(--bg-active)]",
        ].join(" ")}
      >
        <span
          className={[
            "pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm",
            "transition-transform",
            checked ? "translate-x-5" : "translate-x-0",
          ].join(" ")}
        />
      </button>
    </label>
  );
}
