import type { Certification } from "@/types";

interface CertificationListProps {
  certs: Certification[];
}

export function CertificationList({ certs }: CertificationListProps) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {certs.map((cert) => (
        <div
          key={cert.id}
          className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
            bg-[var(--bg-elevated)] p-4"
        >
          <p className="text-sm font-semibold text-[var(--text-primary)]">{cert.name}</p>
          <p className="text-xs text-[var(--text-secondary)]">{cert.provider}</p>
          {cert.earned_date && (
            <p className="mt-1 text-xs text-[var(--text-tertiary)]">
              Earned{" "}
              {new Date(cert.earned_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
            </p>
          )}
        </div>
      ))}
    </div>
  );
}
