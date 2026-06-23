# Jenta Design Tool - Design System & Patterns

This document outlines the core design principles, CSS patterns, and visual identity of the Jenta Design Tool. It serves as a reference for maintaining a consistent aesthetic across the application.

## 1. Visual Theme & Atmosphere

**Core Aesthetic:** Futuristic, AI-Powered, Glassmorphism, and Immersive.

The Jenta Design Tool utilizes a highly modern, sleek interface characterized by:
- **Glassmorphism:** Extensive use of translucent backgrounds with background-blur to create depth without visual clutter.
- **Dynamic Lighting:** Radial gradients and slow-pulsing background glows (`glowPulse`) that simulate a "breathing" AI environment.
- **Fluid Motion:** Elements that float (`animate-float`) and interact with user input (e.g., the mouse-tracking 3D logo).
- **Bilingual Elegance:** Seamless support for both English and Arabic typography using the `Tajawal` font.

---

## 2. Color Palette & Roles

The application uses a robust CSS variable system defined in `index.html` to support both Dark (default) and Light themes.

### Dark Theme (Default)
- **Background Base:** `#0D0F0C` (Deep greenish-black)
- **Background Gradient Start:** `#1F211F` (Slightly lighter center for radial gradient)
- **Accent (Primary):** `#ff0000` (Vibrant Red)
  - *Light:* `#ff8080` | *Dark:* `#c00000`
- **Text Base:** `#e5e7ce` (Off-white with a hint of yellow/green)
- **Text Medium:** `#c2c3b4`
- **Text Secondary:** `#a8a99a`
- **Text Muted:** `#8e8f80`
- **Glass Surfaces:** `rgba(20, 20, 20, 0.5)` with `rgba(229, 231, 206, 0.1)` borders.

### Light Theme
- **Background Base:** `#F8FAFC` (Slate 50)
- **Background Gradient Start:** `#FFFFFF` (Pure White)
- **Accent (Primary):** `#dc2626` (Tailwind Red 600)
- **Text Base:** `#1f2937` (Gray 800)
- **Text Secondary:** `#6b7280` (Gray 500)
- **Glass Surfaces:** `rgba(255, 255, 255, 0.6)` with `rgba(31, 41, 55, 0.05)` borders.

---

## 3. Typography Rules

### Font Family
- **Primary Font:** `Tajawal`, sans-serif. Chosen for its excellent legibility and modern geometric proportions in both Latin and Arabic scripts.

### Hierarchy & Principles
- **Hero Headings:** Extremely large, heavy, and tight. 
  - *Classes:* `text-5xl sm:text-7xl md:text-8xl font-black tracking-tight leading-[0.9]`
- **Section Titles:** Bold and uppercase with wide tracking.
  - *Classes:* `text-sm font-bold uppercase tracking-widest`
- **Body Text:** Medium weight with relaxed line height for readability.
  - *Classes:* `text-base font-medium leading-relaxed`
- **Typewriter Effect:** Used in the hero section to dynamically cycle through keywords ("DESIGN", "STORYBOARD", etc.) with an animated blinking cursor `|`.

---

## 4. Component Stylings

### Glass Cards & Containers
The signature look of the app relies on the `.glass-card` utility class:
```css
.glass-card {
  background: var(--color-glass-bg);
  backdrop-filter: blur(16px);
  border: 1px solid var(--color-glass-border);
  box-shadow: 0 8px 32px 0 var(--color-glass-shadow);
}
```

### Inputs & Forms
Inputs use the `.glass-input` class to blend into the environment while maintaining clear focus states:
- **Default:** Translucent background, subtle border.
- **Hover:** Border color shifts to a semi-transparent accent color.
- **Focus:** Accent color border with a soft `box-shadow` glow.

### Buttons
- **Primary Actions:** Pill-shaped (`rounded-full`), solid accent background, white text, bold.
- **Interactions:** Scale up on hover (`hover:scale-105`), scale down on click (`active:scale-95`).
- **Glows:** Primary buttons cast a colored shadow (`shadow-lg shadow-[var(--color-accent)]/20`).

