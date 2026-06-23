
import React, { useCallback } from 'react';
import { CampaignStudioProject, ImageFile, BrandingResult } from '../types';
import { resizeImage } from '../utils';
import { analyzeProductForCampaign, generateImage, editImage } from '../services/geminiService';
import ImageWorkspace from './ImageWorkspace';
import Button from './ui/Button';
import BrandingResultsGrid from './BrandingResultsGrid';

// Updated for 6 professional scenarios
const CAMPAIGN_SCENARIOS = [
    'Professional Hero Front View (Studio)', 
    'Lifestyle: A person using the product naturally (Human connection)', 
    'Aesthetic Flat Lay Shot (Environment)',
    'Creative Abstract Composition (Artistic)',
    'Dynamic Close-up Detail (Macro)',
    'Outdoor Natural Sunlight Shot (Atmosphere)'
];

const CAMPAIGN_MOODS = [
    { label: 'Original', value: '' },
    { label: 'Minimalist White', value: 'Clean, minimalist white studio aesthetic' },
    { label: 'Dark Luxury', value: 'Dramatic dark luxury with gold accents' },
    { label: 'Pastel Pop', value: 'Soft playful pastel colors' },
    { label: 'Nature Green', value: 'Organic fresh natural green aesthetic' },
    { label: 'Ocean Blue', value: 'Deep serene ocean blue tones' },
    { label: 'Warm Gold', value: 'Warm, golden hour luxury lighting' },
    { label: 'Cyberpunk Neon', value: 'Vibrant neon cyberpunk style' },
];

const EditIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3 text-emerald-500/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
    </svg>
);

const PaletteIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-[var(--color-accent)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
    </svg>
);

