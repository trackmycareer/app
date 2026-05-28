import { Link } from "react-router";
import { Button } from "@/components/Button";

export default function NotFound() {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 p-8">
      <p className="text-6xl font-bold text-[var(--text-tertiary)]">404</p>
      <h1 className="text-xl font-semibold text-[var(--text-primary)]">Page not found</h1>
      <p className="text-sm text-[var(--text-secondary)]">
        The page you&apos;re looking for doesn&apos;t exist or has been moved.
      </p>
      <Link to="/wins">
        <Button variant="secondary">Go to Wins</Button>
      </Link>
    </div>
  );
}
