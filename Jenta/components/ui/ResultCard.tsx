// components/ui/ResultCard.tsx
//
// The single result-image viewer. Replaces the duplicate edit + 2K/4K download
// + fullscreen-portal pattern in:
//   - ResultDisplay.tsx (single image)
//   - ResultsGrid.tsx (photoshoot card)
//   - BrandingResultsGrid.tsx (branding/campaign card)
//
// Scope: ONE card. Callers compose grids of <ResultCard> themselves — this
// component does NOT own the grid wrapper (so it stays free of per-caller
// layout branches per Sandi Metz's "wrong abstraction" rule).
//
// Visually identical to ResultDisplay — same classes, same overlay hints,
// same fullscreen behaviour. Adoption is incremental: existing wrappers can
// delegate to ResultCard without changing their public props.

import React, { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { ImageFile } from '../../types';
import { useImageDownload } from '../../utils/imageDownload';
import Spinner from './Spinner';
import { DownloadIcon, XIcon } from './icons';

export interface ResultCardProps {
  /** The generated image. null shows the empty state. */
  image: ImageFile | null;
  /** When true, shows the loading overlay. */
  isLoading?: boolean;
  /** When true, shows the "Refining..." overlay instead of "Generating...". */
  isEditing?: boolean;
  /** Label shown in the download filenames. */
  fileNamePrefix?: string;
  /** Optional edit handler. When provided, shows the refine-input row. */
  onEdit?: (prompt: string) => void;
  /** Custom label for the loading overlay; defaults to "Generating Design...". */
  loadingLabel?: string;
  /** Aspect of the image frame. Defaults to "aspect-square". */
  frameClassName?: string;
  /** Show the 2K/4K download buttons. Default true. */
  showDownloads?: boolean;
}

const ResultCard: React.FC<ResultCardProps> = ({
  image,
  isLoading = false,
  isEditing = false,
  fileNamePrefix = 'SMART-Studio',
  onEdit,
  loadingLabel,
  frameClassName = 'aspect-square',
  showDownloads = true,
}) => {
  const { download, isDownloading } = useImageDownload(fileNamePrefix);
  const [isFullViewOpen, setIsFullViewOpen] = useState(false);
  const [localPrompt, setLocalPrompt] = useState('');

  // ESC-to-close fullscreen + lock body scroll while open.
  useEffect(() => {
    if (!isFullViewOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsFullViewOpen(false);
    };
    window.addEventListener('keydown', handleKeyDown);
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'auto';
    };
  }, [isFullViewOpen]);

  const handleApplyEdit = () => {
    if (localPrompt.trim() && onEdit) {
      onEdit(localPrompt);
      setLocalPrompt('');
    }
  };

  return (
    <div className="w-full flex flex-col gap-4">
      {/* Image frame */}
      <div
        className={`w-full ${frameClassName} bg-black/10 backdrop-blur-sm rounded-3xl flex items-center justify-center overflow-hidden relative border border-white/5 shadow-inner group`}
      >
        {(isLoading || isEditing) && (
          <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-black/40 backdrop-blur-sm">
            <Spinner />
            <p className="mt-4 text-[10px] font-bold text-[var(--color-accent)] uppercase tracking-[0.2em] animate-pulse">
              {loadingLabel ?? (isEditing ? 'Refining Design...' : 'Generating Design...')}
            </p>
          </div>
        )}

        {!isLoading && !isEditing && !image && (
          <div className="flex flex-col items-center gap-3 opacity-20">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span className="text-xs font-bold uppercase tracking-widest">Result will appear here</span>
          </div>
        )}

        {!isLoading && !isEditing && image && (
          <div className="w-full h-full relative">
            <img
              src={`data:${image.mimeType};base64,${image.base64}`}
              alt="Generated Design"
              className="object-contain w-full h-full cursor-pointer transition-transform duration-700 group-hover:scale-[1.02]"
              onClick={() => setIsFullViewOpen(true)}
            />
            <div className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none flex items-center justify-center">
              <span className="bg-black/60 backdrop-blur-md px-4 py-2 rounded-full border border-white/10 text-[10px] font-bold uppercase tracking-widest text-white">
                Click to Enlarge
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Edit + downloads — only when an image is shown and not loading. */}
      {image && !isLoading && !isEditing && (
        <div className="animate-in slide-in-from-bottom-2 duration-500 space-y-3">
          {onEdit && (
            <div className="relative">
              <input
                type="text"
                value={localPrompt}
                onChange={(e) => setLocalPrompt(e.target.value)}
                placeholder="Describe changes to refine result..."
                className="w-full bg-white/5 border border-white/10 rounded-2xl px-5 py-4 text-xs text-white focus:outline-none focus:border-[var(--color-accent)]/50 transition-all placeholder:text-white/20"
                onKeyDown={(e) => e.key === 'Enter' && handleApplyEdit()}
              />
              <button
                onClick={handleApplyEdit}
                disabled={!localPrompt.trim()}
                className="absolute right-2 top-2 bottom-2 px-6 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white text-[10px] font-black rounded-xl transition-all disabled:opacity-0 shadow-lg shadow-[var(--color-accent)]/20"
              >
                EDIT
              </button>
            </div>
          )}

          {showDownloads && (
            <div className="flex gap-4">
              <button
                onClick={() => download(image, '2k')}
                disabled={isDownloading}
                className="flex-1 inline-flex items-center justify-center bg-white/5 hover:bg-white/10 border border-white/5 text-white font-bold py-4 px-4 rounded-2xl transition-all disabled:opacity-50 text-xs tracking-tighter"
              >
                <DownloadIcon className="h-5 w-5 mr-2" />
                DOWNLOAD 2K
              </button>
              <button
                onClick={() => download(image, '4k')}
                disabled={isDownloading}
                className="flex-1 inline-flex items-center justify-center bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white font-bold py-4 px-4 rounded-2xl transition-all disabled:opacity-50 text-xs tracking-tighter shadow-lg shadow-[var(--color-accent)]/20"
              >
                <DownloadIcon className="h-5 w-5 mr-2" />
                DOWNLOAD 4K
              </button>
            </div>
          )}
        </div>
      )}

      {/* Fullscreen preview portal */}
      {isFullViewOpen && image && createPortal(
        <div
          className="fixed inset-0 bg-black/95 backdrop-blur-xl flex items-center justify-center z-[99999] p-4 animate-in fade-in duration-300 cursor-zoom-out"
          onClick={() => setIsFullViewOpen(false)}
        >
          <img
            src={`data:${image.mimeType};base64,${image.base64}`}
            alt="Full view result"
            className="max-w-full max-h-full object-contain shadow-2xl rounded-2xl"
          />
          <button
            onClick={() => setIsFullViewOpen(false)}
            className="absolute top-6 right-6 text-white/50 hover:text-white transition-all p-3 bg-white/5 rounded-full hover:bg-white/10 backdrop-blur-md"
            aria-label="Close fullscreen"
          >
            <XIcon className="h-8 w-8" />
          </button>
        </div>,
        document.body,
      )}
    </div>
  );
};

export default ResultCard;
