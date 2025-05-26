import { useState, useEffect } from 'react';
import { Photo, getPhotoUrl, deletePhoto } from '../api';
import AuthenticatedImage from './AuthenticatedImage';

interface LightboxProps {
  photo: Photo | null;
  onClose: () => void;
  onDelete?: (photoId: string) => void;
}

const Lightbox = ({ photo, onClose, onDelete }: LightboxProps) => {
  const [isLoading, setIsLoading] = useState(true);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-lg font-medium truncate">{photo.filename}</h3>
              <p className="text-sm text-gray-300">
                {new Date(photo.created_at).toLocaleString()}
              </p>
            </div>
            <div className="flex space-x-3">
              <button
                className="text-white bg-red-600 hover:bg-red-700 px-3 py-1 rounded-md text-sm flex items-center transition-colors"
                onClick={async (e) => {
                  e.stopPropagation();
                  if (!photo || isDeleting) return;

                  if (window.confirm("Are you sure you want to delete this photo? This action cannot be undone.")) {
                    setIsDeleting(true);
                    setError(null);
                    
                    try {
                      await deletePhoto(photo.id);
                      if (onDelete) {
                        onDelete(photo.id);
                      }
                      onClose();
                    } catch (err) {
                      console.error("Failed to delete photo:", err);
                      setError("Failed to delete photo. Please try again.");
                      setIsDeleting(false);
                    }
                  }
                }}
                disabled={isDeleting}
              >
                {isDeleting ? (
                  <span className="flex items-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Deleting...
                  </span>
                ) : (
                  <span className="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    Delete
                  </span>
                )}
              </button>
            </div>
          </div>
          {error && (
            <div className="mt-2 text-sm text-red-300 bg-red-900 bg-opacity-50 p-2 rounded">
              {error}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Lightbox;
