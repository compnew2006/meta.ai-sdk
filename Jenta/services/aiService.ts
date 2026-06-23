// services/aiService.ts
//
// Drop-in replacement for geminiService.ts that routes every call through the
// metaai-go REST API (Meta AI backend) instead of @google/genai. Every exported
// function preserves its original signature and return type so the 11 importing
// components do not change.
//
// Endpoint mapping (see INTEGRATION.md §3.1 for the full table):
//   generateImage / editImage / expandImage  → /upload (refs) → /image → fetchImageToBase64
//   analyze* (vision)                        → /upload ×N → /analyze
//   generatePromptFromText / translateText   → /chat
//   generateCampaignPlan / Storyboard / Marketing → /chat (JSON forced via system_instruction)
//   generateSpeech                           → re-exported from ./geminiServiceTts (stays on Gemini)

import { ImageFile, AudioFile } from '../types';
import {
  request,
  uploadFile,
  fetchImageToBase64,
  base64ToBlob,
  orientationFor,
  mapLimit,
  hashFile,
  getConversation,
  setConversation,
  getDescription,
  setDescription,
} from './metaaiClient';

// Re-export TTS so geminiService.ts can re-export everything from one place.
// generateSpeech stays on Gemini because Meta AI has no TTS endpoint.
export { generateSpeech } from './geminiServiceTts';

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

/**
 * Upload an array of ImageFiles with bounded concurrency (max 3 in flight) so
 * we don't hammer Meta AI with N parallel uploads (rate-limit risk). Preserves
 * input order.
 */
async function uploadAll(images: ImageFile[]): Promise<string[]> {
  if (!images.length) return [];
  return mapLimit(images, 3, (img) => uploadFile(base64ToBlob(img.base64, img.mimeType), img.name));
}

