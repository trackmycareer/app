ALTER TABLE users ADD COLUMN IF NOT EXISTS linkedin_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS github_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS website_url TEXT;

UPDATE users u SET linkedin_url = la.profile_url
FROM linked_accounts la WHERE la.user_id = u.id AND la.provider = 'linkedin';

UPDATE users u SET github_url = la.profile_url
FROM linked_accounts la WHERE la.user_id = u.id AND la.provider = 'github';

UPDATE users u SET website_url = la.profile_url
FROM linked_accounts la WHERE la.user_id = u.id AND la.provider = 'website';

DROP TABLE IF EXISTS linked_accounts;
