import { useState, useEffect } from 'react';
import { getToken } from '../api';

interface AuthenticatedImageProps {
  src: string;
  alt: string;
  className?: string;
  onLoad?: () => void;
  onError?: (error: Error) => void;
}

/**
 * A component that loads images with proper authentication headers
 * Solves the problem that HTML <img> tags can't set Authorization headers
 */
const AuthenticatedImage = ({ src, alt, className, onLoad, onError }: AuthenticatedImageProps) => {
  const [imageSrc, setImageSrc] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // Reset state when src changes
    setIsLoading(true);
    setError(null);
    
    // If there was a previous blob URL, revoke it to free memory
    if (imageSrc && imageSrc.startsWith('blob:')) {
      URL.revokeObjectURL(imageSrc);
      setImageSrc(null);
    }

    const fetchImage = async () => {
      try {
        const token = getToken();
        
        // Fetch the image with the Authorization header
        const response = await fetch(src, {
          headers: {
            ...(token ? { Authorization: `Bearer ${token}` } : {})
          }
        });

        if (!response.ok) {
          throw new Error(`Failed to load image: ${response.status} ${response.statusText}`);
        }

        // Get the image as a blob
        const blob = await response.blob();
        
        // Create a blob URL
        const objectUrl = URL.createObjectURL(blob);
        setImageSrc(objectUrl);
        setIsLoading(false);
        onLoad?.();
      } catch (err) {
        console.error('Failed to load image:', err);
        setIsLoading(false);
        const error = err instanceof Error ? err : new Error(String(err));
        setError(error);
        onError?.(error);
      }
    };

    fetchImage();

    // Clean up the blob URL when the component unmounts or src changes
    return () => {
      if (imageSrc && imageSrc.startsWith('blob:')) {
        URL.revokeObjectURL(imageSrc);
      }
    };
  }, [src, onLoad, onError]);

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center bg-gray-200 ${className}`}>
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error || !imageSrc) {
    return (
      <div className={`flex items-center justify-center bg-gray-100 ${className}`}>
        <div className="text-center p-4">
          <svg className="mx-auto h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span className="block text-sm text-gray-500 mt-1">Failed to load image</span>
        </div>
      </div>
    );
  }

  return <img src={imageSrc} alt={alt} className={className} />;
};

export default AuthenticatedImage;
