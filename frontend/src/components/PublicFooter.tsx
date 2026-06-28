import { LEGAL_URLS } from "@/lib/constants";

interface PublicFooterProps {
  className?: string;
}

// Shared site footer carrying the legal links and the company registration
// details. The legal documents live on the marketing site, so these are
// absolute links to that origin rather than in-app routes.
export function PublicFooter({ className = "" }: PublicFooterProps) {
  return (
    <footer
      className={`border-t border-[var(--border-subtle)] px-4 py-4 text-center text-xs
        text-[var(--text-tertiary)] ${className}`}
    >
      <nav
        aria-label="Legal"
        className="mb-2 flex flex-wrap items-center justify-center gap-x-3 gap-y-1"
      >
        <a
          href={LEGAL_URLS.privacy}
          target="_blank"
          className="transition-colors hover:text-[var(--text-secondary)]"
        >
          Privacy
        </a>
        <span aria-hidden="true">&middot;</span>
        <a
          href={LEGAL_URLS.terms}
          target="_blank"
          className="transition-colors hover:text-[var(--text-secondary)]"
        >
          Terms
        </a>
        <span aria-hidden="true">&middot;</span>
        <a
          href={LEGAL_URLS.cookies}
          target="_blank"
          className="transition-colors hover:text-[var(--text-secondary)]"
        >
          Cookies
        </a>
      </nav>
      <p>&copy; 2026 BH Cloud Labs Ltd. All rights reserved.</p>
      <p className="mt-1">
        BH Cloud Labs Ltd, trading as trackmy.career is registered in England and Wales (No.
        16211348).
      </p>
      <p>Registered Address: The Grange, Grange Road, Great Malvern. WR14 3HA.</p>
    </footer>
  );
}
