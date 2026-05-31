import { useState, useEffect, useMemo, useCallback } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteItem, AutocompleteGroup } from "@/components/Autocomplete";
import { useSkillSearchQuery } from "@/hooks/queries/useSkillSearchQuery";

interface SkillAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onCategorySuggested?: (category: string) => void;
  error?: string;
  required?: boolean;
}

export function SkillAutocomplete({
  value,
  onCategorySuggested,
  ...props
}: SkillAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState("");

  useEffect(() => {
    if (value.length < 2) {
      setDebouncedQuery("");
      return;
    }
    const timer = setTimeout(() => setDebouncedQuery(value), 300);
    return () => clearTimeout(timer);
  }, [value]);

  const { data, isLoading, isError } = useSkillSearchQuery(debouncedQuery);

  const groups = useMemo((): AutocompleteGroup[] => {
    if (!data?.results || isError) return [];
    const userItems = data.results
      .filter((r) => r.source === "user")
      .map((r) => ({
        label: r.name,
        value: r.name,
        metadata: { category: r.category },
      }));
    const commonItems = data.results
      .filter((r) => r.source === "common")
      .map((r) => ({
        label: r.name,
        value: r.name,
        metadata: { category: r.category },
      }));
    const result: AutocompleteGroup[] = [];
    if (userItems.length > 0) result.push({ label: "Your skills", items: userItems });
    if (commonItems.length > 0) result.push({ label: "Common skills", items: commonItems });
    return result;
  }, [data, isError]);

  const handleSelect = useCallback(
    (item: AutocompleteItem) => {
      if (onCategorySuggested && item.metadata?.category) {
        onCategorySuggested(item.metadata.category);
      }
    },
    [onCategorySuggested],
  );

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={isLoading && debouncedQuery.length >= 2}
      hasQueried={debouncedQuery.length >= 2 && !isError}
      ariaLabel="Skill suggestions"
      loadingMessage="Searching for skills..."
      emptyMessage="No skills found"
      onSelect={handleSelect}
    />
  );
}
