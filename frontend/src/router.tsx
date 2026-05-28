import { lazy, Suspense } from "react";
import { createBrowserRouter } from "react-router";
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
const WinsList = lazy(() => import("@/pages/wins/WinsList"));
const WinForm = lazy(() => import("@/pages/wins/WinForm"));
const JobsList = lazy(() => import("@/pages/jobs/JobsList"));
const JobForm = lazy(() => import("@/pages/jobs/JobForm"));
const CertificationsList = lazy(() => import("@/pages/certifications/CertificationsList"));
const CertificationForm = lazy(() => import("@/pages/certifications/CertificationForm"));
const SkillsList = lazy(() => import("@/pages/skills/SkillsList"));
const SkillForm = lazy(() => import("@/pages/skills/SkillForm"));
const ExportPage = lazy(() => import("@/pages/export/ExportPage"));
const Settings = lazy(() => import("@/pages/Settings"));
const AdminUsers = lazy(() => import("@/pages/admin/AdminUsers"));
const AdminSettings = lazy(() => import("@/pages/admin/AdminSettings"));
const AdminBadges = lazy(() => import("@/pages/admin/AdminBadges"));
const Achievements = lazy(() => import("@/pages/gamification/Achievements"));
const ProfileSettings = lazy(() => import("@/pages/profile/ProfileSettings"));
const PublicProfile = lazy(() => import("@/pages/profile/PublicProfile"));
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
    path: "/auth/callback",
    element: (
      <SuspenseWrapper>
        <Callback />
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
            path: "profile",
            element: (
              <SuspenseWrapper>
                <ProfileSettings />
              </SuspenseWrapper>
            ),
          },
          {
            path: "settings",
            element: (
              <SuspenseWrapper>
                <Settings />
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
