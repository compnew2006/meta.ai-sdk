
import React, { useState, useCallback, useRef, useEffect } from 'react';
import { VoiceOverStudioProject, VoiceOverHistoryItem, AudioFile } from '../types';
import { generateSpeech } from '../services/geminiService';
import { VOICES } from '../constants';
import { decodeBase64, decodeAudioData, pcmToWavBlob } from '../utils';

// --- ICONS ---
const PlayIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="currentColor" viewBox="0 0 20 20"><path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z" /></svg>;
const PauseIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="currentColor" viewBox="0 0 20 20"><path d="M5.75 3a.75.75 0 00-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 00.75-.75V3.75A.75.75 0 007.25 3h-1.5zM12.75 3a.75.75 0 00-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 00.75-.75V3.75a.75.75 0 00-.75-.75h-1.5z" /></svg>;
const DownloadIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" /></svg>;
const ReplayIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 110 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd" /></svg>;
const SoundWaveIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" clipRule="evenodd" /><path strokeLinecap="round" strokeLinejoin="round" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>;
const MaleIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" /></svg>;
const FemaleIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 2a1 1 0 00-1 1v1a1 1 0 002 0V3a1 1 0 00-1-1zM4 4a1 1 0 011 1v1a1 1 0 11-2 0V5a1 1 0 011-1zm12 0a1 1 0 011 1v1a1 1 0 11-2 0V5a1 1 0 011-1zM5.4 7.4a1 1 0 00-1.8 0l-.6 3.3a1 1 0 001.8 0l.6-3.3zm11 0a1 1 0 00-1.8 0l-.6 3.3a1 1 0 001.8 0l.6-3.3zM10 12a4 4 0 100 8 4 4 0 000-8zm0 6a2 2 0 110-4 2 2 0 010 4z" clipRule="evenodd" /></svg>;
const AllIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1.5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M3 5a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM3 15a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" /></svg>;
const PreviewPlayIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clipRule="evenodd" /></svg>;
const PreviewPauseIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" /></svg>;
const PreviewLoadingSpinner = () => <div className="animate-spin rounded-full h-5 w-5 border-t-2 border-b-2 border-[var(--color-accent)]"></div>;
// --- END ICONS ---


