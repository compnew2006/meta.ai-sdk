
import React from 'react';
import { CreatorStudioProject, PhotoshootDirectorProject, PromptStudioProject, VoiceOverStudioProject, BrandingStudioProject, CampaignStudioProject, PlanStudioProject, EditStudioProject, StoryboardStudioProject, MarketingStudioProject } from '../types';

const PlusIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
    </svg>
);
const XIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
);

type ProjectUnion = 
  | CreatorStudioProject 
  | PhotoshootDirectorProject 
  | PromptStudioProject 
  | VoiceOverStudioProject 
  | BrandingStudioProject 
  | CampaignStudioProject 
  | PlanStudioProject 
  | EditStudioProject 
  | StoryboardStudioProject 
  | MarketingStudioProject;

interface TabBarProps {
  projects: ProjectUnion[];
  activeProjectIndex: number;
  onSelectTab: (index: number) => void;
  onAddTab: () => void;
  onCloseTab: (index: number) => void;
}

const TabBar: React.FC<TabBarProps> = ({ projects, activeProjectIndex, onSelectTab, onAddTab, onCloseTab }) => {
  return (
    <div className="w-full max-w-7xl flex items-center border-b border-[rgba(var(--color-text-base-rgb,229,231,206),0.1)]">
      <div className="flex items-end -mb-px overflow-x-auto suggestions-scrollbar">
        {projects.map((project, index) => (
          <div
            key={project.id}
            onClick={() => onSelectTab(index)}
            className={`flex-shrink-0 cursor-pointer flex items-center gap-2 px-4 py-2.5 border-t border-l border-r rounded-t-lg transition-all duration-200 group relative ${
              index === activeProjectIndex
                ? 'border-[var(--color-accent)] bg-[rgba(var(--color-accent-rgb),0.3)] text-[var(--color-text-base)]'
                : 'border-transparent text-[var(--color-text-secondary)] hover:bg-[rgba(var(--color-text-base-rgb,229,231,206),0.05)] hover:text-[var(--color-text-base)]'
            }`}
          >
            <span className="text-sm font-medium">{project.name}</span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onCloseTab(index);
              }}
              className="rounded-full p-1 -mr-1 text-[var(--color-text-muted)] hover:text-[var(--color-text-base)] hover:bg-[rgba(var(--color-accent-rgb),0.5)] transition-colors"
              aria-label={`Close ${project.name}`}
            >
              <XIcon />
            </button>
          </div>
        ))}
        <button
          onClick={onAddTab}
          className="ml-1 p-2.5 rounded-t-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-base)] hover:bg-[rgba(var(--color-text-base-rgb,229,231,206),0.1)] transition-colors flex-shrink-0"
          aria-label="Add new project"
        >
          <PlusIcon />
        </button>
      </div>
    </div>
  );
};

export default TabBar;
