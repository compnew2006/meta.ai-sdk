
import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { BrandingResult, ImageFile } from '../types';

const LoadingSpinner = () => (
    <div className="flex justify-center items-center h-full">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
    </div>
);

const ImageResultCard: React.FC<{ 
    result: BrandingResult;
    onEdit?: (prompt: string) => void;
}> = ({ result, onEdit }) => {
    const [isFullViewOpen, setIsFullViewOpen] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const [localPrompt, setLocalPrompt] = useState('');

    if (!result) return null;
    const id = String(result.category || 'Asset');

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
            link.download = `SMART-Studio-${id.replace(/\s+/g, '-')}-${resolution}.png`;
            link.href = canvas.toDataURL('image/png');
            link.click();
            setIsDownloading(false);
        };
        img.onerror = () => setIsDownloading(false);
    };

    const handleApplyEdit = () => {
        if (localPrompt.trim() && onEdit) {
            onEdit(localPrompt);
            setLocalPrompt(''); // Clear after applying
        }
    };

    return (
        <div className="flex flex-col gap-3 group/card animate-in fade-in duration-500">
            <h4 className="text-[10px] font-black text-center text-white/40 uppercase tracking-[0.2em] px-2 truncate">{id}</h4>
            
            <div className="relative w-full aspect-square bg-black/30 backdrop-blur-md rounded-3xl overflow-hidden border border-white/5 group-hover/card:border-[var(--color-accent)]/30 transition-all shadow-xl">
                {(result.isLoading || result.isEditing) ? (
                    <div className="flex flex-col items-center justify-center h-full gap-4 bg-black/40 backdrop-blur-sm">
                        <LoadingSpinner />
                        <span className="text-[9px] font-bold text-white/40 uppercase tracking-widest animate-pulse">
                            {result.isEditing ? 'Refining Design...' : 'Generating...'}
                        </span>
                    </div>
                ) : result.image ? (
                    <img 
                        src={`data:${result.image.mimeType};base64,${result.image.base64}`} 
                        alt={id} 
                        className="object-contain w-full h-full cursor-pointer transition-transform duration-700 group-hover/card:scale-105"
                        onClick={() => setIsFullViewOpen(true)}
                    />
                ) : result.error ? (
                    <div className="p-6 text-center h-full flex items-center justify-center">
                        <p className="text-[10px] text-red-400 font-medium leading-relaxed">{result.error}</p>
                    </div>
                ) : null}
            </div>

            {result.image && !result.isLoading && !result.isEditing && (
                <div className="flex flex-col gap-2">
                    {/* Edit Section - NOW ABOVE BUTTONS */}
                    <div className="px-1 flex flex-col gap-2">
                        <div className="relative group/input">
                            <input 
                                type="text"
                                value={localPrompt}
                                onChange={(e) => setLocalPrompt(e.target.value)}
                                placeholder="Describe changes..."
                                className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2.5 text-[11px] text-white focus:outline-none focus:border-[var(--color-accent)]/50 focus:bg-white/10 transition-all placeholder:text-white/20"
                                onKeyDown={(e) => e.key === 'Enter' && handleApplyEdit()}
                            />
                            <button 
                                onClick={handleApplyEdit}
                                disabled={!localPrompt.trim()}
                                className="absolute right-1 top-1 bottom-1 px-3 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white text-[9px] font-black rounded-lg transition-all disabled:opacity-0 disabled:scale-90 shadow-lg shadow-[var(--color-accent)]/20"
                            >
                                EDIT
                            </button>
                        </div>
                    </div>

                    <div className="flex gap-2 px-1">
                        <button onClick={() => handleDownload('2k')} disabled={isDownloading} className="flex-1 text-[10px] font-black bg-white/5 hover:bg-white/10 py-2.5 rounded-xl border border-white/5 transition-all uppercase tracking-tighter text-white/70">2K</button>
                        <button onClick={() => handleDownload('4k')} disabled={isDownloading} className="flex-1 text-[10px] font-black bg-white/10 hover:bg-[var(--color-accent)] py-2.5 rounded-xl border border-white/5 transition-all uppercase tracking-tighter text-white shadow-sm">4K</button>
                    </div>
                </div>
            )}

            {isFullViewOpen && result.image && createPortal(
                <div className="fixed inset-0 bg-black/95 backdrop-blur-xl flex items-center justify-center z-[99999] p-4 animate-in fade-in cursor-zoom-out" onClick={() => setIsFullViewOpen(false)}>
                    <img src={`data:${result.image.mimeType};base64,${result.image.base64}`} className="max-w-full max-h-full object-contain shadow-2xl rounded-2xl" />
                </div>,
                document.body
            )}
        </div>
    );
};

const BrandingResultsGrid: React.FC<{
  results: BrandingResult[];
  gridClassName?: string;
  onEditResult?: (index: number, prompt: string) => void;
}> = ({ results, gridClassName, onEditResult }) => {
  const safeResults = Array.isArray(results) ? results.filter(r => !!r) : [];
  if (safeResults.length === 0) return null;
  
  return (
    <div className={gridClassName || "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6"}>
      {safeResults.map((result, idx) => (
        <ImageResultCard 
            key={result && result.category ? `${result.category}-${idx}` : idx} 
            result={result} 
            onEdit={(prompt) => onEditResult?.(idx, prompt)}
        />
      ))}
    </div>
  );
};

export default BrandingResultsGrid;
