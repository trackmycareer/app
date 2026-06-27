import { useState, useMemo, useCallback } from "react";
import { useNavigate } from "react-router";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  closestCorners,
} from "@dnd-kit/core";
import type { DragStartEvent, DragEndEvent } from "@dnd-kit/core";
import { sortableKeyboardCoordinates } from "@dnd-kit/sortable";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { Modal } from "@/components/Modal";
import { ConfirmModal } from "@/components/ConfirmModal";
import { SpinnerIcon } from "@/components/icons";
import {
  useApplicationsQuery,
  useMoveApplicationMutation,
  useDeleteApplicationMutation,
} from "@/hooks/queries/useApplicationsQuery";
import { ApplicationColumn } from "./ApplicationColumn";
import { ApplicationCardOverlay } from "./ApplicationCard";
import { APPLICATION_STAGES } from "./stages";
import type { Application, ApplicationStatus } from "@/types";

const COLUMN_PREFIX = "column:";

function emptyGroups(): Record<ApplicationStatus, Application[]> {
  return {
    wishlist: [],
    applied: [],
    screen: [],
    interview: [],
    offer: [],
    accepted: [],
    rejected: [],
  };
}

export default function ApplicationsBoard() {
  const navigate = useNavigate();
  const { data, isLoading, isError } = useApplicationsQuery();
  const moveMutation = useMoveApplicationMutation();
  const deleteMutation = useDeleteApplicationMutation();

  const [activeId, setActiveId] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Application | null>(null);
  const [acceptTarget, setAcceptTarget] = useState<Application | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const byId = useMemo(() => {
    const map = new Map<string, Application>();
    for (const a of data ?? []) map.set(a.id, a);
    return map;
  }, [data]);

  const grouped = useMemo(() => {
    const groups = emptyGroups();
    for (const a of data ?? []) groups[a.status]?.push(a);
    for (const key of Object.keys(groups) as ApplicationStatus[]) {
      groups[key].sort((x, y) => x.sort_order - y.sort_order);
    }
    return groups;
  }, [data]);

  const total = data?.length ?? 0;
  const activeApplication = activeId ? (byId.get(activeId) ?? null) : null;

  // Builds the destination column's id order with the moved card inserted at the
  // drop position, then persists it (the server rewrites that column atomically).
  const applyMove = useCallback(
    (application: Application, destStatus: ApplicationStatus, overId: string | null) => {
      const sourceStatus = application.status;
      const destIds = grouped[destStatus].map((a) => a.id).filter((id) => id !== application.id);

      let insertIndex = destIds.length;
      if (overId && !overId.startsWith(COLUMN_PREFIX)) {
        const idx = destIds.indexOf(overId);
        if (idx !== -1) insertIndex = idx;
      }
      const orderedIds = [
        ...destIds.slice(0, insertIndex),
        application.id,
        ...destIds.slice(insertIndex),
      ];

      const currentIds = grouped[sourceStatus].map((a) => a.id);
      const unchanged =
        destStatus === sourceStatus &&
        orderedIds.length === currentIds.length &&
        orderedIds.every((id, i) => id === currentIds[i]);
      if (unchanged) return;

      moveMutation.mutate({ id: application.id, status: destStatus, ordered_ids: orderedIds });

      if (destStatus === "accepted" && sourceStatus !== "accepted") {
        setAcceptTarget(application);
      }
    },
    [grouped, moveMutation],
  );

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveId(String(event.active.id));
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      setActiveId(null);
      const { active, over } = event;
      if (!over) return;

      const application = byId.get(String(active.id));
      if (!application) return;

      const overId = String(over.id);
      const destStatus = overId.startsWith(COLUMN_PREFIX)
        ? (overId.slice(COLUMN_PREFIX.length) as ApplicationStatus)
        : (byId.get(overId)?.status ?? application.status);

      applyMove(application, destStatus, overId);
    },
    [byId, applyMove],
  );

  const handleStatusChange = useCallback(
    (application: Application, status: ApplicationStatus) => {
      if (status === application.status) return;
      applyMove(application, status, null);
    },
    [applyMove],
  );

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, { onSuccess: () => setDeleteTarget(null) });
  }, [deleteTarget, deleteMutation]);

  const handleSeedJob = useCallback(() => {
    if (!acceptTarget) return;
    navigate("/jobs/new", {
      state: { company: acceptTarget.company, title: acceptTarget.title },
    });
    setAcceptTarget(null);
  }, [acceptTarget, navigate]);

  return (
    <>
      <Topbar title="Applications" />
      <div className="p-4 lg:p-6">
        <div className="mb-5 flex items-center justify-between gap-3">
          <p className="text-sm text-[var(--text-secondary)]">
            Track roles you are pursuing. Drag a card between stages, or use its menu. Private to
            you.
          </p>
          <Button size="sm" onClick={() => navigate("/applications/new")}>
            Add application
          </Button>
        </div>

        {isLoading && (
          <div className="flex items-center justify-center py-16" role="status">
            <SpinnerIcon
              width={28}
              height={28}
              className="animate-spin text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <span className="sr-only">Loading applications...</span>
          </div>
        )}

        {isError && (
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load applications. Please try again later.
          </div>
        )}

        {!isLoading && !isError && (
          <>
            {total === 0 && (
              <p className="mb-4 text-sm text-[var(--text-tertiary)]">
                No applications yet. Add your first one to start tracking it through the pipeline.
              </p>
            )}
            <DndContext
              sensors={sensors}
              collisionDetection={closestCorners}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              onDragCancel={() => setActiveId(null)}
            >
              <div className="flex gap-4 overflow-x-auto pb-4">
                {APPLICATION_STAGES.map((stage) => (
                  <ApplicationColumn
                    key={stage.key}
                    stage={stage}
                    applications={grouped[stage.key]}
                    onEdit={(a) => navigate(`/applications/${a.id}/edit`)}
                    onDelete={setDeleteTarget}
                    onStatusChange={handleStatusChange}
                  />
                ))}
              </div>
              <DragOverlay>
                {activeApplication ? (
                  <ApplicationCardOverlay application={activeApplication} />
                ) : null}
              </DragOverlay>
            </DndContext>
          </>
        )}
      </div>

      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete application"
        message={`Are you sure you want to delete "${deleteTarget?.title} at ${deleteTarget?.company}"? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />

      <Modal
        open={!!acceptTarget}
        onClose={() => setAcceptTarget(null)}
        title="Application accepted"
      >
        <p className="mb-6 text-sm text-[var(--text-secondary)]">
          Add {acceptTarget?.title} at {acceptTarget?.company} to your Career Timeline? You can set
          the start date and other details on the next screen.
        </p>
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={() => setAcceptTarget(null)}>
            Not now
          </Button>
          <Button onClick={handleSeedJob}>Add to timeline</Button>
        </div>
      </Modal>
    </>
  );
}
