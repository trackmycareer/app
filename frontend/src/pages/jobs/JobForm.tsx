import { useState, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate, useLocation } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { CompanyAutocomplete } from "@/components/CompanyAutocomplete";
import { JobTitleAutocomplete } from "@/components/JobTitleAutocomplete";
import { LocationAutocomplete } from "@/components/LocationAutocomplete";
import { Select } from "@/components/Select";
import { SpinnerIcon } from "@/components/icons";
import {
  useJobQuery,
  useCreateJobMutation,
  useUpdateJobMutation,
} from "@/hooks/queries/useJobsQuery";

const EMPLOYMENT_TYPE_OPTIONS = [
  { value: "full_time", label: "Full-time" },
  { value: "part_time", label: "Part-time" },
  { value: "contract", label: "Contract" },
  { value: "freelance", label: "Freelance" },
  { value: "internship", label: "Internship" },
  { value: "education", label: "Education" },
  { value: "volunteer", label: "Volunteer" },
];

const TRANSITION_TYPE_OPTIONS = [
  { value: "", label: "None" },
  { value: "promotion", label: "Promotion" },
  { value: "lateral_move", label: "Lateral move" },
  { value: "company_change", label: "Company change" },
  { value: "first_role", label: "First role" },
];

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

export default function JobForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const routerLocation = useLocation();
  const isEditing = !!id;

  // When creating, an accepted application can pass company/title to prefill.
  const prefill = (routerLocation.state ?? null) as { company?: string; title?: string } | null;

  const { data: existingJob, isLoading: isLoadingJob } = useJobQuery(id ?? "");
  const createMutation = useCreateJobMutation();
  const updateMutation = useUpdateJobMutation();

  const [company, setCompany] = useState(() => (!isEditing && prefill?.company) || "");
  const [title, setTitle] = useState(() => (!isEditing && prefill?.title) || "");
  const [startDate, setStartDate] = useState(todayISO);
  const [endDate, setEndDate] = useState("");
  const [currentlyWorking, setCurrentlyWorking] = useState(true);
  const [employmentType, setEmploymentType] = useState("full_time");
  const [transitionType, setTransitionType] = useState("");
  const [location, setLocation] = useState("");
  const [workMode, setWorkMode] = useState("onsite");
  const [responsibilities, setResponsibilities] = useState("");
  const [notes, setNotes] = useState("");

  const [companyError, setCompanyError] = useState("");
  const [titleError, setTitleError] = useState("");
  const [startDateError, setStartDateError] = useState("");
  const [populated, setPopulated] = useState(false);

  // Populate form when editing
  useEffect(() => {
    if (isEditing && existingJob && !populated) {
      setCompany(existingJob.company);
      setTitle(existingJob.title);
      setStartDate(existingJob.start_date.slice(0, 10));
      if (existingJob.end_date) {
        setEndDate(existingJob.end_date.slice(0, 10));
        setCurrentlyWorking(false);
      } else {
        setEndDate("");
        setCurrentlyWorking(true);
      }
      setEmploymentType(existingJob.employment_type);
      setTransitionType(existingJob.transition_type ?? "");
      setLocation(existingJob.location ?? "");
      setWorkMode(existingJob.work_mode || "onsite");
      setResponsibilities(existingJob.responsibilities ?? "");
      setNotes(existingJob.notes ?? "");
      setPopulated(true);
    }
  }, [isEditing, existingJob, populated]);

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
        setTitleError("Title is required");
        hasError = true;
      } else {
        setTitleError("");
      }

      if (!startDate) {
        setStartDateError("Start date is required");
        hasError = true;
      } else {
        setStartDateError("");
      }

      if (hasError) return;

      const payload = {
        company: trimmedCompany,
        title: trimmedTitle,
        start_date: startDate,
        end_date: currentlyWorking ? null : endDate || null,
        employment_type: employmentType,
        transition_type: transitionType || null,
        location: location.trim() || null,
        work_mode: workMode,
        responsibilities: responsibilities.trim() || null,
        notes: notes.trim() || null,
      };

      if (isEditing && id) {
        updateMutation.mutate({ id, data: payload }, { onSuccess: () => navigate("/jobs") });
      } else {
        createMutation.mutate(payload, {
          onSuccess: () => navigate("/jobs"),
        });
      }
    },
    [
      company,
      title,
      startDate,
      endDate,
      currentlyWorking,
      employmentType,
      transitionType,
      location,
      workMode,
      responsibilities,
      notes,
      isEditing,
      id,
      createMutation,
      updateMutation,
      navigate,
    ],
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingJob) {
    return (
      <>
        <Topbar title="Edit Role" />
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading role...</span>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title={isEditing ? "Edit Role" : "New Role"} />
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
            label="Job title"
            placeholder="Your role or position"
            value={title}
            onChange={(value) => {
              setTitle(value);
              if (titleError) setTitleError("");
            }}
            error={titleError}
            required
          />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextInput
              label="Start date"
              type="date"
              value={startDate}
              onChange={(e) => {
                setStartDate(e.target.value);
                if (startDateError) setStartDateError("");
              }}
              error={startDateError}
              required
            />
            <div className="flex flex-col gap-1.5">
              {!currentlyWorking && (
                <TextInput
                  label="End date"
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                />
              )}
              <label className="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
                <input
                  type="checkbox"
                  checked={currentlyWorking}
                  onChange={(e) => {
                    setCurrentlyWorking(e.target.checked);
                    if (e.target.checked) setEndDate("");
                  }}
                  className="h-4 w-4 rounded border-[var(--border-default)]
                    text-[var(--accent-default)] focus:ring-[var(--accent-default)]"
                />
                I currently work here
              </label>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Select
              label="Employment type"
              value={employmentType}
              onChange={(e) => setEmploymentType(e.target.value)}
            >
              {EMPLOYMENT_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>

            <Select
              label="Transition type"
              value={transitionType}
              onChange={(e) => setTransitionType(e.target.value)}
            >
              {TRANSITION_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>
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
              <option value="onsite">Onsite</option>
              <option value="hybrid">Hybrid</option>
              <option value="remote">Remote</option>
            </Select>
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="job-responsibilities"
              className="text-sm font-medium text-[var(--text-secondary)]"
            >
              Responsibilities
            </label>
            <textarea
              id="job-responsibilities"
              value={responsibilities}
              onChange={(e) => setResponsibilities(e.target.value)}
              placeholder="Describe your key responsibilities..."
              rows={4}
              className="w-full resize-y rounded-[var(--radius-md)] border
                border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                transition-colors focus:outline-none focus:ring-2
                focus:ring-[var(--accent-default)] focus:ring-offset-1
                focus:ring-offset-[var(--bg-base)]"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="job-notes" className="text-sm font-medium text-[var(--text-secondary)]">
              Notes
            </label>
            <textarea
              id="job-notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Any additional notes about this role..."
              rows={3}
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
              onClick={() => navigate("/jobs")}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? "Save changes" : "Add role"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
