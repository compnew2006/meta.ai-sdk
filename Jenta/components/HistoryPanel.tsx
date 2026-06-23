import React, { useState } from 'react';
import { ImageFile, HistoryItem } from '../types';

interface HistoryPanelProps {
  history: HistoryItem[];
  onSelect: (image: ImageFile) => void;
}

const CopyIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
);

const CheckIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
);


const HistoryPanel: React.FC<HistoryPanelProps> = ({ history, onSelect }) => {
  const [copiedPrompt, setCopiedPrompt] = useState<string | null>(null);

  const handleCopy = (e: React.MouseEvent, prompt: string) => {
    e.stopPropagation(); // Prevent onSelect from firing when copying
    navigator.clipboard.writeText(prompt);
    setCopiedPrompt(prompt);
    setTimeout(() => setCopiedPrompt(null), 2000); // Reset after 2 seconds
  };

  return (
    <div className="glass-card rounded-2xl p-4 flex flex-col gap-3">
      <h2 className="text-xl font-bold text-[var(--color-text-base)]">History</h2>
      {history.length === 0 ? (
        <p className="text-sm text-[var(--color-text-secondary)] text-center py-8">
          Your generated images will appear here.
        </p>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {history.map((item, index) => (
            <div
              key={`${item.image.name}-${index}`}
              onClick={() => onSelect(item.image)}
              className="aspect-square rounded-lg overflow-hidden focus:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-offset-gray-800 focus-within:ring-[var(--color-accent)] group relative cursor-pointer"
              aria-label={`View generated image ${index + 1}`}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') onSelect(item.image); }}
            >
              <img
                src={`data:${item.image.mimeType};base64,${item.image.base64}`}
                alt={`Generated image ${index + 1}`}
                className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex flex-col justify-end p-2 pointer-events-none">
                 <p className="text-[var(--color-text-base)] text-xs font-medium [display:-webkit-box] [-webkit-line-clamp:3] [-webkit-box-orient:vertical] overflow-hidden">
                    {item.prompt}
                 </p>
              </div>
              <button
                  onClick={(e) => handleCopy(e, item.prompt)}
                  className="absolute top-2 right-2 z-10 p-1.5 bg-black/40 text-[var(--color-text-base)] rounded-full opacity-0 group-hover:opacity-100 hover:bg-black/60 backdrop-blur-sm transition-all"
                  aria-label="Copy prompt"
              >
                  {copiedPrompt === item.prompt ? <CheckIcon /> : <CopyIcon />}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default HistoryPanel;