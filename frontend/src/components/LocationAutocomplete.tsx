import { useState, useEffect, useMemo } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteGroup } from "@/components/Autocomplete";
import { useLocationSearchQuery } from "@/hooks/queries/useLocationSearchQuery";

interface LocationAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
}

export function LocationAutocomplete({ value, ...props }: LocationAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState("");

  useEffect(() => {
    if (value.length < 2) {
      setDebouncedQuery("");
      return;
    }
    const timer = setTimeout(() => setDebouncedQuery(value), 300);
    return () => clearTimeout(timer);
  }, [value]);

  const { data, isLoading, isError } = useLocationSearchQuery(debouncedQuery);

  const groups = useMemo((): AutocompleteGroup[] => {
    if (!data?.results || isError) return [];
    const userItems = data.results
      .filter((r) => r.source === "user")
      .map((r) => ({ label: r.label, value: r.label }));
    const photonItems = data.results
      .filter((r) => r.source === "photon")
      .map((r) => ({ label: r.label, value: r.label }));
    const result: AutocompleteGroup[] = [];
    if (userItems.length > 0) result.push({ label: "Your locations", items: userItems });
    if (photonItems.length > 0) result.push({ label: "Suggestions", items: photonItems });
    return result;
  }, [data, isError]);

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={isLoading && debouncedQuery.length >= 2}
      hasQueried={debouncedQuery.length >= 2 && !isError}
      ariaLabel="Location suggestions"
      loadingMessage="Searching for locations..."
      emptyMessage="No locations found"
    />
  );
}
