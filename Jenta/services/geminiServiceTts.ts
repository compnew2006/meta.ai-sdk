// services/geminiServiceTts.ts
//
// The ONLY part of Jenta that still calls Google Gemini directly. Meta AI has
// no text-to-speech endpoint, so generateSpeech stays here (hybrid approach).
// All other AI moved to ./aiService (via the metaai-go REST API).
//
// This file is intentionally tiny: one function, one model, one cloud key.
// The key is read from VITE_TTS_API_KEY (preferred) or falls back to the
// legacy process.env.API_KEY / GEMINI_API_KEY that vite.config.ts may define.

import { GoogleGenAI, Modality } from '@google/genai';
import { AudioFile } from '../types';

const TTS_KEY =
  (import.meta.env.VITE_TTS_API_KEY as string | undefined) ||
  (typeof process !== 'undefined' && process.env ? (process.env.API_KEY as string) || (process.env.GEMINI_API_KEY as string) : '') ||
  '';

if (!TTS_KEY) {
  // Defer the hard error until generateSpeech is actually called, so the rest
  // of the app (which no longer needs Gemini) still loads.
  console.warn(
    'geminiServiceTts: no VITE_TTS_API_KEY / API_KEY set — VoiceOverStudio will fail at call time.',
  );
}

const ai = new GoogleGenAI({ apiKey: TTS_KEY });

export async function generateSpeech(
  text: string,
  styleInstructions: string,
  voiceName: string,
): Promise<AudioFile> {
  const model = 'gemini-3.1-flash-preview-tts';
  const prompt = `Speak the following text ${styleInstructions ? '(' + styleInstructions + ')' : ''}: ${text}`;

  try {
    const response = await ai.models.generateContent({
      model,
      contents: [{ parts: [{ text: prompt }] }],
      config: {
        responseModalities: [Modality.AUDIO],
        speechConfig: {
          voiceConfig: {
            prebuiltVoiceConfig: { voiceName },
          },
        },
      },
    });

    const base64Audio = response.candidates?.[0]?.content?.parts?.[0]?.inlineData?.data;
    if (!base64Audio) {
      throw new Error('No audio data returned from the model.');
    }

    return {
      base64: base64Audio,
      name: `voiceover-${Date.now()}.wav`,
    };
  } catch (error) {
    console.error('Error generating speech:', error);
    throw error;
  }
}
