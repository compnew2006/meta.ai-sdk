// components/ResultDisplay.tsx
//
// Thin facade over <ResultCard>. ResultCard now owns the loading / empty /
// image / fullscreen / edit / 2K-4K download logic (previously duplicated here
// verbatim). This wrapper keeps the old prop signature so CreatorStudio and
// any other importer need not change.
//
// Why a facade and not a delete: ResultDisplay is imported by name across the
// studios; collapsing it in one shot would force a multi-file rename. The
// facade lets adoption proceed file-by-file.

import React from 'react';
import { ImageFile } from '../types';
import ResultCard from './ui/ResultCard';

export interface ResultDisplayProps {
  imageFile: ImageFile | null;
  isLoading: boolean;
  onEdit?: (prompt: string) => void;
  isEditing?: boolean;
}

const ResultDisplay: React.FC<ResultDisplayProps> = ({
  imageFile,
  isLoading,
  onEdit,
  isEditing,
}) => (
  <ResultCard
    image={imageFile}
    isLoading={isLoading}
    isEditing={isEditing}
    onEdit={onEdit}
    fileNamePrefix="SMART-Studio-Result"
  />
);

export default ResultDisplay;
