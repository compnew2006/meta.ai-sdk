// utils/imageDownload.ts
//
// The 2K/4K canvas-upscale + download logic. Previously copy-pasted in
// ResultDisplay.tsx (~line 54), ResultsGrid.tsx (~line 23), and
// BrandingResultsGrid.tsx (~line 23) — three near-identical implementations.
// Extracted here as a single hook so ResultCard (and any future caller) uses
// one source of truth.

import { useState, useCallback } from 'react';
import { ImageFile } from '../types';

export type Resolution = '2k' | '4k';

/**
 * Hook returning a `download(image, resolution, fileName)` callback and an
 * `isDownloading` flag. The callback upscales the image to the requested
 * resolution via an offscreen canvas and triggers a browser download.
 *
 * @param fileNamePrefix e.g. "SMART-Studio-Result" — final name is `${prefix}-${resolution}-${timestamp}.png`.
 */
export function useImageDownload(fileNamePrefix = 'SMART-Studio') {
  const [isDownloading, setIsDownloading] = useState(false);

  const download = useCallback(
    (image: ImageFile, resolution: Resolution, fileNamePrefixOverride?: string) => {
      setIsDownloading(true);
      const img = new Image();
      img.src = `data:${image.mimeType};base64,${image.base64}`;
      img.onload = () => {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          setIsDownloading(false);
          return;
        }
        const targetWidth = resolution === '4k' ? 4096 : 2048;
        const aspectRatio = img.width / img.height;
        canvas.width = targetWidth;
        canvas.height = targetWidth / aspectRatio;
        ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
        const link = document.createElement('a');
        link.download = `${fileNamePrefixOverride ?? fileNamePrefix}-${resolution}-${Date.now()}.png`;
        link.href = canvas.toDataURL('image/png');
        link.click();
        setIsDownloading(false);
      };
      img.onerror = () => setIsDownloading(false);
    },
    [fileNamePrefix],
  );

  return { download, isDownloading };
}
