// components/ui/ErrorBanner.tsx
//
// The error display block. Replaces the identical red inline div repeated in
// CreatorStudio (~line 330), PlanStudio (~585), MarketingStudio (~265). Uses
// the --color-error token (added to index.html) instead of hardcoded red-500.

import React from 'react';

export interface ErrorBannerProps {
  message: string;
  /** Extra classes for placement tweaks (e.g. "mt-4"). */
  className?: string;
}

const ErrorBanner: React.FC<ErrorBannerProps> = ({ message, className = '' }) => {
  if (!message) return null;
  return (
    <div
      className={`bg-[var(--color-error-bg)] border border-[var(--color-error-border)] text-[var(--color-error)] px-4 py-3 rounded-xl text-xs font-bold text-center uppercase tracking-tight ${className}`}
      role="alert"
    >
      {message}
    </div>
  );
};

export default ErrorBanner;
