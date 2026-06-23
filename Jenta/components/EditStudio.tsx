
import React, { useState, useRef, useEffect, useCallback } from 'react';
import { EditStudioProject, ImageFile, EditAdjustments, LocalText, GlobalLayer } from '../types';
import { resizeImage } from '../utils';

const ARABIC_FONTS = ['Cairo', 'Tajawal', 'Amiri', 'Reem Kufi', 'Lateef', 'Changa', 'Harmattan', 'Almarai'];
const ENGLISH_FONTS = ['Montserrat', 'Bebas Neue', 'Playfair Display', 'Oswald', 'Rubik', 'Inter', 'Poppins', 'Roboto'];
const FONT_WEIGHTS = [
    { label: 'Thin (300)', value: '300' },
    { label: 'Regular (400)', value: '400' },
    { label: 'Bold (700)', value: '700' },
    { label: 'Black (900)', value: '900' }
];

const LUTS = [
    { name: 'Original', filter: '' },
    { name: 'Warm Sun', filter: 'sepia(0.3) saturate(1.2) hue-rotate(-10deg)' },
    { name: 'Ice Cold', filter: 'saturate(0.8) hue-rotate(180deg) brightness(1.1)' },
    { name: 'Soft Vintage', filter: 'sepia(0.5) contrast(0.8) brightness(1.05)' },
    { name: 'Deep Cinematic', filter: 'contrast(1.3) saturate(0.8) brightness(0.9) hue-rotate(-5deg)' },
    { name: 'Black & White', filter: 'grayscale(1) contrast(1.1)' },
    { name: 'Pastel Glow', filter: 'brightness(1.1) saturate(1.3) contrast(0.9)' },
    { name: 'Neon Night', filter: 'hue-rotate(45deg) saturate(1.5) contrast(1.1)' },
];

const TrashIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>;
const PlusIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>;
const CopyIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2" /></svg>;
const PasteIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /></svg>;
const SaveIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" /></svg>;
const XIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
);
const TextIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h7" /></svg>;
const DownloadIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>;

const createNewText = (): LocalText => ({
    id: Math.random().toString(36).substr(2, 9),
    content: 'نص جديد',
    fontSize: 40,
    color: '#ffffff',
    fontFamily: 'Cairo',
    fontWeight: '700',
    x: 50,
    y: 50,
    isVisible: true,
    rotation: 0,
    letterSpacing: 0,
    lineHeight: 1.1,
    maxWidth: 80 
});

