import { useState, useEffect } from 'react';
import { Photo, getPhotoUrl } from '../api';
import AuthenticatedImage from './AuthenticatedImage';

interface LightboxProps {
  photo: Photo | null;
  onClose: () => void;
}

const Lightbox = ({ photo, onClose }: LightboxProps) => {
  const [isLoading, setIsLoading] = useState(true);

  // Close on escape key
  useEffect(() => {
    const handleEsc = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };
    
    window.addEventListener('keydown', handleEsc);
    
    return () => {
      window.removeEventListener('keydown', handleEsc);
    };
  }, [onClose]);

  // Reset loading state when photo changes
  useEffect(() => {
    setIsLoading(true);
  }, [photo]);

  if (!photo) return null;

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-80 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div 
        className="relative max-w-4xl max-h-full"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          className="absolute top-4 right-4 text-white bg-black bg-opacity-50 rounded-full w-10 h-10 flex items-center justify-center hover:bg-opacity-70 transition-opacity"
          onClick={onClose}
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-white"></div>
          </div>
        )}
        
        <AuthenticatedImage
          src={getPhotoUrl(photo.id)}
          alt={photo.filename}
          className="max-w-full max-h-[80vh] object-contain"
          onLoad={() => setIsLoading(false)}
          onError={(err) => {
            console.error(`Error loading full image: ${photo.id}`, err);
            setIsLoading(false);
          }}
        />
        
        <div className="bg-black bg-opacity-70 text-white p-4 absolute bottom-0 left-0 right-0">
          <h3 className="text-lg font-medium truncate">{photo.filename}</h3>
          <p className="text-sm text-gray-300">
            {new Date(photo.created_at).toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  );
};

export default Lightbox;
