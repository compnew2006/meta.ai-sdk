
import React, { useState, useCallback } from 'react';
import { PromptStudioProject, ImageFile, PromptStudioHistoryItem } from '../types';
import { analyzeImageForPrompt, generatePromptFromText } from '../services/geminiService';
import { resizeImage } from '../utils';
import ImageWorkspace from './ImageWorkspace';

// --- ICONS ---
const CopyIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
);

const CheckIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
);
// --- END ICONS ---


const PromptStudio: React.FC<{
  project: PromptStudioProject;
  setProject: React.Dispatch<React.SetStateAction<PromptStudioProject>>;
}> = ({ project, setProject }) => {
    const [copied, setCopied] = useState(false);

    const TEXT_ONLY_PLACEHOLDER_SVG = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 24 24" fill="none" stroke="#9ca3af" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18h6"></path><path d="M10 22h4"></path><path d="M12 2a7 7 0 0 0-7 7c0 3 2 5 2 7h10c0-2 2-4 2-7a7 7 0 0 0-7-7Z"></path></svg>`;
    const placeholderImageFile: ImageFile = {
        base64: btoa(TEXT_ONLY_PLACEHOLDER_SVG),
        mimeType: 'image/svg+xml',
        name: 'text-idea.svg'
    };

    const handleFileUpload = async (files: File[]) => {
      if (!files || files.length === 0) return;
      setProject(s => ({ ...s, isUploading: true, error: null }));
      
      try {
          const uploaded = await Promise.all(files.map(async file => {
              const resizedFile = await resizeImage(file, 1024, 1024);
              const reader = new FileReader();
              return new Promise<ImageFile>(res => {
                  reader.onloadend = () => res({ base64: (reader.result as string).split(',')[1], mimeType: resizedFile.type, name: resizedFile.name });
                  reader.readAsDataURL(resizedFile);
              });
          }));
          setProject(s => ({ ...s, images: [...s.images, ...uploaded], isUploading: false }));
      } catch (err) {
          setProject(s => ({ ...s, error: "Upload failed", isUploading: false }));
      }
    };

    const handleRemoveImage = (indexToRemove: number) => {
        setProject(s => ({ ...s, images: s.images.filter((_, i) => i !== indexToRemove) }));
    };

    const handleUpdateImage = (index: number, newImage: ImageFile) => {
        setProject(s => {
            const nextImages = [...s.images];
            nextImages[index] = newImage;
            return { ...s, images: nextImages };
        });
    };

    const handleGenerate = useCallback(async () => {
        if (project.images.length === 0 && !project.instructions.trim()) {
            setProject(s => ({ ...s, error: 'Please upload an image or write an idea in the instructions.' }));
            return;
        }
        setProject(s => ({ ...s, isLoading: true, error: null, generatedPrompt: null }));
        try {
            let prompt: string;
            let historyImage: ImageFile;

            if (project.images.length > 0) {
                prompt = await analyzeImageForPrompt(project.images, project.instructions);
                historyImage = project.images[0];
            } else {
                prompt = await generatePromptFromText(project.instructions);
                historyImage = placeholderImageFile;
            }

            const newHistoryItem: PromptStudioHistoryItem = {
                image: historyImage,
                instructions: project.instructions,
                generatedPrompt: prompt,
            };
            setProject(s => ({
                ...s,
                isLoading: false,
                generatedPrompt: prompt,
                history: [newHistoryItem, ...s.history],
                instructions: '', // Clear instructions after generation
            }));
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred.';
            setProject(s => ({ ...s, isLoading: false, error: errorMessage }));
        }
    }, [project.images, project.instructions, setProject]);

    const handleCopy = (text: string) => {
      navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    };

    if (!project) return null;

    return (
        <main className="w-full max-w-4xl flex flex-col gap-4 pt-4 pb-8 mx-auto">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="glass-card rounded-2xl p-4">
                    <h3 className="text-sm font-bold text-[var(--color-text-medium)] mb-3 text-center uppercase">Reference Images</h3>
                    <div className="aspect-square w-full max-w-[300px] mx-auto">
                        <ImageWorkspace
                            id="prompt-studio-workspace"
                            title="Reference Image"
                            images={project.images}
                            onImagesUpload={handleFileUpload}
                            onImageRemove={handleRemoveImage}
                            isUploading={project.isUploading}
                            onImageUpdate={handleUpdateImage}
                        />
                    </div>
                </div>

                <div className="glass-card rounded-2xl p-4 flex flex-col gap-4">
                    <div className="flex-grow w-full">
                        <label htmlFor="instructions" className="block text-sm font-medium text-[var(--color-text-medium)] mb-2">Instructions / Idea</label>
                        <textarea id="instructions" value={project.instructions} onChange={e => setProject({...project, instructions: e.target.value})} rows={6} className="w-full glass-input rounded-md p-3 text-sm" placeholder="e.g., 'A majestic lion wearing a crown in a futuristic city' or describe instructions for your uploaded images..."/>
                    </div>
                    
                    <button
                        onClick={handleGenerate}
                        disabled={project.isLoading || project.isUploading || (project.images.length === 0 && !project.instructions.trim())}
                        className="w-full bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white font-bold py-3 rounded-xl transition-all disabled:opacity-50 shadow-lg"
                    >
                        {project.isLoading ? 'Generating...' : 'Generate Prompt'}
                    </button>
                </div>
            </div>
            
            {project.error && <div className="bg-[rgba(var(--color-accent-rgb),0.2)] border border-[rgba(var(--color-accent-rgb),0.5)] text-[var(--color-accent-light)] px-4 py-3 rounded-lg text-sm" role="alert">{project.error}</div>}

            {project.generatedPrompt && (
                <div className="glass-card p-4 rounded-xl bg-black/20 animate-in fade-in zoom-in-95">
                    <div className="flex justify-between items-center mb-2">
                       <h3 className="text-md font-semibold text-[var(--color-text-base)]">Generated AI Prompt</h3>
                       <button onClick={() => handleCopy(project.generatedPrompt ?? '')} className="flex items-center text-xs px-3 py-1.5 rounded-md bg-[rgba(var(--color-accent-rgb),0.2)] text-[var(--color-accent-light)] hover:text-white transition-colors font-bold">
                           {copied ? <CheckIcon /> : <CopyIcon />}
                           {copied ? 'Copied!' : 'Copy Prompt'}
                       </button>
                    </div>
                    <p className="text-sm text-[var(--color-text-base)] whitespace-pre-wrap font-mono leading-relaxed p-4 bg-black/30 rounded-lg">{project.generatedPrompt}</p>
                </div>
            )}
            
            <div className="glass-card rounded-2xl p-4">
                <h2 className="text-xl font-bold text-[var(--color-text-base)] mb-4">Prompt History</h2>
                {project.history.length === 0 ? (
                    <p className="text-center text-sm text-[var(--color-text-secondary)] py-8">Your prompt history is empty.</p>
                ) : (
                    <div className="space-y-4 max-h-96 overflow-y-auto suggestions-scrollbar pr-2">
                        {project.history.map((item, index) => (
                            <div key={index} className="flex items-start gap-4 p-3 bg-black/20 rounded-lg group">
                               <img src={`data:${item.image.mimeType};base64,${item.image.base64}`} alt="Preview" className="w-20 h-20 rounded-md object-contain flex-shrink-0 bg-black/20"/>
                               <div className="flex-1 overflow-hidden">
                                  <p className="text-xs text-[var(--color-text-base)] font-mono line-clamp-3">{item.generatedPrompt}</p>
                                  {item.instructions && <p className="text-[10px] text-[var(--color-text-muted)] mt-2 italic">"{item.instructions}"</p>}
                               </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </main>
    );
};

export default PromptStudio;
