import { useState, useCallback, useRef } from "react";
import type { DragEvent, ChangeEvent } from "react";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { Select } from "@/components/Select";
import {
  UploadIcon,
  FileTextIcon,
  DownloadIcon,
  SpinnerIcon,
  CloseIcon,
} from "@/components/icons";
import { Modal } from "@/components/Modal";
import { apiClient } from "@/lib/api";
import { toast } from "sonner";
import { ImportReview } from "@/pages/import/ImportReview";
import type { ImportPreview, ImportResult } from "@/types";

const ACCEPTED_EXTENSIONS = [".zip", ".json", ".csv", ".pdf", ".docx"];
const ACCEPTED_MIME_TYPES = [
  "application/zip",
  "application/x-zip-compressed",
  "application/json",
  "text/csv",
  "application/pdf",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
];

const SOURCE_OPTIONS = [
  { value: "", label: "Auto-detect" },
  { value: "linkedin", label: "LinkedIn" },
  { value: "jsonresume", label: "JSON Resume" },
  { value: "csv", label: "CSV" },
  { value: "document", label: "Document (PDF/DOCX)" },
];

const ENTITY_TYPE_OPTIONS = [
  { value: "jobs", label: "Jobs" },
  { value: "certifications", label: "Certifications" },
  { value: "skills", label: "Skills" },
  { value: "wins", label: "Wins" },
];

const TEMPLATE_TYPES = ["jobs", "certifications", "skills", "wins"] as const;