const EditStudio: React.FC<{
    project: EditStudioProject;
    setProject: React.Dispatch<React.SetStateAction<EditStudioProject>>;
}> = ({ project, setProject }) => {

    const [isDownloading, setIsDownloading] = useState<string | null>(null);
    const [draggingId, setDraggingId] = useState<string | null>(null);
    const [draggingSlot, setDraggingSlot] = useState<number | null>(null);
    const [selectedLayerId, setSelectedLayerId] = useState<string | null>(null);
    const [activeSlotIdx, setActiveSlotIdx] = useState<number | null>(null);
    const [selectedTextId, setSelectedTextId] = useState<string | null>(null);
    const [customFonts, setCustomFonts] = useState<string[]>([]);
    const [clipboard, setClipboard] = useState<LocalText | null>(null);
    const [justSavedSlot, setJustSavedSlot] = useState<number | null>(null);
    const containerRefs = useRef<(HTMLDivElement | null)[]>([]);

    const addTextToSlot = useCallback((slotIdx: number, customText?: LocalText) => {
        const newText = customText ? { ...customText, id: Math.random().toString(36).substr(2, 9) } : createNewText();
        setProject(s => {
            const currentTexts = s.localTexts[slotIdx] || [];
            return {
                ...s,
                localTexts: { ...s.localTexts, [slotIdx]: [...currentTexts, newText] }
            };
        });
        setSelectedTextId(newText.id);
        setActiveSlotIdx(slotIdx);
    }, [setProject]);

    const handleCopy = useCallback((textObj: LocalText) => {
        setClipboard({ ...textObj });
    }, []);

    const handlePaste = useCallback((slotIdx: number) => {
        if (clipboard) {
            const pastedText = { ...clipboard, x: clipboard.x + 2, y: clipboard.y + 2 };
            addTextToSlot(slotIdx, pastedText);
        }
    }, [clipboard, addTextToSlot]);

    const handleSaveSlot = (idx: number) => {
        setProject(s => ({
            ...s,
            committedTexts: { ...s.committedTexts, [idx]: JSON.parse(JSON.stringify(s.localTexts[idx] || [])) }
        }));
        setJustSavedSlot(idx);
        setTimeout(() => setJustSavedSlot(null), 2000);
    };

    const wrapText = (ctx: CanvasRenderingContext2D, text: string, maxWidth: number) => {
        const words = text.split(/\s+/);
        let lines = [];
        let currentLine = words[0];

        for (let i = 1; i < words.length; i++) {
            const word = words[i];
            const width = ctx.measureText(currentLine + " " + word).width;
            if (width < maxWidth) {
                currentLine += " " + word;
            } else {
                lines.push(currentLine);
                currentLine = word;
            }
        }
        lines.push(currentLine);
        return lines;
    };

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            const isTyping = ['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName);
            if (isTyping && (e.target as HTMLElement).tagName === 'TEXTAREA') return;

            if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'c') {
                if (activeSlotIdx !== null && selectedTextId) {
                    const texts = project.localTexts[activeSlotIdx] || [];
                    const toCopy = texts.find(t => t.id === selectedTextId);
                    if (toCopy) handleCopy(toCopy);
                }
            }

            if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'v') {
                if (activeSlotIdx !== null && clipboard) {
                    handlePaste(activeSlotIdx);
                }
            }

            if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
                e.preventDefault();
                if (activeSlotIdx !== null) handleSaveSlot(activeSlotIdx);
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [activeSlotIdx, selectedTextId, project.localTexts, clipboard, handleCopy, handlePaste]);

    const handleUpload = async (files: File[]) => {
        if (!files || files.length === 0) return;
        setProject(s => ({ ...s, isUploading: true }));
        try {
            const uploaded = await Promise.all(files.map(async file => {
                const resized = await resizeImage(file, 2048, 2048);
                const reader = new FileReader();
                return new Promise<ImageFile>(res => {
                    reader.onloadend = () => res({ base64: (reader.result as string).split(',')[1], mimeType: resized.type, name: resized.name });
                    reader.readAsDataURL(resized);
                });
            }));
            
            setProject(s => {
                const newBase = [...s.baseImages, ...uploaded];
                const startIdx = s.baseImages.length;
                const newCommitted = { ...s.committedTexts };
                const newLocal = { ...s.localTexts };
                
                uploaded.forEach((_, i) => {
                    newCommitted[startIdx + i] = [];
                    newLocal[startIdx + i] = [];
                });

                return { ...s, baseImages: newBase, committedTexts: newCommitted, localTexts: newLocal, isUploading: false };
            });
        } catch (err) {
            setProject(s => ({ ...s, isUploading: false, error: "Upload failed" }));
        }
    };

    const handleRemoveSlot = (idx: number) => {
        setProject(s => {
            const newBase = s.baseImages.filter((_, i) => i !== idx);
            const newLocal: { [key: number]: LocalText[] } = {};
            const newCommitted: { [key: number]: LocalText[] } = {};

            // Re-index remaining text data
            newBase.forEach((_, newI) => {
                // Determine which old index this maps to
                const oldI = newI < idx ? newI : newI + 1;
                newLocal[newI] = s.localTexts[oldI] || [];
                newCommitted[newI] = s.committedTexts[oldI] || [];
            });

            return { ...s, baseImages: newBase, localTexts: newLocal, committedTexts: newCommitted };
        });
        if (activeSlotIdx === idx) setActiveSlotIdx(null);
    };

    const handleFontUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []) as File[];
        for (const file of files) {
            try {
                const fontName = file.name.split('.')[0];
                const reader = new FileReader();
                reader.onload = async (event) => {
                    const data = event.target?.result;
                    if (data) {
                        const fontFace = new FontFace(fontName, data as ArrayBuffer);
                        const loadedFace = await fontFace.load();
                        document.fonts.add(loadedFace);
                        setCustomFonts(prev => [...new Set([...prev, fontName])]);
                    }
                };
                reader.readAsArrayBuffer(file);
            } catch (err) { console.error(err); }
        }
    };

    const handleGlobalAssetUpload = async (files: File[]) => {
        if (!files || files.length === 0) return;
        try {
            const uploaded = await Promise.all(files.map(async file => {
                const resized = await resizeImage(file, 1024, 1024);
                const reader = new FileReader();
                return new Promise<GlobalLayer>(res => {
                    reader.onloadend = () => {
                        const base64 = (reader.result as string).split(',')[1];
                        res({ 
                            id: Math.random().toString(36).substr(2, 9),
                            file: { base64, mimeType: resized.type, name: resized.name },
                            scale: 20,
                            x: 50,
                            y: 50
                        });
                    };
                    reader.readAsDataURL(resized);
                });
            }));
            setProject(s => ({ ...s, globalLayers: [...s.globalLayers, ...uploaded] }));
        } catch (err) { console.error(err); }
    };

    const updateSlotText = (slotIdx: number, textId: string, updates: Partial<LocalText>) => {
        setProject(s => {
            const texts = s.localTexts[slotIdx] || [];
            const updated = texts.map(t => t.id === textId ? { ...t, ...updates } : t);
            return { ...s, localTexts: { ...s.localTexts, [slotIdx]: updated } };
        });
    };

    const deleteSlotText = (slotIdx: number, textId: string) => {
        setProject(s => {
            const texts = (s.localTexts[slotIdx] || []).filter(t => t.id !== textId);
            return { ...s, localTexts: { ...s.localTexts, [slotIdx]: texts } };
        });
        if (selectedTextId === textId) setSelectedTextId(null);
    };

    const handleDownload = (idx: number, resolution: '2k' | '4k') => {
        const imgFile = project.baseImages[idx];
        const container = containerRefs.current[idx];
        if (!imgFile || !container) return;

        const currentTexts = project.localTexts[idx] || [];
        const committedTexts = project.committedTexts[idx] || [];
        if (JSON.stringify(currentTexts) !== JSON.stringify(committedTexts)) {
            alert("⚠️ فضلاً اضغط على زر 'BAKE' لحفظ التعديلات أولاً لضمان ثباتها عند التحميل.");
            return;
        }

        setIsDownloading(`${idx}-${resolution}`);
        const img = new Image();
        img.src = `data:${imgFile.mimeType};base64,${imgFile.base64}`;
        img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (!ctx) return;
            
            const targetWidth = resolution === '4k' ? 4096 : 2048;
            const imgAspect = img.width / img.height;
            canvas.width = targetWidth;
            canvas.height = targetWidth / imgAspect;

            const previewWidth = container.offsetWidth;
            const scaleMultiplier = canvas.width / previewWidth;

            const activeLut = LUTS.find(l => l.name === project.adjustments.lut);
            const sharpness = project.adjustments.sharpness / 100;
            ctx.filter = `${activeLut?.filter || ''} contrast(${1 + (sharpness - 1) * 0.2}) brightness(${1 + (sharpness - 1) * 0.05})`;
            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            ctx.filter = 'none';

            const renderLayers = async () => {
                for (const layer of project.globalLayers) {
                    await new Promise<void>(res => {
                        const lImg = new Image();
                        lImg.src = `data:${layer.file.mimeType};base64,${layer.file.base64}`;
                        lImg.onload = () => {
                            const scale = (layer.scale / 100) * canvas.width;
                            const aspect = lImg.width / lImg.height;
                            const drawW = scale;
                            const drawH = scale / aspect;
                            ctx.drawImage(lImg, (layer.x / 100) * canvas.width - drawW/2, (layer.y / 100) * canvas.height - drawH/2, drawW, drawH);
                            res();
                        };
                    });
                }
                
                const texts = project.committedTexts[idx] || [];
                texts.forEach(text => {
                    if (!text.isVisible) return;
                    ctx.save();
                    
                    const posX = (text.x / 100) * canvas.width;
                    const posY = (text.y / 100) * canvas.height;
                    
                    ctx.translate(posX, posY);
                    ctx.rotate((text.rotation * Math.PI) / 180);
                    
                    const scaledFontSize = text.fontSize * scaleMultiplier;
                    ctx.font = `${text.fontWeight} ${scaledFontSize}px ${text.fontFamily}`;
                    
                    if ('letterSpacing' in ctx) {
                        (ctx as any).letterSpacing = `${text.letterSpacing * scaleMultiplier}px`;
                    }
                    
                    ctx.fillStyle = text.color;
                    ctx.textAlign = 'center';
                    ctx.textBaseline = 'middle';
                    
                    ctx.shadowColor = 'rgba(0,0,0,0.6)';
                    ctx.shadowBlur = 16 * scaleMultiplier;
                    ctx.shadowOffsetY = 4 * scaleMultiplier;
                    
                    const maxPxWidth = (text.maxWidth / 100) * canvas.width;
                    const lines = wrapText(ctx, text.content, maxPxWidth);
                    
                    const lineHeight = scaledFontSize * text.lineHeight;
                    const totalHeight = lineHeight * (lines.length - 1);
                    
                    lines.forEach((line, i) => {
                        const yOffset = (i * lineHeight) - (totalHeight / 2);
                        ctx.fillText(line, 0, yOffset);
                    });
                    
                    ctx.restore();
                });

                const link = document.createElement('a');
                link.download = `SMART-Studio-Final-${idx + 1}-${resolution}.png`;
                link.href = canvas.toDataURL('image/png');
                link.click();
                setIsDownloading(null);
            };
            renderLayers();
        };
    };

    const activeLut = LUTS.find(l => l.name === project.adjustments.lut);
    const filterStyle = { filter: `${activeLut?.filter || ''} contrast(${1 + (project.adjustments.sharpness / 100 - 1) * 0.2}) brightness(${1 + (project.adjustments.sharpness / 100 - 1) * 0.05})` };

    const selectedLayer = project.globalLayers.find(l => l.id === selectedLayerId);
    const activeSlotTexts = activeSlotIdx !== null ? (project.localTexts[activeSlotIdx] || []) : [];
    const selectedText = activeSlotTexts.find(t => t.id === selectedTextId) || null;

    return (
        <main className="w-full flex flex-col lg:flex-row gap-10 pt-4 pb-20 animate-in fade-in duration-500 items-start">
            
            {/* Control Sidebar */}
            <div className="lg:w-[380px] lg:order-2 flex-shrink-0">
                <div className="glass-card rounded-[2.5rem] p-6 shadow-2xl border border-white/5 sticky top-20 max-h-[85vh] overflow-y-auto suggestions-scrollbar">
                    <h2 className="text-[10px] font-black text-white/40 uppercase tracking-[0.3em] mb-6 flex items-center gap-2.5">
                        <div className="w-1.5 h-1.5 bg-[var(--color-accent)] rounded-full"></div> Design Studio
                    </h2>
                    
                    <div className="space-y-6">
                        <div className="space-y-4 pb-6 border-b border-white/5">
                            <label className="w-full h-12 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl flex items-center justify-center cursor-pointer transition-all group gap-2.5">
                                <PlusIcon />
                                <span className="text-[10px] font-black text-white uppercase tracking-widest">Add Base Photo(s)</span>
                                <input type="file" multiple className="hidden" accept="image/*" onChange={(e) => handleUpload(Array.from(e.target.files || []))} />
                            </label>
                            
                            <div className="flex flex-col gap-1.5">
                                <label className="text-[9px] font-black text-white/30 uppercase tracking-widest ml-1">Color Grade</label>
                                <select 
                                    value={project.adjustments.lut}
                                    onChange={(e) => setProject(s => ({ ...s, adjustments: { ...s.adjustments, lut: e.target.value } }))}
                                    className="w-full bg-black/40 border border-white/10 rounded-xl text-[11px] font-bold p-3 text-white focus:outline-none focus:border-[var(--color-accent)]"
                                >
                                    {LUTS.map(l => <option key={l.name} value={l.name}>{l.name}</option>)}
                                </select>
                            </div>
                        </div>

                        <div className="pb-4 space-y-5">
                            <div className="flex justify-between items-center px-1">
                                <h3 className="text-[10px] font-black text-[var(--color-accent)] uppercase tracking-widest flex items-center gap-2">
                                    <TextIcon /> Typography {activeSlotIdx !== null ? `(Slot 0${activeSlotIdx + 1})` : ''}
                                </h3>
                                <div className="flex gap-1.5">
                                    {clipboard && activeSlotIdx !== null && (
                                        <button onClick={() => handlePaste(activeSlotIdx)} className="p-1.5 bg-emerald-500/20 text-emerald-400 rounded-lg border border-emerald-500/20 hover:scale-105 transition-all" title="Paste"><PasteIcon /></button>
                                    )}
                                    {activeSlotIdx !== null && (
                                        <button onClick={() => addTextToSlot(activeSlotIdx)} className="p-1.5 bg-[var(--color-accent)]/20 text-[var(--color-accent)] rounded-lg border border-[var(--color-accent)]/20 hover:scale-105 transition-all" title="Add Layer"><PlusIcon /></button>
                                    )}
                                </div>
                            </div>
                            
                            {activeSlotIdx !== null ? (
                                <div className="space-y-5 animate-in fade-in slide-in-from-right-2 duration-300">
                                    <div className="flex flex-col gap-1.5 max-h-32 overflow-y-auto suggestions-scrollbar">
                                        {activeSlotTexts.map((txt, i) => (
                                            <div 
                                                key={txt.id} 
                                                onClick={() => setSelectedTextId(txt.id)}
                                                className={`flex items-center justify-between p-2.5 rounded-xl cursor-pointer border transition-all ${selectedTextId === txt.id ? 'bg-[rgba(var(--color-accent-rgb),0.15)] border-[var(--color-accent)]/50' : 'bg-white/5 border-transparent hover:bg-white/10'}`}
                                            >
                                                <div className="flex items-center gap-2">
                                                    <div className={`w-1.5 h-1.5 rounded-full ${selectedTextId === txt.id ? 'bg-[var(--color-accent)] shadow-[0_0_8px_var(--color-accent)]' : 'bg-white/20'}`}></div>
                                                    <span className="text-[10px] font-bold text-white/70 truncate max-w-[120px]">{txt.content || `Layer ${i+1}`}</span>
                                                </div>
                                                <div className="flex gap-1">
                                                    <button onClick={(e) => { e.stopPropagation(); handleCopy(txt); }} className="text-white/40 hover:text-white p-1 transition-colors"><CopyIcon /></button>
                                                    <button onClick={(e) => { e.stopPropagation(); deleteSlotText(activeSlotIdx, txt.id); }} className="text-white/40 hover:text-red-500 p-1 transition-colors"><TrashIcon /></button>
                                                </div>
                                            </div>
                                        ))}
                                    </div>

                                    {selectedText && (
                                        <div className="space-y-4 p-4 bg-black/40 rounded-2xl border border-white/5 animate-in zoom-in-95 duration-200">
                                            <textarea value={selectedText.content} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { content: e.target.value })} className="w-full h-20 bg-black/40 border border-white/10 rounded-xl p-3 text-[11px] text-white focus:border-[var(--color-accent)] focus:outline-none resize-none suggestions-scrollbar leading-relaxed" placeholder="Write here..." />
                                            <div className="grid grid-cols-2 gap-3">
                                                <div className="space-y-1.5">
                                                    <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Font</label>
                                                    <select value={selectedText.fontFamily} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { fontFamily: e.target.value })} className="w-full bg-black/40 border border-white/10 rounded-lg text-[10px] p-2 text-white outline-none focus:border-[var(--color-accent)]">
                                                        {customFonts.map(f => <option key={f} value={f}>{f}</option>)}
                                                        <optgroup label="Standard" className="bg-gray-900">{ARABIC_FONTS.map(f => <option key={f} value={f}>{f}</option>)}{ENGLISH_FONTS.map(f => <option key={f} value={f}>{f}</option>)}</optgroup>
                                                    </select>
                                                </div>
                                                <div className="space-y-1.5">
                                                    <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Color</label>
                                                    <div className="relative h-8 w-full bg-black/40 border border-white/10 rounded-lg overflow-hidden p-1">
                                                        <input type="color" value={selectedText.color} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { color: e.target.value })} className="absolute inset-0 w-full h-full bg-transparent border-none cursor-pointer opacity-0" />
                                                        <div className="w-full h-full rounded-md" style={{ backgroundColor: selectedText.color }}></div>
                                                    </div>
                                                </div>
                                                <div className="space-y-1.5">
                                                    <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Size ({selectedText.fontSize}px)</label>
                                                    <input type="range" min="10" max="250" value={selectedText.fontSize} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { fontSize: parseInt(e.target.value) })} className="w-full accent-[var(--color-accent)] h-1 bg-white/10 rounded-lg appearance-none" />
                                                </div>
                                                <div className="space-y-1.5">
                                                    <label className="text-[9px] font-black text-white/30 uppercase tracking-widest">Box Width ({selectedText.maxWidth}%)</label>
                                                    <input type="range" min="10" max="100" value={selectedText.maxWidth} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { maxWidth: parseInt(e.target.value) })} className="w-full accent-[var(--color-accent)] h-1 bg-white/10 rounded-lg appearance-none" />
                                                </div>
                                            </div>
                                            <div className="space-y-1.5">
                                                <div className="flex justify-between text-[9px] font-black text-white/30 uppercase tracking-widest">
                                                    <span>Line Height ({selectedText.lineHeight.toFixed(1)})</span>
                                                    <span>Rotate ({selectedText.rotation}°)</span>
                                                </div>
                                                <div className="grid grid-cols-2 gap-3">
                                                    <input type="range" min="0.5" max="2.5" step="0.1" value={selectedText.lineHeight} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { lineHeight: parseFloat(e.target.value) })} className="w-full accent-[var(--color-accent)] h-1 bg-white/10 rounded-lg appearance-none mt-2" />
                                                    <input type="range" min="-180" max="180" value={selectedText.rotation} onChange={(e) => updateSlotText(activeSlotIdx, selectedText.id, { rotation: parseInt(e.target.value) })} className="w-full accent-[var(--color-accent)] h-1 bg-white/10 rounded-lg appearance-none mt-2" />
                                                </div>
                                            </div>
                                            <button onClick={() => updateSlotText(activeSlotIdx, selectedText.id, { isVisible: !selectedText.isVisible })} className={`w-full py-2.5 rounded-xl text-[9px] font-black uppercase tracking-[0.2em] transition-all border ${selectedText.isVisible ? 'bg-white/5 text-white/40 border-white/10' : 'bg-[var(--color-accent)] text-white border-[var(--color-accent)]'}`}>
                                                {selectedText.isVisible ? 'Disable Layer' : 'Enable Layer'}
                                            </button>
                                        </div>
                                    )}
                                    <label className="w-full h-10 bg-white/5 hover:bg-white/10 border border-dashed border-white/10 rounded-xl flex items-center justify-center cursor-pointer transition-all gap-2 group">
                                        <span className="text-[9px] font-black text-white/40 uppercase tracking-widest">Custom Font (.ttf)</span>
                                        <input type="file" className="hidden" accept=".ttf,.otf" onChange={handleFontUpload} />
                                    </label>
                                </div>
                            ) : (
                                <div className="p-10 border border-white/5 border-dashed rounded-3xl bg-black/20 text-center">
                                    <p className="text-[10px] text-white/30 font-bold uppercase tracking-widest">Choose a photo to start editing</p>
                                </div>
                            )}
                        </div>

                        <div className="pb-4">
                            <label className="text-[9px] font-black text-white/40 uppercase tracking-widest ml-1 mb-4 block">Graphic Assets</label>
                            <div className="flex flex-wrap gap-2.5 mb-4">
                                {project.globalLayers.map(layer => (
                                    <div key={layer.id} className="relative group/l">
                                        <div onClick={() => setSelectedLayerId(layer.id)} className={`w-12 h-12 rounded-xl border-2 overflow-hidden cursor-pointer transition-all ${selectedLayerId === layer.id ? 'border-[var(--color-accent)] shadow-lg' : 'border-white/10 opacity-60'}`}>
                                            <img src={`data:${layer.file.mimeType};base64,${layer.file.base64}`} className="w-full h-full object-contain" />
                                        </div>
                                        <button onClick={() => setProject(s => ({ ...s, globalLayers: s.globalLayers.filter(l => l.id !== layer.id) }))} className="absolute -top-1 -right-1 bg-red-500 text-white rounded-full p-0.5 opacity-0 group-hover/l:opacity-100 transition-all scale-75 shadow-lg"><XIcon /></button>
                                    </div>
                                ))}
                                <label className="w-12 h-12 rounded-xl border border-dashed border-white/20 flex items-center justify-center cursor-pointer hover:border-white/40 transition-all bg-white/5">
                                    <PlusIcon />
                                    <input type="file" multiple className="hidden" accept="image/*" onChange={(e) => handleGlobalAssetUpload(Array.from(e.target.files || []))} />
                                </label>
                            </div>

                            {selectedLayer && (
                                <div className="space-y-4 p-4 bg-white/5 rounded-2xl border border-white/10 animate-in slide-in-from-top-2">
                                    <div className="grid grid-cols-1 gap-4">
                                        <div className="space-y-1.5">
                                            <div className="flex justify-between text-[9px] font-bold text-white/60"><span>Size</span><span>{selectedLayer.scale}%</span></div>
                                            <input type="range" min="5" max="100" value={selectedLayer.scale} onChange={e => setProject(s => ({...s, globalLayers: s.globalLayers.map(l => l.id === selectedLayerId ? {...l, scale: parseInt(e.target.value)} : l)}))} className="w-full h-1 accent-white bg-white/10 rounded-full appearance-none" />
                                        </div>
                                        <div className="grid grid-cols-2 gap-3">
                                            <div className="space-y-1.5">
                                                <div className="flex justify-between text-[9px] font-bold text-white/60"><span>X</span><span>{selectedLayer.x}%</span></div>
                                                <input type="range" min="0" max="100" value={selectedLayer.x} onChange={e => setProject(s => ({...s, globalLayers: s.globalLayers.map(l => l.id === selectedLayerId ? {...l, x: parseInt(e.target.value)} : l)}))} className="w-full h-1 accent-white bg-white/10 rounded-full appearance-none" />
                                            </div>
                                            <div className="space-y-1.5">
                                                <div className="flex justify-between text-[9px] font-bold text-white/60"><span>Y</span><span>{selectedLayer.y}%</span></div>
                                                <input type="range" min="0" max="100" value={selectedLayer.y} onChange={e => setProject(s => ({...s, globalLayers: s.globalLayers.map(l => l.id === selectedLayerId ? {...l, y: parseInt(e.target.value)} : l)}))} className="w-full h-1 accent-white bg-white/10 rounded-full appearance-none" />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Dynamic Results Grid */}
            <div className="flex-grow lg:order-1 min-w-0 w-full">
                <div className="grid grid-cols-1 md:grid-cols-2 2xl:grid-cols-3 gap-8">
                    {project.baseImages.map((img, idx) => {
                        if (!img) return null;
                        const currentTexts = project.localTexts[idx] || [];
                        const committedTexts = project.committedTexts[idx] || [];
                        const isModified = JSON.stringify(currentTexts) !== JSON.stringify(committedTexts);
                        
                        return (
                            <div key={idx} className="flex flex-col gap-5 animate-in fade-in duration-700" style={{ animationDelay: `${idx * 150}ms` }}>
                                <div className="flex justify-between items-center px-2">
                                    <div className="flex items-center gap-3">
                                        <span className="text-[11px] font-black text-white/40 uppercase tracking-[0.4em]">Slide 0{idx + 1}</span>
                                        {isModified ? (
                                            <span className="text-[9px] font-black text-[var(--color-accent)] uppercase animate-pulse flex items-center gap-1.5 bg-[var(--color-accent)]/10 px-2.5 py-1 rounded-full border border-[var(--color-accent)]/20"><div className="w-1 h-1 bg-[var(--color-accent)] rounded-full"></div> Needs Baking</span>
                                        ) : currentTexts.length > 0 && (
                                            <span className="text-[9px] font-black text-emerald-400 uppercase flex items-center gap-1.5 bg-emerald-500/10 px-2.5 py-1 rounded-full border border-emerald-500/20"><div className="w-1 h-1 bg-emerald-400 rounded-full"></div> Baked & Final</span>
                                        )}
                                    </div>
                                </div>

                                <div 
                                    ref={el => containerRefs.current[idx] = el}
                                    onClick={() => setActiveSlotIdx(idx)}
                                    className={`glass-card rounded-[2rem] overflow-hidden relative bg-black/40 shadow-2xl border-2 transition-all duration-500 cursor-pointer group/slot w-full ${activeSlotIdx === idx ? 'border-[var(--color-accent)] ring-8 ring-[var(--color-accent)]/5 scale-[1.01]' : 'border-white/5 hover:border-white/10'}`}
                                >
                                    <div className={`w-full relative transition-opacity duration-300 ${justSavedSlot === idx ? 'opacity-40' : 'opacity-100'}`}>
                                        <img src={`data:${img.mimeType};base64,${img.base64}`} className="w-full h-auto block select-none pointer-events-none" style={filterStyle} />
                                        {project.globalLayers.map(layer => (
                                            <div key={layer.id} style={{ position: 'absolute', left: `${layer.x}%`, top: `${layer.y}%`, width: `${layer.scale}%`, transform: 'translate(-50%, -50%)', pointerEvents: 'none', zIndex: 5 }}>
                                                <img src={`data:${layer.file.mimeType};base64,${layer.file.base64}`} className="w-full h-auto drop-shadow-2xl" />
                                            </div>
                                        ))}
                                        {currentTexts.map(text => (
                                            text.isVisible && (
                                                <div 
                                                    key={text.id}
                                                    style={{
                                                        position: 'absolute', left: `${text.x}%`, top: `${text.y}%`,
                                                        transform: `translate(-50%, -50%) rotate(${text.rotation}deg)`,
                                                        color: text.color, fontSize: `${text.fontSize}px`,
                                                        fontFamily: text.fontFamily, fontWeight: text.fontWeight,
                                                        letterSpacing: `${text.letterSpacing}px`,
                                                        textAlign: 'center', textShadow: '0 4px 16px rgba(0,0,0,0.6)', zIndex: 10,
                                                        whiteSpace: 'normal', touchAction: 'none', lineHeight: `${text.lineHeight}`,
                                                        width: `${text.maxWidth}%`, display: 'flex', justifyContent: 'center', alignItems: 'center',
                                                        wordWrap: 'break-word', overflowWrap: 'break-word'
                                                    }}
                                                    onPointerDown={(e) => { 
                                                        e.stopPropagation(); 
                                                        setSelectedTextId(text.id);
                                                        setDraggingId(text.id);
                                                        setDraggingSlot(idx);
                                                        setActiveSlotIdx(idx);
                                                        (e.target as HTMLElement).setPointerCapture(e.pointerId); 
                                                    }}
                                                    onPointerMove={(e) => {
                                                        if (draggingId !== text.id || draggingSlot !== idx) return;
                                                        const rect = containerRefs.current[idx]!.getBoundingClientRect();
                                                        const x = ((e.clientX - rect.left) / rect.width) * 100;
                                                        const y = ((e.clientY - rect.top) / rect.height) * 100;
                                                        updateSlotText(idx, text.id, { x: Math.max(0, Math.min(100, x)), y: Math.max(0, Math.min(100, y)) });
                                                    }}
                                                    onPointerUp={() => { setDraggingId(null); setDraggingSlot(null); }}
                                                    className={`px-3 py-2 border-2 rounded-xl cursor-move select-none transition-all ${selectedTextId === text.id ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10' : 'border-transparent hover:border-white/30 hover:bg-white/5'}`}
                                                >
                                                    {text.content}
                                                </div>
                                            )
                                        ))}
                                    </div>
                                </div>

                                <div className="flex gap-2.5 animate-in fade-in duration-500 px-1">
                                    <button onClick={() => handleSaveSlot(idx)} className={`flex-1 flex items-center justify-center gap-2 py-4 rounded-2xl text-[10px] font-black transition-all border ${isModified ? 'bg-emerald-600 text-white border-emerald-500 shadow-xl shadow-emerald-500/20 active:scale-95' : 'bg-white/5 text-white/40 border-white/10 cursor-default'}`}><SaveIcon /> {isModified ? 'BAKE CHANGES' : 'BAKED & FINAL'}</button>
                                    <button onClick={() => handleDownload(idx, '2k')} disabled={isDownloading !== null} className="px-4 bg-white/5 hover:bg-white/10 border border-white/10 text-[10px] font-black rounded-2xl transition-all text-white/70 uppercase">2K</button>
                                    <button onClick={() => handleDownload(idx, '4k')} disabled={isDownloading !== null} className="px-4 bg-[var(--color-accent)] hover:bg-[var(--color-accent-dark)] text-[10px] font-black rounded-2xl transition-all text-white shadow-xl shadow-[var(--color-accent)]/20 uppercase">{isDownloading === `${idx}-4k` ? '...' : '4K'}</button>
                                    <button onClick={() => handleRemoveSlot(idx)} className="p-4 rounded-2xl bg-white/5 text-white/20 hover:text-red-500 hover:bg-red-500/10 transition-all border border-white/10"><TrashIcon /></button>
                                </div>
                            </div>
                        );
                    })}

                    {/* Placeholder for adding more photos in the grid itself */}
                    <label className="flex flex-col gap-5 cursor-pointer group animate-in fade-in duration-700">
                        <div className="px-2"><span className="text-[11px] font-black text-white/20 uppercase tracking-[0.4em]">New Slot</span></div>
                        <div className="w-full aspect-square glass-card rounded-[2rem] border-2 border-dashed border-white/10 flex flex-col items-center justify-center p-12 gap-5 opacity-40 group-hover:opacity-100 group-hover:border-[var(--color-accent)]/30 transition-all duration-500 bg-black/20">
                            <div className="w-16 h-16 rounded-full border-2 border-dashed border-white/40 flex items-center justify-center group-hover:scale-110 group-hover:bg-[var(--color-accent)]/10 transition-all duration-500"><PlusIcon /></div>
                            <span className="text-[11px] font-black uppercase tracking-[0.4em]">Add Photo(s)</span>
                            <input type="file" multiple className="hidden" accept="image/*" onChange={(e) => handleUpload(Array.from(e.target.files || []))} />
                        </div>
                    </label>
                </div>
                
                {project.baseImages.length === 0 && !project.isUploading && (
                    <div className="h-[60vh] w-full flex flex-col items-center justify-center text-center px-10 gap-6 opacity-30">
                         <div className="w-24 h-24 rounded-full border-2 border-dashed border-white/40 flex items-center justify-center scale-150 mb-4"><PlusIcon /></div>
                         <h3 className="text-2xl font-black text-white uppercase tracking-widest">No Photos Uploaded</h3>
                         <p className="max-w-md text-sm font-medium leading-relaxed">Start by uploading the photos you want to edit. Each photo will create its own design workspace.</p>
                    </div>
                )}
            </div>
        </main>
    );
};

export default EditStudio;