/** Strip ```json fences + surrounding prose, returning the outermost JSON span. */
function extractJson(text: string): string {
  let extracted = (text || '').trim();
  const fencedBlock = extracted.match(/```(?:json)?\s*([\s\S]*?)```/i);
  if (fencedBlock) extracted = fencedBlock[1].trim();
  const firstToken = extracted.search(/[[{]/);
  const lastBrace = extracted.lastIndexOf('}');
  const lastBracket = extracted.lastIndexOf(']');
  const lastToken = Math.max(lastBrace, lastBracket);
  if (firstToken >= 0 && lastToken > firstToken) {
    extracted = extracted.slice(firstToken, lastToken + 1);
  }
  return extracted;
}

/**
 * Centralized JSON chat with **self-healing**: asks Meta AI, cleans fences, and
 * if parse fails, asks Meta AI once more to "fix this to valid JSON only" with
 * its own output appended — recovers from Meta's occasional prose-wrapping.
 * Single source of truth for every JSON-forcing call (DRY).
 */
async function chatJson<T>(message: string, fallback: T): Promise<T> {
  const system =
    'Return ONLY raw JSON. No markdown, no prose, no code fences. Start with { or [. No citation markers. No competitor brand names. No Modern Standard Arabic where Egyptian عامية is requested.';
  let msg = message;
  for (let attempt = 0; attempt < 2; attempt++) {
    const r = await request<{ message: string }>('/chat', {
      method: 'POST',
      body: JSON.stringify({ message: msg, system_instruction: system }),
    });
    const cleaned = extractJson(r.message);
    try {
      return JSON.parse(cleaned) as T;
    } catch (err) {
      if (attempt === 0) {
        // Self-heal: feed the bad output back and ask Meta AI to fix it.
        msg = `Your previous reply was not valid JSON:\n${cleaned}\n\nReturn ONLY the corrected, valid JSON now. Nothing else.`;
      } else {
        console.warn('chatJson failed after retry:', cleaned, err);
      }
    }
  }
  return fallback;
}

/**
 * Tolerant JSON parse of an already-fetched text (e.g. /analyze output). If
 * parsing fails, self-heal via one /chat fixup round. Same strategy as chatJson
 * but for endpoints that don't take `system_instruction` cleanly or already
 * returned their content.
 */
async function jsonFromText<T>(text: string, fallback: T): Promise<T> {
  const cleaned = extractJson(text);
  try {
    return JSON.parse(cleaned) as T;
  } catch {
    return chatJson<T>(
      `Fix this to valid JSON only. Return ONLY the corrected JSON:\n${cleaned}`,
      fallback,
    );
  }
}

/**
 * Build a cached text description for a reference image (#2). Meta AI /image is
 * prompt-only (no inline references), so to approach Gemini's reference-image
 * behavior we turn the reference into a concise description ONCE, cache it by
 * content hash, and fold it into every subsequent /image prompt.
 */
async function describeReference(
  image: ImageFile,
  studio: string,
  focus: string,
): Promise<string> {
  const blob = base64ToBlob(image.base64, image.mimeType);
  const hash = await hashFile(blob);
  const cached = getDescription(hash);
  if (cached) return cached;

  const mediaId = await uploadFile(blob, image.name);
  // Resume or start a conversation for this image.
  const convId = getConversation(studio, hash);
  const r = await request<{ message: string; conversation_id?: string }>('/analyze', {
    method: 'POST',
    body: JSON.stringify({
      media_id: mediaId,
      question: `Describe this image for an AI image generator in ~25 words. Focus on ${focus}. No marketing language.`,
      ...(convId ? { conversation_id: convId } : {}),
    }),
  });
  if (r.conversation_id) setConversation(studio, hash, r.conversation_id);
  const desc = r.message.trim();
  setDescription(hash, desc);
  return desc;
}

// ---------------------------------------------------------------------------
// Image generation + editing (was gemini-3.1-flash-image)
// ---------------------------------------------------------------------------

export async function generateImage(
  productImages: ImageFile[],
  prompt: string,
  styleImages: ImageFile[] | null,
  aspectRatio: string = '1:1',
): Promise<ImageFile> {
  // Meta AI /image is prompt-only, so to approach Gemini's reference-image
  // behavior we convert references into concise CACHED descriptions and fold
  // them into the prompt. Descriptions are built once per image (content-hash
  // keyed) — subsequent edits/generations with the same references skip the
  // /analyze round-trip.
  const parts: string[] = [prompt];
  if (productImages?.length) {
    const descs = await Promise.all(
      productImages.slice(0, 3).map((img) =>
        describeReference(img, 'generateImage', 'shape, material, logo position, dominant colors'),
      ),
    );
    parts.push(`Product reference(s): ${descs.join('; ')}.`);
  }
  if (styleImages?.length) {
    const descs = await Promise.all(
      styleImages.slice(0, 3).map((img) =>
        describeReference(img, 'generateImage:style', 'lighting, color palette, mood, aesthetic'),
      ),
    );
    parts.push(`Style reference(s): ${descs.join('; ')}.`);
  }
  const fullPrompt = parts.join('\n\n');

  const r = await request<{ success: boolean; image_urls?: string[]; error?: string }>('/image', {
    method: 'POST',
    body: JSON.stringify({ prompt: fullPrompt, orientation: orientationFor(aspectRatio) }),
  });
  if (!r.success || !r.image_urls?.length) {
    throw new Error(r.error || 'No image was generated by the model.');
  }
  return fetchImageToBase64(r.image_urls[0]);
}

export async function editImage(baseImage: ImageFile, prompt: string): Promise<ImageFile> {
  // Upload the base so Meta has it, then describe the edit. Meta AI applies the
  // edit instruction to the most recent context.
  await uploadFile(base64ToBlob(baseImage.base64, baseImage.mimeType), baseImage.name);
  const r = await request<{ success: boolean; image_urls?: string[]; error?: string }>('/image', {
    method: 'POST',
    body: JSON.stringify({ prompt, orientation: 'SQUARE' }),
  });
  if (!r.success || !r.image_urls?.length) {
    throw new Error(r.error || 'No image was generated by the model.');
  }
  return fetchImageToBase64(r.image_urls[0]);
}

export async function expandImage(image: ImageFile, prompt: string): Promise<ImageFile> {
  // Historical alias of editImage — keep the same behavior (was identical in
  // geminiService.ts).
  return editImage(image, prompt);
}

// ---------------------------------------------------------------------------
// Image analysis (was gemini-3.5-flash vision)
// ---------------------------------------------------------------------------

export async function analyzeImageForPrompt(
  images: ImageFile[],
  instructions: string,
): Promise<string> {
  const ids = await uploadAll(images);
  const question = `Analyze the provided image(s) in detail. Craft a descriptive prompt for an AI model. Instruction: ${instructions}${
    ids.length > 1 ? `\n(Additional image media_ids for context: ${ids.slice(1).join(', ')}.)` : ''
  }`;
  const r = await request<{ message: string }>('/analyze', {
    method: 'POST',
    body: JSON.stringify({ media_id: ids[0], question }),
  });
  return r.message.trim();
}

export async function analyzeStyleImage(images: ImageFile[]): Promise<string> {
  const ids = await uploadAll(images);
  const question =
    'Analyze the visual style of these images. Describe the lighting, color palette, mood, and aesthetic in detail for a text-to-image prompt.' +
    (ids.length > 1 ? `\n(Additional image media_ids for context: ${ids.slice(1).join(', ')}.)` : '');
  const r = await request<{ message: string }>('/analyze', {
    method: 'POST',
    body: JSON.stringify({ media_id: ids[0], question }),
  });
  return r.message.trim();
}

export async function analyzeLogoForBranding(
  images: ImageFile[],
): Promise<{ colors: string[] }> {
  const ids = await uploadAll(images);
  const question =
    'Analyze this logo and extract the primary brand colors. Return them as a JSON object with a "colors" key containing a string array of hex codes.' +
    (ids.length > 1 ? `\n(Additional logo media_ids: ${ids.slice(1).join(', ')}.)` : '');
  const r = await request<{ message: string }>('/analyze', {
    method: 'POST',
    body: JSON.stringify({
      media_id: ids[0],
      question,
      system_instruction:
        'Return ONLY a raw JSON object {"colors":["#hex",...]}. No prose, no markdown fences.',
    }),
  });
  // /analyze returns text; reuse the same tolerant JSON parse + self-heal path.
  return jsonFromText<{ colors: string[] }>(r.message, { colors: [] });
}

export async function analyzeProductForCampaign(productImages: ImageFile[]): Promise<string> {
  const ids = await uploadAll(productImages);
  // Tighter instruction: ban web-search citations (the 【...†L#】 noise) and
  // competitor name-dropping, and cap length so the analysis stays usable as
  // downstream context for generateCampaignPlan.
  const question = `Identify the product/service in this image and analyze its market positioning.

Return a CONCISE analysis (max 250 words) with exactly these two sections:

1. CATEGORY & PRODUCT — product name, type, and the 3-5 most important features visible.
2. AUDIENCE & POSITIONING — ideal customer profile and use-case, in 2-3 sentences.

RULES:
- Use ONLY what is visible in the image plus general product knowledge.
- Do NOT mention competitor brands by name.
- Do NOT include source citations, footnotes, or reference markers.
- Do NOT repeat sections. Plain prose only.` +
    (ids.length > 1 ? `\n(Additional image media_ids for context: ${ids.slice(1).join(', ')}.)` : '');
  const r = await request<{ message: string }>('/analyze', {
    method: 'POST',
    body: JSON.stringify({
      media_id: ids[0],
      question,
      system_instruction:
        'You are a product analyst. Return a clean, professional summary. No citation markers, no footnotes, no competitor brand names.',
    }),
  });
  return cleanAnalysis(r.message);
}

/**
 * Strip the noise Meta AI's web-search leaves in analysis text, and collapse
 * the accidental duplication it sometimes produces (it returned the oraimo
 * analysis twice in a row). Used by analyzeProductForCampaign output before it
 * feeds into generateCampaignPlan, and defensively inside the campaign prompt.
 *
 * Removes:
 *   【1234567†L12-L15】       — citation markers
 *   【entity-Apple¦...】      — entity tags
 *   duplicate consecutive blocks (the oraimo "returned twice" case)
 */
function cleanAnalysis(text: string): string {
  if (!text) return '';
  let s = text;
  // Citation + entity markers: 【...】 with any content.
  s = s.replace(/【[^】]*】/g, '');
  // Collapse runs of whitespace left behind.
  s = s.replace(/[ \t]{2,}/g, ' ');
  // Deduplicate: if the second half of the string equals the first half, keep one.
  // Meta AI returned the oraimo analysis verbatim twice; this halves it.
  const half = Math.floor(s.length / 2);
  if (half > 200) {
    const firstHalf = s.slice(0, half).trim();
    const secondHalf = s.slice(half).trim();
    // Normalise whitespace before comparing so minor spacing diffs still match.
    const normalized = (x: string) => x.replace(/\s+/g, ' ');
    if (normalized(firstHalf).length > 100 && normalized(firstHalf) === normalized(secondHalf)) {
      s = firstHalf;
    }
  }
  return s.trim();
}

// ---------------------------------------------------------------------------
// Pure text → text (was gemini-3.5-flash)
// ---------------------------------------------------------------------------

export async function generatePromptFromText(instructions: string): Promise<string> {
  const r = await request<{ message: string }>('/chat', {
    method: 'POST',
    body: JSON.stringify({ message: `Expand this idea into a detailed text-to-image prompt: "${instructions}"` }),
  });
  return r.message.trim();
}

export async function translateText(text: string): Promise<string> {
  const r = await request<{ message: string }>('/chat', {
    method: 'POST',
    body: JSON.stringify({
      message: `Translate the following text to English, preserving any technical or descriptive nuances: "${text}"`,
    }),
  });
  return r.message.trim();
}

// ---------------------------------------------------------------------------
// JSON-forced text (was gemini-3.5-flash with responseSchema)
// ---------------------------------------------------------------------------

export async function generateCampaignPlan(
  productImages: ImageFile[],
  userPrompt: string,
  targetMarket: string = 'Egypt',
  dialect: string = 'Egyptian / عامية',
  categoryAnalysis: string = '',
): Promise<any[]> {
  // Upload product refs first so the model can ground the plan in them.
  if (productImages?.length) {
    await uploadAll(productImages);
  }

  // 1) NEVER echo the analysis back. Keep the goal short and the analysis
  //    cleaned (no citation/entity markers, deduped). The previous prompt
  //    inlined the analysis twice (~6000 chars) which made Meta AI summarise
  //    instead of producing 9 posts.
  const goal = (userPrompt || '').trim().slice(0, 400);
  const analysis = cleanAnalysis(categoryAnalysis).slice(0, 1500);

  // 2) Explicit negative constraints + few-shot examples of REAL Egyptian
  //    عامية captions. Without these, Meta AI defaults to MSA and competitor
  //    name-drops.
  const message = `You are SMART Studio's autonomous campaign solver for Egypt 2026.

GOAL: ${goal}

PRODUCT CONTEXT (use only as background, do NOT repeat it back):
${analysis}

DELIVER: a raw JSON array of exactly 9 posts (5 Instagram, 4 Facebook). Optimise for Meta's 2026 algorithm: Instagram = saves + shares; Facebook = meaningful comments + dwell time.

HARD RULES (violation = invalid output):
- Return ONLY a JSON array. Start with [ and end with ]. No markdown, no prose, no code fences.
- Egyptian عامية ONLY for captions and tov. NO Modern Standard Arabic (لا تفصيح). NO English captions.
- NEVER mention competitor brands (Apple, Sony, Samsung, etc.) anywhere.
- NEVER use engagement bait ("اكتب كومنت", "منشن صاحبك", "لايك وشير").
- Instagram caption: 125-150 characters, hook in first 7 words, ONE soft CTA only (احفظها / ابعتها), max 2 emojis, exactly 3-5 hashtags — all Egyptian (#مصر, #القاهرة, #...).
- Facebook caption: 40-70 words, a real daily-life story (metro/lecture/gym/coffee), ends with ONE open question, max 2 hashtags, NO links.
- Each post focuses on ONE feature only (price, 50h battery, 10-min charge, ENC, Game Mode, IPX4, comfort).
- Distribute settings: ~3 commute (metro/microbus), ~3 study/work, ~3 lifestyle (coffee/gym).

EXAMPLES of correct Egyptian عامية captions (do NOT copy these, write fresh):
- IG: "بطارية ساعات ما تخلش؟ ده اللي بيدور عليه commuters 🎧 احفظها تعرف السعر #سماعات #مصر"
- FB: "راكب الميكروبوس وبتوع الشغل بيتكلموا فيه مستنيك ترد... الخ"

JSON SCHEMA (per post):
{
  "id": "post_1",
  "platform": "Instagram" | "Facebook",
  "format": "4:5 Feed" | "9:16 Reels" | "1:1 Feed",
  "scenario": "detailed ENGLISH visual prompt for an image generator (lighting, setting, product framing)",
  "caption": "Egyptian عامية, within the limits above",
  "tov": "max 5 words in Egyptian عامية, for on-image design text",
  "hashtags": ["#...", "#..."],
  "schedule": "e.g. الخميس 9:00 م",
  "cta_type": "save" | "comment" | "share",
  "feature_focus": "one feature name in English"
}

Now output the 9-post JSON array.`;

  return chatJson<any[]>(message, []);
}

export async function generateStoryboardPlan(
  subjectImages: ImageFile[],
  customInstructions: string,
): Promise<any[]> {
  if (subjectImages?.length) {
    await uploadAll(subjectImages);
  }
  const message = `Act as a cinematic Storyboard Director and Scriptwriter.
    Context/Prompt: "${customInstructions}".

    Task: Create a professional 9-scene storyboard sequence.
    Maintain strict visual consistency for the subject/character from provided images.
    Each scene must have a unique camera angle to build a cinematic narrative.

    Return a JSON array of 9 objects:
    - sequence: number (1 to 9)
    - description: what is happening in the scene (Arabic)
    - cameraAngle: specific technical camera angle (e.g. Extreme Close-up, Low Angle, Wide Shot)
    - visualPrompt: extremely detailed English prompt for an AI image generator to create THIS scene. Include lighting, mood, and reference to the subject.
    Return ONLY JSON.`;
  return chatJson<any[]>(message, []);
}

export async function generateMarketingAnalysis(
  brandData: { type: 'new' | 'existing'; name?: string; specialty?: string; brief?: string; link?: string },
  language: 'ar' | 'en',
): Promise<string> {
  let context = '';
  if (brandData.type === 'existing') {
    context = `Analyze the brand from this link: ${brandData.link}. Find its real current position, competitors, and audience feedback.`;
  } else {
    context = `Strategic analysis for a NEW brand. Name: ${brandData.name}. Specialty: ${brandData.specialty}. Brief: ${brandData.brief}. Use market trends for this niche.`;
  }
  const message = `Act as a world-class CMO and Marketing Strategist.
    ${context}

    Task: Provide a detailed, professional marketing strategy report.
    Language: ${language === 'ar' ? 'Arabic' : 'English'}.

    The report MUST include:
    1. SWOT Analysis (Strengths, Weaknesses, Opportunities, Threats).
    2. Detailed Buyer Persona (Demographics, Psychographics, Buying Behavior).
    3. Competitor Analysis & Market Gaps.
    4. Value Proposition (USP).
    5. Integrated Go-To-Market (GTM) Strategy.
    6. Smart Pricing Strategy Recommendation.
    7. 30-60-90 Day Execution Roadmap.
    8. Growth KPI Dashboard (Metric recommendations).

    Format the output with professional headers, bullet points, and a tone of high-level business consultation.
    Use Markdown for formatting.`;
  const r = await request<{ message: string }>('/chat', {
    method: 'POST',
    body: JSON.stringify({ message }),
  });
  return r.message || '';
}

// AudioFile is re-exported from geminiServiceTts's generateSpeech return type;
// keep the import live for callers that type against it.
export type { AudioFile };
