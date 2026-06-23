
import React, { useState } from 'react';
import { ImageFile } from '../types';
import ImageUploader from './ImageUploader';
import ImageCropper from './ImageCropper';

interface ImageWorkspaceProps {
  images: ImageFile[]; // Renamed from productImages
  onImagesUpload: (files: File[]) => void; // Renamed from onProductImagesUpload
  onImageRemove: (index: number) => void; // Renamed from onProductImageRemove
  isUploading: boolean;
  onImageUpdate?: (index: number, newImage: ImageFile) => void; // Renamed from onProductImageUpdate
  title?: string;
  id?: string;
}

const PlusIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-[var(--color-text-secondary)] group-hover:text-[var(--color-text-base)] transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
    </svg>
);

const XIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
);

const CropIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
    </svg>
);

const base64ToFile = (base64: string, mimeType: string, fileName: string): File => {
    const byteString = atob(base64);
    const ab = new ArrayBuffer(byteString.length);
    const ia = new Uint8Array(ab);
    for (let i = 0; i < byteString.length; i++) {
        ia[i] = byteString.charCodeAt(i);
    }
    return new File([ab], fileName, { type: mimeType });
}

const ImageWorkspace: React.FC<ImageWorkspaceProps> = ({ 
  images, 
  onImagesUpload, 
  onImageRemove,
  isUploading,
  onImageUpdate,
  title = "Upload Image(s)",
  id = "image-workspace-uploader"
}) => {
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  
  const handleAddMoreUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files && files.length > 0) {
      const imageFiles = Array.from(files).filter((file: File) => file.type.startsWith('image/'));
      if (imageFiles.length > 0) {
          onImagesUpload(imageFiles);
      }
    }
    event.target.value = ''; // Allow re-uploading the same file
  };

  const handleCropConfirm = async (croppedFile: File) => {
    if (editingIndex === null || !onImageUpdate) return;
    
    const reader = new FileReader();
    reader.onloadend = () => {
        const base64String = reader.result as string;
        const newImageFile: ImageFile = {
            base64: base64String.split(',')[1],
            mimeType: croppedFile.type,
            name: croppedFile.name
        };
        onImageUpdate(editingIndex, newImageFile);
        setEditingIndex(null);
    };
    reader.readAsDataURL(croppedFile);
  };

  return (
    <div className="flex flex-col gap-2 w-full">
        {/* Main Uploader Area */}
        <div className="w-full aspect-square relative">
            <div className="absolute inset-0">
                <ImageUploader
                    id={id}
                    title={title}
                    images={images}
                    onFileUpload={onImagesUpload}
                    multiple={true}
                    onRemove={onImageRemove}
                    isUploading={isUploading && images.length === 0}
                    onImageUpdate={onImageUpdate} 
                />
            </div>
        </div>
        
        {/* Thumbnail Bar */}
        {images.length > 0 && (
          <div className="w-full h-16">
            <div className="h-full w-full overflow-x-auto flex flex-row items-center gap-2 suggestions-scrollbar pb-1">
              {images.slice(1).map((image, index) => (
                <div key={index + 1} className="h-14 w-14 flex-shrink-0 relative group/image rounded-md overflow-hidden bg-black/10">
                  <img src={`data:${image.mimeType};base64,${image.base64}`} alt={`Thumbnail ${index + 2}`} className="w-full h-full object-cover" />
                  
                  {/* Overlay Controls for Thumbnails */}
                  <div className="absolute inset-0 bg-black/60 opacity-0 group-hover/image:opacity-100 transition-opacity flex items-center justify-center gap-1">
                      {onImageUpdate && (
                         <button
                            onClick={() => setEditingIndex(index + 1)}
                            className="text-white hover:text-[var(--color-accent)] transition-colors p-1"
                            title="Crop"
                        >
                            <CropIcon />
                        </button>
                      )}
                      <button
                        onClick={() => onImageRemove(index + 1)}
                        className="text-white hover:text-red-500 transition-colors p-1"
                        title="Remove"
                      >
                        <XIcon />
                      </button>
                  </div>
                  
                  {/* Global Full Screen Cropper Triggered from Thumbnail */}
                  {editingIndex === (index + 1) && (
                        <ImageCropper 
                            file={base64ToFile(image.base64, image.mimeType, image.name)}
                            onConfirm={handleCropConfirm}
                            onCancel={() => setEditingIndex(null)}
                        />
                  )}
                </div>
              ))}
              
              {/* Add more button */}
               <label htmlFor={`${id}-add-more`} className="cursor-pointer group h-14 w-14 flex-shrink-0">
                  <div className="h-full w-full rounded-md border border-dashed border-[rgba(var(--color-accent-rgb),0.3)] hover:border-[rgba(var(--color-accent-rgb),0.8)] flex flex-col items-center justify-center text-[var(--color-text-secondary)] hover:text-[var(--color-text-base)] transition-colors">
                      {isUploading ? (
                        <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
                      ) : (
                        <PlusIcon />
                      )}
                  </div>
               </label>
               <input 
                 id={`${id}-add-more`} 
                 type="file" 
                 className="hidden" 
                 accept="image/*" 
                 onChange={handleAddMoreUpload} 
                 multiple 
                 disabled={isUploading}
               />
            </div>
          </div>
        )}
    </div>
  );
};

export default ImageWorkspace;
