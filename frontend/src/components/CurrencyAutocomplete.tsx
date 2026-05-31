import { useState, useEffect, useMemo } from "react";
import { Autocomplete } from "@/components/Autocomplete";
import type { AutocompleteGroup } from "@/components/Autocomplete";

interface CurrencyAutocompleteProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
}

interface Currency {
  code: string;
  name: string;
}

const currencies: Currency[] = [
  { code: "AED", name: "UAE Dirham" },
  { code: "ARS", name: "Argentine Peso" },
  { code: "AUD", name: "Australian Dollar" },
  { code: "BDT", name: "Bangladeshi Taka" },
  { code: "BGN", name: "Bulgarian Lev" },
  { code: "BHD", name: "Bahraini Dinar" },
  { code: "BRL", name: "Brazilian Real" },
  { code: "CAD", name: "Canadian Dollar" },
  { code: "CHF", name: "Swiss Franc" },
  { code: "CLP", name: "Chilean Peso" },
  { code: "CNY", name: "Chinese Yuan" },
  { code: "COP", name: "Colombian Peso" },
  { code: "CZK", name: "Czech Koruna" },
  { code: "DKK", name: "Danish Krone" },
  { code: "EGP", name: "Egyptian Pound" },
  { code: "EUR", name: "Euro" },
  { code: "GBP", name: "Pound Sterling" },
  { code: "GHS", name: "Ghanaian Cedi" },
  { code: "HKD", name: "Hong Kong Dollar" },
  { code: "HRK", name: "Croatian Kuna" },
  { code: "HUF", name: "Hungarian Forint" },
  { code: "IDR", name: "Indonesian Rupiah" },
  { code: "ILS", name: "Israeli Shekel" },
  { code: "INR", name: "Indian Rupee" },
  { code: "ISK", name: "Icelandic Krona" },
  { code: "JPY", name: "Japanese Yen" },
  { code: "KES", name: "Kenyan Shilling" },
  { code: "KRW", name: "South Korean Won" },
  { code: "KWD", name: "Kuwaiti Dinar" },
  { code: "LKR", name: "Sri Lankan Rupee" },
  { code: "MAD", name: "Moroccan Dirham" },
  { code: "MXN", name: "Mexican Peso" },
  { code: "MYR", name: "Malaysian Ringgit" },
  { code: "NGN", name: "Nigerian Naira" },
  { code: "NOK", name: "Norwegian Krone" },
  { code: "NZD", name: "New Zealand Dollar" },
  { code: "OMR", name: "Omani Rial" },
  { code: "PEN", name: "Peruvian Sol" },
  { code: "PHP", name: "Philippine Peso" },
  { code: "PKR", name: "Pakistani Rupee" },
  { code: "PLN", name: "Polish Zloty" },
  { code: "QAR", name: "Qatari Riyal" },
  { code: "RON", name: "Romanian Leu" },
  { code: "RUB", name: "Russian Rouble" },
  { code: "SAR", name: "Saudi Riyal" },
  { code: "SEK", name: "Swedish Krona" },
  { code: "SGD", name: "Singapore Dollar" },
  { code: "THB", name: "Thai Baht" },
  { code: "TRY", name: "Turkish Lira" },
  { code: "TWD", name: "Taiwan Dollar" },
  { code: "TZS", name: "Tanzanian Shilling" },
  { code: "UAH", name: "Ukrainian Hryvnia" },
  { code: "UGX", name: "Ugandan Shilling" },
  { code: "USD", name: "US Dollar" },
  { code: "UYU", name: "Uruguayan Peso" },
  { code: "VND", name: "Vietnamese Dong" },
  { code: "ZAR", name: "South African Rand" },
];

export function CurrencyAutocomplete({ value, ...props }: CurrencyAutocompleteProps) {
  const [debouncedQuery, setDebouncedQuery] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(value), 150);
    return () => clearTimeout(timer);
  }, [value]);

  const groups = useMemo((): AutocompleteGroup[] => {
    const query = debouncedQuery.toLowerCase();
    const matches = (query
      ? currencies.filter(
          (c) =>
            c.code.toLowerCase().includes(query) || c.name.toLowerCase().includes(query),
        )
      : currencies
    )
      .slice(0, 15)
      .map((c) => ({
        label: `${c.code} (${c.name})`,
        value: c.code,
      }));
    if (matches.length === 0) return [];
    return [{ label: "Currencies", items: matches }];
  }, [debouncedQuery]);

  return (
    <Autocomplete
      {...props}
      value={value}
      groups={groups}
      isLoading={false}
      hasQueried
      ariaLabel="Currency suggestions"
      emptyMessage="No currencies found"
    />
  );
}
