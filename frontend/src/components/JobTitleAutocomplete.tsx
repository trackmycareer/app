import { useState, useEffect, useMemo } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteGroup } from "@/components/Autocomplete";
import { useJobTitleSearchQuery } from "@/hooks/queries/useJobTitleSearchQuery";

interface JobTitleAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
}

export function JobTitleAutocomplete({ value, ...props }: JobTitleAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState("");

  useEffect(() => {
    if (value.length < 2) {
      setDebouncedQuery("");
      return;
    }
    const timer = setTimeout(() => setDebouncedQuery(value), 300);
    return () => clearTimeout(timer);
  }, [value]);

  const { data, isLoading, isError } = useJobTitleSearchQuery(debouncedQuery);

  const groups = useMemo((): AutocompleteGroup[] => {
    if (!data?.results || isError) return [];
    const userItems = data.results
      .filter((r) => r.source === "user")
      .map((r) => ({ label: r.title, value: r.title }));
    const commonItems = data.results
      .filter((r) => r.source === "common")
      .map((r) => ({ label: r.title, value: r.title }));
    const result: AutocompleteGroup[] = [];
    if (userItems.length > 0) result.push({ label: "Your titles", items: userItems });
    if (commonItems.length > 0) result.push({ label: "Common titles", items: commonItems });
    return result;
  }, [data, isError]);

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={isLoading && debouncedQuery.length >= 2}
      hasQueried={debouncedQuery.length >= 2 && !isError}
      ariaLabel="Job title suggestions"
      loadingMessage="Searching for job titles..."
      emptyMessage="No job titles found"
    />
  );
}
