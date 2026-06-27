import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Button } from "@/components/Button";
import { Select } from "@/components/Select";
import { GripVerticalIcon } from "@/components/icons";
import { APPLICATION_STAGES } from "./stages";
import type { Application, ApplicationStatus } from "@/types";

function formatApplied(date: string | null): string | null {
  if (!date) return null;
  return new Date(date).toLocaleDateString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

function CardMeta({ application }: { application: Application }) {
  const applied = formatApplied(application.applied_date);
  return (
    <>
      {(application.location || application.work_mode || application.salary) && (
        <div className="mt-2 flex flex-wrap items-center gap-1.5">
          {application.location && (
            <span className="text-xs text-[var(--text-tertiary)]">{application.location}</span>
          )}
          {application.work_mode && application.work_mode !== "onsite" && (
            <span
              className={[
                "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                application.work_mode === "remote"
                  ? "bg-[var(--color-success)]/10 text-[var(--color-success)]"
                  : "bg-[var(--color-info)]/10 text-[var(--color-info)]",
              ].join(" ")}
            >
              {application.work_mode === "remote" ? "Remote" : "Hybrid"}
            </span>
          )}
          {application.salary && (
            <span
              className="inline-flex items-center rounded-full bg-[var(--bg-hover)] px-2 py-0.5
                text-xs font-medium text-[var(--text-secondary)]"
            >
              {application.salary}
            </span>
          )}
        </div>
      )}
      {(applied || application.source) && (
        <div className="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-[var(--text-tertiary)]">
          {applied && (
            <time dateTime={application.applied_date ?? undefined}>Applied {applied}</time>
          )}
          {application.source && <span>via {application.source}</span>}
        </div>
      )}
    </>
  );
}

interface ApplicationCardProps {
  application: Application;
  onEdit: (application: Application) => void;
  onDelete: (application: Application) => void;
  onStatusChange: (application: Application, status: ApplicationStatus) => void;
}

export function ApplicationCard({
  application,
  onEdit,
  onDelete,
  onStatusChange,
}: ApplicationCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: application.id,
    data: { type: "card", status: application.status },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  };

  const label = `${application.title} at ${application.company}`;

  return (
    <li
      ref={setNodeRef}
      style={style}
      className="rounded-[var(--radius-md)] border border-[var(--border-subtle)]
        bg-[var(--bg-elevated)] p-3 shadow-sm"
    >
      <div className="flex items-start gap-2">
        <button
          type="button"
          className="mt-0.5 shrink-0 cursor-grab touch-none rounded-[var(--radius-sm)] p-1
            text-[var(--text-tertiary)] transition-colors hover:bg-[var(--bg-hover)]
            hover:text-[var(--text-secondary)] focus-visible:outline-none focus-visible:ring-2
            focus-visible:ring-[var(--accent-default)] active:cursor-grabbing"
          aria-label={`Drag to reorder ${label}`}
          {...attributes}
          {...listeners}
        >
          <GripVerticalIcon width={16} height={16} aria-hidden="true" />
        </button>
        <div className="min-w-0 flex-1">
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">{application.title}</h3>
          <p className="text-sm text-[var(--text-secondary)]">{application.company}</p>
          <CardMeta application={application} />
        </div>
      </div>

      <div className="mt-3 flex items-center justify-between gap-2">
        <div className="w-32">
          <Select
            value={application.status}
            onChange={(e) => onStatusChange(application, e.target.value as ApplicationStatus)}
            aria-label={`Change stage for ${label}`}
            className="py-1 text-xs"
          >
            {APPLICATION_STAGES.map((stage) => (
              <option key={stage.key} value={stage.key}>
                {stage.label}
              </option>
            ))}
          </Select>
        </div>
        <div className="flex shrink-0 gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(application)}
            aria-label={`Edit ${label}`}
          >
            Edit
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(application)}
            aria-label={`Delete ${label}`}
          >
            Delete
          </Button>
        </div>
      </div>
    </li>
  );
}

// Presentational copy rendered inside the DragOverlay while a card is in flight.
export function ApplicationCardOverlay({ application }: { application: Application }) {
  return (
    <div
      className="rounded-[var(--radius-md)] border border-[var(--accent-default)]/40
        bg-[var(--bg-elevated)] p-3 shadow-lg shadow-stone-900/20"
    >
      <div className="flex items-start gap-2">
        <GripVerticalIcon
          width={16}
          height={16}
          className="mt-0.5 text-[var(--text-tertiary)]"
          aria-hidden="true"
        />
        <div className="min-w-0 flex-1">
          <h3 className="text-sm font-semibold text-[var(--text-primary)]">{application.title}</h3>
          <p className="text-sm text-[var(--text-secondary)]">{application.company}</p>
          <CardMeta application={application} />
        </div>
      </div>
    </div>
  );
}
