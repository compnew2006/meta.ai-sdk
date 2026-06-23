
import React, { useState, useEffect, useRef } from 'react';

interface BannerSliderProps {
  images: string[];
}

const BannerSlider: React.FC<BannerSliderProps> = ({ images }) => {
    const [currentIndex, setCurrentIndex] = useState(0);
    const timeoutRef = useRef<number | null>(null);

    // Reset index if images change drastically (e.g., all deleted)
    useEffect(() => {
        if (currentIndex >= images.length && images.length > 0) {
            setCurrentIndex(0);
        }
    }, [images.length, currentIndex]);

    useEffect(() => {
        if (images.length > 1) {
            timeoutRef.current = window.setInterval(() => {
                setCurrentIndex((prev) => (prev + 1) % images.length);
            }, 10000); // 10 Seconds
        }
        return () => {
            if (timeoutRef.current) clearInterval(timeoutRef.current);
        };
    }, [images.length]);

    if (images.length === 0) {
        return null; 
    }

    return (
        <div className="w-full relative group mb-6 rounded-2xl overflow-hidden shadow-2xl bg-black/20">
            <div className="relative w-full pt-[40%] sm:pt-[30%] md:pt-[25%] overflow-hidden">
                <div 
                    className="absolute inset-0 flex transition-transform duration-700 ease-in-out"
                    style={{ transform: `translateX(-${currentIndex * 100}%)` }}
                >
                    {images.map((img, idx) => {
                        const isVideo = img.startsWith('data:video');
                        return (
                            <div key={idx} className="w-full h-full flex-shrink-0 relative bg-black">
                                {isVideo ? (
                                    <video 
                                        src={img} 
                                        autoPlay 
                                        muted 
                                        loop 
                                        playsInline 
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <img 
                                        src={img.startsWith('data:') ? img : `data:image/jpeg;base64,${img}`} 
                                        alt={`Banner ${idx + 1}`} 
                                        className="w-full h-full object-cover"
                                    />
                                )}
                                <div className="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-transparent pointer-events-none" />
                            </div>
                        );
                    })}
                </div>
                
                {/* Dots Indicator */}
                {images.length > 1 && (
                    <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 flex gap-2 z-10">
                        {images.map((_, idx) => (
                            <button
                                key={idx}
                                onClick={() => setCurrentIndex(idx)}
                                className={`h-1.5 rounded-full transition-all duration-300 shadow-sm ${idx === currentIndex ? 'bg-white w-6' : 'bg-white/40 w-2 hover:bg-white/80'}`}
                                aria-label={`Go to slide ${idx + 1}`}
                            />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default BannerSlider;
