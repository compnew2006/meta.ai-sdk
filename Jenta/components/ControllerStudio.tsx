
import React, { useCallback, useState } from 'react';
import { ControllerStudioProject, ImageFile, ControlCategory, ControllerSlider, HistoryItem } from '../types';
import { resizeImage } from '../utils';
import { editImage } from '../services/geminiService';
import ImageWorkspace from './ImageWorkspace';
import Button from './ui/Button';
import HistoryPanel from './HistoryPanel';

interface ControllerStudioProps {
    project: ControllerStudioProject;
    setProject: React.Dispatch<React.SetStateAction<ControllerStudioProject>>;
}

// Icons
const FaceIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
);

const HeadIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
    </svg>
);

const BodyIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M16 3a1 1 0 011 1v4a1 1 0 001 1h2a1 1 0 011 1v8a1 1 0 01-1 1H4a1 1 0 01-1-1V10a1 1 0 011-1h2a1 1 0 001-1V4a1 1 0 011-1h8z" />
    </svg>
);

const RetouchIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
    </svg>
);

const ResetIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>;
const DownloadIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>;

const ControllerStudio: React.FC<ControllerStudioProps> = ({ project, setProject }) => {
    const [isComparing, setIsComparing] = useState(false);

    const handleFileUpload = async (files: File[]) => {
        if (!files || files.length === 0) return;
        setProject(s => ({ ...s, isUploading: true, error: null }));
        
        try {
            const uploaded = await Promise.all(files.map(async file => {
                const resizedFile = await resizeImage(file, 2048, 2048);
                const reader = new FileReader();
                return new Promise<ImageFile>(res => {
                    reader.onloadend = () => res({ base64: (reader.result as string).split(',')[1], mimeType: resizedFile.type, name: resizedFile.name });
                    reader.readAsDataURL(resizedFile);
                });
            }));
            setProject(s => ({
                ...s,
                sourceImages: [...s.sourceImages, ...uploaded],
                isUploading: false,
            }));
        } catch (e) {
            setProject(s => ({ ...s, isUploading: false, error: "Failed to process images" }));
        }
    };

    const handleRemoveSource = (idx: number) => {
        setProject(s => ({ ...s, sourceImages: s.sourceImages.filter((_, i) => i !== idx), generatedImage: null, error: null }));
    };

    const handleUpdateSource = (idx: number, newImage: ImageFile) => {
        setProject(s => {
            const nextImages = [...s.sourceImages];
            nextImages[idx] = newImage;
            return { ...s, sourceImages: nextImages, generatedImage: null, error: null };
        });
    };

    const handleSliderChange = (id: string, value: number) => {
        setProject(s => ({
            ...s,
            sliders: s.sliders.map(slider => slider.id === id ? { ...slider, value } : slider)
        }));
    };

    const resetSliders = () => {
        setProject(s => ({
            ...s,
            sliders: s.sliders.map(slider => ({ ...slider, value: 0 }))
        }));
    };
    
    const applyPreset = (presetName: string) => {
        resetSliders();
        setProject(s => {
             const newSliders = [...s.sliders];
             // Helper to find and set value
             const setVal = (id: string, val: number) => {
                 const idx = newSliders.findIndex(sl => sl.id === id);
                 if (idx !== -1) newSliders[idx] = { ...newSliders[idx], value: val };
             };

             switch(presetName) {
                 case 'Happy':
                     setVal('smile', 0.8);
                     setVal('eyebrow_raise', 0.2);
                     setVal('eye_direction', 0);
                     break;
                 case 'Sad':
                     setVal('frown', 0.7);
                     setVal('eyebrow_raise', -0.3);
                     setVal('head_pitch', -0.2);
                     break;
                 case 'Surprised':
                     setVal('mouth_open', 0.6);
                     setVal('eyebrow_raise', 0.8);
                     break;
                 case 'Angry':
                     setVal('frown', 0.5);
                     setVal('squint', 0.6);
                     setVal('eyebrow_raise', -0.5);
                     break;
             }
             return { ...s, sliders: newSliders };
        });
    };

    const buildPromptFromSliders = (sliders: ControllerSlider[]): string => {
        const changes: string[] = [];
        
        sliders.forEach(s => {
            if (s.value === 0) return;
            
            const magnitude = Math.abs(s.value);
            const intensity = magnitude < 0.3 ? 'slightly' : magnitude < 0.7 ? 'moderately' : 'strongly';
            
            if (s.id === 'smile') {
                if (s.value > 0) changes.push(`${intensity} smiling`);
                else changes.push(`${intensity} less smiling`);
            } else if (s.id === 'age') {
                if (s.value > 0) changes.push(`make the subject look ${intensity} older`);
                else changes.push(`make the subject look ${intensity} younger`);
            } else if (s.id === 'head_yaw') {
                 if (s.value > 0) changes.push(`turn head ${intensity} to the right`);
                 else changes.push(`turn head ${intensity} to the left`);
            } else if (s.id === 'head_pitch') {
                 if (s.value > 0) changes.push(`tilt head ${intensity} up`);
                 else changes.push(`tilt head ${intensity} down`);
            }
             else {
                changes.push(`${intensity} increase ${s.label.toLowerCase()}`);
            }
        });

        if (changes.length === 0) return "High resolution photo of the subject.";
        return "Edit the image to apply these changes: " + changes.join(", ") + ". Keep the identity and background consistent.";
    };

    const handleGenerate = async () => {
        if (project.sourceImages.length === 0) return;

        setProject(s => ({ ...s, isGenerating: true, error: null }));
        
        const prompt = buildPromptFromSliders(project.sliders);

        try {
            const result = await editImage(project.sourceImages[0], prompt);
            setProject(s => ({
                ...s,
                generatedImage: result,
                isGenerating: false,
                history: [{ image: result, prompt: prompt }, ...s.history]
            }));
        } catch (err) {
             const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred.';
             setProject(s => ({ ...s, isGenerating: false, error: errorMessage }));
        }
    };

    const handleExport = () => {
        if (!project.generatedImage) return;
        const link = document.createElement('a');
        link.download = `SMART-Studio-controller-${Date.now()}.png`;
        link.href = `data:${project.generatedImage.mimeType};base64,${project.generatedImage.base64}`;
        link.click();
    };

    const activeSliders = project.sliders.filter(s => s.category === project.activeCategory);

    return (
        <main className="w-full grid grid-cols-1 lg:grid-cols-12 gap-6 pt-4 pb-12 items-start">
            
            {/* --- LEFT COLUMN: UPLOAD & HISTORY --- */}
            <div className="lg:col-span-3 flex flex-col gap-4 order-1">
                <div className="glass-card p-4 rounded-2xl">
                    <h3 className="text-sm font-bold text-[var(--color-text-medium)] mb-3">1. Person Image</h3>
                    <p className="text-xs text-[var(--color-text-secondary)] mb-3">Select a clear photo of your subject.</p>
                    <div className="aspect-[3/4] w-full">
                        <ImageWorkspace
                            id="controller-source-uploader"
                            title="Reference Image"
                            images={project.sourceImages}
                            onImagesUpload={handleFileUpload}
                            onImageRemove={handleRemoveSource}
                            isUploading={project.isUploading}
                            onImageUpdate={handleUpdateSource}
                        />
                    </div>
                </div>
                
                <HistoryPanel history={project.history} onSelect={(img) => setProject(s => ({ ...s, generatedImage: img }))} />
            </div>

            {/* --- MIDDLE COLUMN: PREVIEW --- */}
            <div className="lg:col-span-5 flex flex-col gap-4 order-2">
                <div className="glass-card p-4 rounded-2xl h-full min-h-[500px] flex flex-col">
                    <div className="flex-grow bg-black/30 rounded-xl border border-[rgba(var(--color-text-base-rgb),0.1)] relative overflow-hidden flex items-center justify-center group select-none min-h-[350px]">
                        {project.isGenerating && (
                             <div className="absolute inset-0 flex items-center justify-center z-20 bg-black/50 backdrop-blur-sm">
                                 <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
                             </div>
                        )}
                        
                        {(() => {
                            const hasSource = project.sourceImages.length > 0;
                            const hasGenerated = !!project.generatedImage;
                            
                            if (!hasSource) {
                                return <div className="text-[var(--color-text-muted)] text-sm">Upload an image to start</div>;
                            }
                            
                            const showOriginal = isComparing || !hasGenerated;
                            const currentImage = showOriginal ? project.sourceImages[0] : project.generatedImage;
                            
                            if (!currentImage) return null;

                            return (
                                <>
                                    <img 
                                        src={`data:${currentImage.mimeType};base64,${currentImage.base64}`} 
                                        className="w-full h-full object-contain" 
                                        alt="Preview" 
                                        draggable={false}
                                    />
                                    <div className="absolute top-4 left-4 px-3 py-1 bg-black/60 backdrop-blur-md rounded-full border border-white/10 text-xs font-bold text-white shadow-lg transition-all">
                                        {showOriginal ? 'Original' : 'Generated'}
                                    </div>
                                </>
                            );
                        })()}
                    </div>

                    {project.error && <div className="mt-4 bg-[rgba(var(--color-accent-rgb),0.2)] border border-[rgba(var(--color-accent-rgb),0.5)] text-[var(--color-accent-light)] px-3 py-2 rounded-lg text-xs text-center">{project.error}</div>}

                    <div className="mt-4 grid grid-cols-2 gap-3">
                        <Button
                            onClick={handleGenerate}
                            disabled={project.sourceImages.length === 0 || project.isGenerating}
                            loading={project.isGenerating}
                            variant="primary"
                            className="col-span-2 py-3 px-4 shadow-lg"
                        >
                            {project.isGenerating ? 'Generating...' : 'Generate'}
                        </Button>

                        {project.generatedImage ? (
                             <>
                                <button
                                    onMouseDown={() => setIsComparing(true)}
                                    onMouseUp={() => setIsComparing(false)}
                                    onMouseLeave={() => setIsComparing(false)}
                                    onTouchStart={() => setIsComparing(true)}
                                    onTouchEnd={() => setIsComparing(false)}
                                    className="bg-blue-600/80 hover:bg-blue-600 text-white border border-blue-500/30 font-medium py-3 px-4 rounded-xl transition-all select-none text-sm shadow-md"
                                >
                                    Hold to Compare
                                </button>
                                <button 
                                    onClick={handleExport}
                                    className="bg-green-600/80 hover:bg-green-600 text-white border border-green-500/30 font-medium py-3 px-4 rounded-xl transition-all flex items-center justify-center gap-2 text-sm shadow-md"
                                >
                                    <DownloadIcon />
                                    Export Image
                                </button>
                             </>
                        ) : (
                            <button 
                                onClick={resetSliders}
                                className="col-span-2 bg-[rgba(var(--color-text-base-rgb),0.1)] hover:bg-[rgba(var(--color-text-base-rgb),0.2)] text-[var(--color-text-base)] font-medium py-3 px-4 rounded-xl transition-all flex items-center justify-center gap-2"
                            >
                                <ResetIcon /> Reset All
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* --- RIGHT COLUMN: CONTROLS --- */}
            <div className="lg:col-span-4 flex flex-col gap-4 order-3">
                <div className="glass-card rounded-2xl p-1 flex mb-2">
                     {(['Face', 'Head', 'Body', 'Retouch'] as ControlCategory[]).map(cat => (
                         <button
                            key={cat}
                            onClick={() => setProject(s => ({ ...s, activeCategory: cat }))}
                            className={`flex-1 flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-medium transition-all ${project.activeCategory === cat ? 'bg-[var(--color-accent)] text-white shadow-md' : 'text-[var(--color-text-secondary)] hover:bg-[rgba(var(--color-text-base-rgb),0.05)] hover:text-[var(--color-text-base)]'}`}
                         >
                             {cat === 'Face' && <FaceIcon />}
                             {cat === 'Head' && <HeadIcon />}
                             {cat === 'Body' && <BodyIcon />}
                             {cat === 'Retouch' && <RetouchIcon />}
                             {cat}
                         </button>
                     ))}
                </div>

                <div className="glass-card rounded-2xl p-5 flex flex-col gap-6 h-full">
                    <div className="flex flex-col gap-5 overflow-y-auto max-h-[600px] pr-2 suggestions-scrollbar">
                        {activeSliders.map(slider => (
                            <div key={slider.id} className="flex flex-col gap-2">
                                <div className="flex justify-between items-center">
                                    <label htmlFor={slider.id} className="text-sm font-medium text-[var(--color-text-base)]">{slider.label}</label>
                                    <span className="text-xs font-mono bg-[rgba(var(--color-accent-rgb),0.1)] text-[var(--color-accent-light)] px-2 py-0.5 rounded">{slider.value.toFixed(2)}</span>
                                </div>
                                <input 
                                    type="range" 
                                    id={slider.id}
                                    min={slider.min} 
                                    max={slider.max} 
                                    step={slider.step}
                                    value={slider.value}
                                    onChange={(e) => handleSliderChange(slider.id, parseFloat(e.target.value))}
                                    className="w-full h-2 bg-[rgba(var(--color-text-base-rgb),0.1)] rounded-lg appearance-none cursor-pointer accent-[var(--color-accent)]"
                                />
                            </div>
                        ))}
                    </div>
                </div>
            </div>

        </main>
    );
};

export default ControllerStudio;
