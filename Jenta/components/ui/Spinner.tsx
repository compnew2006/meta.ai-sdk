// components/ui/Spinner.tsx
//
// The single loading indicator. Replaces 5+ redefined LoadingSpinner
// functions and ~10 inline `animate-spin rounded-full ... border-t-2 border-b-2
// border-[var(--color-accent)]` strings across the studios.

import React from 'react';

export interface SpinnerProps {
  /** Tailwind size classes for the spinning ring, e.g. "h-12 w-12". */
  size?: string;
  /** Container height class; defaults to filling the parent. */
  containerClassName?: string;
}

const Spinner: React.FC<SpinnerProps> = ({
  size = 'h-12 w-12',
  containerClassName = 'flex justify-center items-center h-full',
}) => (
  <div className={containerClassName}>
    <div className={`animate-spin rounded-full ${size} border-t-2 border-b-2 border-[var(--color-accent)]`} />
  </div>
);

export default Spinner;
