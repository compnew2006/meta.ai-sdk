
import React, { useCallback, useState } from 'react';
import { StoryboardStudioProject, ImageFile, StoryboardScene } from '../types';
import { resizeImage } from '../utils';
import { generateStoryboardPlan, generateImage } from '../services/geminiService';
import ImageWorkspace from './ImageWorkspace';

const DirectorIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
);

const GridIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
    </svg>
);

const MagicIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
    </svg>
);

const DownloadIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
    </svg>
);

const StoryboardStudio: React.FC<{
    project: StoryboardStudioProject;
    setProject: React.Dispatch<React.SetStateAction<StoryboardStudioProject>>;
}> = ({ project, setProject }) => {

    const handleFileUpload = async (files: File[]) => {
        if (!files || files.length === 0) return;
        setProject(s => ({ ...s, isUploading: true, error: null }));
        try {
            const uploaded = await Promise.all(files.map(async file => {
                const resized = await resizeImage(file, 2048, 2048);
                const reader = new FileReader();
                return new Promise<ImageFile>(res => {
                    reader.onloadend = () => res({ base64: (reader.result as string).split(',')[1], mimeType: resized.type, name: resized.name });
                    reader.readAsDataURL(resized);
                });
            }));
            setProject(s => ({
                ...s,
                subjectImages: [...s.subjectImages, ...uploaded],
                isUploading: false,
                gridImage: null,
                scenes: []
            }));
        } catch (err) {
            setProject(s => ({ ...s, isUploading: false, error: "Upload failed" }));
        }
    };

    const handleRemoveSubject = (idx: number) => {
        setProject(s => ({ ...s, subjectImages: s.subjectImages.filter((_, i) => i !== idx), scenes: [], gridImage: null }));
    };

    const onCreatePlan = async () => {
        setProject(s => ({ ...s, isGeneratingPlan: true, error: null, scenes: [], gridImage: null }));
        try {
            const plan = await generateStoryboardPlan(project.subjectImages, project.customInstructions);
            const scenes: StoryboardScene[] = plan.map(p => ({
                ...p,
                id: Math.random().toString(36).substr(2, 9),
                image: null,
                isLoading: false,
                error: null
            }));
            setProject(s => ({ ...s, scenes, isGeneratingPlan: false }));
        } catch (err) {
            setProject(s => ({ ...s, isGeneratingPlan: false, error: "Failed to generate storyboard plan" }));
        }
    };

    const onGenerateGrid = async () => {
        if (project.scenes.length === 0) return;
        setProject(s => ({ ...s, isGeneratingGrid: true, error: null }));
        try {
            const prompt = `A professional storyboard grid layout containing 9 numbered panels. Each panel showing a unique scene of the story: ${project.customInstructions}. High contrast, cinematic lighting, sketching style mixed with digital art. Maintain identity from subject image. Aspect Ratio: ${project.aspectRatio}. 8k resolution.`;
            const image = await generateImage(project.subjectImages, prompt, null, project.aspectRatio);
            setProject(s => ({ ...s, gridImage: image, isGeneratingGrid: false }));
        } catch (err) {
            setProject(s => ({ ...s, isGeneratingGrid: false, error: "Grid generation failed" }));
        }
    };

    const onGenerateSceneImage = async (sceneId: string) => {
        const sceneIdx = project.scenes.findIndex(s => s.id === sceneId);
        if (sceneIdx === -1) return;

        setProject(s => {
            const next = [...s.scenes];
            next[sceneIdx] = { ...next[sceneIdx], isLoading: true, error: null };
            return { ...s, scenes: next };
        });

        try {
            const scene = project.scenes[sceneIdx];
            const textConstraint = "STRICTLY PRESERVE all original branding/features from provided subject images. NO EXTRA text.";
            const finalPrompt = `Professional Cinema Shot. ${scene.cameraAngle}. ${scene.visualPrompt}. HIGH-END COMMERCIAL QUALITY. 8k, Photorealistic. ${textConstraint}`;
            
            const image = await generateImage(project.subjectImages, finalPrompt, null, project.aspectRatio);
            
            setProject(s => {
                const next = [...s.scenes];
                next[sceneIdx] = { ...next[sceneIdx], image, isLoading: false };
                return { ...s, scenes: next };
            });
        } catch (err) {
            setProject(s => {
                const next = [...s.scenes];
                next[sceneIdx] = { ...next[sceneIdx], isLoading: false, error: "Failed to generate HQ image" };
                return { ...s, scenes: next };
            });
        }
    };

    const handleDownload = (image: ImageFile, label: string) => {
        const link = document.createElement('a');
        link.href = `data:${image.mimeType};base64,${image.base64}`;
        link.download = `SMART-Studio-Storyboard-${label.replace(/\s+/g, '-')}-${Date.now()}.png`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    if (!project) return null;

    return (
        <main className="w-full flex flex-col gap-8 pt-4 pb-12 animate-in fade-in duration-700">
            {/* Control Section */}
            <div className="glass-card rounded-[2.5rem] p-8 shadow-2xl">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4">
                    <h2 className="text-3xl font-black text-white tracking-tighter flex items-center">
                        <DirectorIcon /> CINEMATIC STORYBOARD DIRECTOR
                    </h2>
                    <div className="flex bg-black/40 rounded-full p-1 border border-white/10">
                        <button 
                            onClick={() => setProject(s => ({ ...s, aspectRatio: '16:9' }))}
                            className={`px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-all ${project.aspectRatio === '16:9' ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-white/40 hover:text-white/60'}`}
                        >
                            Landscape (16:9)
                        </button>
                        <button 
                            onClick={() => setProject(s => ({ ...s, aspectRatio: '9:16' }))}
                            className={`px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-all ${project.aspectRatio === '9:16' ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-white/40 hover:text-white/60'}`}
                        >
                            Portrait (9:16)
                        </button>
                    </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                    <div className="lg:col-span-3">
                        <label className="text-[10px] font-black text-white/40 uppercase tracking-[0.2em] mb-3 block text-center">Reference Subject</label>
                        <ImageWorkspace
                            id="storyboard-subject-uploader"
                            images={project.subjectImages}
                            onImagesUpload={handleFileUpload}
                            onImageRemove={handleRemoveSubject}
                            isUploading={project.isUploading}
                        />
                    </div>

                    <div className="lg:col-span-9 flex flex-col gap-6">
                        <div className="flex flex-col gap-2 bg-white/5 p-6 rounded-3xl border border-white/5">
                            <label className="text-xs font-bold text-[var(--color-accent)] uppercase tracking-widest">Story Vision / Ad Script</label>
                            <textarea
                                value={project.customInstructions}
                                onChange={(e) => setProject(s => ({ ...s, customInstructions: e.target.value }))}
                                placeholder="e.g. 'A futuristic robot exploring a secret garden, discovering a glowing flower. Focus on emotional cinematic lighting.'"
                                className="w-full bg-transparent border-none p-0 text-lg font-medium focus:ring-0 placeholder:text-white/20 min-h-[100px] suggestions-scrollbar"
                            />
                        </div>

                        <div className="flex flex-col sm:flex-row gap-4">
                            <button
                                onClick={onCreatePlan}
                                disabled={project.isGeneratingPlan || !project.customInstructions.trim()}
                                className="flex-1 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white font-black py-4 rounded-2xl shadow-xl transition-all active:scale-[0.98] disabled:opacity-30 text-base uppercase tracking-widest"
                            >
                                {project.isGeneratingPlan ? 'Plotting Script...' : 'Generate 9-Scene Plan'}
                            </button>
                            {project.scenes.length > 0 && (
                                <button
                                    onClick={onGenerateGrid}
                                    disabled={project.isGeneratingGrid}
                                    className="flex-1 bg-white text-black hover:bg-gray-200 font-black py-4 rounded-2xl shadow-xl transition-all active:scale-[0.98] disabled:opacity-30 text-base uppercase tracking-widest flex items-center justify-center gap-2"
                                >
                                    <GridIcon />
                                    {project.isGeneratingGrid ? 'Creating Preview...' : 'Generate Storyboard Grid'}
                                </button>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Grid Image Preview */}
            {project.gridImage && (
                <div className="animate-in fade-in duration-1000">
                    <div className="flex items-center gap-4 mb-4">
                        <h3 className="text-sm font-black text-white/40 uppercase tracking-widest">Storyboard Preview Grid</h3>
                        <div className="h-px flex-grow bg-white/10"></div>
                        <button 
                            onClick={() => handleDownload(project.gridImage!, 'Preview-Grid')} 
                            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-full text-[10px] font-black transition-all shadow-lg uppercase tracking-widest"
                        >
                            <DownloadIcon />
                            Download Grid
                        </button>
                    </div>
                    <div className={`glass-card rounded-[2rem] overflow-hidden border border-white/10 shadow-2xl relative ${project.aspectRatio === '9:16' ? 'max-w-md mx-auto aspect-[9/16]' : 'aspect-video'}`}>
                        <img src={`data:${project.gridImage.mimeType};base64,${project.gridImage.base64}`} alt="Storyboard Grid" className="w-full h-full object-cover" />
                        <div className="absolute top-4 left-4 bg-black/60 px-3 py-1 rounded-full text-[9px] font-black text-white/80 border border-white/10 backdrop-blur-md">3x3 LOW-RES PREVIEW</div>
                        <div className="absolute bottom-6 right-6 opacity-0 group-hover:opacity-100 transition-opacity">
                             <button 
                                onClick={() => handleDownload(project.gridImage!, 'Preview-Grid')}
                                className="p-4 bg-emerald-500 text-white rounded-full shadow-2xl hover:scale-110 transition-transform"
                             >
                                <DownloadIcon />
                             </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Detailed Scenes List */}
            {project.scenes.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mt-4">
                    {project.scenes.map((scene, idx) => (
                        <div key={scene.id} className="glass-card rounded-[2rem] overflow-hidden flex flex-col border border-white/5 group hover:border-[var(--color-accent)]/30 transition-all shadow-2xl animate-in slide-in-from-bottom-8 duration-700" style={{ animationDelay: `${idx * 100}ms` }}>
                            <div className={`relative bg-black/40 flex items-center justify-center overflow-hidden ${project.aspectRatio === '9:16' ? 'aspect-[9/16]' : 'aspect-video'}`}>
                                {scene.isLoading ? (
                                    <div className="flex flex-col items-center gap-3">
                                        <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
                                        <span className="text-[10px] font-bold text-white/30 tracking-widest uppercase">Developing HQ Shot...</span>
                                    </div>
                                ) : scene.image ? (
                                    <div className="w-full h-full relative group/img">
                                        <img src={`data:${scene.image.mimeType};base64,${scene.image.base64}`} alt={`Scene ${idx+1}`} className="w-full h-full object-cover" />
                                        <div className="absolute inset-0 bg-black/50 opacity-0 group-hover/img:opacity-100 transition-opacity flex items-center justify-center gap-4">
                                            <button 
                                                onClick={() => handleDownload(scene.image!, `Scene-${idx+1}`)} 
                                                className="p-4 bg-emerald-500 text-white rounded-full hover:scale-110 transition-transform shadow-xl flex items-center justify-center"
                                                title="Download Scene Image"
                                            >
                                                <DownloadIcon />
                                            </button>
                                            <button 
                                                onClick={() => onGenerateSceneImage(scene.id)} 
                                                className="p-4 bg-white text-black rounded-full hover:scale-110 transition-transform shadow-xl flex items-center justify-center"
                                                title="Regenerate Scene"
                                            >
                                                <DirectorIcon />
                                            </button>
                                        </div>
                                    </div>
                                ) : (
                                    <div className="flex flex-col items-center gap-4 px-8 text-center">
                                        <div className="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center border border-white/5 text-white/20">
                                            <MagicIcon />
                                        </div>
                                        <button 
                                            onClick={() => onGenerateSceneImage(scene.id)}
                                            className="px-6 py-2 bg-white/10 hover:bg-white/20 text-white text-[10px] font-black rounded-full transition-colors uppercase tracking-widest border border-white/10"
                                        >
                                            Generate HQ Shot
                                        </button>
                                    </div>
                                )}
                                <div className="absolute top-4 left-4 bg-black/60 backdrop-blur-md px-3 py-1 rounded-full text-[10px] font-black text-white/80 border border-white/10">
                                    SCENE 0{scene.sequence}
                                </div>
                            </div>

                            <div className="p-6 flex flex-col gap-4">
                                <div>
                                    <label className="text-[9px] font-black text-[var(--color-accent)] uppercase tracking-widest mb-1 block">Camera Strategy</label>
                                    <div className="text-xs font-bold text-white/90">{scene.cameraAngle}</div>
                                </div>
                                <div>
                                    <label className="text-[9px] font-black text-white/30 uppercase tracking-widest mb-1 block">Description</label>
                                    <p className="text-sm text-white/70 leading-relaxed font-medium">{scene.description}</p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {project.error && (
                <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-sm text-center font-bold">
                    {project.error}
                </div>
            )}
        </main>
    );
};

export default StoryboardStudio;
