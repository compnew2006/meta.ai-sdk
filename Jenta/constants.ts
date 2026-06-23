
import { LightingStyle, CameraPerspective, AspectRatio, ControllerSlider, AppView } from './types';

/**
 * Single source of truth for the studio list. Drives the desktop nav, the
 * mobile nav, and the renderContent() switch — replacing three previously
 * unsynced hand-typed lists (which had drifted: Branding and Controller were
 * imported and had project state but were never reachable from the UI).
 *
 * `label`   = desktop nav text.
 * `short`   = mobile nav text (kept short for the horizontal pill bar).
 */
export const STUDIOS: { view: AppView; label: string; short: string }[] = [
  { view: 'creator_studio',      label: 'Creator Studio', short: 'Creator' },
  { view: 'plan_studio',         label: 'Plan (NEW)',      short: 'Plan' },
  { view: 'campaign_studio',     label: 'Campaign',        short: 'Campaign' },
  { view: 'storyboard_studio',   label: 'Storyboard',      short: 'Storyboard' },
  { view: 'marketing_studio',    label: 'Marketing',       short: 'Marketing' },
  { view: 'photoshoot_director', label: 'Photoshoot',      short: 'Photoshoot' },
  { view: 'prompt_studio',       label: 'Prompt',          short: 'Prompt' },
  { view: 'edit_studio',         label: 'Edit (PRO)',      short: 'Edit' },
  { view: 'video_studio',        label: 'Video Studio',    short: 'Video' },
  { view: 'voice_over_studio',   label: 'Voice Over',      short: 'Voice' },
  { view: 'branding_studio',     label: 'Branding',        short: 'Branding' },
  { view: 'controller_studio',   label: 'Controller',      short: 'Control' },
];

export const LIGHTING_STYLES: { value: LightingStyle; label: string }[] = [
  { value: 'Natural Light', label: 'Natural Light' },
  { value: 'Studio Light', label: 'Studio Light' },
  { value: 'Golden Hour', label: 'Golden Hour' },
  { value: 'Blue Hour', label: 'Blue Hour' },
  { value: 'Cinematic', label: 'Cinematic' },
  { value: 'Dramatic', label: 'Dramatic' },
];

export const CAMERA_PERSPECTIVES: { value: CameraPerspective; label: string }[] = [
  { value: 'Front View', label: 'Front View' },
  { value: 'Top View', label: 'Top View' },
  { value: 'Side View', label: 'Side View' },
  { value: '45° Angle', label: '45° Angle' },
  { value: 'Close-up', label: 'Close-up' },
  { value: 'Macro Shot', label: 'Macro Shot' },
];

export const ASPECT_RATIOS: { value: AspectRatio; label: string }[] = [
  { value: '16:9', label: 'Landscape (16:9)' },
  { value: '9:16', label: 'Portrait (9:16)' },
  { value: '4:3', label: 'Classic (4:3)' },
  { value: '3:4', label: 'Classic Portrait (3:4)' },
  { value: '1:1', label: 'Square (1:1)' },
];

// --- New constants for Photoshoot Director ---
export const MAX_SHOT_SELECTION = 6;

export const SHOT_TYPES: { category: string; types: string[] }[] = [
  {
    category: 'Shot Angle',
    types: [
        'Close-Up', 'Medium Shot', 'Full Shot', 'High Angle',
        'Low Angle', 'Dutch Angle', 'Top-Down Shot', 'Macro Shot',
        'Eye-Level Shot', 'Worm\'s-Eye View', 'Detailed Texture Shot', 'Symmetrical Front-On',
        'Dynamic 3/4 Angle', 'Hero Shot (Slightly Low Angle)'
    ]
  },
  {
    category: 'Product in Action / Use Case',
    types: [
        'Lifestyle: Model Interacting with Product',
        'Dynamic Motion Shot (Pouring/Spraying)',
        'Close-up on Product Usage',
        'Product in its Intended Setting (e.g., Kitchen, Gym)',
        'Product as part of a Daily Routine',
        'With hands, showing usage',
        'On a creative desk',
        'In a travel setting (e.g., backpack)',
        'As part of a flat lay composition',
        'During a fitness activity',
        'In a cozy home environment',
        'Unboxing experience',
        'On Cafe Table (Hands Nearby)',
        'Held by Model (Close-up)',
    ]
  },
  {
    category: 'Environment & Style',
    types: [
        'On a Modern Kitchen Counter', 'On a Sandy Beach', 'On a Rustic Wooden Table', 'In a Lush Green Forest',
        'On a windowsill (morning light)', 'In nature (with dew drops)', 'Splash Shot',
        'Minimalist Studio (Gradient Background)', 'On a Marble Surface', 'Floating in Mid-Air (Surreal)', 'Amidst Urban Cityscape (Bokeh)',
        'On a bed of flowers', 'With Geometric Shapes & Shadows', 'Neon-lit Cyberpunk Setting', 'Luxury Velvet Background',
        'Submerged in Water', 'Industrial Concrete Background', 'Packaging Shot'
    ]
  }
];

