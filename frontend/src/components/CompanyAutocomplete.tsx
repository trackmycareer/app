import { useState, useEffect, useMemo } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteGroup } from "@/components/Autocomplete";
import { useCompanySearchQuery } from "@/hooks/queries/useCompanySearchQuery";

interface CompanyAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
}

export function CompanyAutocomplete({ value, ...props }: CompanyAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState("");

  useEffect(() => {
    if (value.length < 2) {
      setDebouncedQuery("");
      return;
    }
    const timer = setTimeout(() => setDebouncedQuery(value), 300);
    return () => clearTimeout(timer);
  }, [value]);

  const { data, isLoading, isError } = useCompanySearchQuery(debouncedQuery);

  const groups = useMemo((): AutocompleteGroup[] => {
    if (!data?.results || isError) return [];
    const userItems = data.results
      .filter((r) => r.source === "user")
      .map((r) => ({ label: r.name, value: r.name }));
    const chItems = data.results
      .filter((r) => r.source === "companies_house")
      .map((r) => ({
        label: r.name,
        value: r.name,
        badge: r.status
          ? {
              text: r.status,
              variant: (r.status.toLowerCase() === "active" ? "success" : "neutral") as
                | "success"
                | "neutral",
            }
          : undefined,
      }));
    const result: AutocompleteGroup[] = [];
    if (userItems.length > 0) result.push({ label: "Your companies", items: userItems });
    if (chItems.length > 0) result.push({ label: "Companies House", items: chItems });
    return result;
  }, [data, isError]);

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={isLoading && debouncedQuery.length >= 2}
      hasQueried={debouncedQuery.length >= 2 && !isError}
      ariaLabel="Company suggestions"
      loadingMessage="Searching for companies..."
      emptyMessage="No companies found"
    />
  );
}
