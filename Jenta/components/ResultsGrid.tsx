
import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { PhotoshootResult, ShotType } from '../types';

const LoadingSpinner = () => (
    <div className="flex justify-center items-center h-full">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
    </div>
);

const ImageResult: React.FC<{ 
    result: PhotoshootResult;
    onEdit?: (prompt: string) => void;
}> = ({ result, onEdit }) => {
    const [isFullViewOpen, setIsFullViewOpen] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const [localPrompt, setLocalPrompt] = useState('');

    if (!result) return null;
    const shotType = result.shotType;

    const handleDownload = (resolution: '2k' | '4k') => {
        if (!result.image) return;
        setIsDownloading(true);
        const img = new Image();
        img.src = `data:${result.image.mimeType};base64,${result.image.base64}`;
        img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (!ctx) { setIsDownloading(false); return; }
            const targetWidth = resolution === '4k' ? 4096 : 2048;
            canvas.width = targetWidth;
            canvas.height = targetWidth / (img.width / img.height);
            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            const link = document.createElement('a');
            link.download = `SMART-Studio-${shotType.replace(/\s+/g, '-')}-${resolution}.png`;
            link.href = canvas.toDataURL('image/png');
            link.click();
            setIsDownloading(false);
        };
        img.onerror = () => setIsDownloading(false);
    };

    const handleApplyEdit = () => {
        if (localPrompt.trim() && onEdit) {
            onEdit(localPrompt);
            setLocalPrompt('');
        }
    };

    return (
        <div className="flex flex-col gap-2 group/card">
            <h4 className="text-[11px] font-bold text-center text-[var(--color-text-muted)] uppercase tracking-widest truncate">{shotType}</h4>
            
            <div className="relative w-full aspect-square bg-black/20 backdrop-blur-sm rounded-2xl overflow-hidden border border-white/5 group-hover/card:border-[var(--color-accent)]/30 transition-all">
                {result.isLoading || result.isEditing ? (
                    <div className="flex flex-col items-center justify-center h-full gap-3 bg-black/20">
                        <LoadingSpinner />
                        <span className="text-[9px] font-bold text-white/20 uppercase tracking-widest animate-pulse">
                            {result.isEditing ? 'Refining Shot...' : 'Capturing...'}
                        </span>
                    </div>
                ) : result.image ? (
                    <img 
                        src={`data:${result.image.mimeType};base64,${result.image.base64}`} 
                        alt={shotType} 
                        className="object-contain w-full h-full cursor-pointer"
                        onClick={() => setIsFullViewOpen(true)}
                    />
                ) : result.error ? (
                    <div className="p-4 text-center">
                        <p className="text-[10px] text-red-300">{result.error}</p>
                    </div>
                ) : null}
            </div>

            {result.image && !result.isLoading && !result.isEditing && (
                <div className="flex flex-col gap-2 mt-1">
                    {/* Edit Section */}
                    <div className="px-1 relative">
                        <input 
                            type="text"
                            value={localPrompt}
                            onChange={(e) => setLocalPrompt(e.target.value)}
                            placeholder="Edit shot..."
                            className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-[10px] text-white focus:outline-none focus:border-[var(--color-accent)]/50 transition-all"
                            onKeyDown={(e) => e.key === 'Enter' && handleApplyEdit()}
                        />
                        <button 
                            onClick={handleApplyEdit}
                            disabled={!localPrompt.trim()}
                            className="absolute right-2 top-1 bottom-1 px-3 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white text-[8px] font-black rounded-lg transition-all disabled:opacity-0"
                        >
                            APPLY
                        </button>
                    </div>

                    <div className="flex gap-1.5 px-1">
                        <button onClick={() => handleDownload('2k')} disabled={isDownloading} className="flex-1 text-[9px] font-bold bg-white/5 hover:bg-white/10 py-2 rounded-lg border border-white/5 transition-colors uppercase">2K</button>
                        <button onClick={() => handleDownload('4k')} disabled={isDownloading} className="flex-1 text-[9px] font-bold bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] py-2 rounded-lg transition-colors uppercase text-white">4K</button>
                    </div>
                </div>
            )}

            {isFullViewOpen && result.image && createPortal(
                <div className="fixed inset-0 bg-black/95 backdrop-blur-md flex items-center justify-center z-[9999] p-4 animate-in fade-in cursor-zoom-out" onClick={() => setIsFullViewOpen(false)}>
                    <img src={`data:${result.image.mimeType};base64,${result.image.base64}`} className="max-w-full max-h-full object-contain shadow-2xl" />
                </div>,
                document.body
            )}
        </div>
    );
};

const ResultsGrid: React.FC<{
  results: PhotoshootResult[];
  onEditResult?: (index: number, prompt: string) => void;
}> = ({ results, onEditResult }) => {
  const safeResults = Array.isArray(results) ? results.filter(r => !!r) : [];
  if (safeResults.length === 0) return null;
  
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      {safeResults.map((result, idx) => (
        <ImageResult 
            key={`${result.shotType}-${idx}`} 
            result={result} 
            onEdit={(p) => onEditResult?.(idx, p)}
        />
      ))}
    </div>
  );
};

export default ResultsGrid;
