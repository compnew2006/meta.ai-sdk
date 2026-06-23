
export const resizeImage = (file: File, maxWidth: number, maxHeight: number): Promise<File> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = (event) => {
      if (!event.target?.result) {
        return reject(new Error('Could not read file.'));
      }
      const img = new Image();
      img.src = event.target.result as string;
      img.onload = () => {
        const canvas = document.createElement('canvas');
        let { width, height } = img;

        if (width > height) {
          if (width > maxWidth) {
            height *= maxWidth / width;
            width = maxWidth;
          }
        } else {
          if (height > maxHeight) {
            width *= maxHeight / height;
            height = maxHeight;
          }
        }
        
        canvas.width = width;
        canvas.height = height;

        const ctx = canvas.getContext('2d');
        if (!ctx) {
          return reject(new Error('Could not get canvas context'));
        }
        
        // Ensure transparency is preserved for PNGs
        ctx.clearRect(0, 0, width, height);
        ctx.drawImage(img, 0, 0, width, height);
        
        const mimeType = file.type === 'image/png' ? 'image/png' : 'image/jpeg';
        const quality = 0.9;
        
        canvas.toBlob((blob) => {
          if (!blob) {
            return reject(new Error('Canvas toBlob failed'));
          }
          const newFileName = file.name;
          const newFile = new File([blob], newFileName, {
            type: mimeType,
            lastModified: Date.now(),
          });
          resolve(newFile);
        }, mimeType, quality);
      };
      img.onerror = (err) => reject(err);
    };
    reader.onerror = (err) => reject(err);
  });
};

export function decodeBase64(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const len = binaryString.length;
  const bytes = new Uint8Array(len);
  for (let i = 0; i < len; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

// Custom function to decode raw PCM data into an AudioBuffer for playback
export async function decodeAudioData(
  data: Uint8Array,
  ctx: AudioContext,
  sampleRate: number,
  numChannels: number,
): Promise<AudioBuffer> {
  const dataInt16 = new Int16Array(data.buffer);
  const frameCount = dataInt16.length / numChannels;
  const buffer = ctx.createBuffer(numChannels, frameCount, sampleRate);

  for (let channel = 0; channel < numChannels; channel++) {
    const channelData = buffer.getChannelData(channel);
    for (let i = 0; i < frameCount; i++) {
      channelData[i] = dataInt16[i * numChannels + channel] / 32768.0;
    }
  }
  return buffer;
}


// Function to convert raw PCM data (as Float32Array) to a WAV Blob for download
export function pcmToWavBlob(pcmData: Float32Array, sampleRate: number, numChannels: number): Blob {
    const headerSize = 44;
    const dataSize = pcmData.length * 2; // 16-bit samples
    const fileSize = headerSize + dataSize;
    const buffer = new ArrayBuffer(fileSize);
    const view = new DataView(buffer);

    let offset = 0;

    const writeString = (view: DataView, offset: number, string: string) => {
        for (let i = 0; i < string.length; i++) {
            view.setUint8(offset + i, string.charCodeAt(i));
        }
    };

    // RIFF chunk descriptor
    writeString(view, offset, 'RIFF'); offset += 4;
    view.setUint32(offset, fileSize - 8, true); offset += 4;
    writeString(view, offset, 'WAVE'); offset += 4;

    // "fmt " sub-chunk
    writeString(view, offset, 'fmt '); offset += 4;
    view.setUint32(offset, 16, true); offset += 4; // Sub-chunk size
    view.setUint16(offset, 1, true); offset += 2; // Audio format (1 for PCM)
    view.setUint16(offset, numChannels, true); offset += 2;
    view.setUint32(offset, sampleRate, true); offset += 4;
    view.setUint32(offset, sampleRate * numChannels * 2, true); offset += 4; // Byte rate
    view.setUint16(offset, numChannels * 2, true); offset += 2; // Block align
    view.setUint16(offset, 16, true); offset += 2; // Bits per sample

    // "data" sub-chunk
    writeString(view, offset, 'data'); offset += 4;
    view.setUint32(offset, dataSize, true); offset += 4;

    // Write PCM data
    for (let i = 0; i < pcmData.length; i++, offset += 2) {
        let s = Math.max(-1, Math.min(1, pcmData[i]));
        view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7FFF, true);
    }

    return new Blob([view], { type: 'audio/wav' });
}
