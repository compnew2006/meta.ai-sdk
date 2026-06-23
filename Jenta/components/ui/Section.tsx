// components/ui/Section.tsx
//
// The glass-card + tiny-uppercase-label wrapper. Replaces dozens of inline
// `<div className="glass-card rounded-2xl p-5 flex flex-col gap-4 shadow-xl
// border border-white/5">` + `<h3 className="text-[10px] font-black
// text-white/40 uppercase tracking-widest">` pairs across the studios.
//
// The class strings are taken verbatim from the most-used existing copy, so
// adopting Section is visually identical to the inline pattern it replaces.

import React from 'react';

export interface SectionProps {
  /** Small uppercase label rendered above the children. Omit for an unlabeled card. */
  title?: string;
  /** Optional leading icon node for the label. */
  icon?: React.ReactNode;
  /** Padding; "md" = p-4, "lg" = p-5 (default), "xl" = p-6. */
  padding?: 'md' | 'lg' | 'xl';
  /** Corner radius; default 2xl matches the dominant existing style. */
  radius?: 'xl' | '2xl' | '3xl';
  /** Extra classes for the outer card (e.g. "min-h-[400px]"). */
  className?: string;
  /** Extra classes for the inner content wrapper. */
  bodyClassName?: string;
  children: React.ReactNode;
}

const PADDING: Record<NonNullable<SectionProps['padding']>, string> = {
  md: 'p-4',
  lg: 'p-5',
  xl: 'p-6',
};

const RADIUS: Record<NonNullable<SectionProps['radius']>, string> = {
  xl: 'rounded-xl',
  '2xl': 'rounded-2xl',
  '3xl': 'rounded-3xl',
};

const Section: React.FC<SectionProps> = ({
  title,
  icon,
  padding = 'lg',
  radius = '2xl',
  className = '',
  bodyClassName = 'flex flex-col gap-4',
  children,
}) => (
  <div className={`glass-card ${RADIUS[radius]} ${PADDING[padding]} shadow-xl border border-white/5 ${className}`}>
    {title && (
      <div className="flex items-center gap-1.5 mb-1">
        {icon && <span className="text-white/40">{icon}</span>}
        <h3 className="text-[10px] font-black text-white/40 uppercase tracking-widest">{title}</h3>
      </div>
    )}
    <div className={bodyClassName}>{children}</div>
  </div>
);

export default Section;
