// services/geminiService.ts
//
// BACKEND SHIM — Jenta's AI now runs through the metaai-go REST API (Meta AI).
// This file re-exports the new implementation (./aiService) so the 11
// components that `import { ... } from '../services/geminiService'` keep working
// unchanged.
//
// To roll back to Gemini, restore the previous contents of this file from git.
// See INTEGRATION.md for the full mapping.

export * from './aiService';
