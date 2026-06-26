import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { ConfirmModal } from "@/components/ConfirmModal";
import { UserDetailDrawer } from "@/components/UserDetailDrawer";
import { SpinnerIcon, UsersIcon } from "@/components/icons";
import { apiClient } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { User } from "@/types";

const PAGE_SIZE = 20;

const adminKeys = {
  all: ["admin"] as const,
  users: () => [...adminKeys.all, "users"] as const,
  userList: (params: Record<string, string | number>) => [...adminKeys.users(), params] as const,
  stats: () => [...adminKeys.all, "stats"] as const,
};

export default function AdminUsers() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((s) => s.user);

  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null);
  const [toggleTarget, setToggleTarget] = useState<User | null>(null);
  const [detailUserId, setDetailUserId] = useState<string | null>(null);

  const params: Record<string, string | number> = {
    limit: PAGE_SIZE,
    offset,
  };
  if (search) params.search = search;

  const { data, isLoading, isError } = useQuery({
    queryKey: adminKeys.userList(params),
    queryFn: () => apiClient.admin.users.list(params).then((res) => res.data),
  });

  const users: User[] = (data as { data?: { users?: User[] } })?.data?.users ?? [];
  const total: number = (data as { data?: { total?: number } })?.data?.total ?? 0;

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.admin.users.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.users() });
      toast.success("User deleted");
      setDeleteTarget(null);
      setDetailUserId(null);
    },
  });

  const toggleAdminMutation = useMutation({
    mutationFn: ({ id, isAdmin }: { id: string; isAdmin: boolean }) =>
      apiClient.admin.users.toggleAdmin(id, isAdmin),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.users() });
      toast.success("Admin status updated");
      setToggleTarget(null);
    },
  });

  const handleSearchChange = useCallback((value: string) => {
    setSearch(value);
    setOffset(0);
  }, []);

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id);
  }, [deleteTarget, deleteMutation]);

  const handleToggleAdmin = useCallback(() => {
    if (!toggleTarget) return;
    toggleAdminMutation.mutate({
      id: toggleTarget.id,
      isAdmin: !toggleTarget.is_admin,
    });
  }, [toggleTarget, toggleAdminMutation]);

  const handleCloseDetail = useCallback(() => setDetailUserId(null), []);

  const hasMore = offset + PAGE_SIZE < total;
  const hasPrevious = offset > 0;

  return (
    <>
      <Topbar title="Users" />
      <div className="mx-auto max-w-5xl space-y-6 p-4 lg:p-6">
        {/* Search */}
        <section aria-label="Search users" className="flex flex-col gap-3 sm:flex-row">
          <div className="flex-1">
            <TextInput
              placeholder="Search by name or email..."
              value={search}
              onChange={(e) => handleSearchChange(e.target.value)}
              aria-label="Search users"
            />
          </div>
          {!isLoading && !isError && (
            <span className="flex items-center text-sm text-[var(--text-tertiary)]">
              {total} {total === 1 ? "user" : "users"}
            </span>
          )}
        </section>

        {/* Users table */}
        <section aria-label="User list">
          {isLoading && (
            <div className="flex items-center justify-center py-16" role="status">
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Loading users...</span>
            </div>
          )}

          {isError && (
            <div
              className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
                bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
              role="alert"
            >
              Failed to load users. Please try again later.
            </div>
          )}

          {!isLoading && !isError && users.length === 0 && (
            <div className="py-16 text-center">
              <UsersIcon
                width={48}
                height={48}
                className="mx-auto mb-4 text-[var(--text-tertiary)]"
                aria-hidden="true"
              />
              <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
                No users found
              </h2>
              <p className="text-sm text-[var(--text-secondary)]">
                {search ? "Try adjusting your search term." : "No users have registered yet."}
              </p>
            </div>
          )}

          {!isLoading && !isError && users.length > 0 && (
            <>
              {/* Mobile card layout */}
              <div className="space-y-3 sm:hidden">
                {users.map((u) => {
                  const isSelf = currentUser?.id === u.id;
                  return (
                    <div
                      key={u.id}
                      className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                      bg-[var(--bg-surface)] p-4"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0 flex-1">
                          <button
                            type="button"
                            onClick={() => setDetailUserId(u.id)}
                            className="inline-flex min-h-11 items-center rounded-[var(--radius-sm)]
                            text-left font-medium text-[var(--text-primary)] underline
                            decoration-[var(--border-strong)] underline-offset-2
                            hover:text-[var(--accent-default)] hover:decoration-[var(--accent-default)]
                            focus-visible:outline-none focus-visible:ring-2
                            focus-visible:ring-[var(--accent-default)] focus-visible:ring-offset-2
                            focus-visible:ring-offset-[var(--bg-surface)]"
                            aria-label={`View details for ${u.name}`}
                          >
                            {u.name}
                          </button>
                          <p className="truncate text-sm text-[var(--text-secondary)]">{u.email}</p>
                        </div>
                        {u.is_admin && (
                          <span
                            className="inline-flex shrink-0 items-center rounded-full
                            bg-[var(--accent-default)]/10 px-2 py-0.5 text-xs
                            font-medium text-[var(--accent-default)]"
                          >
                            Admin
                          </span>
                        )}
                      </div>
                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-[var(--text-tertiary)]">
                        <span
                          className="inline-flex items-center rounded-full bg-[var(--bg-elevated)]
                          px-2 py-0.5 font-medium text-[var(--text-secondary)]"
                        >
                          {u.provider}
                        </span>
                        <span>
                          {new Date(u.created_at).toLocaleDateString("en-GB", {
                            day: "numeric",
                            month: "short",
                            year: "numeric",
                          })}
                        </span>
                      </div>
                      <div className="mt-3 flex gap-2 border-t border-[var(--border-subtle)] pt-3">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setToggleTarget(u)}
                          disabled={isSelf}
                          aria-label={
                            u.is_admin ? `Remove admin from ${u.name}` : `Make admin for ${u.name}`
                          }
                        >
                          {u.is_admin ? "Remove admin" : "Make admin"}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleteTarget(u)}
                          disabled={isSelf}
                          aria-label={`Delete ${u.name}`}
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Desktop table layout */}
              <div className="hidden overflow-x-auto sm:block">
                <table className="w-full text-left text-sm">
                  <thead>
                    <tr
                      className="border-b border-[var(--border-default)]
                      text-xs uppercase tracking-wider text-[var(--text-tertiary)]"
                    >
                      <th className="px-4 py-3 font-medium" scope="col">
                        Name
                      </th>
                      <th className="px-4 py-3 font-medium" scope="col">
                        Email
                      </th>
                      <th className="px-4 py-3 font-medium" scope="col">
                        Provider
                      </th>
                      <th className="px-4 py-3 font-medium" scope="col">
                        Role
                      </th>
                      <th className="px-4 py-3 font-medium" scope="col">
                        Created
                      </th>
                      <th className="px-4 py-3 font-medium text-right" scope="col">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u) => {
                      const isSelf = currentUser?.id === u.id;
                      return (
                        <tr
                          key={u.id}
                          className="border-b border-[var(--border-subtle)] transition-colors
                          hover:bg-[var(--bg-hover)]"
                        >
                          <td className="px-4 py-3">
                            <button
                              type="button"
                              onClick={() => setDetailUserId(u.id)}
                              className="rounded-[var(--radius-sm)] py-1 text-left font-medium
                              text-[var(--text-primary)] underline decoration-[var(--border-strong)]
                              underline-offset-2 hover:text-[var(--accent-default)]
                              hover:decoration-[var(--accent-default)] focus-visible:outline-none
                              focus-visible:ring-2 focus-visible:ring-[var(--accent-default)]
                              focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--bg-base)]"
                              aria-label={`View details for ${u.name}`}
                            >
                              {u.name}
                            </button>
                          </td>
                          <td className="px-4 py-3 text-[var(--text-secondary)]">{u.email}</td>
                          <td className="px-4 py-3">
                            <span
                              className="inline-flex items-center rounded-full
                              bg-[var(--bg-elevated)] px-2 py-0.5 text-xs
                              font-medium text-[var(--text-secondary)]"
                            >
                              {u.provider}
                            </span>
                          </td>
                          <td className="px-4 py-3">
                            {u.is_admin ? (
                              <span
                                className="inline-flex items-center rounded-full
                                bg-[var(--accent-default)]/10 px-2 py-0.5 text-xs
                                font-medium text-[var(--accent-default)]"
                              >
                                Admin
                              </span>
                            ) : (
                              <span className="text-xs text-[var(--text-tertiary)]">User</span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-[var(--text-tertiary)]">
                            {new Date(u.created_at).toLocaleDateString("en-GB", {
                              day: "numeric",
                              month: "short",
                              year: "numeric",
                            })}
                          </td>
                          <td className="px-4 py-3 text-right">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setToggleTarget(u)}
                                disabled={isSelf}
                                aria-label={
                                  u.is_admin
                                    ? `Remove admin from ${u.name}`
                                    : `Make admin for ${u.name}`
                                }
                              >
                                {u.is_admin ? "Remove admin" : "Make admin"}
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setDeleteTarget(u)}
                                disabled={isSelf}
                                aria-label={`Delete ${u.name}`}
                              >
                                Delete
                              </Button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </>
          )}

          {/* Pagination */}
          {!isLoading && !isError && users.length > 0 && (hasPrevious || hasMore) && (
            <nav aria-label="Pagination" className="flex items-center justify-between pt-4">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                disabled={!hasPrevious}
              >
                Previous
              </Button>
              <span className="text-xs text-[var(--text-tertiary)]">
                Showing {offset + 1} to {Math.min(offset + PAGE_SIZE, total)} of {total}
              </span>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setOffset(offset + PAGE_SIZE)}
                disabled={!hasMore}
              >
                Next
              </Button>
            </nav>
          )}
        </section>
      </div>

      {/* User detail drawer */}
      <UserDetailDrawer
        userId={detailUserId}
        open={!!detailUserId}
        onClose={handleCloseDetail}
        isSelf={currentUser?.id === detailUserId}
        onToggleAdmin={setToggleTarget}
        onDelete={setDeleteTarget}
        confirmOpen={!!deleteTarget || !!toggleTarget}
      />

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete user"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? All their data will be permanently removed. This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />

      {/* Toggle admin confirmation modal */}
      <ConfirmModal
        open={!!toggleTarget}
        onClose={() => setToggleTarget(null)}
        onConfirm={handleToggleAdmin}
        title={toggleTarget?.is_admin ? "Remove admin privileges" : "Grant admin privileges"}
        message={
          toggleTarget?.is_admin
            ? `Are you sure you want to remove admin privileges from "${toggleTarget?.name}"?`
            : `Are you sure you want to grant admin privileges to "${toggleTarget?.name}"? They will have full access to all administrative functions.`
        }
        confirmLabel={toggleTarget?.is_admin ? "Remove admin" : "Make admin"}
        loading={toggleAdminMutation.isPending}
      />
    </>
  );
}