// --- New constants for Voice Over Studio ---
export const VOICES: { value: string; label: string; description: string; gender: 'Male' | 'Female' }[] = [
  { value: 'Kore', label: 'Kore', description: 'Professional & Clear', gender: 'Female' },
  { value: 'Puck', label: 'Puck', description: 'Energetic & Youthful', gender: 'Male' },
  { value: 'Charon', label: 'Charon', description: 'Deep & Authoritative', gender: 'Male' },
  { value: 'Fenrir', label: 'Fenrir', description: 'Warm & Narrative', gender: 'Male' },
  { value: 'Zephyr', label: 'Zephyr', description: 'Calm & Soothing', gender: 'Male' },
  { value: 'Despina', label: 'Despina', description: 'Clear & Melodic', gender: 'Female' },
  { value: 'Orus', label: 'Orus', description: 'Crisp & Announcer', gender: 'Male' },
  { value: 'Leda', label: 'Leda', description: 'Elegant & Sophisticated', gender: 'Female' },
  { value: 'Gacrux', label: 'Gacrux', description: 'Powerful & Bold', gender: 'Male' },
  { value: 'Umbriel', label: 'Umbriel', description: 'Grounded & Natural', gender: 'Female' },
];

// --- Constants for Controller Studio ---
export const CONTROLLER_SLIDERS: ControllerSlider[] = [
    // Face Category
    { id: 'smile', label: 'Smile', value: 0, min: -1, max: 1, step: 0.1, category: 'Face' },
    { id: 'frown', label: 'Frown', value: 0, min: 0, max: 1, step: 0.1, category: 'Face' },
    { id: 'mouth_open', label: 'Mouth Open', value: 0, min: 0, max: 1, step: 0.1, category: 'Face' },
    { id: 'wink_left', label: 'Wink Left', value: 0, min: 0, max: 1, step: 0.1, category: 'Face' },
    { id: 'wink_right', label: 'Wink Right', value: 0, min: 0, max: 1, step: 0.1, category: 'Face' },
    { id: 'eyebrow_raise', label: 'Eyebrow Raise', value: 0, min: -1, max: 1, step: 0.1, category: 'Face' },
    { id: 'squint', label: 'Squint', value: 0, min: 0, max: 1, step: 0.1, category: 'Face' },
    { id: 'eye_direction', label: 'Eye Direction', value: 0, min: -1, max: 1, step: 0.1, category: 'Face' },
    { id: 'age', label: 'Age', value: 0, min: -1, max: 1, step: 0.1, category: 'Face' },
    
    // Head Category
    { id: 'head_pitch', label: 'Head Pitch (Up/Down)', value: 0, min: -1, max: 1, step: 0.1, category: 'Head' },
    { id: 'head_yaw', label: 'Head Yaw (Left/Right)', value: 0, min: -1, max: 1, step: 0.1, category: 'Head' },
    { id: 'head_roll', label: 'Head Roll', value: 0, min: -1, max: 1, step: 0.1, category: 'Head' },

    // Body Category
    { id: 'body_turn', label: 'Body Turn', value: 0, min: -1, max: 1, step: 0.1, category: 'Body' },
    { id: 'shoulder_shrug', label: 'Shoulder Shrug', value: 0, min: 0, max: 1, step: 0.1, category: 'Body' },

    // Retouch Category
    { id: 'skin_smooth', label: 'Skin Smoothing', value: 0, min: 0, max: 1, step: 0.1, category: 'Retouch' },
    { id: 'brightness', label: 'Brightness', value: 0, min: -1, max: 1, step: 0.1, category: 'Retouch' },
    { id: 'contrast', label: 'Contrast', value: 0, min: -1, max: 1, step: 0.1, category: 'Retouch' },
    { id: 'sharpness', label: 'Sharpness', value: 0, min: 0, max: 1, step: 0.1, category: 'Retouch' },
];
