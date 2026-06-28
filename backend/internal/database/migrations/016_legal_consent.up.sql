-- Record acceptance of the Terms of Service and Privacy Policy.
-- Both columns are left NULL for existing users on purpose: they predate the
-- published policies, so they are treated as not-yet-consented and are shown a
-- one-time acceptance prompt on their next visit. New registrations set these
-- at account creation. The same mechanism re-prompts everyone when the policy
-- version (auth.TermsVersion) changes.
ALTER TABLE users ADD COLUMN terms_accepted_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN terms_version TEXT;
