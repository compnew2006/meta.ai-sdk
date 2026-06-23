
import React, { useState, useRef, useEffect } from 'react';
import { CustomizationOptions } from '../types';
import { LIGHTING_STYLES, CAMERA_PERSPECTIVES } from '../constants';

interface CustomizationPanelProps {
  options: CustomizationOptions;
  setOptions: (options: CustomizationOptions) => void;
}

interface CustomSelectProps {
  label: string;
  icon: React.ReactNode;
  value: string;
  options: { value: string; label: string }[];
  onChange: (value: string) => void;
}

const CustomSelect: React.FC<CustomSelectProps> = ({ label, icon, value, options, onChange }) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const selectedLabel = options.find(opt => opt.value === value)?.label || value;

  return (
    <div className="flex-1 flex flex-col gap-2 relative" ref={containerRef}>
       <label 
            className="text-xs font-bold text-[var(--color-text-medium)] uppercase tracking-wider ml-1 flex items-center gap-1.5 whitespace-nowrap cursor-pointer"
            onClick={() => setIsOpen(!isOpen)}
        >
            {icon}
            {label}
        </label>
        <div className="relative w-full">
            <button
                onClick={() => setIsOpen(!isOpen)}
                className={`w-full text-left appearance-none bg-black/20 border text-[var(--color-text-base)] rounded-xl px-3 py-2.5 pr-8 outline-none transition-all text-sm font-medium flex items-center justify-between
                ${isOpen 
                    ? 'border-[var(--color-accent)] ring-2 ring-[var(--color-accent)] bg-[rgba(var(--color-accent-rgb),0.05)]' 
                    : 'border-[rgba(var(--color-text-base-rgb),0.1)] hover:border-[var(--color-accent)]'
                }`}
            >
                <span className="truncate">{selectedLabel}</span>
                <svg xmlns="http://www.w3.org/2000/svg" className={`h-4 w-4 text-[var(--color-accent)] transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                </svg>
            </button>

            {isOpen && (
                <div className="absolute top-full left-0 mt-2 w-full bg-black/90 backdrop-blur-md border border-[rgba(var(--color-text-base-rgb),0.1)] rounded-xl shadow-2xl z-50 max-h-60 overflow-y-auto suggestions-scrollbar animate-in fade-in zoom-in-95 duration-100 origin-top">
                    {options.map((opt) => (
                        <div
                            key={opt.value}
                            onClick={() => {
                                onChange(opt.value);
                                setIsOpen(false);
                            }}
                            className={`px-4 py-2.5 text-sm cursor-pointer transition-colors border-b border-white/5 last:border-0
                                ${opt.value === value 
                                    ? 'text-[var(--color-accent)] font-bold bg-[rgba(var(--color-accent-rgb),0.1)]' 
                                    : 'text-[var(--color-text-base)] hover:bg-[var(--color-accent)] hover:text-white'
                                }`}
                        >
                            {opt.label}
                        </div>
                    ))}
                </div>
            )}
        </div>
    </div>
  );
};

const CustomizationPanel: React.FC<CustomizationPanelProps> = ({ options, setOptions }) => {
  return (
    <div className="glass-card rounded-2xl p-4 relative z-20">
      <div className="flex flex-row gap-3">
        
        {/* Lighting Style Section */}
        <CustomSelect
            label="Lighting"
            icon={(
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-[var(--color-accent)]" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.706-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 14.95l.707-.707a1 1 0 10-1.414-1.414l-.707.707a1 1 0 001.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 100 2h1z" clipRule="evenodd" />
                </svg>
            )}
            value={options.lightingStyle}
            options={LIGHTING_STYLES}
            onChange={(val) => setOptions({ ...options, lightingStyle: val as any })}
        />

        {/* Camera Perspective Section */}
        <CustomSelect
            label="Angle"
            icon={(
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-[var(--color-accent)]" viewBox="0 0 20 20" fill="currentColor">
                     <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
            )}
            value={options.cameraPerspective}
            options={CAMERA_PERSPECTIVES}
            onChange={(val) => setOptions({ ...options, cameraPerspective: val as any })}
        />

      </div>
    </div>
  );
};

export default CustomizationPanel;
