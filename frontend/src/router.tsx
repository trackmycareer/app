import { lazy, Suspense } from "react";
import { createBrowserRouter, Navigate } from "react-router";
import { RequireAuth } from "@/components/RequireAuth";
import { RequireAdmin } from "@/components/RequireAdmin";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { PageLoader } from "@/components/PageLoader";

// Lazy load all pages
const App = lazy(() => import("@/App"));
const Dashboard = lazy(() => import("@/pages/Dashboard"));
const Login = lazy(() => import("@/pages/Login"));
const Register = lazy(() => import("@/pages/Register"));
const Callback = lazy(() => import("@/pages/auth/Callback"));
const LinkCallback = lazy(() => import("@/pages/auth/LinkCallback"));
const VerifyEmailRequired = lazy(() => import("@/pages/auth/VerifyEmailRequired"));
const VerifyEmail = lazy(() => import("@/pages/auth/VerifyEmail"));
const ConfirmEmailChange = lazy(() => import("@/pages/auth/ConfirmEmailChange"));
const ForgotPassword = lazy(() => import("@/pages/auth/ForgotPassword"));
const ResetPassword = lazy(() => import("@/pages/auth/ResetPassword"));
const WinsList = lazy(() => import("@/pages/wins/WinsList"));
const WinForm = lazy(() => import("@/pages/wins/WinForm"));
const JobsList = lazy(() => import("@/pages/jobs/JobsList"));
const JobForm = lazy(() => import("@/pages/jobs/JobForm"));
const ApplicationsBoard = lazy(() => import("@/pages/applications/ApplicationsBoard"));
const ApplicationForm = lazy(() => import("@/pages/applications/ApplicationForm"));
const CertificationsList = lazy(() => import("@/pages/certifications/CertificationsList"));
const CertificationForm = lazy(() => import("@/pages/certifications/CertificationForm"));
const SkillsList = lazy(() => import("@/pages/skills/SkillsList"));
const SkillForm = lazy(() => import("@/pages/skills/SkillForm"));
const ImportPage = lazy(() => import("@/pages/import/ImportPage"));
const ExportPage = lazy(() => import("@/pages/export/ExportPage"));
const AdminUsers = lazy(() => import("@/pages/admin/AdminUsers"));
const AdminSettings = lazy(() => import("@/pages/admin/AdminSettings"));
const AdminBadges = lazy(() => import("@/pages/admin/AdminBadges"));
const Achievements = lazy(() => import("@/pages/gamification/Achievements"));
const NotificationsList = lazy(() => import("@/pages/notifications/NotificationsList"));
const ProfileSettings = lazy(() => import("@/pages/profile/ProfileSettings"));
const PublicProfile = lazy(() => import("@/pages/profile/PublicProfile"));
const Security = lazy(() => import("@/pages/security/Security"));
const Support = lazy(() => import("@/pages/support/Support"));
const SupportThankYou = lazy(() => import("@/pages/support/ThankYou"));
const NotFound = lazy(() => import("@/pages/NotFound"));

function SuspenseWrapper({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary>
      <Suspense fallback={<PageLoader />}>{children}</Suspense>
    </ErrorBoundary>
  );
}

