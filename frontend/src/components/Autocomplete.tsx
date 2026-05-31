import { useState, useEffect, useRef, useCallback, useId } from "react";
import type { KeyboardEvent } from "react";

export interface AutocompleteItem {
  label: string;
  value: string;
  badge?: { text: string; variant: "success" | "neutral" };
  metadata?: Record<string, string>;
}

export interface AutocompleteGroup {
  label: string;
  items: AutocompleteItem[];
}

export interface AutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onSelect?: (item: AutocompleteItem) => void;
  error?: string;
  required?: boolean;
  groups: AutocompleteGroup[];
  isLoading: boolean;
  hasQueried: boolean;
  ariaLabel: string;
  loadingMessage?: string;
  emptyMessage?: string;
}

export function Autocomplete({
  label,
  placeholder,
  value,
  onChange,
  onSelect,
  error,
  required,
  groups,
  isLoading,
  hasQueried,
  ariaLabel,
  loadingMessage = "Searching...",
  emptyMessage = "No results found",
}: AutocompleteProps) {
  const generatedId = useId();
  const inputId = `autocomplete-input-${generatedId}`;
  const listboxId = `autocomplete-listbox-${generatedId}`;
  const errorId = `${inputId}-error`;

  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);

  // Flatten all items for keyboard navigation
  const allItems = groups.flatMap((g) => g.items);
  const hasResults = allItems.length > 0;

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleMouseDown(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleMouseDown);
    return () => document.removeEventListener("mousedown", handleMouseDown);
  }, []);

  // Reset highlight when groups change
  useEffect(() => {
    setHighlightedIndex(-1);
  }, [groups]);

  // Scroll highlighted option into view for keyboard users
  useEffect(() => {
    if (highlightedIndex < 0) return;
    const el = document.getElementById(`autocomplete-option-${generatedId}-${highlightedIndex}`);
    el?.scrollIntoView({ block: "nearest" });
  }, [highlightedIndex, generatedId]);

  const selectItem = useCallback(
    (item: AutocompleteItem) => {
      onChange(item.value);
      onSelect?.(item);
      setIsOpen(false);
      setHighlightedIndex(-1);
      inputRef.current?.focus();
    },
    [onChange, onSelect],
  );

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange(e.target.value);
      setIsOpen(true);
    },
    [onChange],
  );

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      if (!isOpen || !hasResults) {
        if (e.key === "ArrowDown" && hasQueried && hasResults) {
          e.preventDefault();
          setIsOpen(true);
          setHighlightedIndex(0);
        }
        return;
      }

      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setHighlightedIndex((prev) => (prev + 1) % allItems.length);
          break;
        case "ArrowUp":
          e.preventDefault();
          setHighlightedIndex((prev) => (prev - 1 + allItems.length) % allItems.length);
          break;
        case "Enter":
          if (highlightedIndex >= 0 && highlightedIndex < allItems.length) {
            e.preventDefault();
            selectItem(allItems[highlightedIndex]);
          }
          break;
        case "Escape":
          e.preventDefault();
          setIsOpen(false);
          setHighlightedIndex(-1);
          break;
      }
    },
    [isOpen, hasResults, hasQueried, allItems, highlightedIndex, selectItem],
  );

  const handleFocus = useCallback(() => {
    if (hasQueried) {
      setIsOpen(true);
    }
  }, [hasQueried]);

  const showDropdown = isOpen && hasQueried;

  // Build the highlighted option ID for aria-activedescendant
  const highlightedOptionId =
    highlightedIndex >= 0 ? `autocomplete-option-${generatedId}-${highlightedIndex}` : undefined;

  const inputClassName = [
    "w-full rounded-[var(--radius-md)] border bg-[var(--bg-surface)] px-3 py-2",
    "text-sm text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]",
    "transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-1",
    "focus:ring-offset-[var(--bg-base)]",
    "focus:shadow-[0_0_0_3px] focus:shadow-[var(--accent-default)]/15",
    error
      ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
      : "border-[var(--border-default)] focus:ring-[var(--accent-default)]",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ]
    .filter(Boolean)
    .join(" ");

  function renderBadge(badge: AutocompleteItem["badge"]) {
    if (!badge) return null;

    return (
      <span
        className={[
          "ml-2 inline-flex shrink-0 items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium leading-none",
          badge.variant === "success"
            ? "bg-[var(--color-success)]/15 text-[var(--color-success)]"
            : "bg-[var(--bg-elevated)] text-[var(--text-tertiary)]",
        ].join(" ")}
      >
        {badge.text}
      </span>
    );
  }

  function renderOption(item: AutocompleteItem, flatIdx: number) {
    const optionId = `autocomplete-option-${generatedId}-${flatIdx}`;
    const isHighlighted = flatIdx === highlightedIndex;

    return (
      <li
        key={`${item.value}-${flatIdx}`}
        id={optionId}
        role="option"
        aria-selected={item.value === value}
        className={[
          "flex min-w-0 cursor-pointer items-center overflow-hidden px-3 py-2 text-sm",
          isHighlighted
            ? "bg-[var(--accent-default)]/10 text-[var(--text-primary)]"
            : "text-[var(--text-primary)] hover:bg-[var(--bg-surface-hover,var(--border-subtle))]",
        ].join(" ")}
        onMouseDown={(e) => {
          e.preventDefault();
          selectItem(item);
        }}
        onMouseEnter={() => setHighlightedIndex(flatIdx)}
      >
        <span className="min-w-0 truncate">{item.label}</span>
        {renderBadge(item.badge)}
      </li>
    );
  }

  // Track flat index across groups
  let flatIndex = 0;

  return (
    <div ref={containerRef} className="relative flex flex-col gap-1.5">
      {label && (
        <label htmlFor={inputId} className="text-sm font-medium text-[var(--text-secondary)]">
          {label}
        </label>
      )}
      <input
        ref={inputRef}
        id={inputId}
        type="text"
        role="combobox"
        aria-expanded={showDropdown}
        aria-haspopup="listbox"
        aria-controls={listboxId}
        aria-autocomplete="list"
        aria-activedescendant={
          showDropdown && highlightedOptionId ? highlightedOptionId : undefined
        }
        aria-label={label ? undefined : ariaLabel}
        aria-invalid={!!error}
        aria-describedby={error ? errorId : undefined}
        placeholder={placeholder}
        value={value}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        onFocus={handleFocus}
        required={required}
        className={inputClassName}
        autoComplete="off"
      />

      <div className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {isLoading && hasQueried && loadingMessage}
        {!isLoading && hasQueried && showDropdown && !hasResults && emptyMessage}
        {!isLoading &&
          hasQueried &&
          showDropdown &&
          hasResults &&
          `${allItems.length} suggestion${allItems.length === 1 ? "" : "s"} available`}
      </div>

      <ul
        id={listboxId}
        role="listbox"
        aria-label={ariaLabel}
        className={
          showDropdown
            ? "absolute top-full left-0 z-50 mt-1 max-h-60 w-full overflow-auto rounded-[var(--radius-md)] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-lg"
            : "hidden"
        }
      >
        {showDropdown && (
          <>
            {isLoading && (
              <li role="presentation" className="px-3 py-3 text-sm text-[var(--text-tertiary)]">
                {loadingMessage}
              </li>
            )}

            {!isLoading && !hasResults && (
              <li role="presentation" className="px-3 py-3 text-sm text-[var(--text-tertiary)]">
                {emptyMessage}
              </li>
            )}

            {!isLoading &&
              hasResults &&
              groups.map((group) => {
                if (group.items.length === 0) return null;
                const groupItems = group.items.map((item) => {
                  const idx = flatIndex;
                  flatIndex++;
                  return renderOption(item, idx);
                });
                return (
                  <li key={group.label} role="presentation" className="list-none">
                    <span
                      aria-hidden="true"
                      role="presentation"
                      className="block px-3 pt-2 pb-1 text-xs font-semibold uppercase text-[var(--text-tertiary)]"
                    >
                      {group.label}
                    </span>
                    <ul role="group" aria-label={group.label} className="list-none">
                      {groupItems}
                    </ul>
                  </li>
                );
              })}
          </>
        )}
      </ul>

      {error && (
        <p id={errorId} className="text-xs text-[var(--color-error)]" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
