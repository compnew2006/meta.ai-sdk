
import React, { useState, useRef, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { expandImage } from '../services/geminiService';
import { ImageFile } from '../types';

interface ImageCropperProps {
  file: File;
  onConfirm: (editedFile: File) => void;
  onCancel: () => void;
}

type AspectRatio = {
    label: string;
    value: number; // width / height
    icon: React.ReactNode;
};

const RATIOS: AspectRatio[] = [
    { 
        label: '1:1', 
        value: 1, 
        icon: <div className="w-5 h-5 border-2 border-current rounded-sm"></div> 
    },
    { 
        label: '4:5', 
        value: 0.8, 
        icon: <div className="w-4 h-5 border-2 border-current rounded-sm"></div> 
    },
    { 
        label: '16:9', 
        value: 1.7777, 
        icon: <div className="w-6 h-3.5 border-2 border-current rounded-sm"></div> 
    },
    { 
        label: '9:16', 
        value: 0.5625, 
        icon: <div className="w-3.5 h-6 border-2 border-current rounded-sm"></div> 
    },
    { 
        label: '4:3', 
        value: 1.3333, 
        icon: <div className="w-5 h-4 border-2 border-current rounded-sm"></div> 
    },
    { 
        label: 'Free', 
        value: 0, 
        icon: <div className="w-5 h-5 border-2 border-dashed border-current rounded-sm"></div> 
    }
];

// Icons
const CheckIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>;
const XIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>;
const ZoomInIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7" /></svg>;
const ZoomOutIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM13 10H7" /></svg>;
const ExpandIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M4 8V4m0 0h4M4 4l5 5m11 5l-5-5m5 5v-4m0 4h-4M4 16v4m0 0h4m-4 0l5-5m11 1V4" /></svg>;
const CropIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>;
const MagicIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M5 2a1 1 0 011 1v1h1a1 1 0 010 2H6v1a1 1 0 11-2 0V6H3a1 1 0 110-2h1V3a1 1 0 011-1zm12 10a1 1 0 011 1v1h1a1 1 0 110 2h-1v1a1 1 0 11-2 0v-1h-1a1 1 0 110-2h1v-1a1 1 0 011-1zM10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm0 14a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM4.156 5.156a1 1 0 011.414-1.414l.707.707a1 1 0 01-1.414 1.414l-.707-.707zm11.314 11.314a1 1 0 011.414-1.414l.707.707a1 1 0 01-1.414 1.414l-.707-.707zm-8.485 2.122a1 1 0 01-1.414 1.414l-.707-.707a1 1 0 011.414-1.414l.707.707zM15.844 5.156a1 1 0 011.414 1.414l-.707.707a1 1 0 01-1.414-1.414l.707-.707z" clipRule="evenodd" /></svg>;

const ImageCropper: React.FC<ImageCropperProps> = ({ file, onConfirm, onCancel }) => {
    const [activeTool, setActiveTool] = useState<'crop' | 'expand'>('crop');
    const [selectedRatio, setSelectedRatio] = useState<number>(1);
    const [imageSrc, setImageSrc] = useState<string | null>(null);
    const [isGenerating, setIsGenerating] = useState(false);
    const [expandPrompt, setExpandPrompt] = useState("Seamlessly fill the empty space to extend the background naturally.");
    const [error, setError] = useState<string | null>(null);

    // Image State
    // We track offset (x, y) relative to center of viewport
    const [offset, setOffset] = useState({ x: 0, y: 0 });
    const [zoom, setZoom] = useState(1);
    
    // Interaction Refs
    const isDraggingRef = useRef(false);
    const lastPointerRef = useRef({ x: 0, y: 0 });
    const containerRef = useRef<HTMLDivElement>(null);
    const imgRef = useRef<HTMLImageElement>(null);
    const naturalSizeRef = useRef({ w: 0, h: 0 });

    useEffect(() => {
        const objectUrl = URL.createObjectURL(file);
        setImageSrc(objectUrl);
        document.body.style.overflow = 'hidden'; // Lock background scroll
        return () => {
            URL.revokeObjectURL(objectUrl);
            document.body.style.overflow = 'auto';
        };
    }, [file]);

    // Reset when ratio or tool changes
    useEffect(() => {
        setOffset({ x: 0, y: 0 });
        setZoom(1);
    }, [selectedRatio, activeTool]);

    // Calculate Viewport Size based on container and ratio
    const getViewportRect = () => {
        if (!containerRef.current) return { width: 0, height: 0, left: 0, top: 0 };
        
        const containerW = containerRef.current.clientWidth;
        const containerH = containerRef.current.clientHeight;
        const padding = 40; // Space around the crop box

        const availableW = containerW - (padding * 2);
        const availableH = containerH - (padding * 2);

        let vpW, vpH;

        // If ratio is 0 (Free), we default to image ratio or square, 
        // but for simplicity in this strict editor, let's treat Free as fitting the image initial ratio
        const effectiveRatio = selectedRatio === 0 
            ? (naturalSizeRef.current.w / naturalSizeRef.current.h || 1) 
            : selectedRatio;

        if (availableW / effectiveRatio <= availableH) {
            vpW = availableW;
            vpH = availableW / effectiveRatio;
        } else {
            vpH = availableH;
            vpW = availableH * effectiveRatio;
        }

        return {
            width: vpW,
            height: vpH,
            left: (containerW - vpW) / 2,
            top: (containerH - vpH) / 2
        };
    };

    const handlePointerDown = (e: React.PointerEvent) => {
        e.preventDefault();
        isDraggingRef.current = true;
        lastPointerRef.current = { x: e.clientX, y: e.clientY };
    };

    const handlePointerMove = (e: React.PointerEvent) => {
        if (!isDraggingRef.current) return;
        e.preventDefault();

        const dx = e.clientX - lastPointerRef.current.x;
        const dy = e.clientY - lastPointerRef.current.y;
        lastPointerRef.current = { x: e.clientX, y: e.clientY };

        setOffset(prev => ({
            x: prev.x + dx,
            y: prev.y + dy
        }));
    };

    const handlePointerUp = () => {
        isDraggingRef.current = false;
        if (activeTool === 'crop') {
            snapToBounds();
        }
    };

    const handleWheel = (e: React.WheelEvent) => {
        e.stopPropagation();
        const delta = -e.deltaY * 0.001; 
        const newZoom = Math.max(0.1, Math.min(zoom + delta, 5));
        setZoom(newZoom);
    };

    // Correct bounds to ensure image covers viewport in Crop mode
    const snapToBounds = () => {
        if (!naturalSizeRef.current.w) return;
        // Strict boundary logic would go here.
        // For now, we rely on the visual feedback of the dark overlay.
    };

    const getRenderData = () => {
        const viewport = getViewportRect();
        // Calculate Image Scale to Fit Viewport initially
        const imgAspect = naturalSizeRef.current.w / naturalSizeRef.current.h;
        const vpAspect = viewport.width / viewport.height;

        let baseScale;
        if (activeTool === 'crop') {
             // Contain base scale ensures user can see whole image initially
             if (imgAspect > vpAspect) {
                baseScale = viewport.width / naturalSizeRef.current.w;
             } else {
                baseScale = viewport.height / naturalSizeRef.current.h;
             }
        } else {
             // Expand mode: Start smaller (70% of fit) to see space
             if (imgAspect > vpAspect) {
                baseScale = (viewport.width * 0.7) / naturalSizeRef.current.w;
             } else {
                baseScale = (viewport.height * 0.7) / naturalSizeRef.current.h;
             }
        }

        const currentScale = baseScale * zoom;
        
        return {
            viewport,
            baseScale,
            currentScale
        };
    };

    const renderData = getRenderData();

    const generateResult = async (mode: 'crop' | 'expand') => {
        if (!imgRef.current || !naturalSizeRef.current.w) return;

        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        // High resolution output
        const targetWidth = 2048;
        const targetHeight = targetWidth / (renderData.viewport.width / renderData.viewport.height);
        
        canvas.width = targetWidth;
        canvas.height = targetHeight;

        // Background
        if (mode === 'expand') {
            ctx.fillStyle = '#FFFFFF'; 
            ctx.fillRect(0, 0, canvas.width, canvas.height);
        } else {
            ctx.fillStyle = '#000000';
            ctx.fillRect(0, 0, canvas.width, canvas.height);
        }
        
        const scaleFactor = targetWidth / renderData.viewport.width;
        
        const imgRenderW = naturalSizeRef.current.w * renderData.currentScale;
        const imgRenderH = naturalSizeRef.current.h * renderData.currentScale;

        // Calculate position relative to canvas center
        const drawX = (canvas.width - (imgRenderW * scaleFactor)) / 2 + (offset.x * scaleFactor);
        const drawY = (canvas.height - (imgRenderH * scaleFactor)) / 2 + (offset.y * scaleFactor);
        const drawW = imgRenderW * scaleFactor;
        const drawH = imgRenderH * scaleFactor;

        ctx.imageSmoothingEnabled = true;
        ctx.imageSmoothingQuality = 'high';
        
        ctx.drawImage(imgRef.current, drawX, drawY, drawW, drawH);

        if (mode === 'crop') {
            canvas.toBlob((blob) => {
                if (blob) {
                    onConfirm(new File([blob], file.name, { type: 'image/jpeg' }));
                }
            }, 'image/jpeg', 0.95);
        } else {
            setIsGenerating(true);
            const base64 = canvas.toDataURL('image/png').split(',')[1];
            const imageFile: ImageFile = { base64, mimeType: 'image/png', name: 'expand_input.png' };
            try {
                const result = await expandImage(imageFile, expandPrompt);
                const byteString = atob(result.base64);
                const ab = new ArrayBuffer(byteString.length);
                const ia = new Uint8Array(ab);
                for (let i = 0; i < byteString.length; i++) ia[i] = byteString.charCodeAt(i);
                onConfirm(new File([ab], 'expanded.png', { type: 'image/png' }));
            } catch (err) {
                setError(err instanceof Error ? err.message : "Expansion failed");
                setIsGenerating(false);
            }
        }
    };

    if (!imageSrc) return null;

    // Use CreatePortal to render at document root level, ensuring it covers everything
    return createPortal(
        <div className="fixed inset-0 z-[99999] bg-[#090909] text-white flex flex-col animate-in fade-in duration-200">
            
            {/* 1. Header Toolbar */}
            <div className="h-16 flex items-center justify-between px-6 border-b border-white/10 bg-[#121212] flex-shrink-0 z-50">
                <div className="flex items-center gap-1">
                     <button onClick={onCancel} className="p-2 hover:bg-white/10 rounded-full transition-colors text-white/60 hover:text-white">
                        <XIcon />
                     </button>
                     <span className="font-semibold ml-2 hidden sm:block">Edit Image</span>
                </div>

                <div className="flex bg-black/40 rounded-lg p-1 border border-white/10">
                    <button 
                        onClick={() => setActiveTool('crop')}
                        className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-all ${activeTool === 'crop' ? 'bg-[var(--color-accent)] text-white shadow-md' : 'text-white/60 hover:text-white hover:bg-white/5'}`}
                    >
                        <CropIcon /> <span className="hidden sm:inline">Crop</span>
                    </button>
                    <button 
                        onClick={() => setActiveTool('expand')}
                        className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-all ${activeTool === 'expand' ? 'bg-[var(--color-accent)] text-white shadow-md' : 'text-white/60 hover:text-white hover:bg-white/5'}`}
                    >
                        <ExpandIcon /> <span className="hidden sm:inline">Expand (AI)</span>
                    </button>
                </div>

                <button 
                    onClick={() => generateResult(activeTool)}
                    disabled={isGenerating}
                    className="flex items-center gap-2 px-6 py-2 bg-white text-black font-bold rounded-full hover:bg-gray-200 transition-colors disabled:opacity-50 text-sm sm:text-base"
                >
                    {isGenerating ? (
                        <div className="w-5 h-5 border-2 border-black border-t-transparent rounded-full animate-spin"></div>
                    ) : (
                        <CheckIcon />
                    )}
                    {activeTool === 'crop' ? 'Done' : 'Generate'}
                </button>
            </div>

            {/* 2. Main Canvas / Viewport */}
            <div className="flex-grow relative overflow-hidden bg-[#050505] cursor-move select-none touch-none flex items-center justify-center">
                 <div 
                    ref={containerRef}
                    className="absolute inset-4 sm:inset-8 md:inset-12 flex items-center justify-center"
                    onPointerDown={handlePointerDown}
                    onPointerMove={handlePointerMove}
                    onPointerUp={handlePointerUp}
                    onPointerLeave={handlePointerUp}
                    onWheel={handleWheel}
                 >
                     {/* The Viewport Frame (Fixed Center) */}
                     <div 
                        style={{ 
                            width: renderData.viewport.width, 
                            height: renderData.viewport.height,
                            boxShadow: '0 0 0 9999px rgba(0, 0, 0, 0.85)' // The dark overlay
                        }} 
                        className={`relative z-20 pointer-events-none transition-all duration-300 border-2 border-white/30 ${activeTool === 'expand' ? 'border-dashed border-[var(--color-accent)]/70' : ''}`}
                     >
                         {/* Grid Lines for Crop */}
                         {activeTool === 'crop' && (
                             <div className="absolute inset-0 grid grid-cols-3 grid-rows-3 opacity-40">
                                 <div className="border-r border-b border-white/50"></div>
                                 <div className="border-r border-b border-white/50"></div>
                                 <div className="border-b border-white/50"></div>
                                 <div className="border-r border-b border-white/50"></div>
                                 <div className="border-r border-b border-white/50"></div>
                                 <div className="border-b border-white/50"></div>
                                 <div className="border-r border-white/50"></div>
                                 <div className="border-r border-white/50"></div>
                             </div>
                         )}
                         {/* Checkerboard for Expand */}
                         {activeTool === 'expand' && (
                            <div className="absolute inset-0 opacity-20 -z-10" 
                                style={{
                                    backgroundImage: `linear-gradient(45deg, #444 25%, transparent 25%), linear-gradient(-45deg, #444 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #444 75%), linear-gradient(-45deg, transparent 75%, #444 75%)`,
                                    backgroundSize: `20px 20px`,
                                    backgroundPosition: `0 0, 0 10px, 10px -10px, -10px 0px`
                                }}
                            />
                         )}
                     </div>

                     {/* The Movable Image (Behind the frame's shadow but logically in same container) */}
                     <img 
                        ref={imgRef}
                        src={imageSrc}
                        alt="Crop target"
                        className="absolute max-w-none max-h-none pointer-events-none z-10 transition-transform duration-75 ease-out"
                        style={{
                            width: naturalSizeRef.current.w,
                            height: naturalSizeRef.current.h,
                            transform: `translate(${offset.x}px, ${offset.y}px) scale(${renderData.currentScale})`,
                            opacity: isGenerating ? 0.5 : 1
                        }}
                        onLoad={(e) => {
                             naturalSizeRef.current = { w: e.currentTarget.naturalWidth, h: e.currentTarget.naturalHeight };
                             setZoom(1.0001); // Force re-render
                        }}
                     />
                 </div>
                 
                 {isGenerating && (
                    <div className="absolute inset-0 z-50 flex flex-col items-center justify-center pointer-events-auto">
                        <div className="bg-black/80 px-8 py-6 rounded-2xl border border-white/10 backdrop-blur-md flex flex-col items-center">
                            <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-[var(--color-accent)] mb-3"></div>
                            <p className="text-white font-medium">Generating AI Fill...</p>
                        </div>
                    </div>
                 )}

                 {error && (
                    <div className="absolute top-4 left-1/2 -translate-x-1/2 bg-red-500/90 text-white px-6 py-2 rounded-full shadow-xl text-sm font-medium animate-in slide-in-from-top-4 z-50">
                        {error}
                        <button onClick={() => setError(null)} className="ml-3 font-bold opacity-80 hover:opacity-100">✕</button>
                    </div>
                 )}
            </div>

            {/* 3. Footer Controls */}
            <div className="bg-[#121212] border-t border-white/10 flex-shrink-0 flex flex-col z-50">
                
                {/* Zoom Slider Bar */}
                <div className="flex items-center justify-center gap-4 px-8 py-4 border-b border-white/5">
                    <ZoomOutIcon />
                    <input 
                        type="range"
                        min={0.1}
                        max={activeTool === 'crop' ? 3 : 1.5} // Limit expand zoom out
                        step={0.01}
                        value={zoom}
                        onChange={(e) => setZoom(parseFloat(e.target.value))}
                        className="w-full max-w-md h-1.5 bg-white/20 rounded-lg appearance-none cursor-pointer accent-[var(--color-accent)] hover:bg-white/30 transition-colors"
                    />
                    <ZoomInIcon />
                    <span className="text-xs font-mono text-white/50 w-10 text-right">{Math.round(zoom * 100)}%</span>
                </div>

                <div className="flex flex-col md:flex-row h-auto md:h-28">
                    {/* Aspect Ratios */}
                    <div className="flex-1 overflow-x-auto suggestions-scrollbar border-b md:border-b-0 md:border-r border-white/10 p-4">
                        <h4 className="text-[10px] font-bold text-white/40 uppercase mb-2 text-center md:text-left">Aspect Ratio</h4>
                        <div className="flex md:grid md:grid-cols-3 lg:grid-cols-6 gap-3 justify-center md:justify-start">
                            {RATIOS.map((ratio) => (
                                <button
                                    key={ratio.label}
                                    onClick={() => setSelectedRatio(ratio.value)}
                                    className={`flex-shrink-0 flex flex-col items-center justify-center gap-1.5 w-16 h-16 rounded-xl border transition-all ${
                                        selectedRatio === ratio.value 
                                        ? 'bg-white text-black border-white shadow-[0_0_15px_rgba(255,255,255,0.3)]' 
                                        : 'bg-white/5 border-transparent text-white/60 hover:bg-white/10 hover:text-white'
                                    }`}
                                >
                                    {ratio.icon}
                                    <span className="text-[10px] font-bold">{ratio.label}</span>
                                </button>
                            ))}
                        </div>
                    </div>

                    {/* Expand Prompt (Conditional) */}
                    {activeTool === 'expand' && (
                        <div className="flex-1 p-4 animate-in fade-in slide-in-from-right-4">
                             <h4 className="text-[10px] font-bold text-white/40 uppercase mb-2 flex items-center gap-2">
                                <MagicIcon /> AI Prompt
                             </h4>
                             <div className="relative">
                                <textarea
                                    value={expandPrompt}
                                    onChange={(e) => setExpandPrompt(e.target.value)}
                                    className="w-full h-16 bg-black/40 border border-white/10 rounded-lg p-3 text-sm text-white/90 placeholder-white/30 focus:outline-none focus:border-[var(--color-accent)] resize-none leading-relaxed"
                                />
                                <div className="absolute bottom-2 right-2 text-[10px] text-white/30">
                                    Describe the fill area
                                </div>
                             </div>
                        </div>
                    )}
                    
                    {/* Placeholder for balance if crop selected */}
                    {activeTool === 'crop' && (
                         <div className="flex-1 flex items-center justify-center text-white/20 text-sm italic p-4">
                            Drag image to position • Scroll to zoom
                         </div>
                    )}
                </div>
            </div>
        </div>,
        document.body
    );
};

export default ImageCropper;
