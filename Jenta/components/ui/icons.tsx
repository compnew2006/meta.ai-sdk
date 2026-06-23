// components/ui/icons.tsx
//
// Single source of truth for inline SVG icons. Each icon is an arrow-function
// component accepting standard SVG props (className, onClick, etc.). The path
// data is taken from the most-used existing copy in the codebase so adoption
// is visually identical, not a redesign.
//
// Replaces ~50 duplicate *Icon definitions across 14 files. Adoption is
// incremental — files keep working until they swap their local definition for
// an import from here.

import React from 'react';

type IconProps = React.SVGProps<SVGSVGElement> & { size?: number | string };

const baseSvgProps = (size: number | string = '1em'): React.SVGProps<SVGSVGElement> => ({
  xmlns: 'http://www.w3.org/2000/svg',
  fill: 'none',
  viewBox: '0 0 24 24',
  stroke: 'currentColor',
  strokeWidth: 2,
  style: { width: size, height: size, ...({} as React.CSSProperties) },
});

export const DownloadIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
  </svg>
);

export const CopyIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
  </svg>
);

export const CheckIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
  </svg>
);

export const TrashIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
  </svg>
);

export const PlusIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
  </svg>
);

export const XIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
  </svg>
);

export const CropIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M7 3v12a2 2 0 002 2h12M3 7h12a2 2 0 012 2v12" />
  </svg>
);

export const MagicIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
  </svg>
);

export const ArrowRightIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3" />
  </svg>
);

export const ChevronLeftIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
  </svg>
);

export const ChevronRightIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
  </svg>
);

export const SparkleIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
  </svg>
);

export const GlobeIcon: React.FC<IconProps> = ({ size, ...rest }) => (
  <svg {...baseSvgProps(size)} {...rest}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z M3.6 9h16.8 M3.6 15h16.8 M12 3a15 15 0 010 18 M12 3a15 15 0 000 18" />
  </svg>
);
