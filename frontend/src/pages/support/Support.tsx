import { useState } from "react";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { HeartIcon } from "@/components/icons";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";

export default function Support() {
  const user = useAuthStore((s) => s.user);
  const isSupporter = user?.is_one_time_supporter || user?.is_subscriber;

  const [loading, setLoading] = useState<"one_time" | "subscription" | "portal" | null>(null);

  const handleCheckout = async (type: "one_time" | "subscription") => {
    if (loading) return;
    setLoading(type);
    try {
      const res = await apiClient.support.getCheckoutUrl(type);
      window.open(res.data.data.checkout_url, "_blank", "noopener,noreferrer");
    } catch {
      toast.error("Could not open the checkout page. Please try again later.");
    } finally {
      setLoading(null);
    }
  };

  const handleManage = async () => {
    if (loading) return;
    setLoading("portal");
    try {
      const res = await apiClient.support.getPortalUrl();
      window.open(res.data.data.portal_url, "_blank", "noopener,noreferrer");
    } catch {
      toast.error("Could not open the management portal. Please try again later.");
    } finally {
      setLoading(null);
    }
  };

  return (
    <>
      <Topbar title="Support" />
      <div className="mx-auto max-w-2xl space-y-8 p-4 lg:p-6">
        <div className="space-y-2">
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Support trackmy.career</h1>
          <p className="text-sm text-[var(--text-secondary)]">
            trackmy.career is free to use. If you find it useful, consider supporting its
            development. Every contribution helps keep the project running and improving.
          </p>
        </div>

        {isSupporter && (
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--accent-warm)]/30
              bg-[var(--accent-warm)]/5 p-5"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-3">
                <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--accent-warm)]/10">
                  <HeartIcon
                    width={16}
                    height={16}
                    fill="currentColor"
                    className="text-[var(--accent-warm)]"
                    aria-hidden="true"
                  />
                </div>
                <div>
                  <p className="text-sm font-semibold text-[var(--text-primary)]">
                    You're a supporter
                  </p>
                  <p className="mt-0.5 text-xs text-[var(--text-secondary)]">
                    {user?.is_subscriber && user?.is_one_time_supporter
                      ? "Monthly subscriber and one-time supporter"
                      : user?.is_subscriber
                        ? "Monthly subscriber"
                        : "One-time supporter"}
                  </p>
                </div>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={handleManage}
                loading={loading === "portal"}
                disabled={loading !== null}
              >
                Manage
              </Button>
            </div>
          </section>
        )}

        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
        >
          <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
            What supporters get
          </h2>
          <ul className="space-y-3">
            <li className="flex items-start gap-3">
              <div
                className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full
                  bg-[var(--accent-warm)]/10"
              >
                <HeartIcon
                  width={12}
                  height={12}
                  className="text-[var(--accent-warm)]"
                  aria-hidden="true"
                />
              </div>
              <span className="text-sm text-[var(--text-secondary)]">
                <strong className="font-medium text-[var(--text-primary)]">
                  Exclusive Supporter badge
                </strong>{" "}
                displayed on your profile, achievements page, and public profile.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <div
                className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full
                  bg-[var(--accent-warm)]/10"
              >
                <HeartIcon
                  width={12}
                  height={12}
                  className="text-[var(--accent-warm)]"
                  aria-hidden="true"
                />
              </div>
              <span className="text-sm text-[var(--text-secondary)]">
                <strong className="font-medium text-[var(--text-primary)]">
                  Gold tier recognition
                </strong>{" "}
                with a unique badge that stands out from regular achievement badges.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <div
                className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full
                  bg-[var(--accent-warm)]/10"
              >
                <HeartIcon
                  width={12}
                  height={12}
                  className="text-[var(--accent-warm)]"
                  aria-hidden="true"
                />
              </div>
              <span className="text-sm text-[var(--text-secondary)]">
                <strong className="font-medium text-[var(--text-primary)]">
                  Visible on your public profile
                </strong>{" "}
                so visitors can see you support the tools you use.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <div
                className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full
                  bg-[var(--accent-warm)]/10"
              >
                <HeartIcon
                  width={12}
                  height={12}
                  className="text-[var(--accent-warm)]"
                  aria-hidden="true"
                />
              </div>
              <span className="text-sm text-[var(--text-secondary)]">
                <strong className="font-medium text-[var(--text-primary)]">Custom domain</strong> to
                host your public career profile at a web address you own, instead of a shared
                trackmy.career link.
              </span>
            </li>
          </ul>
        </section>

        <div className="grid gap-4 sm:grid-cols-2">
          <div
            className="flex flex-col rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
          >
            <div className="mb-4 flex-1 space-y-2">
              <h3 className="text-base font-semibold text-[var(--text-primary)]">One-time</h3>
              {user?.is_one_time_supporter ? (
                <p className="text-sm text-[var(--accent-warm)]">
                  You've made a one-time contribution. Thank you!
                </p>
              ) : (
                <p className="text-sm text-[var(--text-secondary)]">
                  A single contribution. Your Supporter badge stays on your profile permanently.
                </p>
              )}
            </div>
            <Button
              onClick={() => handleCheckout("one_time")}
              loading={loading === "one_time"}
              disabled={loading !== null}
              className="w-full"
            >
              {user?.is_one_time_supporter ? "Support again" : "Support once"}
            </Button>
          </div>

          <div
            className={[
              "flex flex-col rounded-[var(--radius-xl)] bg-[var(--bg-surface)] p-5",
              user?.is_subscriber
                ? "border border-[var(--accent-warm)]/30 ring-1 ring-[var(--accent-warm)]/10"
                : "border border-[var(--border-default)]",
            ].join(" ")}
          >
            <div className="mb-4 flex-1 space-y-2">
              <h3 className="text-base font-semibold text-[var(--text-primary)]">Monthly</h3>
              {user?.is_subscriber ? (
                <p className="text-sm text-[var(--accent-warm)]">
                  You're an active monthly subscriber. Thank you!
                </p>
              ) : (
                <p className="text-sm text-[var(--text-secondary)]">
                  Ongoing support. Your badge stays active for as long as you're subscribed.
                </p>
              )}
            </div>
            {user?.is_subscriber ? (
              <Button
                variant="secondary"
                onClick={handleManage}
                loading={loading === "portal"}
                disabled={loading !== null}
                className="w-full"
              >
                Manage subscription
              </Button>
            ) : (
              <Button
                onClick={() => handleCheckout("subscription")}
                loading={loading === "subscription"}
                disabled={loading !== null}
                variant="primary"
                className="w-full"
              >
                Support monthly
              </Button>
            )}
          </div>
        </div>

        <p className="text-center text-xs text-[var(--text-tertiary)]">
          Payments are processed securely by Polar.
        </p>
      </div>
    </>
  );
}
