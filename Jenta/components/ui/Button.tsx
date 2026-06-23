// components/ui/Button.tsx
//
// The single button primitive. Variants map 1:1 to the className signatures
// already used inline across the studios, so adopting it is visually identical
// (not a redesign). Replaces:
//   - primary CTA: bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] ... (~15x)
//   - ghost/secondary: bg-white/5 hover:bg-white/10 ... (~10x)
//   - danger: bg-red-500/80 hover:bg-red-600 ... (~3x)
//   - success: bg-emerald-600 ... (~2x)

import React from 'react';
import Spinner from './Spinner';

export type ButtonVariant = 'primary' | 'ghost' | 'danger' | 'success';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** When true, disables the button and renders an inline spinner. */
  loading?: boolean;
  /** Stretch to fill the parent's width (w-full). */
  fullWidth?: boolean;
}

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary:
    'bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white shadow-lg shadow-[var(--color-accent)]/20',
  ghost:
    'bg-white/5 hover:bg-white/10 border border-white/5 text-[var(--color-text-base)]',
  danger:
    'bg-[var(--color-error)] hover:opacity-90 text-white',
  success:
    'bg-[var(--color-success)] hover:opacity-90 text-white shadow-lg',
};

const SIZE_CLASSES: Record<ButtonSize, string> = {
  sm: 'px-3 py-1.5 text-xs rounded-lg',
  md: 'px-4 py-2.5 text-sm rounded-xl',
  lg: 'px-8 py-3 text-lg rounded-2xl',
};

const Button: React.FC<ButtonProps> = ({
  variant = 'primary',
  size = 'md',
  loading = false,
  fullWidth = false,
  disabled,
  className = '',
  children,
  ...rest
}) => {
  const isDisabled = disabled || loading;
  return (
    <button
      {...rest}
      disabled={isDisabled}
      className={`font-bold transition-all transform active:scale-95 disabled:opacity-50 disabled:transform-none inline-flex items-center justify-center gap-2 ${
        VARIANT_CLASSES[variant]
      } ${SIZE_CLASSES[size]} ${fullWidth ? 'w-full' : ''} ${className}`}
    >
      {loading && <Spinner size="h-4 w-4" containerClassName="inline-flex" />}
      {children}
    </button>
  );
};

export default Button;
