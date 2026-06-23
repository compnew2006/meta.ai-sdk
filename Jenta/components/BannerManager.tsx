
import React from 'react';
import { resizeImage } from '../utils';

interface BannerManagerProps {
    isOpen: boolean;
    onClose: () => void;
    images: string[];
    onAddImage: (dataUri: string) => void;
    onRemoveImage: (index: number) => void;
}

const TrashIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
    </svg>
);

const PlusIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
    </svg>
);

const BannerManager: React.FC<BannerManagerProps> = ({ isOpen, onClose, images, onAddImage, onRemoveImage }) => {
    if (!isOpen) return null;

    const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            const file = e.target.files[0];
            
            const isVideo = file.type.startsWith('video/');
            const isGif = file.type === 'image/gif';

            try {
                if (isVideo || isGif) {
                    // For video and GIFs, we don't want to resize to preserve animation/data
                    // Just read as Data URL directly
                    const reader = new FileReader();
                    reader.onloadend = () => {
                        if (typeof reader.result === 'string') {
                            onAddImage(reader.result);
                        }
                    };
                    reader.readAsDataURL(file);
                } else if (file.type.startsWith('image/')) {
                    // Resize static images to save space
                    const resized = await resizeImage(file, 1920, 1080);
                    const reader = new FileReader();
                    reader.onloadend = () => {
                        if (typeof reader.result === 'string') {
                            onAddImage(reader.result);
                        }
                    };
                    reader.readAsDataURL(resized);
                }
            } catch (err) {
                console.error("Banner upload failed", err);
            }
        }
        e.target.value = '';
    };

    return (
        <div className="fixed inset-0 z-[100] flex items-end sm:items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onClick={onClose}>
            <div className="glass-card w-full max-w-xl rounded-2xl p-6 relative animate-in slide-in-from-bottom-10 fade-in duration-300" onClick={e => e.stopPropagation()}>
                <div className="flex justify-between items-center mb-4 border-b border-[var(--color-glass-border)] pb-3">
                    <div>
                        <h3 className="text-xl font-bold text-[var(--color-text-base)]">Manage Banners</h3>
                        <p className="text-xs text-[var(--color-text-secondary)]">Add or remove slides (Images, GIFs, or Videos).</p>
                    </div>
                    <button onClick={onClose} className="p-2 rounded-full hover:bg-[rgba(var(--color-text-base-rgb),0.1)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-base)] transition-colors">
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 max-h-[60vh] overflow-y-auto suggestions-scrollbar p-1">
                    {images.map((img, idx) => {
                        const isVideo = img.startsWith('data:video');
                        return (
                            <div key={idx} className="relative aspect-video rounded-lg overflow-hidden group border border-[var(--color-glass-border)] bg-black/50">
                                {isVideo ? (
                                    <video 
                                        src={img} 
                                        className="w-full h-full object-cover" 
                                        muted 
                                        loop 
                                        playsInline 
                                        onMouseOver={e => e.currentTarget.play()} 
                                        onMouseOut={e => e.currentTarget.pause()}
                                    />
                                ) : (
                                    <img 
                                        src={img.startsWith('data:') ? img : `data:image/jpeg;base64,${img}`} 
                                        alt={`Banner ${idx + 1}`} 
                                        className="w-full h-full object-cover"
                                    />
                                )}
                                <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                                    <button 
                                        onClick={() => onRemoveImage(idx)}
                                        className="p-2 rounded-full bg-red-500/80 text-white hover:bg-red-600 transition-transform hover:scale-110"
                                        title="Delete Slide"
                                    >
                                        <TrashIcon />
                                    </button>
                                </div>
                            </div>
                        );
                    })}
                    <label className="relative aspect-video rounded-lg border-2 border-dashed border-[rgba(var(--color-accent-rgb),0.3)] hover:border-[var(--color-accent)] flex flex-col items-center justify-center cursor-pointer transition-all bg-[rgba(var(--color-accent-rgb),0.05)] hover:bg-[rgba(var(--color-accent-rgb),0.1)] group">
                        <div className="text-[var(--color-text-secondary)] group-hover:text-[var(--color-accent)] transition-colors transform group-hover:scale-110 duration-300">
                            <PlusIcon />
                        </div>
                        <span className="text-xs mt-2 font-medium text-[var(--color-text-secondary)] group-hover:text-[var(--color-text-base)]">Add Slide</span>
                        <input type="file" className="hidden" accept="image/*,video/*" onChange={handleFileUpload} />
                    </label>
                </div>
            </div>
        </div>
    );
};

export default BannerManager;
