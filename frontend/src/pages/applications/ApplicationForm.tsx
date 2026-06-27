import { useState, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { CompanyAutocomplete } from "@/components/CompanyAutocomplete";
import { JobTitleAutocomplete } from "@/components/JobTitleAutocomplete";
import { LocationAutocomplete } from "@/components/LocationAutocomplete";
import { SpinnerIcon } from "@/components/icons";
import {
  useApplicationQuery,
  useCreateApplicationMutation,
  useUpdateApplicationMutation,
} from "@/hooks/queries/useApplicationsQuery";
import { APPLICATION_STAGES } from "./stages";
import type { ApplicationStatus } from "@/types";

export default function ApplicationForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  const { data: existing, isLoading: isLoadingApplication } = useApplicationQuery(id ?? "");
  const createMutation = useCreateApplicationMutation();
  const updateMutation = useUpdateApplicationMutation();

  const [company, setCompany] = useState("");
  const [title, setTitle] = useState("");
  const [status, setStatus] = useState<ApplicationStatus>("wishlist");
  const [appliedDate, setAppliedDate] = useState("");
  const [location, setLocation] = useState("");
  const [workMode, setWorkMode] = useState("");
  const [jobUrl, setJobUrl] = useState("");
  const [source, setSource] = useState("");
  const [salary, setSalary] = useState("");
  const [notes, setNotes] = useState("");

  const [companyError, setCompanyError] = useState("");
  const [titleError, setTitleError] = useState("");
  const [populated, setPopulated] = useState(false);

  useEffect(() => {
    if (isEditing && existing && !populated) {
      setCompany(existing.company);
      setTitle(existing.title);
      setStatus(existing.status);
      setAppliedDate(existing.applied_date ? existing.applied_date.slice(0, 10) : "");
      setLocation(existing.location ?? "");
      setWorkMode(existing.work_mode ?? "");
      setJobUrl(existing.job_url ?? "");
      setSource(existing.source ?? "");
      setSalary(existing.salary ?? "");
      setNotes(existing.notes ?? "");
      setPopulated(true);
    }
  }, [isEditing, existing, populated]);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();

      const trimmedCompany = company.trim();
      const trimmedTitle = title.trim();
      let hasError = false;

      if (!trimmedCompany) {
        setCompanyError("Company is required");
        hasError = true;
      } else {
        setCompanyError("");
      }

      if (!trimmedTitle) {
        setTitleError("Role is required");
        hasError = true;
      } else {
        setTitleError("");
      }

      if (hasError) return;

      const payload = {
        company: trimmedCompany,
        title: trimmedTitle,
        status,
        applied_date: appliedDate || null,
        location: location.trim() || null,
        work_mode: workMode || null,
        job_url: jobUrl.trim() || null,
        source: source.trim() || null,
        salary: salary.trim() || null,
        notes: notes.trim() || null,
      };

      if (isEditing && id) {
        updateMutation.mutate(
          { id, data: payload },
          { onSuccess: () => navigate("/applications") },
        );
      } else {
        createMutation.mutate(payload, { onSuccess: () => navigate("/applications") });
      }
    },
    [
      company,
      title,
      status,
      appliedDate,
      location,
      workMode,
      jobUrl,
      source,
      salary,
      notes,
      isEditing,
      id,
      createMutation,
      updateMutation,
      navigate,
    ],
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingApplication) {
    return (
      <>
        <Topbar title="Edit Application" />
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading application...</span>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title={isEditing ? "Edit Application" : "New Application"} />
      <div className="mx-auto max-w-2xl p-4 lg:p-6">
        <form onSubmit={handleSubmit} className="space-y-5">
          <CompanyAutocomplete
            label="Company"
            placeholder="Company name"
            value={company}
            onChange={(value) => {
              setCompany(value);
              if (companyError) setCompanyError("");
            }}
            error={companyError}
            required
          />

          <JobTitleAutocomplete
            label="Role"
            placeholder="The role you are applying for"
            value={title}
            onChange={(value) => {
              setTitle(value);
              if (titleError) setTitleError("");
            }}
            error={titleError}
            required
          />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Select
              label="Stage"
              value={status}
              onChange={(e) => setStatus(e.target.value as ApplicationStatus)}
            >
              {APPLICATION_STAGES.map((stage) => (
                <option key={stage.key} value={stage.key}>
                  {stage.label}
                </option>
              ))}
            </Select>
            <TextInput
              label="Date applied"
              type="date"
              value={appliedDate}
              onChange={(e) => setAppliedDate(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <LocationAutocomplete
              label="Location"
              placeholder="e.g. London, UK"
              value={location}
              onChange={(value) => setLocation(value)}
            />
            <Select
              label="Work mode"
              value={workMode}
              onChange={(e) => setWorkMode(e.target.value)}
            >
              <option value="">Not specified</option>
              <option value="onsite">Onsite</option>
              <option value="hybrid">Hybrid</option>
              <option value="remote">Remote</option>
            </Select>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextInput
              label="Salary"
              placeholder="e.g. £60k to £70k"
              value={salary}
              onChange={(e) => setSalary(e.target.value)}
            />
            <TextInput
              label="Source"
              placeholder="e.g. LinkedIn, referral"
              value={source}
              onChange={(e) => setSource(e.target.value)}
            />
          </div>

          <TextInput
            label="Job posting URL"
            type="url"
            placeholder="https://..."
            value={jobUrl}
            onChange={(e) => setJobUrl(e.target.value)}
          />

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="application-notes"
              className="text-sm font-medium text-[var(--text-secondary)]"
            >
              Notes
            </label>
            <textarea
              id="application-notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Contacts, interview prep, anything worth remembering..."
              rows={4}
              className="w-full resize-y rounded-[var(--radius-md)] border
                border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                transition-colors focus:outline-none focus:ring-2
                focus:ring-[var(--accent-default)] focus:ring-offset-1
                focus:ring-offset-[var(--bg-base)]"
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate("/applications")}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? "Save changes" : "Add application"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
