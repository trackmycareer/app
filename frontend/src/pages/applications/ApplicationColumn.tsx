import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { ApplicationCard } from "./ApplicationCard";
import type { StageDef } from "./stages";
import type { Application, ApplicationStatus } from "@/types";

interface ApplicationColumnProps {
  stage: StageDef;
  applications: Application[];
  onEdit: (application: Application) => void;
  onDelete: (application: Application) => void;
  onStatusChange: (application: Application, status: ApplicationStatus) => void;
}

export function ApplicationColumn({
  stage,
  applications,
  onEdit,
  onDelete,
  onStatusChange,
}: ApplicationColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: `column:${stage.key}`,
    data: { type: "column", status: stage.key },
  });

  return (
    <section
      className="flex w-72 shrink-0 flex-col rounded-[var(--radius-lg)] bg-[var(--bg-surface)]"
      aria-label={`${stage.label}, ${applications.length} ${applications.length === 1 ? "application" : "applications"}`}
    >
      <header className="flex items-center gap-2 border-b border-[var(--border-subtle)] px-3 py-2.5">
        <span
          className="h-2 w-2 rounded-full"
          style={{ backgroundColor: stage.accent }}
          aria-hidden="true"
        />
        <h2 className="text-sm font-semibold text-[var(--text-primary)]">{stage.label}</h2>
        <span
          className="ml-auto rounded-full bg-[var(--bg-elevated)] px-2 py-0.5 text-xs
            text-[var(--text-tertiary)]"
        >
          {applications.length}
        </span>
      </header>

      <SortableContext items={applications.map((a) => a.id)} strategy={verticalListSortingStrategy}>
        <ul
          ref={setNodeRef}
          className={[
            "flex min-h-24 flex-1 flex-col gap-2 p-2 transition-colors",
            isOver ? "bg-[var(--accent-subtle)]" : "",
          ].join(" ")}
        >
          {applications.map((application) => (
            <ApplicationCard
              key={application.id}
              application={application}
              onEdit={onEdit}
              onDelete={onDelete}
              onStatusChange={onStatusChange}
            />
          ))}
          {applications.length === 0 && (
            <li
              className="rounded-[var(--radius-md)] border border-dashed border-[var(--border-default)]
                p-4 text-center text-xs text-[var(--text-tertiary)]"
            >
              Drop here
            </li>
          )}
        </ul>
      </SortableContext>
    </section>
  );
}
