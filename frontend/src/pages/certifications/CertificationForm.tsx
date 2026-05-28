import { useState, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { SpinnerIcon } from "@/components/icons";
import {
  useCertQuery,
  useCreateCertMutation,
  useUpdateCertMutation,
} from "@/hooks/queries/useCertsQuery";

const STATUS_OPTIONS = [
  { value: "planning", label: "Planning" },
  { value: "studying", label: "Studying" },
  { value: "scheduled", label: "Scheduled" },
  { value: "passed", label: "Passed" },
  { value: "expired", label: "Expired" },
];

export default function CertificationForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  const { data: existingCert, isLoading: isLoadingCert } = useCertQuery(
    id ?? ""
  );
  const createMutation = useCreateCertMutation();
  const updateMutation = useUpdateCertMutation();

  const [name, setName] = useState("");
  const [provider, setProvider] = useState("");
  const [status, setStatus] = useState("planning");
  const [earnedDate, setEarnedDate] = useState("");
  const [expiryDate, setExpiryDate] = useState("");
  const [cost, setCost] = useState("");
  const [currency, setCurrency] = useState("GBP");
  const [credentialUrl, setCredentialUrl] = useState("");
  const [studyNotes, setStudyNotes] = useState("");
  const [studyProgress, setStudyProgress] = useState(0);
  const [nameError, setNameError] = useState("");
  const [providerError, setProviderError] = useState("");
  const [populated, setPopulated] = useState(false);

  // Populate form when editing
  useEffect(() => {
    if (isEditing && existingCert && !populated) {
      setName(existingCert.name);
      setProvider(existingCert.provider);
      setStatus(existingCert.status);
      setEarnedDate(existingCert.earned_date ?? "");
      setExpiryDate(existingCert.expiry_date ?? "");
      setCost(
        existingCert.cost !== null && existingCert.cost !== undefined
          ? String(existingCert.cost)
          : ""
      );
      setCurrency(existingCert.currency || "GBP");
      setCredentialUrl(existingCert.credential_url ?? "");
      setStudyNotes(existingCert.study_notes ?? "");
      setStudyProgress(existingCert.study_progress);
      setPopulated(true);
    }
  }, [isEditing, existingCert, populated]);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();

      const trimmedName = name.trim();
      const trimmedProvider = provider.trim();
      let hasError = false;

      if (!trimmedName) {
        setNameError("Name is required");
        hasError = true;
      } else {
        setNameError("");
      }

      if (!trimmedProvider) {
        setProviderError("Provider is required");
        hasError = true;
      } else {
        setProviderError("");
      }

      if (hasError) return;

      const payload: Record<string, unknown> = {
        name: trimmedName,
        provider: trimmedProvider,
        status,
        currency: currency.trim() || "GBP",
        study_progress: status === "studying" ? studyProgress : 0,
      };

      if (earnedDate) payload.earned_date = earnedDate;
      if (expiryDate) payload.expiry_date = expiryDate;
      if (cost !== "") payload.cost = parseFloat(cost);
      if (credentialUrl.trim()) payload.credential_url = credentialUrl.trim();
      if (studyNotes.trim()) payload.study_notes = studyNotes.trim();

      if (isEditing && id) {
        updateMutation.mutate(
          { id, data: payload },
          { onSuccess: () => navigate("/certifications") }
        );
      } else {
        createMutation.mutate(payload, {
          onSuccess: () => navigate("/certifications"),
        });
      }
    },
    [
      name,
      provider,
      status,
      earnedDate,
      expiryDate,
      cost,
      currency,
      credentialUrl,
      studyNotes,
      studyProgress,
      isEditing,
      id,
      createMutation,
      updateMutation,
      navigate,
    ]
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingCert) {
    return (
      <>
        <Topbar title="Edit Certification" />
        <div
          className="flex items-center justify-center py-16"
          role="status"
        >
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading certification...</span>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar
        title={isEditing ? "Edit Certification" : "New Certification"}
      />
      <div className="mx-auto max-w-2xl p-4 lg:p-6">
        <form onSubmit={handleSubmit} className="space-y-5">
          <TextInput
            label="Name"
            placeholder="e.g. AWS Solutions Architect"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              if (nameError) setNameError("");
            }}
            error={nameError}
            required
          />

          <TextInput
            label="Provider"
            placeholder="e.g. Amazon Web Services"
            value={provider}
            onChange={(e) => {
              setProvider(e.target.value);
              if (providerError) setProviderError("");
            }}
            error={providerError}
            required
          />

          <Select
            label="Status"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </Select>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextInput
              label="Earned date"
              type="date"
              value={earnedDate}
              onChange={(e) => setEarnedDate(e.target.value)}
            />
            <TextInput
              label="Expiry date"
              type="date"
              value={expiryDate}
              onChange={(e) => setExpiryDate(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextInput
              label="Cost"
              type="number"
              min="0"
              step="0.01"
              placeholder="0.00"
              value={cost}
              onChange={(e) => setCost(e.target.value)}
            />
            <TextInput
              label="Currency"
              placeholder="GBP"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              maxLength={3}
            />
          </div>

          <TextInput
            label="Credential URL"
            type="url"
            placeholder="https://..."
            value={credentialUrl}
            onChange={(e) => setCredentialUrl(e.target.value)}
          />

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="cert-study-notes"
              className="text-sm font-medium text-[var(--text-secondary)]"
            >
              Study notes
            </label>
            <textarea
              id="cert-study-notes"
              value={studyNotes}
              onChange={(e) => setStudyNotes(e.target.value)}
              placeholder="Notes, resources, study plan..."
              rows={4}
              className="w-full resize-y rounded-[var(--radius-md)] border
                border-[var(--border-default)] bg-[var(--bg-surface)]
                px-3 py-2 text-sm text-[var(--text-primary)]
                placeholder:text-[var(--text-tertiary)]
                transition-colors focus:outline-none focus:ring-2
                focus:ring-[var(--accent-default)] focus:ring-offset-1
                focus:ring-offset-[var(--bg-base)]"
            />
          </div>

          {status === "studying" && (
            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="cert-study-progress"
                className="text-sm font-medium text-[var(--text-secondary)]"
              >
                Study progress ({studyProgress}%)
              </label>
              <input
                id="cert-study-progress"
                type="range"
                min={0}
                max={100}
                step={1}
                value={studyProgress}
                onChange={(e) =>
                  setStudyProgress(parseInt(e.target.value, 10))
                }
                className="h-2 w-full cursor-pointer appearance-none
                  rounded-full bg-[var(--border-subtle)]
                  accent-[var(--accent-default)]"
              />
              <div className="flex justify-between text-xs text-[var(--text-tertiary)]">
                <span>0%</span>
                <span>100%</span>
              </div>
            </div>
          )}

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate("/certifications")}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? "Save changes" : "Add certification"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
