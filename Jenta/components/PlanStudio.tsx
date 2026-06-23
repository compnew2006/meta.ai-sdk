
import React, { useCallback, useState, useEffect } from 'react';
import { PlanStudioProject, ImageFile, PlanIdea } from '../types';
import { resizeImage } from '../utils';
import { generateCampaignPlan, generateImage, analyzeProductForCampaign } from '../services/geminiService';
import ImageWorkspace from './ImageWorkspace';
import Button from './ui/Button';
import { BrandLogoDataUri } from './ui/BrandLogo';

const MagicIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
    </svg>
);

const ExportIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 17v-2a4 4 0 014-4h2m3 2v-3a2 2 0 00-2-2H9a2 2 0 00-2 2v3m0 0l-3-3m3 3l3-3" />
    </svg>
);

const GlobeIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-[var(--color-accent)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
);

const ChatIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-[var(--color-accent)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
    </svg>
);

const TARGET_MARKETS = [
    'Egypt', 'Saudi Arabia (KSA)', 'United Arab Emirates (UAE)', 'Oman (Sultanate of Oman)', 'The Gulf (General)', 'Global / International', 'Europe', 'North America'
];

// New dynamic dialect structure
const BASE_LANGUAGES = [
    { id: 'egy', label: 'Egyptian / مصرية' },
    { id: 'sau', label: 'Saudi / سعودية' },
    { id: 'oma', label: 'Omani / عمانية' },
    { id: 'gulf', label: 'Gulf / خليجية عامة' },
    { id: 'msa', label: 'Modern Standard / فصحى' },
    { id: 'lev', label: 'Levantine / شامي' },
    { id: 'eng', label: 'English / إنجليزية' }
];

const FORMALITY_OPTIONS = [
    { id: 'slang', label: 'Slang / عامية دارجة' },
    { id: 'formal', label: 'Formal / رسمي - فصحى' }
];


