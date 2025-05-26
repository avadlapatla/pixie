import { useState, useEffect } from 'react';
import { useAlbums } from '../../context/AlbumsContext';
import { Photo, getPhotos, getThumbnailUrl } from '../../api';
import AuthenticatedImage from '../AuthenticatedImage';

interface SelectPhotosModalProps {
  isOpen: boolean;
  onClose: () => void;
  albumId: string;
}

const SelectPhotosModal = ({ isOpen, onClose, albumId }: SelectPhotosModalProps) => {
  const { addToAlbum, albumPhotos } = useAlbums();
  const [allPhotos, setAllPhotos] = useState<Photo[]>([]);
  const [selectedPhotos, setSelectedPhotos] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [addingPhotos, setAddingPhotos] = useState(false);
  
  // Fetch all photos when modal opens
  useEffect(() => {
    if (isOpen) {
      fetchAllPhotos();
    }
  }, [isOpen]);

  // Reset selected photos when modal closes
  useEffect(() => {
    if (!isOpen) {
      setSelectedPhotos([]);
    }
  }, [isOpen]);

  const fetchAllPhotos = async () => {
    setLoading(true);
    try {
      const photos = await getPhotos();
      
      // Filter out photos already in the album
      const albumPhotoIds = albumPhotos.map(photo => photo.id);
      const filteredPhotos = photos.filter(photo => !albumPhotoIds.includes(photo.id));
      
      setAllPhotos(filteredPhotos);
      setError(null);
    } catch (err) {
      console.error('Error fetching photos:', err);
      setError('Failed to load photos. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const togglePhotoSelection = (photoId: string) => {
    setSelectedPhotos(prevSelected => {
      if (prevSelected.includes(photoId)) {
        return prevSelected.filter(id => id !== photoId);
      } else {
        return [...prevSelected, photoId];
      }
    });
  };

  const handleAddPhotos = async () => {
    if (selectedPhotos.length === 0) return;
    
    setAddingPhotos(true);
    try {
      await addToAlbum(albumId, selectedPhotos);
      onClose();
    } catch (err) {
      console.error('Error adding photos to album:', err);
      setError('Failed to add photos to album. Please try again.');
    } finally {
      setAddingPhotos(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 z-50 flex items-center justify-center p-4">
      <div 
        className="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] flex flex-col"
        onClick={e => e.stopPropagation()}
      >
        <div className="p-4 border-b flex justify-between items-center">
          <h2 className="text-lg font-medium">Add photos to album</h2>
          <button 
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700"
            aria-label="Close modal"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div className="p-4 flex-1 overflow-auto">
          {loading ? (
            <div className="h-64 flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : error ? (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
              <span className="block sm:inline">{error}</span>
            </div>
          ) : allPhotos.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 mx-auto text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <p>No new photos available to add to this album.</p>
            </div>
          ) : (
            <>
              <div className="mb-4">
                <p className="text-gray-700">
                  Select photos to add to this album. <span className="text-blue-600">{selectedPhotos.length}</span> selected.
                </p>
              </div>
            
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-2">
                {allPhotos.map((photo) => {
                  const isSelected = selectedPhotos.includes(photo.id);
                  const thumbnailUrl = getThumbnailUrl(photo);
                  
                  return (
                    <div 
                      key={photo.id} 
                      className={`aspect-square relative overflow-hidden rounded cursor-pointer ${
                        isSelected ? 'ring-4 ring-blue-500' : 'hover:opacity-90'
                      }`}
                      onClick={() => togglePhotoSelection(photo.id)}
                    >
                      <AuthenticatedImage 
                        src={thumbnailUrl} 
                        alt={photo.filename} 
                        className="w-full h-full object-cover"
                      />
                      
                      {/* Selection checkbox */}
                      <div className="absolute top-2 right-2">
                        <div className={`h-6 w-6 rounded-full ${
                          isSelected ? 'bg-blue-500' : 'bg-white bg-opacity-70'
                        } flex items-center justify-center`}>
                          {isSelected && (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-white" viewBox="0 0 20 20" fill="currentColor">
                              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                            </svg>
                          )}
                        </div>
                      </div>
                      
                      {/* Photo filename */}
                      <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black to-transparent text-white text-xs truncate">
                        {photo.filename}
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </div>
        
        <div className="p-4 border-t flex justify-end space-x-3">
          <button
            onClick={onClose}
            className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            Cancel
          </button>
          <button
            onClick={handleAddPhotos}
            disabled={selectedPhotos.length === 0 || addingPhotos}
            className={`px-4 py-2 rounded-md shadow-sm text-sm font-medium text-white 
              ${selectedPhotos.length === 0 
                ? 'bg-blue-400 cursor-not-allowed' 
                : 'bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
              }`}
          >
            {addingPhotos ? (
              <span className="flex items-center">
                <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Adding...
              </span>
            ) : `Add ${selectedPhotos.length} photo${selectedPhotos.length !== 1 ? 's' : ''}`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default SelectPhotosModal;