### Navigation (Tab Bar)
- **Desktop:** Clean text links with a bottom border indicating the active state (`border-[var(--color-accent)]`).
- **Mobile:** Horizontally scrollable container (`overflow-x-auto`) with hidden scrollbars (`.suggestions-scrollbar`) and chevron navigation buttons.

---

## 5. Layout Principles

### Grid & Container
- **Max Width:** Content is constrained to `max-w-7xl` (1280px) and centered (`mx-auto`).
- **Spacing:** Generous padding (`px-4` on mobile, `sm:px-6`, `lg:px-8` on larger screens).
- **Structure:** Heavy reliance on Flexbox (`flex`, `flex-col`, `items-center`) and CSS Grid (`grid-cols-1 md:grid-cols-3`) for responsive alignment.

### Whitespace Philosophy
- Elements are given ample room to breathe. The hero section uses `min-h-[60vh]` to ensure the main value proposition is front and center without crowding.
- Sections are separated by large margins (e.g., `mt-20`).

---

## 6. Depth & Elevation

### Shadow Philosophy
Instead of standard drop shadows, depth is created through:
1.  **Background Glows:** A fixed pseudo-element (`body::before`) creates a massive, slow-pulsing radial gradient behind all content.
    ```css
    @keyframes glowPulse {
      0%, 100% { transform: scale(1); opacity: 0.7; }
      50% { transform: scale(1.5); opacity: 1; }
    }
    ```
2.  **Backdrop Blurs:** Stacking `.glass-card` elements over the glowing background creates natural, dynamic depth.
3.  **Floating Elements:** Hero images use `@keyframes float` for a continuous, smooth vertical oscillation.
4.  **Interactive Parallax:** The main 3D logo tracks the user's mouse movement, moving slightly in the opposite direction of the cursor to simulate 3D space.

---

## 7. Do's and Don'ts

### Do
- **Do** use CSS variables (`var(--color-...)`) for all colors to ensure theme switching works flawlessly.
- **Do** use `backdrop-blur` for overlapping UI elements (like sticky navbars) to maintain context of the content underneath.
- **Do** wrap horizontal lists in `.suggestions-scrollbar` to hide ugly default scrollbars on Windows/Linux.
- **Do** ensure interactive elements have both `hover:` and `active:` states for tactile feedback.

### Don't
- **Don't** use solid, opaque background colors for cards or panels; it breaks the glassmorphism illusion.
- **Don't** use standard Tailwind gray colors directly (e.g., `text-gray-500`); always use the semantic variables (`var(--color-text-secondary)`).
- **Don't** clutter the UI with heavy borders; rely on spacing, subtle 10% opacity borders, and blurs to separate content.

---

## 8. Responsive Behavior

### Breakpoints
- **Mobile First:** Default classes target mobile devices.
- **`sm` (640px):** Adjust padding and slightly increase typography size.
- **`md` (768px):** Switch from single-column to multi-column grids (e.g., footer links).
- **`lg` (1024px):** Show desktop navigation, reveal decorative floating elements (like the hero 3D logo).

### Touch Targets & Mobile Nav
- Mobile navigation uses a sticky top bar with horizontal scrolling.
- Buttons and tabs have generous padding (`py-3`, `px-4`) to ensure they are easily tappable on touch devices.

---

## 9. Agent Prompt Guide

When asking an AI agent to build new components for this project, use the following guidelines:

**Quick Color Reference:**
- Primary Action: `bg-[var(--color-accent)] text-white`
- Card Container: `className="glass-card rounded-2xl p-6"`
- Input Field: `className="glass-input w-full rounded-xl p-3"`
- Primary Text: `text-[var(--color-text-base)]`
- Secondary Text: `text-[var(--color-text-secondary)]`

**Example Component Prompt:**
> "Create a new settings panel component. It should use the `.glass-card` class for its container, have a bold `Tajawal` heading using `var(--color-text-base)`, and include a primary save button that is pill-shaped, uses the accent color, and scales up on hover (`hover:scale-105 active:scale-95`). Inputs should use the `.glass-input` class."
