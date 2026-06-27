export const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";
export const APP_DOMAIN = import.meta.env.VITE_APP_DOMAIN || "localhost";

// Published version of the Terms of Service and Privacy Policy. A signed-in user
// whose accepted version differs from this is asked to accept again. Keep in
// sync with the backend constant legal.Version in
// app/backend/internal/legal/legal.go whenever the policies change.
export const LEGAL_LAST_UPDATED = "2026-06-27";

// The marketing site hosts the legal documents. The dashboard runs on a separate
// origin (app.trackmy.career) and links out to them.
export const SITE_URL = import.meta.env.VITE_SITE_URL || "https://trackmy.career";
export const LEGAL_URLS = {
  terms: `${SITE_URL}/terms`,
  privacy: `${SITE_URL}/privacy`,
  cookies: `${SITE_URL}/cookies`,
} as const;
