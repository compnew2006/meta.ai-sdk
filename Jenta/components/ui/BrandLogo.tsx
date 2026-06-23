// components/ui/BrandLogo.tsx
//
// Single source of truth for the SMART Studio brand mark. Used by:
//   - App.tsx hero (size="lg") and top-nav (size="sm")
//   - Plan/Marketing print-window reports (BrandLogoDataUri — a base64 SVG so
//     the printed PDF shows the brand with no external request and no author
//     attribution).
//
// Replaces the 3 previously-hardcoded `LOGO_IMAGE_URL = "https://i.ibb.co/..."`
// constants which pointed to the original author's personal image host.

import React from 'react';

export type BrandLogoSize = 'sm' | 'md' | 'lg';

const SIZE_CLASSES: Record<BrandLogoSize, string> = {
  sm: 'h-10 w-auto',
  md: 'w-48 h-48 md:w-64 md:h-64',
  lg: 'w-64 h-64 md:w-96 md:h-96',
};

/**
 * Inline SVG brand mark. Identical paths/gradients to the former hero SVG, now
 * parameterised by size so the nav and hero share one source.
 *
 * NOTE: gradient/filter ids are globally unique (`smart-glass`, `smart-shine`,
 * `smart-core`, `smart-blur`). If multiple BrandLogo instances render in the
 * same document (e.g. hero + nav both on the landing page), the duplicate ids
 * still resolve correctly because they reference identical defs — browsers
 * pick the first match, and all matches are identical.
 */
export const BrandLogo: React.FC<{ size?: BrandLogoSize; className?: string }> = ({
  size = 'md',
  className = '',
}) => (
  <div role="img" aria-label="SMART Studio" className={`object-contain ${SIZE_CLASSES[size]} ${className}`}>
    <svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%">
      <defs>
        <linearGradient id="smart-glass" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#ff8080" />
          <stop offset="45%" stopColor="#ff0000" />
          <stop offset="100%" stopColor="#c00000" />
        </linearGradient>
        <linearGradient id="smart-shine" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#ffffff" stopOpacity="0.5" />
          <stop offset="55%" stopColor="#ffffff" stopOpacity="0" />
        </linearGradient>
        <radialGradient id="smart-core" cx="50%" cy="38%" r="60%">
          <stop offset="0%" stopColor="#ffffff" stopOpacity="0.28" />
          <stop offset="100%" stopColor="#ffffff" stopOpacity="0" />
        </radialGradient>
        <filter id="smart-blur" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="14" />
        </filter>
      </defs>
      <ellipse cx="256" cy="468" rx="150" ry="22" fill="#ff0000" opacity="0.28" filter="url(#smart-blur)" />
      <rect x="96" y="96" width="320" height="320" rx="84" fill="url(#smart-glass)" />
      <rect x="96" y="96" width="320" height="320" rx="84" fill="url(#smart-core)" />
      <rect x="96" y="96" width="320" height="160" rx="84" fill="url(#smart-shine)" />
      <rect x="96" y="96" width="320" height="320" rx="84" fill="none" stroke="#ffffff" strokeOpacity="0.4" strokeWidth="3" />
      <text x="256" y="300" fontFamily="'Tajawal','Segoe UI',system-ui,sans-serif" fontSize="150" fontWeight="800" fill="#ffffff" textAnchor="middle" letterSpacing="-6">S</text>
      <text x="256" y="372" fontFamily="'Tajawal','Segoe UI',system-ui,sans-serif" fontSize="46" fontWeight="700" fill="#ffffff" fillOpacity="0.95" textAnchor="middle" letterSpacing="4">SMART</text>
    </svg>
  </div>
);

/**
 * The same SVG as a `data:image/svg+xml;base64,…` URI for embedding inside
 * generated print-window HTML (Plan/Marketing reports). Lets the printed PDF
 * show the brand with zero external image requests.
 *
 * Computed once at module load. The SVG is the minimal <svg>…</svg> string
 * with no whitespace noise; base64 keeps Arabic/UTF-8 safe inside HTML attrs.
 */
const SVG_SOURCE = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><defs><linearGradient id="sg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="#ff8080"/><stop offset="45%" stop-color="#ff0000"/><stop offset="100%" stop-color="#c00000"/></linearGradient><radialGradient id="sc" cx="50%" cy="38%" r="60%"><stop offset="0%" stop-color="#ffffff" stop-opacity="0.28"/><stop offset="100%" stop-color="#ffffff" stop-opacity="0"/></radialGradient></defs><rect x="96" y="96" width="320" height="320" rx="84" fill="url(#sg)"/><rect x="96" y="96" width="320" height="320" rx="84" fill="url(#sc)"/><rect x="96" y="96" width="320" height="320" rx="84" fill="none" stroke="#ffffff" stroke-opacity="0.4" stroke-width="3"/><text x="256" y="300" font-family="sans-serif" font-size="150" font-weight="800" fill="#ffffff" text-anchor="middle" letter-spacing="-6">S</text><text x="256" y="372" font-family="sans-serif" font-size="46" font-weight="700" fill="#ffffff" fill-opacity="0.95" text-anchor="middle" letter-spacing="4">SMART</text></svg>`;

export const BrandLogoDataUri: string =
  'data:image/svg+xml;base64,' +
  typeof window !== 'undefined'
    ? window.btoa(SVG_SOURCE)
    : Buffer.from(SVG_SOURCE).toString('base64');
