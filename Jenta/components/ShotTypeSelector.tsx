import React from 'react';
import { SHOT_TYPES, MAX_SHOT_SELECTION } from '../constants';
import { ShotType } from '../types';

interface ShotTypeSelectorProps {
  selected: ShotType[];
  onChange: (selected: ShotType[]) => void;
  customStylePrompt: string;
  onCustomStylePromptChange: (prompt: string) => void;
}

const ShotTypeButton: React.FC<{
  type: ShotType,
  isSelected: boolean,
  onClick: () => void,
}> = ({ type, isSelected, onClick }) => {
  return (
    <button
      onClick={onClick}
      className={`w-full text-center text-sm px-2 py-2.5 rounded-lg transition-all duration-200 border ${
        isSelected
          ? 'bg-[var(--color-accent)] text-white border-[var(--color-accent)] shadow-md shadow-[var(--color-accent)]/20'
          : 'bg-black/20 border-transparent hover:border-[var(--color-accent)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-base)]'
      }`}
    >
      {type}
    </button>
  );
};


const ShotTypeSelector: React.FC<ShotTypeSelectorProps> = ({ selected, onChange, customStylePrompt, onCustomStylePromptChange }) => {

  const handleToggle = (type: ShotType) => {
    const isSelected = selected.includes(type);
    if (isSelected) {
      onChange(selected.filter(t => t !== type));
    } else {
      if (selected.length < MAX_SHOT_SELECTION) {
        onChange([...selected, type]);
      }
    }
  };

  return (
    <div className="glass-card rounded-2xl p-4">
      <h3 className="text-lg font-bold text-[var(--color-text-base)] mb-1">2. Select Shot Types</h3>
      <p className="text-sm text-[var(--color-text-secondary)] mb-3">
        Choose up to {MAX_SHOT_SELECTION} types. ({selected.length}/{MAX_SHOT_SELECTION})
      </p>

      <div className="flex flex-col gap-4">
        {SHOT_TYPES.map(section => (
            <div key={section.category}>
                <h4 className="text-sm font-semibold text-[var(--color-text-medium)] mb-2">{section.category}</h4>
                <div className="grid grid-cols-2 gap-2">
                    {section.types.map(type => (
                        <ShotTypeButton
                            key={type}
                            type={type}
                            isSelected={selected.includes(type)}
                            onClick={() => handleToggle(type)}
                        />
                    ))}
                </div>
            </div>
        ))}
      </div>

      <div className="mt-4 pt-4 border-t border-[rgba(var(--color-text-base-rgb),0.1)]">
        <h4 className="text-sm font-semibold text-[var(--color-text-medium)] mb-2">Custom Style Prompt (Optional)</h4>
        <textarea
            value={customStylePrompt}
            onChange={(e) => onCustomStylePromptChange(e.target.value)}
            rows={3}
            className="w-full glass-input rounded-md p-3 text-sm leading-relaxed transition-all"
            placeholder="e.g., 'Minimalist aesthetic, with a soft pastel color palette and dramatic long shadows.'"
        />
      </div>
    </div>
  );
};

export default ShotTypeSelector;