const PlanStudio: React.FC<{
    project: PlanStudioProject;
    setProject: React.Dispatch<React.SetStateAction<PlanStudioProject>>;
}> = ({ project, setProject }) => {

    const [isDownloading, setIsDownloading] = useState<string | null>(null);
    const [selectedLang, setSelectedLang] = useState('egy');
    const [selectedFormality, setSelectedFormality] = useState('slang');

    // Update dialect in project when internal selects change
    useEffect(() => {
        const langLabel = BASE_LANGUAGES.find(l => l.id === selectedLang)?.label || 'Arabic';
        const formLabel = FORMALITY_OPTIONS.find(f => f.id === selectedFormality)?.label || 'Slang';
        
        // Special mapping for requested "Saudi native pure"
        let finalDialect = `${langLabel} (${formLabel})`;
        if (selectedLang === 'sau' && selectedFormality === 'slang') {
            finalDialect = "اللهجة السعودية الدارجة الصرفة (Native Pure Saudi Slang)";
        } else if (selectedLang === 'oma' && selectedFormality === 'slang') {
            finalDialect = "اللهجة العمانية الدارجة (Omani Slang)";
        }
        
        setProject(s => ({ ...s, dialect: finalDialect }));
    }, [selectedLang, selectedFormality, setProject]);

    // Trigger Category Analysis when images change
    useEffect(() => {
        if (project.productImages.length > 0 && !project.categoryAnalysis && !project.isAnalyzingCategory) {
            const runAnalysis = async () => {
                setProject(s => ({ ...s, isAnalyzingCategory: true }));
                try {
                    const analysis = await analyzeProductForCampaign(project.productImages);
                    setProject(s => ({ ...s, categoryAnalysis: analysis, isAnalyzingCategory: false }));
                } catch (err) {
                    setProject(s => ({ ...s, isAnalyzingCategory: false }));
                }
            };
            runAnalysis();
        }
    }, [project.productImages.length, project.categoryAnalysis, project.isAnalyzingCategory, setProject]);

    const handleFileUpload = (target: 'product') => async (files: File[]) => {
        if (!files || files.length === 0) return;
        setProject(s => ({ ...s, isUploading: true, error: null, categoryAnalysis: null }));
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
                productImages: [...s.productImages, ...uploaded],
                isUploading: false
            }));
        } catch (err) {
            setProject(s => ({ ...s, isUploading: false, error: "Upload failed" }));
        }
    };

    const onCreatePlan = async () => {
        if (!project.prompt.trim()) {
            setProject(s => ({ ...s, error: 'Please describe your goal or campaign vision.' }));
            return;
        }
        setProject(s => ({ ...s, isGeneratingPlan: true, error: null }));
        try {
            const plan = await generateCampaignPlan(
                project.productImages,
                project.prompt,
                project.targetMarket,
                project.dialect,
                project.categoryAnalysis || ""
            );
            const ideas: PlanIdea[] = plan.map(p => ({
                ...p,
                image: null,
                isLoadingImage: false,
                imageError: null
            }));
            setProject(s => ({ ...s, ideas, isGeneratingPlan: false }));
        } catch (err) {
            setProject(s => ({ ...s, isGeneratingPlan: false, error: "Plan generation failed" }));
        }
    };

    const onGenerateIdeaImage = async (ideaId: string) => {
        const ideaIdx = project.ideas.findIndex(i => i.id === ideaId);
        if (ideaIdx === -1) return;

        setProject(s => {
            const next = [...s.ideas];
            next[ideaIdx] = { ...next[ideaIdx], isLoadingImage: true, imageError: null };
            return { ...s, ideas: next };
        });

        try {
            const textConstraint = "STRICTLY PRESERVE all original branding from the product images if provided. NO EXTRA generated text in the scene.";
            const categoryContext = project.categoryAnalysis ? `Product Category context: ${project.categoryAnalysis}.` : '';
            const finalPrompt = `Professional commercial photography for social media. ${categoryContext} Scenario: ${project.ideas[ideaIdx].scenario}. Style: Photorealistic, high-end commercial shot. ${textConstraint}`;
            
            // Map the format of the idea to a valid AspectRatio
            const format = project.ideas[ideaIdx].format || "3:4";
            const mapFormatToAspectRatio = (f: string): string => {
                const lower = f.toLowerCase();
                if (lower.includes("9:16")) return "9:16";
                if (lower.includes("16:9")) return "16:9";
                if (lower.includes("1:1")) return "1:1";
                return "3:4"; // Default or closest for 4:5 Feed
            };
            const matchedAspectRatio = mapFormatToAspectRatio(format);
            
            const image = await generateImage(project.productImages, finalPrompt, null, matchedAspectRatio);
            
            setProject(s => {
                const next = [...s.ideas];
                next[ideaIdx] = { ...next[ideaIdx], image, isLoadingImage: false };
                return { ...s, ideas: next };
            });
        } catch (err) {
            setProject(s => {
                const next = [...s.ideas];
                next[ideaIdx] = { ...next[ideaIdx], isLoadingImage: false, imageError: "Failed to generate image" };
                return { ...s, ideas: next };
            });
        }
    };

    const handleDownload = (image: ImageFile, label: string, resolution: '2k' | '4k' | 'original' = 'original') => {
        if (resolution === 'original') {
            const link = document.createElement('a');
            link.href = `data:${image.mimeType};base64,${image.base64}`;
            link.download = `SMART-Studio-Plan-${label.replace(/\s+/g, '-')}-${Date.now()}.png`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            return;
        }

        setIsDownloading(`${label}-${resolution}`);
        const img = new Image();
        img.src = `data:${image.mimeType};base64,${image.base64}`;
        img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (!ctx) {
                setIsDownloading(null);
                return;
            };

            const targetWidth = resolution === '4k' ? 4096 : 2048;
            const aspectRatio = img.width / img.height;
            
            canvas.width = targetWidth;
            canvas.height = targetWidth / aspectRatio;

            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

            const link = document.createElement('a');
            link.download = `SMART-Studio-Plan-${label.replace(/\s+/g, '-')}-${resolution}-${Date.now()}.png`;
            link.href = canvas.toDataURL('image/png');
            link.click();
            setIsDownloading(null);
        };
        img.onerror = () => setIsDownloading(null);
    };

    const updateIdea = (id: string, field: keyof PlanIdea, value: string) => {
        setProject(s => ({
            ...s,
            ideas: s.ideas.map(i => i.id === id ? { ...i, [field]: value } : i)
        }));
    };

    const handleExportFullReport = () => {
        const printWindow = window.open('', '_blank');
        if (!printWindow) return;

        const ideasRows = project.ideas.map((idea, idx) => `
            <tr>
                <td style="padding: 15px; border-bottom: 1px solid #eee; text-align: center; font-weight: bold; width: 50px;">${idx + 1}</td>
                <td style="padding: 15px; border-bottom: 1px solid #eee; font-weight: bold; color: #ff0000; width: 150px;">
                    ${idea.platform || 'Meta'} <br/>
                    <span style="font-size: 11px; color: #666; font-weight: normal;">${idea.format || ''}</span>
                </td>
                <td style="padding: 15px; border-bottom: 1px solid #eee; font-weight: bold; color: #333; width: 120px;">${idea.tov}</td>
                <td style="padding: 15px; border-bottom: 1px solid #eee; line-height: 1.6;">
                    ${idea.caption}
                    ${idea.hashtags && idea.hashtags.length > 0 ? `<div style="margin-top: 5px; color: #ff0000; font-size: 11px; font-weight: bold;">${idea.hashtags.join(' ')}</div>` : ''}
                </td>
                <td style="padding: 15px; border-bottom: 1px solid #eee; font-size: 11px; color: #666; width: 120px;">
                    <strong>${idea.schedule}</strong> <br/>
                    <span style="font-size: 10px; text-transform: uppercase;">CTA: ${idea.cta_type || 'share'}</span>
                </td>
            </tr>
        `).join('');

        printWindow.document.write(`
            <html>
            <head>
                <title>SMART Studio Campaign Plan - ${project.name}</title>
                <link href="https://fonts.googleapis.com/css2?family=Tajawal:wght@400;700&display=swap" rel="stylesheet">
                <style>
                    body { font-family: 'Tajawal', sans-serif; direction: rtl; padding: 40px; color: #333; }
                    .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 3px solid #ff0000; padding-bottom: 20px; margin-bottom: 30px; }
                    .logo { height: 60px; }
                    .title-box h1 { margin: 0; color: #000; font-size: 28px; }
                    .title-box p { margin: 5px 0 0 0; color: #666; }
                    table { width: 100%; border-collapse: collapse; margin-top: 20px; }
                    th { background: #f9f9f9; padding: 15px; text-align: right; border-bottom: 2px solid #eee; font-size: 14px; text-transform: uppercase; }
                    .footer { margin-top: 50px; text-align: center; font-size: 13px; color: #666; border-top: 1px solid #eee; padding-top: 25px; }
                    .footer a { color: #ff0000; text-decoration: none; font-weight: bold; border-bottom: 1px dashed #ff0000; }
                    @media print {
                        .no-print { display: none; }
                        body { padding: 0; }
                    }
                </style>
            </head>
            <body>
                <div class="header">
                    <div class="title-box">
                        <h1>خطة الحملة الإعلانية الذكية</h1>
                        <p>بواسطة SMART Studio PRO</p>
                    </div>
                    <img src="${BrandLogoDataUri}" class="logo" />
                </div>
                
                <div style="margin-bottom: 30px; background: #fff5f5; padding: 20px; border-radius: 10px; border-right: 5px solid #ff0000;">
                    <h3 style="margin-top: 0; color: #ff0000;">تفاصيل الحملة:</h3>
                    <p><strong>السوق المستهدف:</strong> ${project.targetMarket}</p>
                    <p><strong>اللهجة:</strong> ${project.dialect}</p>
                    <p><strong>الرؤية:</strong> ${project.prompt}</p>
                </div>

                <table>
                    <thead>
                        <tr>
                            <th>#</th>
                            <th>المنصة والنسق</th>
                            <th>نص التصميم (Hook)</th>
                            <th>نص المنشور والوسوم</th>
                            <th>موعد النشر والهدف</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${ideasRows}
                    </tbody>
                </table>

                <div class="footer">
                    تم إنشاء هذا التقرير آلياً بواسطة SMART Studio. جميع الحقوق محفوظة © ${new Date().getFullYear()} SMART Studio.
                </div>

                <div class="no-print" style="position: fixed; bottom: 30px; left: 30px;">
                    <button onclick="window.print()" style="background: #ff0000; color: white; border: none; padding: 18px 35px; border-radius: 50px; font-weight: bold; cursor: pointer; box-shadow: 0 15px 30px rgba(255,0,0,0.4); font-size: 16px; transition: transform 0.2s;">
                        تأكيد وتحميل كـ PDF / طباعة
                    </button>
                </div>
            </body>
            </html>
        `);
        printWindow.document.close();
    };

    return (
        <main className="w-full flex flex-col gap-8 pt-4 pb-12 animate-in fade-in duration-700">
            <div className="glass-card rounded-[2.5rem] p-8 shadow-2xl">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4">
                    <h2 className="text-3xl font-black text-white tracking-tighter flex items-center">
                        <MagicIcon /> STRATEGIC CAMPAIGN PLANNER
                    </h2>
                    <div className="flex gap-3">
                        {project.ideas.length > 0 && (
                            <button
                                onClick={handleExportFullReport}
                                className="px-8 py-3 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-white rounded-full text-sm font-black uppercase tracking-widest flex items-center shadow-xl shadow-[var(--color-accent)]/20 transition-all active:scale-95 border-2 border-white/10"
                            >
                                <ExportIcon /> Download Full Campaign (PDF)
                            </button>
                        )}
                        <div className="px-4 py-2.5 bg-white/5 rounded-full border border-white/10 text-[10px] font-bold text-white/50 uppercase tracking-widest flex items-center">
                            Output Size: 3:4
                        </div>
                    </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                    <div className="lg:col-span-4 flex flex-col gap-6">
                        <div className="flex flex-col gap-4">
                            <label className="text-[10px] font-black text-white/40 uppercase tracking-[0.2em] ml-1">Product Reference (Optional)</label>
                            <ImageWorkspace
                                id="plan-product-up"
                                images={project.productImages}
                                onImagesUpload={handleFileUpload('product')}
                                onImageRemove={(i) => setProject(s => ({ ...s, productImages: s.productImages.filter((_, idx) => idx !== i), categoryAnalysis: null }))}
                                isUploading={project.isUploading}
                            />
                        </div>
                    </div>

                    <div className="lg:col-span-8 flex flex-col gap-6">
                        <div className="flex flex-col gap-2 bg-white/5 p-6 rounded-3xl border border-white/5">
                            <label className="text-xs font-bold text-[var(--color-accent)] uppercase tracking-widest">Campaign Goal & Brand Vision</label>
                            <textarea
                                value={project.prompt}
                                onChange={(e) => setProject(s => ({ ...s, prompt: e.target.value }))}
                                placeholder="e.g. 'Launching a limited edition luxury perfume for women. Focus on mystery and elegance.'"
                                className="w-full bg-transparent border-none p-0 text-lg font-medium focus:ring-0 placeholder:text-white/20 min-h-[100px] suggestions-scrollbar"
                            />
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div className="flex flex-col gap-2 bg-black/20 p-4 rounded-2xl border border-white/5">
                                <label className="text-[10px] font-black text-white/40 uppercase tracking-widest flex items-center gap-2">
                                    <GlobeIcon /> Target Market
                                </label>
                                <select 
                                    value={project.targetMarket}
                                    onChange={(e) => setProject(s => ({ ...s, targetMarket: e.target.value }))}
                                    className="bg-transparent border-none p-0 text-sm font-bold text-white focus:ring-0 cursor-pointer"
                                >
                                    {TARGET_MARKETS.map(m => <option key={m} value={m} className="bg-gray-900">{m}</option>)}
                                </select>
                            </div>
                            
                            <div className="flex flex-col gap-2 bg-black/20 p-4 rounded-2xl border border-white/5">
                                <label className="text-[10px] font-black text-white/40 uppercase tracking-widest flex items-center gap-2">
                                    <ChatIcon /> Language / البلد
                                </label>
                                <select 
                                    value={selectedLang}
                                    onChange={(e) => setSelectedLang(e.target.value)}
                                    className="bg-transparent border-none p-0 text-sm font-bold text-white focus:ring-0 cursor-pointer"
                                >
                                    {BASE_LANGUAGES.map(l => <option key={l.id} value={l.id} className="bg-gray-900">{l.label}</option>)}
                                </select>
                            </div>

                            <div className="flex flex-col gap-2 bg-black/20 p-4 rounded-2xl border border-white/5">
                                <label className="text-[10px] font-black text-white/40 uppercase tracking-widest flex items-center gap-2">
                                    <ChatIcon /> Style / نوع اللهجة
                                </label>
                                <select 
                                    value={selectedFormality}
                                    onChange={(e) => setSelectedFormality(e.target.value)}
                                    className="bg-transparent border-none p-0 text-sm font-bold text-white focus:ring-0 cursor-pointer"
                                    disabled={selectedLang === 'msa'} // MSA is always formal
                                >
                                    {FORMALITY_OPTIONS.map(f => <option key={f.id} value={f.id} className="bg-gray-900">{f.label}</option>)}
                                </select>
                            </div>
                        </div>

                        <div className={`transition-all duration-500 overflow-hidden ${project.productImages.length > 0 ? 'opacity-100 max-h-96' : 'opacity-0 max-h-0'}`}>
                            <div className="bg-blue-500/5 border border-blue-500/20 rounded-2xl p-5">
                                <h4 className="text-[10px] font-black text-blue-500/70 uppercase tracking-[0.2em] mb-2">Market Intelligence</h4>
                                {project.isAnalyzingCategory ? (
                                    <div className="flex items-center gap-3 text-blue-400/60 animate-pulse">
                                        <div className="w-4 h-4 rounded-full border-2 border-current border-t-transparent animate-spin"></div>
                                        <span className="text-[11px] font-bold">Analyzing product positioning...</span>
                                    </div>
                                ) : (
                                    <p className="text-[11px] text-white/70 leading-relaxed italic">
                                        {project.categoryAnalysis || "Identify product category to get started..."}
                                    </p>
                                )}
                            </div>
                        </div>

                        <Button
                            onClick={onCreatePlan}
                            disabled={project.isGeneratingPlan || !project.prompt.trim()}
                            loading={project.isGeneratingPlan}
                            fullWidth
                            variant="primary"
                            className="font-black py-5 shadow-xl uppercase tracking-widest"
                        >
                            {project.isGeneratingPlan ? 'ORCHESTRATING STRATEGY...' : 'CRAFT 9 LOCALIZED POSTS'}
                        </Button>
                    </div>
                </div>
            </div>

            {project.ideas.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                    {project.ideas.map((idea, idx) => {
                        const is916 = idea.format?.includes("9:16");
                        const is11 = idea.format?.includes("1:1");
                        const aspectClass = is916 ? "aspect-[9/16]" : is11 ? "aspect-square" : "aspect-[3/4]";

                        return (
                            <div key={idea.id} className="glass-card rounded-[2rem] overflow-hidden flex flex-col border border-white/5 group hover:border-[var(--color-accent)]/30 transition-all shadow-2xl animate-in slide-in-from-bottom-8 duration-700" style={{ animationDelay: `${idx * 100}ms` }}>
                                <div className={`${aspectClass} bg-black/40 relative overflow-hidden flex items-center justify-center`}>
                                    {idea.isLoadingImage ? (
                                        <div className="flex flex-col items-center gap-3">
                                            <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
                                            <span className="text-[10px] font-bold text-white/30 tracking-widest uppercase">Generating Visual...</span>
                                        </div>
                                    ) : idea.image ? (
                                        <div className="w-full h-full relative group/img">
                                            <img src={`data:${idea.image.mimeType};base64,${idea.image.base64}`} alt="Post Visual" className="w-full h-full object-cover" />
                                            <div className="absolute inset-0 bg-black/40 opacity-0 group-hover/img:opacity-100 transition-opacity flex items-center justify-center gap-2">
                                                <button 
                                                    onClick={() => handleDownload(idea.image!, `Post-${idx+1}`, '2k')}
                                                    className="p-3 bg-white text-black rounded-full hover:bg-[var(--color-accent)] hover:text-white transition-all transform hover:scale-110 shadow-xl"
                                                    title="Download 2K"
                                                >
                                                    <span className="text-[10px] font-black">2K</span>
                                                </button>
                                                <button 
                                                    onClick={() => handleDownload(idea.image!, `Post-${idx+1}`, '4k')}
                                                    className="p-3 bg-white text-black rounded-full hover:bg-[var(--color-accent)] hover:text-white transition-all transform hover:scale-110 shadow-xl"
                                                    title="Download 4K"
                                                >
                                                    <span className="text-[10px] font-black">4K</span>
                                                </button>
                                                <button 
                                                    onClick={() => onGenerateIdeaImage(idea.id)}
                                                    className="p-3 bg-black/60 text-white rounded-full hover:bg-black/80 transition-all transform hover:scale-110 border border-white/10"
                                                    title="Regenerate"
                                                >
                                                    <MagicIcon />
                                                </button>
                                            </div>
                                        </div>
                                    ) : (
                                        <div className="flex flex-col items-center gap-4 px-8 text-center">
                                            <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center border border-white/5">
                                                <MagicIcon />
                                            </div>
                                            <button 
                                                onClick={() => onGenerateIdeaImage(idea.id)}
                                                className="px-6 py-2 bg-white/10 hover:bg-white/20 text-white text-xs font-bold rounded-full transition-colors uppercase tracking-widest border border-white/10"
                                            >
                                                Generate Image
                                            </button>
                                        </div>
                                    )}
                                    <div className="absolute top-4 left-4 flex flex-wrap gap-1.5 z-10">
                                        <div className="bg-black/60 backdrop-blur-md px-3 py-1 rounded-full text-[10px] font-black text-white/80 border border-white/10">
                                            POST 0{idx + 1}
                                        </div>
                                        {idea.platform && (
                                            <div className="bg-[var(--color-accent)]/80 backdrop-blur-md px-3 py-1 rounded-full text-[10px] font-black text-white border border-white/10">
                                                {idea.platform.toUpperCase()}
                                            </div>
                                        )}
                                        {idea.format && (
                                            <div className="bg-white/10 backdrop-blur-md px-3 py-1 rounded-full text-[10px] font-bold text-white/80 border border-white/10">
                                                {idea.format}
                                            </div>
                                        )}
                                    </div>
                                </div>

                                <div className="p-6 flex flex-col gap-4">
                                    <div className="space-y-1">
                                        <div className="flex justify-between items-center">
                                            <label className="text-[9px] font-black text-[var(--color-accent)] uppercase tracking-widest">Caption ({project.dialect})</label>
                                        </div>
                                        <textarea
                                            value={idea.caption}
                                            onChange={(e) => updateIdea(idea.id, 'caption', e.target.value)}
                                            className="w-full bg-black/20 rounded-xl p-3 text-sm text-white/90 border border-white/5 focus:border-[var(--color-accent)]/50 focus:ring-0 resize-none suggestions-scrollbar h-24"
                                        />
                                        {idea.hashtags && idea.hashtags.length > 0 && (
                                            <div className="flex flex-wrap gap-1 mt-1.5">
                                                {idea.hashtags.map((tag, tIdx) => (
                                                    <span key={tIdx} className="text-[9px] text-[var(--color-accent)] font-bold bg-[var(--color-accent)]/5 px-2 py-0.5 rounded-md border border-[var(--color-accent)]/10">
                                                        {tag}
                                                    </span>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                    <div className="grid grid-cols-2 gap-3">
                                        <div className="space-y-1">
                                            <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Design Hook / Text</label>
                                            <input
                                                value={idea.tov}
                                                onChange={(e) => updateIdea(idea.id, 'tov', e.target.value)}
                                                className="w-full bg-black/20 rounded-xl px-3 py-2 text-[11px] text-white/70 border border-white/5"
                                            />
                                        </div>
                                        <div className="space-y-1 flex flex-col justify-end">
                                            <div className="flex gap-1.5 h-full pt-1">
                                                <button 
                                                    onClick={() => handleDownload(idea.image!, `Post-${idx+1}`, '2k')}
                                                    disabled={!idea.image || isDownloading === `Post-${idx+1}-2k`}
                                                    className="flex-1 bg-white/5 hover:bg-white/10 border border-white/5 text-[10px] font-black rounded-lg transition-all text-white/60 disabled:opacity-20 flex items-center justify-center"
                                                >
                                                    {isDownloading === `Post-${idx+1}-2k` ? '...' : '2K'}
                                                </button>
                                                <button 
                                                    onClick={() => handleDownload(idea.image!, `Post-${idx+1}`, '4k')}
                                                    disabled={!idea.image || isDownloading === `Post-${idx+1}-4k`}
                                                    className="flex-1 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-[10px] font-black rounded-lg transition-all text-white disabled:opacity-20 flex items-center justify-center shadow-lg shadow-[var(--color-accent)]/20"
                                                >
                                                    {isDownloading === `Post-${idx+1}-4k` ? '...' : '4K'}
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-2 gap-3 text-[10px]">
                                        <div className="bg-white/5 p-3 rounded-xl border border-white/5 text-white/70">
                                            <span className="text-white/30 block uppercase tracking-wider text-[8px] font-black">Focus Feature</span>
                                            <span className="truncate block font-semibold">{idea.feature_focus || 'General'}</span>
                                        </div>
                                        <div className="bg-white/5 p-3 rounded-xl border border-white/5 text-white/70">
                                            <span className="text-white/30 block uppercase tracking-wider text-[8px] font-black">CTA Goal</span>
                                            <span className="font-bold text-[var(--color-accent)] block truncate">
                                                {idea.cta_type ? idea.cta_type.toUpperCase() : 'SHARE'}
                                            </span>
                                        </div>
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Visual Scenario (AI Prompt)</label>
                                        <textarea
                                            value={idea.scenario}
                                            onChange={(e) => updateIdea(idea.id, 'scenario', e.target.value)}
                                            className="w-full bg-black/10 rounded-xl px-3 py-2 text-[10px] text-white/50 border border-dashed border-white/10 focus:border-white/30 focus:ring-0 resize-none"
                                        />
                                    </div>
                                </div>
                            </div>
                        );
                    })}
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

export default PlanStudio;
