import { useState, useEffect, useCallback, useMemo } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { SpinnerIcon, CloseIcon } from "@/components/icons";
import {
  useSkillQuery,
  useCreateSkillMutation,
  useUpdateSkillMutation,
  useAddEvidenceMutation,
  useRemoveEvidenceMutation,
} from "@/hooks/queries/useSkillsQuery";
import { useWinsQuery } from "@/hooks/queries/useWinsQuery";
import { useCertsQuery } from "@/hooks/queries/useCertsQuery";
import { useJobsQuery } from "@/hooks/queries/useJobsQuery";
import type { SkillEvidence } from "@/types";

const PROFICIENCY_OPTIONS = [
  { value: 1, label: "Beginner" },
  { value: 2, label: "Intermediate" },
  { value: 3, label: "Advanced" },
  { value: 4, label: "Expert" },
];

function EvidenceTypeLabel({ type }: { type: string }) {
  const labels: Record<string, string> = {
    win: "Win",
    certification: "Certification",
    job: "Job",
  };
  return (
    <span
      className="inline-flex items-center rounded-full bg-[var(--bg-elevated)] px-2 py-0.5
        text-xs font-medium text-[var(--text-secondary)]"
    >
      {labels[type] ?? type}
    </span>
  );
}

export default function SkillForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  const { data: existingSkill, isLoading: isLoadingSkill } = useSkillQuery(id ?? "");
  const createMutation = useCreateSkillMutation();
  const updateMutation = useUpdateSkillMutation();
  const addEvidenceMutation = useAddEvidenceMutation();
  const removeEvidenceMutation = useRemoveEvidenceMutation();

  // Load related entities for evidence linking
  const { data: winsData } = useWinsQuery({ limit: 200 });
  const { data: certsData } = useCertsQuery({});
  const { data: jobsData } = useJobsQuery();

  const wins = useMemo(() => {
    if (!winsData) return [];
    const raw = winsData as unknown;
    if (Array.isArray(raw)) return raw;
    if (raw && typeof raw === "object" && "wins" in raw) {
      return (raw as { wins: unknown[] }).wins ?? [];
    }
    return [];
  }, [winsData]);

  const certs = useMemo(() => {
    if (!certsData) return [];
    if (Array.isArray(certsData)) return certsData;
    if (certsData && typeof certsData === "object" && "certifications" in certsData) {
      return (certsData as { certifications: unknown[] }).certifications ?? [];
    }
    return [];
  }, [certsData]);

  const jobs = useMemo(() => {
    if (!jobsData) return [];
    const raw = jobsData as unknown;
    if (Array.isArray(raw)) return raw;
    if (raw && typeof raw === "object" && "data" in raw) {
      return (raw as { data: unknown[] }).data ?? [];
    }
    return [];
  }, [jobsData]);

  const [name, setName] = useState("");
  const [category, setCategory] = useState("");
  const [proficiency, setProficiency] = useState(1);
  const [notes, setNotes] = useState("");
  const [nameError, setNameError] = useState("");
  const [populated, setPopulated] = useState(false);

  // Evidence add form
  const [addEvidenceType, setAddEvidenceType] = useState("win");
  const [addEvidenceId, setAddEvidenceId] = useState("");

  // Populate form when editing
  useEffect(() => {
    if (isEditing && existingSkill && !populated) {
      setName(existingSkill.name);
      setCategory(existingSkill.category ?? "");
      setProficiency(existingSkill.proficiency);
      setNotes(existingSkill.notes ?? "");
      setPopulated(true);
    }
  }, [isEditing, existingSkill, populated]);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = name.trim();
      if (!trimmed) {
        setNameError("Name is required");
        return;
      }
      setNameError("");

      const payload = {
        name: trimmed,
        category: category.trim() || undefined,
        proficiency,
        notes: notes.trim() || undefined,
      };

      if (isEditing && id) {
        updateMutation.mutate(
          { id, data: payload },
          { onSuccess: () => navigate("/skills") },
        );
      } else {
        createMutation.mutate(payload, {
          onSuccess: () => navigate("/skills"),
        });
      }
    },
    [name, category, proficiency, notes, isEditing, id, createMutation, updateMutation, navigate],
  );

  const handleAddEvidence = useCallback(() => {
    if (!id || !addEvidenceId) return;
    addEvidenceMutation.mutate({
      skillId: id,
      evidenceType: addEvidenceType,
      evidenceId: addEvidenceId,
    }, {
      onSuccess: () => setAddEvidenceId(""),
    });
  }, [id, addEvidenceType, addEvidenceId, addEvidenceMutation]);

  const handleRemoveEvidence = useCallback(
    (evidence: SkillEvidence) => {
      if (!id) return;
      removeEvidenceMutation.mutate({
        skillId: id,
        evidenceId: evidence.id,
      });
    },
    [id, removeEvidenceMutation],
  );

  // Build evidence source options based on selected type
  const evidenceOptions = useMemo(() => {
    const alreadyLinked = new Set(
      (existingSkill?.evidence ?? [])
        .filter((e) => e.evidence_type === addEvidenceType)
        .map((e) => e.evidence_id),
    );

    if (addEvidenceType === "win") {
      return (wins as { id: string; title: string }[])
        .filter((w) => !alreadyLinked.has(w.id))
        .map((w) => ({ value: w.id, label: w.title }));
    }
    if (addEvidenceType === "certification") {
      return (certs as { id: string; name: string }[])
        .filter((c) => !alreadyLinked.has(c.id))
        .map((c) => ({ value: c.id, label: c.name }));
    }
    if (addEvidenceType === "job") {
      return (jobs as { id: string; company: string; title: string }[])
        .filter((j) => !alreadyLinked.has(j.id))
        .map((j) => ({ value: j.id, label: `${j.title} at ${j.company}` }));
    }
    return [];
  }, [addEvidenceType, wins, certs, jobs, existingSkill]);

  // Look up display name for an evidence item
  const getEvidenceLabel = useCallback(
    (evidence: SkillEvidence): string => {
      if (evidence.evidence_type === "win") {
        const w = (wins as { id: string; title: string }[]).find(
          (item) => item.id === evidence.evidence_id,
        );
        return w?.title ?? evidence.evidence_id;
      }
      if (evidence.evidence_type === "certification") {
        const c = (certs as { id: string; name: string }[]).find(
          (item) => item.id === evidence.evidence_id,
        );
        return c?.name ?? evidence.evidence_id;
      }
      if (evidence.evidence_type === "job") {
        const j = (jobs as { id: string; company: string; title: string }[]).find(
          (item) => item.id === evidence.evidence_id,
        );
        return j ? `${j.title} at ${j.company}` : evidence.evidence_id;
      }
      return evidence.evidence_id;
    },
    [wins, certs, jobs],
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingSkill) {
    return (
      <>
        <Topbar title="Edit Skill" />
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading skill...</span>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title={isEditing ? "Edit Skill" : "New Skill"} />
      <div className="mx-auto max-w-2xl p-4 lg:p-6">
        <form onSubmit={handleSubmit} className="space-y-5">
          <TextInput
            label="Name"
            placeholder="e.g. TypeScript, System Design, AWS"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              if (nameError) setNameError("");
            }}
            error={nameError}
            required
          />

          <TextInput
            label="Category"
            placeholder="e.g. Programming Languages, Cloud, Soft Skills"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          />

          {/* Proficiency segmented control */}
          <fieldset>
            <legend className="mb-1.5 text-sm font-medium text-[var(--text-secondary)]">
              Proficiency
            </legend>
            <div
              className="inline-flex rounded-[var(--radius-md)] border border-[var(--border-default)]
                bg-[var(--bg-surface)] p-0.5"
              role="radiogroup"
              aria-label="Proficiency level"
            >
              {PROFICIENCY_OPTIONS.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  role="radio"
                  aria-checked={proficiency === opt.value}
                  onClick={() => setProficiency(opt.value)}
                  className={`rounded-[var(--radius-sm)] px-3 py-1.5 text-sm font-medium
                    transition-all ${
                      proficiency === opt.value
                        ? "bg-[var(--accent-default)] text-white shadow-sm"
                        : "text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
                    }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </fieldset>

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="skill-notes"
              className="text-sm font-medium text-[var(--text-secondary)]"
            >
              Notes
            </label>
            <textarea
              id="skill-notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Any additional notes about this skill..."
              rows={4}
              className="w-full resize-y rounded-[var(--radius-md)] border
                border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]
                focus:ring-offset-1 focus:ring-offset-[var(--bg-base)]"
            />
          </div>

          {/* Evidence section (edit mode only) */}
          {isEditing && existingSkill && (
            <section
              aria-label="Skill evidence"
              className="space-y-3 rounded-[var(--radius-lg)] border border-[var(--border-default)]
                bg-[var(--bg-base)] p-4"
            >
              <h3 className="text-sm font-semibold text-[var(--text-primary)]">
                Evidence
              </h3>

              {/* Existing evidence items */}
              {existingSkill.evidence.length === 0 && (
                <p className="text-sm text-[var(--text-tertiary)]">
                  No evidence linked yet. Link wins, certifications, or jobs below.
                </p>
              )}
              {existingSkill.evidence.length > 0 && (
                <ul className="space-y-2" aria-label="Linked evidence">
                  {existingSkill.evidence.map((ev) => (
                    <li
                      key={ev.id}
                      className="flex items-center justify-between gap-2 rounded-[var(--radius-md)]
                        border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-3 py-2"
                    >
                      <div className="flex items-center gap-2 overflow-hidden">
                        <EvidenceTypeLabel type={ev.evidence_type} />
                        <span className="truncate text-sm text-[var(--text-primary)]">
                          {getEvidenceLabel(ev)}
                        </span>
                      </div>
                      <button
                        type="button"
                        onClick={() => handleRemoveEvidence(ev)}
                        disabled={removeEvidenceMutation.isPending}
                        className="shrink-0 rounded-[var(--radius-sm)] p-1 text-[var(--text-tertiary)]
                          transition-colors hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]
                          disabled:opacity-50"
                        aria-label={`Remove evidence: ${getEvidenceLabel(ev)}`}
                      >
                        <CloseIcon width={14} height={14} />
                      </button>
                    </li>
                  ))}
                </ul>
              )}

              {/* Add evidence form */}
              <div className="flex flex-col gap-2 pt-2 sm:flex-row sm:items-end">
                <div className="w-full sm:w-40">
                  <Select
                    label="Type"
                    value={addEvidenceType}
                    onChange={(e) => {
                      setAddEvidenceType(e.target.value);
                      setAddEvidenceId("");
                    }}
                  >
                    <option value="win">Win</option>
                    <option value="certification">Certification</option>
                    <option value="job">Job</option>
                  </Select>
                </div>
                <div className="flex-1">
                  <Select
                    label="Item"
                    value={addEvidenceId}
                    onChange={(e) => setAddEvidenceId(e.target.value)}
                    disabled={evidenceOptions.length === 0}
                  >
                    <option value="">
                      {evidenceOptions.length === 0
                        ? "No items available"
                        : "Select an item..."}
                    </option>
                    {evidenceOptions.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </Select>
                </div>
                <Button
                  type="button"
                  size="sm"
                  onClick={handleAddEvidence}
                  disabled={!addEvidenceId}
                  loading={addEvidenceMutation.isPending}
                >
                  Link
                </Button>
              </div>
            </section>
          )}

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate("/skills")}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? "Save changes" : "Create skill"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
