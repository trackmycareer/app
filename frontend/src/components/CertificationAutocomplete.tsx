import { useState, useEffect, useMemo, useCallback } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteItem, AutocompleteGroup } from "@/components/Autocomplete";
import { useCertSearchQuery } from "@/hooks/queries/useCertSearchQuery";

interface CertificationAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onProviderSuggested?: (provider: string) => void;
  error?: string;
  required?: boolean;
}

export function CertificationAutocomplete({
  value,
  onProviderSuggested,
  ...props
}: CertificationAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState("");

  useEffect(() => {
    if (value.length < 2) {
      setDebouncedQuery("");
      return;
    }
    const timer = setTimeout(() => setDebouncedQuery(value), 300);
    return () => clearTimeout(timer);
  }, [value]);

  const { data, isLoading, isError } = useCertSearchQuery(debouncedQuery);

  const groups = useMemo((): AutocompleteGroup[] => {
    if (!data?.results || isError) return [];
    const userItems = data.results
      .filter((r) => r.source === "user")
      .map((r) => ({
        label: r.name,
        value: r.name,
        metadata: { provider: r.provider },
      }));
    const commonItems = data.results
      .filter((r) => r.source === "common")
      .map((r) => ({
        label: r.name,
        value: r.name,
        metadata: { provider: r.provider },
      }));
    const result: AutocompleteGroup[] = [];
    if (userItems.length > 0) result.push({ label: "Your certifications", items: userItems });
    if (commonItems.length > 0) result.push({ label: "Common certifications", items: commonItems });
    return result;
  }, [data, isError]);

  const handleSelect = useCallback(
    (item: AutocompleteItem) => {
      if (onProviderSuggested && item.metadata?.provider) {
        onProviderSuggested(item.metadata.provider);
      }
    },
    [onProviderSuggested],
  );

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={isLoading && debouncedQuery.length >= 2}
      hasQueried={debouncedQuery.length >= 2 && !isError}
      ariaLabel="Certification suggestions"
      loadingMessage="Searching for certifications..."
      emptyMessage="No certifications found"
      onSelect={handleSelect}
    />
  );
}
