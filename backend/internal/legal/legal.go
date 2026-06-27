// Package legal holds shared constants for the published legal documents
// (Terms of Service, Privacy Policy and Cookie Policy).
package legal

// Version identifies the current published version of the Terms of Service and
// Privacy Policy. It is recorded against each user when they accept. Bumping it
// re-prompts every user for fresh consent on their next visit. Keep it in sync
// with LEGAL_LAST_UPDATED in app/frontend/src/lib/constants.ts whenever the
// policies change.
const Version = "2026-06-27"
