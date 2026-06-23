
import React, { useState, useEffect, useRef } from 'react';

const PaletteIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
    </svg>
);

const CheckIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
);

const SunIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fillRule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.706-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 14.95l.707-.707a1 1 0 10-1.414-1.414l-.707.707a1 1 0 001.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 100 2h1z" clipRule="evenodd" />
    </svg>
);

const MoonIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
    </svg>
);


interface Theme {
    id: string;
    name: string;
    icon: React.ReactNode;
}

const THEMES: Theme[] = [
    { id: 'dark', name: 'Dark Mode', icon: <MoonIcon /> },
    { id: 'light', name: 'Light Mode', icon: <SunIcon /> },
];

interface ThemeSwitcherProps {
    currentTheme: string;
    onThemeChange: (themeId: string) => void;
    dropUp?: boolean;
}

const ThemeSwitcher: React.FC<ThemeSwitcherProps> = ({ currentTheme, onThemeChange, dropUp = false }) => {
    const [isOpen, setIsOpen] = useState(false);
    const switcherRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (switcherRef.current && !switcherRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    const handleThemeSelect = (themeId: string) => {
        onThemeChange(themeId);
        setIsOpen(false);
    };

    return (
        <div className="relative" ref={switcherRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="bg-black/20 hover:bg-black/40 text-[var(--color-text-medium)] hover:text-[var(--color-text-base)] font-bold p-2 sm:p-3 rounded-full transition-all duration-300 ease-in-out shadow-lg hover:shadow-xl transform hover:scale-110"
                aria-label="Change theme"
                aria-haspopup="true"
                aria-expanded={isOpen}
            >
                <PaletteIcon />
            </button>
            {isOpen && (
                <div className={`absolute ${dropUp ? 'bottom-full mb-2' : 'top-full mt-2'} right-0 w-48 glass-card rounded-lg shadow-2xl p-2 z-50`}>
                    <ul className="flex flex-col gap-1">
                        {THEMES.map((theme) => (
                            <li key={theme.id}>
                                <button
                                    onClick={() => handleThemeSelect(theme.id)}
                                    className={`w-full flex items-center justify-between text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                        currentTheme === theme.id 
                                        ? 'bg-[rgba(var(--color-accent-rgb),0.3)] text-[var(--color-text-base)] font-semibold' 
                                        : 'text-[var(--color-text-medium)] hover:bg-[rgba(var(--color-text-base-rgb,229,231,206),0.1)] hover:text-[var(--color-text-base)]'
                                    }`}
                                >
                                    <div className="flex items-center gap-3">
                                        {theme.icon}
                                        {theme.name}
                                    </div>
                                    {currentTheme === theme.id && <CheckIcon />}
                                </button>
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
};

export default ThemeSwitcher;