export const router = createBrowserRouter([
  // Public routes
  {
    path: "/login",
    element: (
      <SuspenseWrapper>
        <Login />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/register",
    element: (
      <SuspenseWrapper>
        <Register />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/auth/:provider/callback",
    element: (
      <SuspenseWrapper>
        <Callback />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/auth/link/:provider/callback",
    element: (
      <SuspenseWrapper>
        <LinkCallback />
      </SuspenseWrapper>
    ),
  },

  {
    path: "/forgot-password",
    element: (
      <SuspenseWrapper>
        <ForgotPassword />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/reset-password",
    element: (
      <SuspenseWrapper>
        <ResetPassword />
      </SuspenseWrapper>
    ),
  },

  {
    path: "/verify-email-required",
    element: (
      <SuspenseWrapper>
        <VerifyEmailRequired />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/verify-email",
    element: (
      <SuspenseWrapper>
        <VerifyEmail />
      </SuspenseWrapper>
    ),
  },
  {
    path: "/confirm-email-change",
    element: (
      <SuspenseWrapper>
        <ConfirmEmailChange />
      </SuspenseWrapper>
    ),
  },

  // Public profile (no auth, no app layout)
  {
    path: "/u/:username",
    element: (
      <SuspenseWrapper>
        <PublicProfile />
      </SuspenseWrapper>
    ),
  },

  // Protected routes
  {
    element: <RequireAuth />,
    children: [
      {
        element: (
          <SuspenseWrapper>
            <App />
          </SuspenseWrapper>
        ),
        children: [
          {
            index: true,
            element: (
              <SuspenseWrapper>
                <Dashboard />
              </SuspenseWrapper>
            ),
          },
          {
            path: "wins",
            element: (
              <SuspenseWrapper>
                <WinsList />
              </SuspenseWrapper>
            ),
          },
          {
            path: "wins/new",
            element: (
              <SuspenseWrapper>
                <WinForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "wins/:id/edit",
            element: (
              <SuspenseWrapper>
                <WinForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "jobs",
            element: (
              <SuspenseWrapper>
                <JobsList />
              </SuspenseWrapper>
            ),
          },
          {
            path: "jobs/new",
            element: (
              <SuspenseWrapper>
                <JobForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "jobs/:id/edit",
            element: (
              <SuspenseWrapper>
                <JobForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "applications",
            element: (
              <SuspenseWrapper>
                <ApplicationsBoard />
              </SuspenseWrapper>
            ),
          },
          {
            path: "applications/new",
            element: (
              <SuspenseWrapper>
                <ApplicationForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "applications/:id/edit",
            element: (
              <SuspenseWrapper>
                <ApplicationForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "certifications",
            element: (
              <SuspenseWrapper>
                <CertificationsList />
              </SuspenseWrapper>
            ),
          },
          {
            path: "certifications/new",
            element: (
              <SuspenseWrapper>
                <CertificationForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "certifications/:id/edit",
            element: (
              <SuspenseWrapper>
                <CertificationForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "skills",
            element: (
              <SuspenseWrapper>
                <SkillsList />
              </SuspenseWrapper>
            ),
          },
          {
            path: "skills/new",
            element: (
              <SuspenseWrapper>
                <SkillForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "skills/:id/edit",
            element: (
              <SuspenseWrapper>
                <SkillForm />
              </SuspenseWrapper>
            ),
          },
          {
            path: "achievements",
            element: (
              <SuspenseWrapper>
                <Achievements />
              </SuspenseWrapper>
            ),
          },
          {
            path: "notifications",
            element: (
              <SuspenseWrapper>
                <NotificationsList />
              </SuspenseWrapper>
            ),
          },
          {
            path: "profile",
            element: (
              <SuspenseWrapper>
                <ProfileSettings />
              </SuspenseWrapper>
            ),
          },
          {
            path: "security",
            element: (
              <SuspenseWrapper>
                <Security />
              </SuspenseWrapper>
            ),
          },
          {
            path: "settings",
            element: <Navigate to="/profile" replace />,
          },
          {
            path: "import",
            element: (
              <SuspenseWrapper>
                <ImportPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: "export",
            element: (
              <SuspenseWrapper>
                <ExportPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: "support",
            element: (
              <SuspenseWrapper>
                <Support />
              </SuspenseWrapper>
            ),
          },
          {
            path: "support/thank-you",
            element: (
              <SuspenseWrapper>
                <SupportThankYou />
              </SuspenseWrapper>
            ),
          },
          {
            element: <RequireAdmin />,
            children: [
              {
                path: "admin/users",
                element: (
                  <SuspenseWrapper>
                    <AdminUsers />
                  </SuspenseWrapper>
                ),
              },
              {
                path: "admin/settings",
                element: (
                  <SuspenseWrapper>
                    <AdminSettings />
                  </SuspenseWrapper>
                ),
              },
              {
                path: "admin/badges",
                element: (
                  <SuspenseWrapper>
                    <AdminBadges />
                  </SuspenseWrapper>
                ),
              },
            ],
          },
          {
            path: "*",
            element: (
              <SuspenseWrapper>
                <NotFound />
              </SuspenseWrapper>
            ),
          },
        ],
      },
    ],
  },
]);
