import { useState, useCallback } from "react";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { DownloadIcon } from "@/components/icons";
import { apiClient } from "@/lib/api";
import { toast } from "sonner";

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

export default function ExportPage() {
  const [jsonLoading, setJsonLoading] = useState(false);
  const [mdLoading, setMdLoading] = useState(false);

  const handleExportJSON = useCallback(async () => {
    setJsonLoading(true);
    try {
      const res = await apiClient.export.json();
      const blob = new Blob([res.data as BlobPart], { type: "application/json" });
      const today = new Date().toISOString().slice(0, 10);
      downloadBlob(blob, `career-export-${today}.json`);
      toast.success("JSON export downloaded");
    } catch {
      toast.error("Failed to export JSON. Please try again.");
    } finally {
      setJsonLoading(false);
    }
  }, []);

  const handleExportMarkdown = useCallback(async () => {
    setMdLoading(true);
    try {
      const res = await apiClient.export.markdown();
      const blob = new Blob([res.data as BlobPart], { type: "text/markdown" });
      const today = new Date().toISOString().slice(0, 10);
      downloadBlob(blob, `career-export-${today}.md`);
      toast.success("Markdown export downloaded");
    } catch {
      toast.error("Failed to export Markdown. Please try again.");
    } finally {
      setMdLoading(false);
    }
  }, []);

  return (
    <>
      <Topbar title="Export" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
        >
          <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">
            Export your data
          </h2>
          <p className="mb-6 text-sm text-[var(--text-secondary)]">
            Download a complete copy of your career data. The export includes your wins,
            jobs, certifications, skills, and tags.
          </p>

          <div className="space-y-4">
            {/* JSON export */}
            <div
              className="flex items-center justify-between rounded-[var(--radius-lg)]
                border border-[var(--border-subtle)] p-4"
            >
              <div>
                <h3 className="text-sm font-medium text-[var(--text-primary)]">
                  JSON format
                </h3>
                <p className="mt-0.5 text-xs text-[var(--text-tertiary)]">
                  Structured data format. Ideal for backups or importing into other tools.
                </p>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={handleExportJSON}
                loading={jsonLoading}
                icon={<DownloadIcon width={16} height={16} />}
              >
                Download JSON
              </Button>
            </div>

            {/* Markdown export */}
            <div
              className="flex items-center justify-between rounded-[var(--radius-lg)]
                border border-[var(--border-subtle)] p-4"
            >
              <div>
                <h3 className="text-sm font-medium text-[var(--text-primary)]">
                  Markdown format
                </h3>
                <p className="mt-0.5 text-xs text-[var(--text-tertiary)]">
                  Human-readable document. Useful for sharing, printing, or adding to a portfolio.
                </p>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={handleExportMarkdown}
                loading={mdLoading}
                icon={<DownloadIcon width={16} height={16} />}
              >
                Download Markdown
              </Button>
            </div>
          </div>
        </section>
      </div>
    </>
  );
}
