import { useState, useEffect, useRef, useCallback, useId, useMemo } from "react";
import type { KeyboardEvent } from "react";
import { SearchIcon } from "@/components/icons";
import { EMOJI_CATEGORIES, searchEmojis } from "@/components/emoji-data";
import type { EmojiItem } from "@/components/emoji-data";

export interface EmojiPickerProps {
  value: string;
  onChange: (emoji: string) => void;
  label?: string;
  error?: string;
  required?: boolean;
  placeholder?: string;
}

const GRID_COLS = 7;

export function EmojiPicker({
  value,
  onChange,
  label,
  error,
  placeholder = "🏅",
}: EmojiPickerProps) {
  const generatedId = useId();
  const triggerId = `emoji-trigger-${generatedId}`;
  const popoverId = `emoji-popover-${generatedId}`;
  const searchId = `emoji-search-${generatedId}`;
  const errorId = `emoji-error-${generatedId}`;
  const tabpanelId = `emoji-tabpanel-${generatedId}`;

  const containerRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const gridRef = useRef<HTMLDivElement>(null);

  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState(EMOJI_CATEGORIES[0].id);
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [announcement, setAnnouncement] = useState("");

  const activeCategoryData = EMOJI_CATEGORIES.find((c) => c.id === activeCategory);
  const displayedEmojis = useMemo<EmojiItem[]>(
    () => (searchQuery ? searchEmojis(searchQuery) : (activeCategoryData?.emojis ?? [])),
    [searchQuery, activeCategoryData],
  );

  const open = useCallback(() => {
    setIsOpen(true);
    setSearchQuery("");
    setFocusedIndex(-1);
    setActiveCategory(EMOJI_CATEGORIES[0].id);
    requestAnimationFrame(() => {
      searchRef.current?.focus();
    });
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setSearchQuery("");
    setFocusedIndex(-1);
    triggerRef.current?.focus();
  }, []);

  const selectEmoji = useCallback(
    (emoji: EmojiItem) => {
      onChange(emoji.emoji);
      setAnnouncement(`Selected ${emoji.name}`);
      close();
    },
    [onChange, close],
  );

  // Click-outside
  useEffect(() => {
    if (!isOpen) return;
    function handleMouseDown(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        close();
      }
    }
    document.addEventListener("mousedown", handleMouseDown);
    return () => document.removeEventListener("mousedown", handleMouseDown);
  }, [isOpen, close]);

  // Focus trap
  useEffect(() => {
    if (!isOpen) return;
    const container = containerRef.current;
    if (!container) return;

    function handleTab(e: globalThis.KeyboardEvent) {
      if (e.key !== "Tab") return;
      const focusable = container!.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
      );
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }

    container.addEventListener("keydown", handleTab);
    return () => container.removeEventListener("keydown", handleTab);
  }, [isOpen]);

  // Escape key
  useEffect(() => {
    if (!isOpen) return;
    function handleEscape(e: globalThis.KeyboardEvent) {
      if (e.key === "Escape") close();
    }
    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [isOpen, close]);

  // Reset focused index when displayed emojis change
  useEffect(() => {
    setFocusedIndex(-1);
  }, [searchQuery, activeCategory]);

  // Announce search result counts for screen readers
  useEffect(() => {
    if (!isOpen || !searchQuery) return;
    const count = displayedEmojis.length;
    setAnnouncement(
      count === 0
        ? "No emojis found"
        : `${count} emoji${count === 1 ? "" : "s"} found`,
    );
  }, [isOpen, searchQuery, displayedEmojis.length]);

  // Scroll focused emoji into view
  useEffect(() => {
    if (focusedIndex < 0 || !gridRef.current) return;
    const buttons = gridRef.current.querySelectorAll<HTMLElement>('[role="gridcell"] > button');
    buttons[focusedIndex]?.scrollIntoView({ block: "nearest" });
  }, [focusedIndex]);

  const handleSearchKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "ArrowDown" && displayedEmojis.length > 0) {
        e.preventDefault();
        setFocusedIndex(0);
        const firstBtn = gridRef.current?.querySelector<HTMLElement>('[role="gridcell"] > button');
        firstBtn?.focus();
      }
    },
    [displayedEmojis.length],
  );

  const handleTabKeyDown = useCallback(
    (e: KeyboardEvent<HTMLDivElement>) => {
      const ids = EMOJI_CATEGORIES.map((c) => c.id);
      const currentIdx = ids.indexOf(activeCategory);
      let nextIdx = currentIdx;

      switch (e.key) {
        case "ArrowRight":
          e.preventDefault();
          nextIdx = (currentIdx + 1) % ids.length;
          break;
        case "ArrowLeft":
          e.preventDefault();
          nextIdx = (currentIdx - 1 + ids.length) % ids.length;
          break;
        case "Home":
          e.preventDefault();
          nextIdx = 0;
          break;
        case "End":
          e.preventDefault();
          nextIdx = ids.length - 1;
          break;
        default:
          return;
      }

      setActiveCategory(ids[nextIdx]);
      requestAnimationFrame(() => {
        const tablist = e.currentTarget;
        const tabs = tablist.querySelectorAll<HTMLElement>('[role="tab"]');
        tabs[nextIdx]?.focus();
      });
    },
    [activeCategory],
  );

  const handleGridKeyDown = useCallback(
    (e: KeyboardEvent<HTMLDivElement>) => {
      const total = displayedEmojis.length;
      if (total === 0) return;

      let nextIndex = focusedIndex;

      switch (e.key) {
        case "ArrowRight":
          e.preventDefault();
          nextIndex = focusedIndex + 1 < total ? focusedIndex + 1 : focusedIndex;
          break;
        case "ArrowLeft":
          e.preventDefault();
          nextIndex = focusedIndex > 0 ? focusedIndex - 1 : 0;
          break;
        case "ArrowDown":
          e.preventDefault();
          nextIndex = focusedIndex + GRID_COLS < total ? focusedIndex + GRID_COLS : focusedIndex;
          break;
        case "ArrowUp":
          e.preventDefault();
          if (focusedIndex - GRID_COLS >= 0) {
            nextIndex = focusedIndex - GRID_COLS;
          } else {
            searchRef.current?.focus();
            setFocusedIndex(-1);
            return;
          }
          break;
        case "Enter":
        case " ":
          e.preventDefault();
          if (focusedIndex >= 0 && focusedIndex < total) {
            selectEmoji(displayedEmojis[focusedIndex]);
          }
          return;
        case "Home":
          e.preventDefault();
          nextIndex = 0;
          break;
        case "End":
          e.preventDefault();
          nextIndex = total - 1;
          break;
        default:
          return;
      }

      setFocusedIndex(nextIndex);
      const buttons = gridRef.current?.querySelectorAll<HTMLElement>('[role="gridcell"] > button');
      buttons?.[nextIndex]?.focus();
    },
    [focusedIndex, displayedEmojis, selectEmoji],
  );

  const triggerClassName = [
    "flex w-full items-center gap-2 rounded-[var(--radius-md)] border bg-[var(--bg-surface)] px-3 py-2",
    "text-left transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-1",
    "focus:ring-offset-[var(--bg-base)]",
    "focus:shadow-[0_0_0_3px] focus:shadow-[var(--accent-default)]/15",
    error
      ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
      : "border-[var(--border-default)] focus:ring-[var(--accent-default)]",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div ref={containerRef} className="relative flex flex-col gap-1.5">
      {label && (
        <label htmlFor={triggerId} className="text-sm font-medium text-[var(--text-secondary)]">
          {label}
        </label>
      )}

      <button
        ref={triggerRef}
        id={triggerId}
        type="button"
        onClick={() => (isOpen ? close() : open())}
        aria-haspopup="dialog"
        aria-expanded={isOpen}
        aria-controls={isOpen ? popoverId : undefined}
        aria-describedby={error ? errorId : undefined}
        className={triggerClassName}
      >
        <span className="text-2xl leading-none">{value || placeholder}</span>
        <span className="text-xs text-[var(--text-tertiary)]">
          {value ? "Change emoji" : "Choose emoji"}
        </span>
      </button>

      {isOpen && (
        <div
          id={popoverId}
          role="dialog"
          aria-modal="true"
          aria-label="Choose an emoji"
          className="absolute top-full left-0 z-50 mt-1 w-[min(21rem,calc(100vw-2rem))] rounded-[var(--radius-lg)] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-lg"
        >
          {/* Search */}
          <div className="flex items-center gap-2 border-b border-[var(--border-subtle)] px-3 py-2">
            <SearchIcon
              width={16}
              height={16}
              className="shrink-0 text-[var(--text-tertiary)]"
              aria-hidden="true"
            />
            <input
              ref={searchRef}
              id={searchId}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleSearchKeyDown}
              placeholder="Search emojis..."
              aria-label="Search emojis"
              autoComplete="off"
              className="w-full bg-transparent text-sm text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)] focus:outline-none"
            />
          </div>

          {/* Category tabs */}
          {!searchQuery && (
            <div
              role="tablist"
              aria-label="Emoji categories"
              tabIndex={-1}
              onKeyDown={handleTabKeyDown}
              className="flex gap-0.5 overflow-x-auto border-b border-[var(--border-subtle)] px-2 py-1.5"
            >
              {EMOJI_CATEGORIES.map((cat) => (
                <button
                  key={cat.id}
                  type="button"
                  role="tab"
                  tabIndex={cat.id === activeCategory ? 0 : -1}
                  aria-selected={cat.id === activeCategory}
                  aria-controls={tabpanelId}
                  title={cat.label}
                  onClick={() => setActiveCategory(cat.id)}
                  className={[
                    "shrink-0 rounded-[var(--radius-sm)] px-1.5 py-1 text-base transition-colors",
                    cat.id === activeCategory
                      ? "bg-[var(--accent-default)]/15 text-[var(--text-primary)]"
                      : "text-[var(--text-tertiary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]",
                  ].join(" ")}
                >
                  {cat.icon}
                </button>
              ))}
            </div>
          )}

          {/* Tabpanel with category label and emoji grid */}
          <div
            id={tabpanelId}
            role={searchQuery ? undefined : "tabpanel"}
            aria-label={searchQuery ? undefined : activeCategoryData?.label}
          >
            <div className="px-3 pt-2 pb-1 text-xs font-semibold text-[var(--text-tertiary)]">
              {searchQuery ? `Results for "${searchQuery}"` : activeCategoryData?.label}
            </div>

            {/* Emoji grid */}
            <div
              ref={gridRef}
              role="grid"
              aria-label="Emojis"
              tabIndex={-1}
              onKeyDown={handleGridKeyDown}
              className="max-h-52 overflow-y-auto px-2 pb-2"
            >
            {displayedEmojis.length > 0 ? (
              <div className="grid grid-cols-7 gap-0.5">
                {displayedEmojis.map((_emoji, idx) => {
                  const row = Math.floor(idx / GRID_COLS);
                  const isFirstInRow = idx % GRID_COLS === 0;
                  const rowEnd = Math.min((row + 1) * GRID_COLS, displayedEmojis.length);
                  const rowItems = displayedEmojis.slice(row * GRID_COLS, rowEnd);

                  return isFirstInRow ? (
                    <div
                      key={`row-${row}`}
                      role="row"
                      className="col-span-7 grid grid-cols-7 gap-0.5"
                    >
                      {rowItems.map((rowEmoji, colIdx) => {
                        const flatIdx = row * GRID_COLS + colIdx;
                        return (
                          <div key={rowEmoji.emoji + flatIdx} role="gridcell">
                            <button
                              type="button"
                              tabIndex={flatIdx === focusedIndex ? 0 : -1}
                              aria-label={rowEmoji.name}
                              title={rowEmoji.name}
                              onClick={() => selectEmoji(rowEmoji)}
                              onFocus={() => setFocusedIndex(flatIdx)}
                              className={[
                                "flex h-9 w-full items-center justify-center rounded-[var(--radius-sm)] text-xl transition-colors",
                                "hover:bg-[var(--bg-hover)]",
                                flatIdx === focusedIndex
                                  ? "bg-[var(--accent-default)]/10 ring-2 ring-[var(--accent-default)]"
                                  : "",
                              ]
                                .filter(Boolean)
                                .join(" ")}
                            >
                              {rowEmoji.emoji}
                            </button>
                          </div>
                        );
                      })}
                    </div>
                  ) : null;
                })}
              </div>
            ) : (
              <p className="py-6 text-center text-sm text-[var(--text-tertiary)]">
                No emojis found
              </p>
            )}
            </div>
          </div>
        </div>
      )}

      {/* Live region for screen readers */}
      <div className="sr-only" role="status" aria-live="polite" aria-atomic="true">
        {announcement}
      </div>

      {error && (
        <p id={errorId} className="text-xs text-[var(--color-error)]" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
