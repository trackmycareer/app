import { Component } from "react";
import type { ErrorInfo, ReactNode } from "react";
import { AlertTriangleIcon } from "@/components/icons";
import { Button } from "@/components/Button";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("ErrorBoundary caught:", error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex min-h-[400px] flex-col items-center justify-center gap-4 p-8">
          <AlertTriangleIcon
            width={48}
            height={48}
            className="text-[var(--color-warning)]"
          />
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            Something went wrong
          </h2>
          <p className="max-w-md text-center text-sm text-[var(--text-secondary)]">
            {this.state.error?.message || "An unexpected error occurred."}
          </p>
          <Button variant="secondary" onClick={this.handleReset}>
            Try again
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}