const CampaignStudio: React.FC<{
    project: CampaignStudioProject;
    setProject: React.Dispatch<React.SetStateAction<CampaignStudioProject>>;
}> = ({ project, setProject }) => {

    if (!project) return null;
    const productImages = project.productImages || [];
    const results = project.results || [];
    const mode = project.mode || 'auto';

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
            const combined = [...productImages, ...uploaded];
            setProject(s => ({ ...s, productImages: combined, isUploading: false, isAnalyzing: true }));
            const analysis = await analyzeProductForCampaign(combined);
            setProject(s => ({ ...s, productAnalysis: analysis, isAnalyzing: false }));
        } catch (err) {
            setProject(s => ({ ...s, isUploading: false, isAnalyzing: false, error: "Upload failed" }));
        }
    };

    const handleRemoveProduct = (idx: number) => {
        setProject(s => ({ ...s, productImages: s.productImages.filter((_, i) => i !== idx) }));
    };

    const handleUpdateProduct = (idx: number, newImage: ImageFile) => {
        setProject(s => {
            const nextImages = [...s.productImages];
            nextImages[idx] = newImage;
            return { ...s, productImages: nextImages };
        });
    };

    const handleCustomIdeaChange = (index: number, value: string) => {
        setProject(s => {
            const nextIdeas = [...s.customIdeas];
            nextIdeas[index] = value;
            return { ...s, customIdeas: nextIdeas };
        });
    };

    const onGenerate = useCallback(async () => {
        if (productImages.length === 0) return;
        
        if (mode === 'custom' && project.customIdeas.every(idea => !idea.trim())) {
            setProject(s => ({ ...s, error: 'Please write at least one idea description.' }));
            return;
        }

        setProject(s => ({ ...s, isGenerating: true, error: null, results: [] }));
        
        try {
            let analysis = project.productAnalysis || await analyzeProductForCampaign(productImages);
            
            const scenarios = mode === 'auto' 
                ? CAMPAIGN_SCENARIOS 
                : project.customIdeas.filter(idea => idea.trim().length > 0);

            const initial = scenarios.map(scenario => ({ 
                scenario, 
                image: null, 
                isLoading: true, 
                error: null, 
                editPrompt: '', 
                isEditing: false 
            }));

            setProject(s => ({ ...s, results: initial as any }));

            const promises = scenarios.map((scenario) => {
                const moodV = mode === 'auto' ? (CAMPAIGN_MOODS.find(m => m.label === project.selectedMood)?.value || '') : '';
                const backgroundInfo = project.customPrompt ? ` Style details: ${project.customPrompt}.` : '';
                
                const textConstraint = "STRICTLY PRESERVE all original text, labels, and branding on the product. DO NOT erase or modify existing writing. NO EXTRA generated text in the scene environment.";

                const prompt = mode === 'auto'
                    ? `Professional Commercial Photography: ${analysis}. Scenario: ${scenario}. Style: ${moodV}.${backgroundInfo} PHOTOREALISTIC, HIGH-RESOLUTION, CLEAN IMAGE. ${textConstraint}`
                    : `Professional Product Idea Shoot: ${analysis}. Idea: ${scenario}.${backgroundInfo} PHOTOREALISTIC, STRICT IDENTITY PRESERVATION. ${textConstraint}`;
                
                return generateImage(productImages, prompt, null)
                    .then(image => ({ scenario, image }))
                    .catch(error => ({ scenario, error: error.message }));
            });

            const completed = await Promise.all(promises);
            
            setProject(s => {
                const nextResults = [...s.results];
                completed.forEach(res => {
                    const idx = nextResults.findIndex(r => r && r.scenario === res.scenario);
                    if (idx !== -1) {
                        if ('image' in res) nextResults[idx] = { ...nextResults[idx], image: res.image, isLoading: false };
                        else nextResults[idx] = { ...nextResults[idx], error: res.error, isLoading: false };
                    }
                });
                return { ...s, results: nextResults };
            });
        } catch (err) {
            setProject(s => ({ ...s, error: 'Generation failed', isGenerating: false }));
        } finally {
            setProject(s => ({ ...s, isGenerating: false }));
        }
    }, [productImages, project.productAnalysis, project.selectedMood, project.customPrompt, project.customIdeas, mode, setProject]);

    const handleEditResult = async (index: number, editPromptText: string) => {
        const resultToEdit = results[index];
        if (!resultToEdit || !resultToEdit.image) return;

        setProject(s => {
            const nextResults = [...s.results];
            nextResults[index] = { ...nextResults[index], isEditing: true, error: null };
            return { ...s, results: nextResults };
        });

        try {
            const updatedImage = await editImage(resultToEdit.image, editPromptText);
            setProject(s => {
                const nextResults = [...s.results];
                nextResults[index] = { ...nextResults[index], image: updatedImage, isEditing: false };
                return { ...s, results: nextResults };
            });
        } catch (err) {
            console.error("Edit failed:", err);
            setProject(s => {
                const nextResults = [...s.results];
                nextResults[index] = { ...nextResults[index], isEditing: false, error: err instanceof Error ? err.message : "Edit failed" };
                return { ...s, results: nextResults };
            });
        }
    };

    const safeGridResults: BrandingResult[] = results.map(r => ({
        category: r.scenario as any,
        image: r.image || null,
        isLoading: !!r.isLoading,
        isEditing: !!r.isEditing,
        error: r.error || null,
        editPrompt: r.editPrompt || ''
    }));

    return (
        <main className="w-full flex flex-col gap-8 pt-4 pb-12 animate-in fade-in duration-700">
            {/* Mode Switcher */}
            <div className="flex justify-center">
                <div className="bg-black/20 p-1 rounded-2xl border border-white/5 flex gap-1">
                    <button 
                        onClick={() => setProject(s => ({ ...s, mode: 'auto' }))}
                        className={`px-8 py-2 rounded-xl text-sm font-bold transition-all ${mode === 'auto' ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-white/40 hover:text-white/60'}`}
                    >
                        Auto Scenarios
                    </button>
                    <button 
                        onClick={() => setProject(s => ({ ...s, mode: 'custom' }))}
                        className={`px-8 py-2 rounded-xl text-sm font-bold transition-all ${mode === 'custom' ? 'bg-[var(--color-accent)] text-white shadow-lg' : 'text-white/40 hover:text-white/60'}`}
                    >
                        Custom Ideas
                    </button>
                </div>
            </div>

            {/* Control Panel */}
            <div className="glass-card rounded-[2rem] p-6 md:p-8 flex flex-col lg:flex-row gap-8 shadow-2xl">
                {/* Left: Product Selection */}
                <div className="w-full lg:w-72 flex-shrink-0">
                    <h3 className="text-xs font-bold text-white/40 uppercase tracking-[0.2em] mb-4 text-center">Product Ref.</h3>
                    <div className="w-full relative">
                        <ImageWorkspace
                            id="campaign-product-uploader"
                            images={productImages}
                            onImagesUpload={handleFileUpload}
                            onImageRemove={handleRemoveProduct}
                            isUploading={project.isUploading}
                            onImageUpdate={handleUpdateProduct}
                        />
                    </div>
                </div>

                {/* Right: Campaign Controls */}
                <div className="flex-grow flex flex-col gap-6">
                    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                        <div>
                            <h2 className="text-2xl font-black text-white tracking-tight">
                                {mode === 'auto' ? 'Social Media Campaign Studio' : 'Custom Idea Studio'}
                            </h2>
                            <p className="text-sm text-white/50 mt-1">
                                {mode === 'auto' 
                                    ? 'AI will generate 6 unique Social Media feed posts based on professional use cases.' 
                                    : 'Describe up to 6 custom ideas to generate specialized campaign posts.'}
                            </p>
                        </div>
                        <Button
                            onClick={onGenerate}
                            disabled={productImages.length === 0 || project.isGenerating}
                            loading={project.isGenerating}
                            variant="primary"
                            className="font-black py-4 px-8 shadow-xl whitespace-nowrap"
                        >
                            {project.isGenerating
                                ? 'Generating...'
                                : mode === 'auto' ? 'Generate 6 Feed Posts' : 'Generate 6 Custom Posts'}
                        </Button>
                    </div>

                    {/* Shared Style/Background Prompt */}
                    <div className="flex flex-col gap-2 bg-white/5 p-4 rounded-2xl border border-white/5">
                        <label className="text-[10px] font-black text-[var(--color-accent)] uppercase tracking-widest flex items-center gap-2">
                            <PaletteIcon /> Design Theme & Aesthetics
                        </label>
                        <textarea 
                            value={project.customPrompt} 
                            onChange={(e) => setProject(s => ({ ...s, customPrompt: e.target.value }))} 
                            placeholder="Specify aesthetics (e.g., 'Luxury marble surfaces with soft rim lighting' or 'Minimalist geometric shadows')" 
                            className="w-full bg-transparent border-none p-0 text-sm font-medium focus:ring-0 placeholder:text-white/20 resize-none min-h-[60px]"
                        />
                    </div>

                    {mode === 'auto' ? (
                        /* Auto Mode Mood Selector */
                        <div className="flex flex-col gap-3">
                            <label className="text-[10px] font-black text-white/30 uppercase tracking-widest ml-1">Quick Mood Preset</label>
                            <div className="flex flex-wrap gap-2">
                                {CAMPAIGN_MOODS.map(mood => (
                                    <button 
                                        key={mood.label} 
                                        onClick={() => setProject(s => ({ ...s, selectedMood: mood.label }))} 
                                        className={`px-5 py-2 text-xs font-bold rounded-full transition-all border ${
                                            project.selectedMood === mood.label 
                                            ? 'bg-[var(--color-accent)] text-white border-[var(--color-accent)] shadow-lg shadow-[var(--color-accent)]/20' 
                                            : 'bg-black/20 text-white/60 border-white/5 hover:border-white/20 hover:text-white'
                                        }`}
                                    >
                                        {mood.label}
                                    </button>
                                ))}
                            </div>
                        </div>
                    ) : (
                        /* Custom Ideas Mode Inputs - 6 inputs in grid */
                        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                            {[0, 1, 2, 3, 4, 5].map((idx) => (
                                <div key={idx} className="flex flex-col gap-2">
                                    <label className="text-[10px] font-black text-white/30 uppercase tracking-widest ml-1">Custom Idea {idx + 1}</label>
                                    <textarea 
                                        value={project.customIdeas[idx]}
                                        onChange={(e) => handleCustomIdeaChange(idx, e.target.value)}
                                        placeholder={`e.g. "Product being unboxed by a stylish hand..."`}
                                        className="w-full h-20 glass-input rounded-2xl p-4 text-xs font-medium focus:ring-2 focus:ring-[var(--color-accent)] border-white/10 resize-none"
                                    />
                                </div>
                            ))}
                        </div>
                    )}

                    {/* Analysis Section */}
                    <div className={`transition-all duration-500 overflow-hidden ${project.productAnalysis || project.isAnalyzing ? 'opacity-100 max-h-96' : 'opacity-0 max-h-0'}`}>
                        <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-2xl p-5 group/analysis">
                            <div className="flex justify-between items-center mb-3">
                                <h4 className="text-[10px] font-black text-emerald-500/70 uppercase tracking-[0.2em]">AI Product Intelligence</h4>
                                {!project.isAnalyzing && project.productAnalysis && (
                                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10">
                                        <EditIcon />
                                        <span className="text-[8px] font-black text-emerald-500/60 uppercase tracking-tighter">Editable</span>
                                    </div>
                                )}
                            </div>
                            
                            {project.isAnalyzing ? (
                                <div className="flex items-center gap-3 text-emerald-400/60 animate-pulse py-2">
                                    <div className="w-4 h-4 rounded-full border-2 border-current border-t-transparent animate-spin"></div>
                                    <span className="text-xs font-bold uppercase tracking-widest">Analyzing your product...</span>
                                </div>
                            ) : (
                                <textarea 
                                    value={project.productAnalysis || ''}
                                    onChange={(e) => setProject(s => ({ ...s, productAnalysis: e.target.value }))}
                                    rows={4}
                                    className="w-full bg-transparent border-none p-0 text-[11px] font-medium text-emerald-400/90 leading-relaxed focus:ring-0 resize-none suggestions-scrollbar placeholder:text-emerald-500/20"
                                    placeholder="AI analysis result will appear here. You can edit it to refine how the AI describes your product..."
                                />
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Results Grid */}
            {results.length > 0 && (
                <div className="flex flex-col gap-6 animate-in slide-in-from-bottom-4 duration-1000">
                    <div className="flex items-center gap-4">
                        <h3 className="text-xl font-bold text-white tracking-tight">Campaign Gallery ({results.length} results)</h3>
                        <div className="h-px flex-grow bg-white/5"></div>
                    </div>
                    <BrandingResultsGrid 
                        results={safeGridResults}
                        gridClassName="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6"
                        onEditResult={handleEditResult}
                    />
                </div>
            )}
            
            {project.error && (
                <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl text-sm text-center">
                    {project.error}
                </div>
            )}
        </main>
    );
};

export default CampaignStudio;
