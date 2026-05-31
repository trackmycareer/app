CREATE UNIQUE INDEX idx_linked_accounts_unique_provider_id
ON linked_accounts(provider, provider_id)
WHERE provider_id IS NOT NULL;
