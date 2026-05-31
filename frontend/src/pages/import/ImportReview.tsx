import { useState, useCallback } from "react";
import type { ImportPreview, ImportResult, JobPreview, CertPreview, SkillPreview, WinPreview } from "@/types";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { CollapsibleSection } from "@/components/CollapsibleSection";
import { JobCard } from "./components/JobCard";
import { CertCard } from "./components/CertCard";
import { SkillCard } from "./components/SkillCard";
import { WinCard } from "./components/WinCard";
import {
  AlertTriangleIcon,
  BuildingIcon,
  AwardIcon,
  ZapIcon,
  TrophyIcon,
  CheckCircleIcon,
  UserCircleIcon,
} from "@/components/icons";

interface ImportReviewProps {
  preview: ImportPreview;
  onConfirm: (preview: ImportPreview) => void;
  onCancel: () => void;
  confirming: boolean;
  result: ImportResult | null;
}

export function ImportReview({ preview, onConfirm, onCancel, confirming, result }: ImportReviewProps) {
  const [jobs, setJobs] = useState<JobPreview[]>(preview.jobs);
  const [certs, setCerts] = useState<CertPreview[]>(preview.certifications);
  const [skills, setSkills] = useState<SkillPreview[]>(preview.skills);
  const [wins, setWins] = useState<WinPreview[]>(preview.wins ?? []);
  const [profileName, setProfileName] = useState(preview.profile?.name ?? "");
  const [profileBio, setProfileBio] = useState(preview.profile?.bio ?? "");

  const removeJob = useCallback((index: number) => {
    setJobs((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const removeCert = useCallback((index: number) => {
    setCerts((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const removeSkill = useCallback((index: number) => {
    setSkills((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const removeWin = useCallback((index: number) => {
    setWins((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const totalItems = jobs.length + certs.length + skills.length + wins.length;

  const handleConfirm = useCallback(() => {
    const updatedPreview: ImportPreview = {
      ...preview,
      jobs,
      certifications: certs,
      skills,
      wins,
      profile: preview.profile
        ? { name: profileName, bio: profileBio }
        : undefined,
    };
    onConfirm(updatedPreview);
  }, [preview, jobs, certs, skills, wins, profileName, profileBio, onConfirm]);

  // Success state
  if (result) {
    const totalCreated =
      result.jobs_created + result.certifications_created +
      result.skills_created + result.wins_created;

    return (
      <div className="animate-fade-in-up space-y-6">
        <div
          className="rounded-[var(--radius-xl)] border border-[var(--color-success)]/30
            bg-[var(--color-success)]/5 p-8 text-center"
        >
          <CheckCircleIcon
            width={48}
            height={48}
            className="mx-auto mb-4 text-[var(--color-success)]"
            aria-hidden="true"
          />
          <h2 className="mb-2 text-xl font-bold text-[var(--text-primary)]">
            Import complete
          </h2>
          <p className="mb-6 text-sm text-[var(--text-secondary)]">
            Successfully imported {totalCreated} item{totalCreated !== 1 ? "s" : ""} into your
            career profile.
          </p>

          <div className="mx-auto mb-6 grid max-w-sm grid-cols-2 gap-3">
            {result.jobs_created > 0 && (
              <div
                className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                  bg-[var(--bg-surface)] p-3"
              >
                <p className="text-2xl font-bold text-[var(--accent-default)]">
                  {result.jobs_created}
                </p>
                <p className="text-xs text-[var(--text-tertiary)]">
                  role{result.jobs_created !== 1 ? "s" : ""}
                </p>
              </div>
            )}
            {result.certifications_created > 0 && (
              <div
                className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                  bg-[var(--bg-surface)] p-3"
              >
                <p className="text-2xl font-bold text-[var(--accent-default)]">
                  {result.certifications_created}
                </p>
                <p className="text-xs text-[var(--text-tertiary)]">
                  certification{result.certifications_created !== 1 ? "s" : ""}
                </p>
              </div>
            )}
            {result.skills_created > 0 && (
              <div
                className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                  bg-[var(--bg-surface)] p-3"
              >
                <p className="text-2xl font-bold text-[var(--accent-default)]">
                  {result.skills_created}
                </p>
                <p className="text-xs text-[var(--text-tertiary)]">
                  skill{result.skills_created !== 1 ? "s" : ""}
                </p>
              </div>
            )}
            {result.wins_created > 0 && (
              <div
                className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                  bg-[var(--bg-surface)] p-3"
              >
                <p className="text-2xl font-bold text-[var(--accent-default)]">
                  {result.wins_created}
                </p>
                <p className="text-xs text-[var(--text-tertiary)]">
                  win{result.wins_created !== 1 ? "s" : ""}
                </p>
              </div>
            )}
            {result.profile_updated && (
              <div
                className="col-span-2 rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                  bg-[var(--bg-surface)] p-3"
              >
                <p className="text-sm font-medium text-[var(--accent-default)]">
                  Profile updated
                </p>
              </div>
            )}
          </div>

          <Button size="md" onClick={() => window.location.assign("/")}>
            View your dashboard
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="animate-fade-in space-y-5">
      {/* Warnings */}
      {preview.warnings.length > 0 && (
        <div
          className="rounded-[var(--radius-lg)] border border-[var(--color-warning)]/30
            bg-[var(--color-warning)]/5 p-4"
          role="alert"
        >
          <div className="mb-2 flex items-center gap-2">
            <AlertTriangleIcon
              width={18}
              height={18}
              className="flex-shrink-0 text-[var(--color-warning)]"
              aria-hidden="true"
            />
            <p className="text-sm font-medium text-[var(--color-warning)]">
              The following items may need your attention
            </p>
          </div>
          <ul className="ml-6 list-disc space-y-1">
            {preview.warnings.map((warning, i) => (
              <li key={i} className="text-sm text-[var(--text-secondary)]">
                {warning}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Profile section */}
      {preview.profile && (
        <div
          className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
            bg-[var(--bg-surface)] p-5"
        >
          <div className="mb-4 flex items-center gap-3">
            <UserCircleIcon
              width={20}
              height={20}
              className="text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <h3 className="text-sm font-semibold text-[var(--text-primary)]">Profile</h3>
            <span
              className="inline-flex items-center rounded-full bg-[var(--accent-default)]/15 px-2
                py-0.5 text-xs font-medium text-[var(--accent-default)]"
            >
              Will update your profile
            </span>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <TextInput
              label="Name"
              value={profileName}
              onChange={(e) => setProfileName(e.target.value)}
            />
            <TextInput
              label="Bio"
              value={profileBio}
              onChange={(e) => setProfileBio(e.target.value)}
            />
          </div>
        </div>
      )}

      {/* Jobs section */}
      {jobs.length > 0 && (
        <CollapsibleSection
          title={`${jobs.length} role${jobs.length !== 1 ? "s" : ""} found`}
          icon={<BuildingIcon width={18} height={18} />}
          count={jobs.length}
        >
          <div className="space-y-3">
            {jobs.map((job, i) => (
              <JobCard key={`${job.company}-${job.title}-${i}`} job={job} index={i} onRemove={() => removeJob(i)} />
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* Certifications section */}
      {certs.length > 0 && (
        <CollapsibleSection
          title={`${certs.length} certification${certs.length !== 1 ? "s" : ""} found`}
          icon={<AwardIcon width={18} height={18} />}
          count={certs.length}
        >
          <div className="space-y-3">
            {certs.map((cert, i) => (
              <CertCard key={`${cert.name}-${cert.provider}-${i}`} cert={cert} index={i} onRemove={() => removeCert(i)} />
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* Skills section */}
      {skills.length > 0 && (
        <CollapsibleSection
          title={`${skills.length} skill${skills.length !== 1 ? "s" : ""} found`}
          icon={<ZapIcon width={18} height={18} />}
          count={skills.length}
        >
          <div className="space-y-2">
            {skills.map((skill, i) => (
              <SkillCard key={`${skill.name}-${i}`} skill={skill} index={i} onRemove={() => removeSkill(i)} />
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* Wins section */}
      {wins.length > 0 && (
        <CollapsibleSection
          title={`${wins.length} win${wins.length !== 1 ? "s" : ""} found`}
          icon={<TrophyIcon width={18} height={18} />}
          count={wins.length}
        >
          <div className="space-y-3">
            {wins.map((win, i) => (
              <WinCard key={`${win.title}-${i}`} win={win} index={i} onRemove={() => removeWin(i)} />
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* Empty state after removing everything */}
      {totalItems === 0 && !preview.profile && (
        <div className="py-12 text-center">
          <p className="text-sm text-[var(--text-secondary)]">
            All items have been removed. Upload a different file or cancel.
          </p>
        </div>
      )}

      {/* Sticky summary bar */}
      <div
        className="sticky bottom-0 -mx-4 border-t border-[var(--border-subtle)]
          bg-[var(--bg-surface)]/95 px-4 py-4 backdrop-blur-sm lg:-mx-6 lg:px-6"
      >
        <div className="flex items-center justify-between gap-4">
          <p className="text-sm text-[var(--text-secondary)]">
            {totalItems > 0
              ? `Import ${jobs.length > 0 ? `${jobs.length} role${jobs.length !== 1 ? "s" : ""}` : ""}${
                  jobs.length > 0 && certs.length > 0 ? ", " : ""
                }${certs.length > 0 ? `${certs.length} certification${certs.length !== 1 ? "s" : ""}` : ""}${
                  (jobs.length > 0 || certs.length > 0) && skills.length > 0 ? ", " : ""
                }${skills.length > 0 ? `${skills.length} skill${skills.length !== 1 ? "s" : ""}` : ""}${
                  (jobs.length > 0 || certs.length > 0 || skills.length > 0) && wins.length > 0 ? ", " : ""
                }${wins.length > 0 ? `${wins.length} win${wins.length !== 1 ? "s" : ""}` : ""}`
              : "Nothing to import"}
          </p>
          <div className="flex items-center gap-3">
            <Button variant="secondary" size="sm" onClick={onCancel} disabled={confirming}>
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={handleConfirm}
              loading={confirming}
              disabled={totalItems === 0 && !preview.profile}
            >
              Confirm import
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