function detectSource(file: File): string {
  const ext = file.name.toLowerCase().split(".").pop();
  if (ext === "zip") return "linkedin";
  if (ext === "json") return "jsonresume";
  if (ext === "csv") return "csv";
  if (ext === "pdf" || ext === "docx") return "document";
  return "";
}

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

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default function ImportPage() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [source, setSource] = useState("");
  const [entityType, setEntityType] = useState("jobs");
  const [dragOver, setDragOver] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<ImportPreview | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [result, setResult] = useState<ImportResult | null>(null);
  const [templateLoading, setTemplateLoading] = useState<string | null>(null);
  const [linkedinHelpOpen, setLinkedinHelpOpen] = useState(false);

  const detectedSource = file ? detectSource(file) : "";
  const effectiveSource = source || detectedSource;
  const showEntityType = effectiveSource === "csv";

  const validateFile = useCallback((f: File): boolean => {
    const ext = `.${f.name.toLowerCase().split(".").pop()}`;
    if (!ACCEPTED_EXTENSIONS.includes(ext) && !ACCEPTED_MIME_TYPES.includes(f.type)) {
      setError(
        `Unsupported file type. Please upload a ZIP, JSON, CSV, PDF, or DOCX file.`
      );
      return false;
    }
    if (f.size > 50 * 1024 * 1024) {
      setError("File is too large. Maximum size is 50 MB.");
      return false;
    }
    return true;
  }, []);

  const handleFileSelect = useCallback(
    (f: File) => {
      setError("");
      if (validateFile(f)) {
        setFile(f);
        setSource("");
      }
    },
    [validateFile],
  );

  const handleDragOver = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragOver(false);
  }, []);

  const handleDrop = useCallback(
    (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setDragOver(false);
      const droppedFile = e.dataTransfer.files[0];
      if (droppedFile) handleFileSelect(droppedFile);
    },
    [handleFileSelect],
  );

  const handleInputChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const selected = e.target.files?.[0];
      if (selected) handleFileSelect(selected);
    },
    [handleFileSelect],
  );

  const handleUpload = useCallback(async () => {
    if (!file) return;
    setUploading(true);
    setError("");
    try {
      const res = await apiClient.import.preview(
        file,
        effectiveSource || undefined,
        showEntityType ? entityType : undefined,
      );
      setPreview(res.data.data);
    } catch (err) {
      const message =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        "Failed to process file. Please check the format and try again.";
      setError(message);
    } finally {
      setUploading(false);
    }
  }, [file, effectiveSource, showEntityType, entityType]);

  const handleConfirm = useCallback(async (updatedPreview: ImportPreview) => {
    setConfirming(true);
    try {
      const res = await apiClient.import.confirm(updatedPreview);
      setResult(res.data.data);
      toast.success("Import completed successfully");
    } catch {
      toast.error("Failed to confirm import. Please try again.");
    } finally {
      setConfirming(false);
    }
  }, []);

  const handleCancel = useCallback(() => {
    setPreview(null);
    setResult(null);
    setFile(null);
    setSource("");
    setError("");
  }, []);

  const handleClearFile = useCallback(() => {
    setFile(null);
    setSource("");
    setError("");
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }, []);

  const handleDownloadTemplate = useCallback(async (type: string) => {
    setTemplateLoading(type);
    try {
      const res = await apiClient.import.downloadTemplate(type);
      const blob = new Blob([res.data as BlobPart], { type: "text/csv" });
      downloadBlob(blob, `${type}-template.csv`);
      toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} template downloaded`);
    } catch {
      toast.error("Failed to download template. Please try again.");
    } finally {
      setTemplateLoading(null);
    }
  }, []);

  return (
    <>
      <Topbar title="Import" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        {/* Review state */}
        {preview && (
          <ImportReview
            preview={preview}
            onConfirm={handleConfirm}
            onCancel={handleCancel}
            confirming={confirming}
            result={result}
          />
        )}

        {/* Upload state */}
        {!preview && (
          <div className="animate-fade-in space-y-6">
            {/* Header section */}
            <section
              className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
                bg-[var(--bg-surface)] p-5"
            >
              <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">
                Import your career data
              </h2>
              <p className="mb-6 text-sm text-[var(--text-secondary)]">
                Upload a{" "}
                <button
                  type="button"
                  onClick={() => setLinkedinHelpOpen(true)}
                  className="font-medium text-[var(--accent-default)] underline decoration-[var(--accent-default)]/30
                    underline-offset-2 transition-colors hover:text-[var(--accent-bright)]
                    hover:decoration-[var(--accent-bright)]/50"
                >
                  LinkedIn data export
                </button>
                , JSON Resume, CSV file, or document (PDF, DOCX) to populate your career profile.
                We will extract your positions, education, volunteering, certifications, skills,
                projects, publications, organisations, recommendations, endorsements, and LinkedIn
                Learning courses automatically.
              </p>

              <Modal
                open={linkedinHelpOpen}
                onClose={() => setLinkedinHelpOpen(false)}
                title="How to download your LinkedIn data"
              >
                <ol className="list-inside list-decimal space-y-3 text-sm text-[var(--text-secondary)]">
                  <li>
                    Open LinkedIn and click your <strong className="text-[var(--text-primary)]">profile picture</strong> in the top right corner.
                  </li>
                  <li>
                    Select <strong className="text-[var(--text-primary)]">Settings &amp; Privacy</strong>.
                  </li>
                  <li>
                    Go to <strong className="text-[var(--text-primary)]">Data privacy</strong> in the left sidebar.
                  </li>
                  <li>
                    Under "How LinkedIn uses your data", click{" "}
                    <strong className="text-[var(--text-primary)]">Download your data</strong>.
                  </li>
                  <li>
                    Select <strong className="text-[var(--text-primary)]">Download larger data archive</strong> to
                    ensure your positions, certifications, and skills are included. Alternatively,
                    choose <strong className="text-[var(--text-primary)]">"Want something in particular?"</strong> and
                    tick <strong className="text-[var(--text-primary)]">Profile</strong> (and any
                    other categories you see that relate to your career history).
                  </li>
                  <li>
                    Click <strong className="text-[var(--text-primary)]">Request archive</strong>.
                    LinkedIn will email you when the file is ready (usually within a few minutes).
                  </li>
                  <li>
                    Download the ZIP file from the email link, then upload it here.
                  </li>
                </ol>
                <p className="mt-4 text-xs text-[var(--text-tertiary)]">
                  We extract positions, education, volunteering, certifications, skills, projects,
                  publications, organisations, recommendations, endorsements, and LinkedIn Learning
                  courses from the export. Everything else in the archive is ignored.
                </p>
              </Modal>

              {/* Drop zone */}
              <div
                role="button"
                tabIndex={0}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    fileInputRef.current?.click();
                  }
                }}
                className={[
                  "relative flex cursor-pointer flex-col items-center justify-center gap-3",
                  "rounded-[var(--radius-lg)] border-2 border-dashed p-8 transition-all duration-200",
                  dragOver
                    ? "border-[var(--accent-default)] bg-[var(--accent-default)]/5"
                    : "border-[var(--border-default)] hover:border-[var(--accent-default)]/50 hover:bg-[var(--bg-elevated)]",
                  uploading ? "pointer-events-none opacity-60" : "",
                ].join(" ")}
                aria-label="Upload file. Drop your file here or click to browse."
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept={ACCEPTED_EXTENSIONS.join(",")}
                  onChange={handleInputChange}
                  className="sr-only"
                  aria-hidden="true"
                  tabIndex={-1}
                />

                {uploading ? (
                  <>
                    <SpinnerIcon
                      width={32}
                      height={32}
                      className="animate-spin text-[var(--accent-default)]"
                      aria-hidden="true"
                    />
                    <p className="text-sm font-medium text-[var(--text-primary)]">
                      Analysing your file...
                    </p>
                  </>
                ) : (
                  <>
                    <UploadIcon
                      width={32}
                      height={32}
                      className={
                        dragOver
                          ? "text-[var(--accent-default)]"
                          : "text-[var(--text-tertiary)]"
                      }
                      aria-hidden="true"
                    />
                    <div className="text-center">
                      <p className="text-sm font-medium text-[var(--text-primary)]">
                        Drop your file here, or click to browse
                      </p>
                      <p className="mt-1 text-xs text-[var(--text-tertiary)]">
                        ZIP (LinkedIn), JSON (JSON Resume), CSV, PDF, DOCX
                      </p>
                    </div>
                  </>
                )}
              </div>

              {/* Selected file details */}
              {file && !uploading && (
                <div className="mt-4 animate-fade-in-up space-y-4">
                  <div
                    className="flex items-center gap-3 rounded-[var(--radius-lg)] border
                      border-[var(--border-subtle)] bg-[var(--bg-elevated)] px-4 py-3"
                  >
                    <FileTextIcon
                      width={20}
                      height={20}
                      className="flex-shrink-0 text-[var(--accent-default)]"
                      aria-hidden="true"
                    />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-[var(--text-primary)]">
                        {file.name}
                      </p>
                      <p className="text-xs text-[var(--text-tertiary)]">
                        {formatFileSize(file.size)}
                        {detectedSource && (
                          <span>
                            {" "}
                            &middot; Detected as{" "}
                            <span className="font-medium text-[var(--accent-default)]">
                              {SOURCE_OPTIONS.find((s) => s.value === detectedSource)?.label ??
                                detectedSource}
                            </span>
                          </span>
                        )}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleClearFile();
                      }}
                      className="flex-shrink-0 rounded-[var(--radius-sm)] p-1
                        text-[var(--text-tertiary)] transition-colors hover:bg-[var(--bg-hover)]
                        hover:text-[var(--text-primary)]"
                      aria-label="Remove selected file"
                    >
                      <CloseIcon width={16} height={16} />
                    </button>
                  </div>

                  {/* Source override */}
                  <Select
                    label="Source (override auto-detection)"
                    value={source}
                    onChange={(e) => setSource(e.target.value)}
                  >
                    {SOURCE_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </Select>

                  {/* Entity type for CSV */}
                  {showEntityType && (
                    <Select
                      label="What does this CSV contain?"
                      value={entityType}
                      onChange={(e) => setEntityType(e.target.value)}
                    >
                      {ENTITY_TYPE_OPTIONS.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </Select>
                  )}

                  <Button
                    size="md"
                    onClick={handleUpload}
                    icon={<UploadIcon width={16} height={16} />}
                  >
                    Upload and preview
                  </Button>
                </div>
              )}

              {/* Error message */}
              {error && (
                <div
                  className="mt-4 rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
                    bg-[var(--color-error)]/5 p-4 text-center"
                  role="alert"
                >
                  <p className="text-sm text-[var(--color-error)]">{error}</p>
                  <button
                    type="button"
                    onClick={handleClearFile}
                    className="mt-2 text-sm font-medium text-[var(--accent-default)]
                      transition-colors hover:text-[var(--accent-bright)]"
                  >
                    Try again
                  </button>
                </div>
              )}
            </section>

            {/* Template downloads */}
            <section
              className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
                bg-[var(--bg-surface)] p-5"
            >
              <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">
                Or start from a template
              </h2>
              <p className="mb-4 text-sm text-[var(--text-secondary)]">
                Download a CSV template, fill it in with your data, then upload the CSV above.
              </p>

              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {TEMPLATE_TYPES.map((type) => (
                  <Button
                    key={type}
                    variant="secondary"
                    size="sm"
                    onClick={() => handleDownloadTemplate(type)}
                    loading={templateLoading === type}
                    icon={<DownloadIcon width={14} height={14} />}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </Button>
                ))}
              </div>
            </section>
          </div>
        )}
      </div>
    </>
  );
}
