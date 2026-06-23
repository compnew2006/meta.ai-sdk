
import React, { useCallback } from 'react';
import { ImageFile, ShotType, PhotoshootDirectorProject } from '../types';
import { resizeImage } from '../utils';
import { generateImage, editImage } from '../services/geminiService';
import ImageWorkspace from './ImageWorkspace';
import ShotTypeSelector from './ShotTypeSelector';
import ResultsGrid from './ResultsGrid';

interface PhotoshootDirectorProps {
  project: PhotoshootDirectorProject;
  setProject: React.Dispatch<React.SetStateAction<PhotoshootDirectorProject>>;
}

const PhotoshootDirector: React.FC<PhotoshootDirectorProps> = ({ project, setProject }) => {

    const onGenerate = useCallback(async () => {
    if (!project || project.productImages.length === 0 || project.selectedShotTypes.length === 0) {
      setProject(s => ({ ...s, error: 'Please upload a product image and select at least one shot type.' }));
      return;
    }

    setProject(s => ({
      ...s,
      isGenerating: true,
      error: null,
      results: s.selectedShotTypes.map(shotType => ({
        shotType,
        image: null,
        isLoading: true,
        error: null,
        editPrompt: '',
        isEditing: false,
      })),
    }));

    const generationPromises = project.selectedShotTypes.map(shotType => {
      const textProtection = "STRICTLY PRESERVE all original text, labels, and branding on the product. DO NOT erase original writing. NO EXTRA generated text in the scene.";
      let prompt = `A high-resolution, professional photograph of the subject from the provided image. The desired shot is: '${shotType}'. The background should be clean, non-distracting, and complementary to the subject. ${textProtection}`;
      if (project.customStylePrompt) {
        prompt += ` Additional style requirements: ${project.customStylePrompt}`;
      }
      return generateImage(project.productImages, prompt, null)
        .then(image => ({ status: 'fulfilled' as const, value: { shotType, image } }))
        .catch(error => ({ status: 'rejected' as const, reason: { shotType, error } }));
    });

    const settledResults = await Promise.all(generationPromises);

    settledResults.forEach(result => {
      if (result.status === 'fulfilled') {
        const { shotType, image } = result.value;
        setProject(s => ({
          ...s,
          results: s.results.map(r => r.shotType === shotType ? { ...r, image, isLoading: false } : r),
        }));
      } else {
        const { shotType, error } = result.reason;
        console.error(`Failed to generate image for ${shotType}:`, error);
        setProject(s => ({
          ...s,
          results: s.results.map(r => r.shotType === shotType ? { ...r, error: error.message || 'Generation failed', isLoading: false } : r),
        }));
      }
    });
    
    setProject(s => ({ ...s, isGenerating: false }));
  }, [project, setProject]);

  const handleEditResult = async (index: number, prompt: string) => {
      const result = project.results[index];
      if (!result || !result.image) return;

      setProject(s => {
          const newResults = [...s.results];
          newResults[index] = { ...newResults[index], isEditing: true, error: null };
          return { ...s, results: newResults };
      });

      try {
          const updated = await editImage(result.image, prompt);
          setProject(s => {
              const newResults = [...s.results];
              newResults[index] = { ...newResults[index], image: updated, isEditing: false };
              return { ...s, results: newResults };
          });
      } catch (err) {
          setProject(s => {
              const newResults = [...s.results];
              newResults[index] = { ...newResults[index], isEditing: false, error: err instanceof Error ? err.message : 'Edit failed' };
              return { ...s, results: newResults };
          });
      }
  };

  const handleFileUpload = async (files: File[]) => {
    if (!files || files.length === 0) return;

    setProject(s => ({ ...s, isUploading: true, error: null }));
    let currentError: string | null = null;
    
    const filePromises = files.map(file => {
      return new Promise<ImageFile | null>(async (resolve) => {
        if (!file.type.startsWith('image/')) {
          if (!currentError) currentError = `File '${file.name}' is not a supported image type.`;
          resolve(null);
          return;
        }
        try {
          const resizedFile = await resizeImage(file, 2048, 2048);
          const reader = new FileReader();
          reader.onloadend = () => {
            const base64String = reader.result as string;
            resolve({
              base64: base64String.split(',')[1],
              mimeType: resizedFile.type,
              name: resizedFile.name
            });
          };
          reader.onerror = () => {
            if (!currentError) currentError = `Error reading file '${resizedFile.name}'.`;
            resolve(null);
          };
          reader.readAsDataURL(resizedFile);
        } catch (err) {
          console.error(`Error processing ${file.name}:`, err);
          if (!currentError) currentError = `Could not process file '${file.name}'.`;
          resolve(null);
        }
      });
    });

    const results = await Promise.all(filePromises);
    const validImages = results.filter((img): img is ImageFile => img !== null);

    setProject(s => ({
      ...s,
      productImages: [...s.productImages, ...validImages],
      error: currentError,
      isUploading: false,
    }));
  };

  const handleRemoveImage = (indexToRemove: number) => {
    setProject(s => ({
      ...s,
      productImages: s.productImages.filter((_, index) => index !== indexToRemove),
    }));
  };

  const handleImageUpdate = (index: number, newImage: ImageFile) => {
      setProject(s => {
          const newImages = [...s.productImages];
          if (index >= 0 && index < newImages.length) {
              newImages[index] = newImage;
          }
          return { ...s, productImages: newImages };
      });
  };

  if (!project) {
    return (
        <main className="w-full max-w-7xl flex items-center justify-center gap-8 pt-8 pb-12 flex-grow">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-[var(--color-accent)]"></div>
            <p className="text-[var(--color-text-secondary)]">Loading Photoshoot Director...</p>
        </main>
    );
  }

  return (
    <main className="w-full flex flex-col gap-4 pt-4 pb-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 flex-grow">
        {/* Left Column: Controls */}
        <div className="lg:col-span-1 flex flex-col gap-4">
          <div className="glass-card rounded-2xl p-4">
            <h3 className="text-lg font-bold text-[var(--color-text-base)] mb-4">1. Upload Image(s)</h3>
            <ImageWorkspace
              id="photoshoot-product-uploader"
              images={project.productImages}
              onImagesUpload={handleFileUpload}
              onImageRemove={handleRemoveImage}
              isUploading={project.isUploading}
              onImageUpdate={handleImageUpdate}
            />
          </div>
          <ShotTypeSelector
            selected={project.selectedShotTypes}
            onChange={(selected) => setProject(s => ({ ...s, selectedShotTypes: selected }))}
            customStylePrompt={project.customStylePrompt}
            onCustomStylePromptChange={(prompt) => setProject(s => ({...s, customStylePrompt: prompt }))}
          />
          <div className="w-full px-4 py-2">
            <button
                onClick={onGenerate}
                disabled={project.isGenerating || project.isUploading || project.productImages.length === 0 || project.selectedShotTypes.length === 0}
                className="w-full bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-dark)] hover:from-[var(--color-accent-dark)] hover:to-[var(--color-accent-darker)] text-[var(--color-text-base)] font-bold py-3 px-8 rounded-lg text-lg transition-all duration-300 ease-in-out disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-2xl hover:shadow-[var(--color-accent)]/20 disabled:shadow-none transform hover:-translate-y-1 disabled:transform-none"
            >
                {project.isGenerating ? 'Generating...' : `Generate ${project.selectedShotTypes.length} Shot${project.selectedShotTypes.length === 1 ? '' : 's'}`}
            </button>
          </div>
        </div>

        {/* Right Column: Results */}
        <div className="lg:col-span-2 flex flex-col gap-6">
          <div className="flex flex-col flex-grow glass-card rounded-2xl p-4">
              <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-bold text-[var(--color-text-base)]">3. Generated Results</h3>
              </div>
               {project.error && <div className="mb-4 bg-[rgba(var(--color-accent-rgb),0.2)] border border-[rgba(var(--color-accent-rgb),0.5)] text-[var(--color-accent-light)] px-4 py-3 rounded-lg" role="alert">{project.error}</div>}
              <ResultsGrid 
                results={project.results} 
                onEditResult={handleEditResult}
              />
          </div>
        </div>
      </div>
    </main>
  );
};

export default PhotoshootDirector;
