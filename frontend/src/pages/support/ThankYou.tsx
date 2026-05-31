import { Link } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { HeartIcon } from "@/components/icons";

export default function SupportThankYou() {
  return (
    <>
      <Topbar title="Thank You" />
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-6 p-8 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-[var(--accent-warm)]/10">
          <HeartIcon width={32} height={32} className="text-[var(--accent-warm)]" aria-hidden="true" />
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">
            Thank you for your support!
          </h1>
          <p className="max-w-md text-sm text-[var(--text-secondary)]">
            Your contribution helps keep trackmy.career running and improving.
            Your Supporter badge will appear shortly once the payment is confirmed.
          </p>
        </div>

        <div className="flex gap-3">
          <Link to="/achievements">
            <Button variant="secondary">View achievements</Button>
          </Link>
          <Link to="/">
            <Button>Go to dashboard</Button>
          </Link>
        </div>
      </div>
    </>
  );
}
