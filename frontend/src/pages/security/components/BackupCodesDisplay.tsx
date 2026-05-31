import { useState, useCallback } from "react";
import { toast } from "sonner";
import { Button } from "@/components/Button";
import { CopyIcon, CheckIcon, DownloadIcon, AlertTriangleIcon } from "@/components/icons";

interface BackupCodesDisplayProps {
  codes: string[];
  onDone?: () => void;
}

export function BackupCodesDisplay({ codes, onDone }: BackupCodesDisplayProps) {
  const [copied, setCopied] = useState(false);

  const handleCopyAll = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(codes.join("\n"));
      setCopied(true);
      toast.success("Backup codes copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy to clipboard");
    }
  }, [codes]);

  const handleDownload = useCallback(() => {
    const content = [
      "trackmy.career Backup Codes",
      "============================",
      "",
      "Store these codes somewhere safe.",
      "Each code can only be used once.",
      "",
      ...codes,
      "",
      `Generated: ${new Date().toISOString()}`,
    ].join("\n");

    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "trackmy-career-backup-codes.txt";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }, [codes]);

  return (
    <div className="space-y-4">
      <div
        className="flex items-start gap-2 rounded-[var(--radius-md)] border border-[var(--color-warning)]/30
          bg-[var(--color-warning)]/5 p-3"
      >
        <AlertTriangleIcon
          width={16}
          height={16}
          className="mt-0.5 shrink-0 text-[var(--color-warning)]"
        />
        <p className="text-sm text-[var(--text-secondary)]">
          Store these codes somewhere safe. Each code can only be used once. If you lose access to
          your authenticator app or passkey, you can use these codes to sign in.
        </p>
      </div>

      <div
        className="grid grid-cols-2 gap-2 rounded-[var(--radius-md)] border border-[var(--border-default)]
          bg-[var(--bg-elevated)] p-4"
      >
        {codes.map((code) => (
          <code
            key={code}
            className="rounded-[var(--radius-sm)] bg-[var(--bg-base)] px-3 py-2 text-center
              font-mono text-sm text-[var(--text-primary)]"
          >
            {code}
          </code>
        ))}
      </div>

      <div className="flex gap-3">
        <Button
          type="button"
          variant="secondary"
          size="sm"
          icon={
            copied ? (
              <CheckIcon width={14} height={14} />
            ) : (
              <CopyIcon width={14} height={14} />
            )
          }
          onClick={handleCopyAll}
        >
          {copied ? "Copied" : "Copy all"}
        </Button>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          icon={<DownloadIcon width={14} height={14} />}
          onClick={handleDownload}
        >
          Download
        </Button>
      </div>

      {onDone && (
        <div className="flex justify-end pt-2">
          <Button type="button" onClick={onDone}>
            Done
          </Button>
        </div>
      )}
    </div>
  );
}