const VoiceOverStudio: React.FC<{
  project: VoiceOverStudioProject;
  setProject: React.Dispatch<React.SetStateAction<VoiceOverStudioProject>>;
}> = ({ project, setProject }) => {

    const audioContextRef = useRef<AudioContext | null>(null);
    const audioSourceRef = useRef<AudioBufferSourceNode | null>(null);
    const previewAudioSourceRef = useRef<AudioBufferSourceNode | null>(null);

    const styleSuggestions = [
      { label: "Default", value: "" },
      { label: "Saudi (Enthusiastic)", value: "in an enthusiastic Saudi dialect" },
      { label: "Egyptian (Enthusiastic)", value: "in an enthusiastic Egyptian dialect" },
    ];

    // Initialize AudioContext on component mount
    useEffect(() => {
        if (!audioContextRef.current && typeof window !== 'undefined') {
            audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)({ sampleRate: 24000 });
        }
        // Cleanup on unmount
        return () => {
            audioSourceRef.current?.stop();
            audioSourceRef.current = null;
            previewAudioSourceRef.current?.stop();
            previewAudioSourceRef.current = null;
        };
    }, []);

    // Handle voice selection when filter changes to prevent invalid project
    useEffect(() => {
        const filteredVoices = project.voiceGenderFilter === 'All'
            ? VOICES
            : VOICES.filter(v => v.gender === project.voiceGenderFilter);
        
        const isCurrentVoiceInFilteredList = filteredVoices.some(v => v.value === project.selectedVoice);

        if (!isCurrentVoiceInFilteredList && filteredVoices.length > 0) {
            setProject(s => ({ ...s, selectedVoice: filteredVoices[0].value }));
        }
    }, [project.voiceGenderFilter, project.selectedVoice, setProject]);
    
    const handleGenerate = useCallback(async () => {
        if (!project.text.trim()) {
            setProject(s => ({ ...s, error: 'Please enter some text to generate audio.' }));
            return;
        }

        setProject(s => ({ ...s, isLoading: true, error: null, generatedAudio: null }));
        try {
            const audio = await generateSpeech(project.text, project.styleInstructions, project.selectedVoice);
            const newHistoryItem: VoiceOverHistoryItem = {
                audio,
                text: project.text,
                style: project.styleInstructions,
                voice: project.selectedVoice,
            };
            setProject(s => ({
                ...s,
                isLoading: false,
                generatedAudio: audio,
                history: [newHistoryItem, ...s.history]
            }));
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred.';
            setProject(s => ({ ...s, isLoading: false, error: errorMessage }));
        }
    }, [project.text, project.styleInstructions, project.selectedVoice, setProject]);

    const playAudio = useCallback(async (audioFile: AudioFile) => {
        if (!audioContextRef.current) return;
        
        // Stop any currently playing audio
        if (audioSourceRef.current) {
            audioSourceRef.current.stop();
        }

        setProject(s => ({ ...s, isPlaying: true }));

        const audioCtx = audioContextRef.current;
        const pcmBytes = decodeBase64(audioFile.base64);
        const audioBuffer = await decodeAudioData(pcmBytes, audioCtx, 24000, 1);
        
        const source = audioCtx.createBufferSource();
        source.buffer = audioBuffer;
        source.connect(audioCtx.destination);
        source.start();
        
        source.onended = () => {
            setProject(s => ({ ...s, isPlaying: false }));
            audioSourceRef.current = null;
        };
        audioSourceRef.current = source;
    }, [setProject]);

    const stopAudio = useCallback(() => {
        if (audioSourceRef.current) {
            audioSourceRef.current.stop();
            // onended event will handle project update
        }
    }, []);
    
    const handlePreview = useCallback(async (e: React.MouseEvent, voiceName: string) => {
        e.stopPropagation();
    
        if (project.previewPlayingVoice === voiceName && previewAudioSourceRef.current) {
            previewAudioSourceRef.current.stop();
            return;
        }
    
        if (previewAudioSourceRef.current) {
            previewAudioSourceRef.current.stop();
        }
    
        setProject(s => ({ ...s, previewLoadingVoice: voiceName, previewPlayingVoice: null, error: null }));
        
        try {
            const sampleText = "Hello, this is a preview of the selected voice.";
            const audio = await generateSpeech(sampleText, '', voiceName);
            
            if (!audioContextRef.current) return;
    
            const audioCtx = audioContextRef.current;
            const pcmBytes = decodeBase64(audio.base64);
            const audioBuffer = await decodeAudioData(pcmBytes, audioCtx, 24000, 1);
            
            const source = audioCtx.createBufferSource();
            source.buffer = audioBuffer;
            source.connect(audioCtx.destination);
            
            setProject(s => s.previewLoadingVoice === voiceName ? { ...s, previewLoadingVoice: null, previewPlayingVoice: voiceName } : s);
            
            source.start();
            
            source.onended = () => {
                setProject(s => s.previewPlayingVoice === voiceName ? { ...s, previewPlayingVoice: null } : s);
                if (previewAudioSourceRef.current === source) {
                    previewAudioSourceRef.current = null;
                }
            };
            previewAudioSourceRef.current = source;
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred during preview.';
            setProject(s => ({ ...s, error: errorMessage, previewLoadingVoice: null }));
        }
    }, [project.previewPlayingVoice, setProject]);


    const handleDownload = useCallback(async (audioFile: AudioFile) => {
        if (!audioContextRef.current) return;

        const pcmBytes = decodeBase64(audioFile.base64);
        const audioBuffer = await decodeAudioData(pcmBytes, audioContextRef.current, 24000, 1);
        const pcmFloat32 = audioBuffer.getChannelData(0);
        const wavBlob = pcmToWavBlob(pcmFloat32, 24000, 1);

        const url = URL.createObjectURL(wavBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `SMART-Studio-voice-over-${Date.now()}.wav`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }, []);


    if (!project) {
        return (
            <main className="w-full max-w-4xl flex items-center justify-center gap-8 pt-8 pb-12 flex-grow">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
                <p className="text-[var(--color-text-secondary)]">Loading Voice Over Studio...</p>
            </main>
        );
    }

    const filteredVoices = project.voiceGenderFilter === 'All'
        ? VOICES
        : VOICES.filter(v => v.gender === project.voiceGenderFilter);
    
    return (
        <main className="w-full max-w-5xl flex flex-col gap-4 pt-4 pb-8 mx-auto">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Left: Controls */}
                <div className="glass-card rounded-2xl p-4 space-y-4">
                    <div>
                        <label htmlFor="vo-text" className="block text-sm font-medium text-[var(--color-text-medium)] mb-2">Text to Generate</label>
                        <textarea id="vo-text" value={project.text} onChange={e => setProject({...project, text: e.target.value})} rows={6} className="w-full glass-input rounded-md p-3 text-sm leading-relaxed" placeholder="Enter your script here..."/>
                    </div>
                    <div>
                        <label htmlFor="vo-style" className="block text-sm font-medium text-[var(--color-text-medium)] mb-2">Style Instructions / Dialect</label>
                        <div className="w-full glass-input rounded-md p-3 flex items-baseline focus-within:border-[rgba(var(--color-accent-rgb),0.5)] focus-within:shadow-[0_0_0_2px_rgba(var(--color-accent-rgb),0.2)]">
                            <span className="text-sm text-[var(--color-text-secondary)] mr-2 select-none">Read aloud in a warm and friendly tone:</span>
                            <textarea 
                                id="vo-style" 
                                value={project.styleInstructions} 
                                onChange={e => setProject({...project, styleInstructions: e.target.value})} 
                                rows={1} 
                                className="w-full bg-transparent focus:outline-none text-sm text-[var(--color-text-base)] resize-none" 
                                placeholder="e.g., in an enthusiastic dialect..."
                            />
                        </div>
                        <div className="pt-2">
                            <div className="flex flex-wrap items-center gap-2">
                                <span className="text-xs text-[var(--color-text-muted)]">Suggestions:</span>
                                {styleSuggestions.map((suggestion, index) => (
                                    <button
                                    key={index}
                                    onClick={() => setProject(s => ({ ...s, styleInstructions: suggestion.value }))}
                                    className="text-xs px-3 py-1.5 bg-[rgba(var(--color-text-base-rgb,229,231,206),0.05)] hover:bg-[rgba(var(--color-text-base-rgb,229,231,206),0.1)] rounded-full transition-colors text-[var(--color-text-secondary)] hover:text-[var(--color-text-base)]"
                                    >
                                    {suggestion.label}
                                    </button>
                                ))}
                            </div>
                        </div>
                    </div>
                    <div>
                        <div className="flex justify-between items-center mb-2">
                            <label className="block text-sm font-medium text-[var(--color-text-medium)]">Choose a Voice</label>
                            <div className="flex items-center gap-1 p-0.5 bg-black/20 rounded-full">
                                {(['All', 'Male', 'Female'] as const).map(filter => (
                                    <button
                                        key={filter}
                                        onClick={() => setProject(s => ({...s, voiceGenderFilter: filter}))}
                                        className={`flex items-center text-xs px-3 py-1 rounded-full transition-colors ${project.voiceGenderFilter === filter ? 'bg-[var(--color-accent)] text-white' : 'text-[var(--color-text-secondary)] hover:text-white'}`}
                                    >
                                        {filter === 'Male' && <MaleIcon />}
                                        {filter === 'Female' && <FemaleIcon />}
                                        {filter === 'All' && <AllIcon />}
                                        {filter}
                                    </button>
                                ))}
                            </div>
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                            {filteredVoices.map(voice => (
                                <div
                                    key={voice.value}
                                    onClick={() => {
                                        if (previewAudioSourceRef.current) {
                                            previewAudioSourceRef.current.stop();
                                        }
                                        setProject(s => ({ ...s, selectedVoice: voice.value }));
                                    }}
                                    role="button"
                                    tabIndex={0}
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter' || e.key === ' ') {
                                            e.preventDefault();
                                            if (previewAudioSourceRef.current) {
                                                previewAudioSourceRef.current.stop();
                                            }
                                            setProject(s => ({ ...s, selectedVoice: voice.value }));
                                        }
                                    }}
                                    className={`w-full text-left p-3 rounded-lg border-2 transition-all cursor-pointer ${project.selectedVoice === voice.value ? 'border-[var(--color-accent)] bg-[rgba(var(--color-accent-rgb),0.2)]' : 'border-transparent bg-black/20 hover:border-[rgba(var(--color-accent-rgb),0.5)]'}`}
                                >
                                    <div className="flex justify-between items-center gap-2">
                                        <div>
                                            <p className="font-semibold text-[var(--color-text-base)]">{voice.label}</p>
                                            <p className="text-xs text-[var(--color-text-secondary)]">{voice.description}</p>
                                        </div>
                                        <button 
                                            onClick={(e) => handlePreview(e, voice.value)}
                                            className="p-2 rounded-full hover:bg-[rgba(var(--color-text-base-rgb),0.1)] transition-colors z-10 relative flex-shrink-0"
                                            aria-label={`Preview voice ${voice.label}`}
                                        >
                                            {project.previewLoadingVoice === voice.value ? <PreviewLoadingSpinner /> :
                                             project.previewPlayingVoice === voice.value ? <PreviewPauseIcon /> :
                                             <PreviewPlayIcon />}
                                        </button>
                                    </div>
                                </div>
                            ))}
                             {filteredVoices.length === 0 && (
                                <p className="col-span-full text-center text-sm text-[var(--color-text-muted)] py-4">
                                    No {project.voiceGenderFilter.toLowerCase()} voices available.
                                </p>
                            )}
                        </div>
                    </div>

                    <button
                        onClick={handleGenerate}
                        disabled={project.isLoading || !project.text.trim()}
                        className="w-full bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-dark)] hover:from-[var(--color-accent-dark)] hover:to-[var(--color-accent-darker)] text-[var(--color-text-base)] font-bold py-3 px-8 rounded-lg text-lg transition-all duration-300 ease-in-out disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-2xl hover:shadow-[var(--color-accent)]/20 disabled:shadow-none transform hover:-translate-y-1 disabled:transform-none"
                    >
                        {project.isLoading ? 'Generating...' : 'Generate Audio'}
                    </button>
                    
                    {project.error && <div className="bg-[rgba(var(--color-accent-rgb),0.2)] border border-[rgba(var(--color-accent-rgb),0.5)] text-[var(--color-accent-light)] px-4 py-3 rounded-lg text-sm" role="alert">{project.error}</div>}
                </div>

                {/* Right: Results & History */}
                <div className="flex flex-col gap-4">
                    <div className="glass-card rounded-2xl p-4">
                        <h3 className="text-lg font-bold text-[var(--color-text-base)] mb-4">Result</h3>
                        <div className="min-h-[120px] flex items-center justify-center bg-black/20 rounded-lg p-4">
                            {project.isLoading && <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>}
                            {!project.isLoading && !project.generatedAudio && <p className="text-[var(--color-text-muted)] text-sm">Your generated audio will appear here.</p>}
                            {!project.isLoading && project.generatedAudio && (
                                <div className="w-full flex flex-col sm:flex-row items-center gap-4">
                                    <button onClick={() => project.isPlaying ? stopAudio() : playAudio(project.generatedAudio as AudioFile)} className="flex-shrink-0 w-14 h-14 flex items-center justify-center rounded-full bg-[var(--color-accent)] text-white hover:bg-[var(--color-accent-dark)] transition-colors">
                                        {project.isPlaying ? <PauseIcon /> : <PlayIcon />}
                                    </button>
                                    <div className="flex-grow w-full flex flex-col gap-2">
                                        <p className="text-sm text-center sm:text-left text-[var(--color-text-base)] font-medium">Playback</p>
                                        <button onClick={() => handleDownload(project.generatedAudio as AudioFile)} className="w-full flex items-center justify-center text-sm px-3 py-2 rounded-md bg-[rgba(var(--color-text-base-rgb,229,231,206),0.1)] hover:bg-[rgba(var(--color-text-base-rgb,229,231,206),0.2)] text-[var(--color-text-base)] transition-colors font-semibold">
                                            <DownloadIcon /> Download .WAV
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                    <div className="glass-card rounded-2xl p-4 flex-grow flex flex-col">
                        <h3 className="text-lg font-bold text-[var(--color-text-base)] mb-2">History</h3>
                        {project.history.length === 0 ? (
                            <div className="flex-grow flex items-center justify-center">
                                <p className="text-center text-sm text-[var(--color-text-secondary)] py-8">Your audio history is empty.</p>
                            </div>
                        ) : (
                            <div className="space-y-2 max-h-96 overflow-y-auto suggestions-scrollbar pr-2">
                                {project.history.map((item, index) => (
                                    <div key={index} className="flex items-center gap-3 p-2 bg-black/20 rounded-lg">
                                       <button onClick={() => playAudio(item.audio)} className="w-8 h-8 flex-shrink-0 flex items-center justify-center rounded-full bg-[rgba(var(--color-accent-rgb),0.3)] text-white hover:bg-[rgba(var(--color-accent-rgb),0.5)] transition-colors"><PlayIcon /></button>
                                       <div className="flex-1 overflow-hidden">
                                          <p className="text-sm text-[var(--color-text-base)] truncate" title={item.text}>"{item.text}"</p>
                                          <p className="text-xs text-[var(--color-text-muted)]">{item.voice} - {item.style || 'Default Style'}</p>
                                       </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </main>
    );
};

export default VoiceOverStudio;
