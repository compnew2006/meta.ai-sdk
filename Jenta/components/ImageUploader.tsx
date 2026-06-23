
import React, { useState } from 'react';
import { ImageFile } from '../types';
import ImageCropper from './ImageCropper';

interface ImageUploaderProps {
  onFileUpload: (files: File[]) => void;
  images: ImageFile[];
  title: string;
  id: string;
  multiple?: boolean;
  onRemove?: (index: number) => void;
  isUploading?: boolean;
  onImageUpdate?: (index: number, newImage: ImageFile) => void;
}

const UploadIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-[var(--color-text-secondary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
    </svg>
);

const XIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
);

const CropIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
    </svg>
);

const LoadingSpinnerSmall = () => (
    <div className="flex justify-center items-center">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
    </div>
);

// Helper to convert base64 back to File for cropping
const base64ToFile = (base64: string, mimeType: string, fileName: string): File => {
    const byteString = atob(base64);
    const ab = new ArrayBuffer(byteString.length);
    const ia = new Uint8Array(ab);
    for (let i = 0; i < byteString.length; i++) {
        ia[i] = byteString.charCodeAt(i);
    }
    return new File([ab], fileName, { type: mimeType });
}

const ImageUploader: React.FC<ImageUploaderProps> = ({ onFileUpload, images, title, id, multiple = false, onRemove, isUploading = false, onImageUpdate }) => {
  const [isDragging, setIsDragging] = useState(false);
  const [editingImageIndex, setEditingImageIndex] = useState<number | null>(null);

  const handleDragEnter = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.items && e.dataTransfer.items.length > 0) {
      setIsDragging(true);
    }
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      const files: File[] = [];
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        const file = e.dataTransfer.files[i];
        if (file.type.startsWith('image/')) {
          files.push(file);
        }
      }
      
      if (files.length > 0) {
        onFileUpload(files);
      }
      e.dataTransfer.clearData();
    }
  };

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files && files.length > 0) {
      const imageFiles: File[] = [];
      for (let i = 0; i < files.length; i++) {
        const file = files[i];
        if (file.type.startsWith('image/')) {
          imageFiles.push(file);
        }
      }
      
      if (imageFiles.length > 0) {
        onFileUpload(imageFiles);
      }
    }
    event.target.value = ''; // Allow re-uploading the same file
  };

  const handleCropConfirm = async (croppedFile: File) => {
    if (editingImageIndex === null || !onImageUpdate) return;
    
    const reader = new FileReader();
    reader.onloadend = () => {
        const base64String = reader.result as string;
        const newImageFile: ImageFile = {
            base64: base64String.split(',')[1],
            mimeType: croppedFile.type,
            name: croppedFile.name
        };
        onImageUpdate(editingImageIndex, newImageFile);
        setEditingImageIndex(null);
    };
    reader.readAsDataURL(croppedFile);
  };

  return (
    <div 
      className="w-full h-full min-h-[150px] relative overflow-hidden rounded-xl bg-black/5 group/container"
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
    >
      <label htmlFor={id} className="cursor-pointer group block w-full h-full">
        <div className={`w-full h-full relative border-2 rounded-xl text-center transition-all duration-300 bg-black/10 backdrop-blur-sm flex flex-col justify-center items-center ${isDragging ? 'border-solid border-[var(--color-accent)]' : 'border-dashed border-[rgba(var(--color-accent-rgb),0.3)] group-hover:border-[rgba(var(--color-accent-rgb),0.8)]'}`}>
          {images.length > 0 ? (
            <div className="absolute inset-0 w-full h-full group/image">
                <img src={`data:${images[0].mimeType};base64,${images[0].base64}`} alt={title} className="w-full h-full object-contain rounded-xl" />
                
                {/* Controls Overlay */}
                <div className="absolute top-2 right-2 flex gap-2 opacity-0 group-hover/image:opacity-100 transition-opacity z-10">
                    {/* Edit Button */}
                    {onImageUpdate && (
                        <button
                            onClick={(e) => { e.preventDefault(); e.stopPropagation(); setEditingImageIndex(0); }}
                            className="bg-black/60 text-white rounded-full p-2 hover:bg-[var(--color-accent)] transition-colors backdrop-blur-sm shadow-lg border border-white/10"
                            title="Crop Image"
                        >
                            <CropIcon />
                        </button>
                    )}
                    {/* Remove Button */}
                    {onRemove && (
                        <button
                            onClick={(e) => { e.preventDefault(); e.stopPropagation(); onRemove(0); }}
                            className="bg-black/60 text-white rounded-full p-2 hover:bg-red-500 transition-colors backdrop-blur-sm shadow-lg border border-white/10"
                            title="Remove Image"
                        >
                            <XIcon />
                        </button>
                    )}
                </div>
            </div>
          ) : isUploading ? (
            <div className="flex flex-col items-center p-4">
              <LoadingSpinnerSmall />
              <p className="mt-2 text-sm text-[var(--color-text-secondary)]">Processing...</p>
            </div>
          ) : (
            <div className="flex flex-col items-center p-4">
              <UploadIcon />
              <p className="mt-2 text-sm text-[var(--color-text-secondary)] group-hover:text-[var(--color-text-base)] transition-colors">{title}</p>
              <p className="mt-1 text-xs text-[var(--color-text-muted)]">or drag and drop</p>
            </div>
          )}
        </div>
      </label>
      <input id={id} type="file" className="hidden" accept="image/*" onChange={handleInputChange} multiple={multiple} disabled={isUploading} />
      
       {isDragging && !isUploading && (
          <div className="absolute inset-0 bg-black/60 flex flex-col items-center justify-center rounded-xl backdrop-blur-sm pointer-events-none z-10">
              <UploadIcon />
              <p className="text-lg text-[var(--color-text-base)] mt-2 font-semibold">Drop image to upload</p>
          </div>
      )}

      {/* Inline Cropper Overlay */}
      {editingImageIndex !== null && images[editingImageIndex] && (
          <ImageCropper 
            file={base64ToFile(images[editingImageIndex].base64, images[editingImageIndex].mimeType, images[editingImageIndex].name)}
            onConfirm={handleCropConfirm}
            onCancel={() => setEditingImageIndex(null)}
          />
      )}
    </div>
  );
};

export default ImageUploader;
