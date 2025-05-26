import { useEffect, useState } from 'react';
import { useAlbums } from '../../context/AlbumsContext';
import { Photo } from '../../api';
import { getThumbnailUrl } from '../../api';
import AuthenticatedImage from '../AuthenticatedImage';
import SelectPhotosModal from './SelectPhotosModal';

interface AlbumViewProps {
  onPhotoClick: (photo: Photo) => void;
}

const AlbumView = ({ onPhotoClick }: AlbumViewProps) => {
  const { selectedAlbum, albumPhotos, loading, error, updateExistingAlbum, removeFromAlbum } = useAlbums();
  const [isAddingPhotos, setIsAddingPhotos] = useState(false);
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [albumName, setAlbumName] = useState('');
  const [albumDescription, setAlbumDescription] = useState('');

  // Initialize form values when album changes
  useEffect(() => {
    if (selectedAlbum) {
      setAlbumName(selectedAlbum.name);
      setAlbumDescription(selectedAlbum.description || '');
    }
  }, [selectedAlbum]);

  const handleSaveTitle = async () => {
    if (selectedAlbum && albumName.trim()) {
      await updateExistingAlbum(
        selectedAlbum.id, 
        albumName.trim(),
        albumDescription.trim() || undefined
      );
      setIsEditingTitle(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveTitle();
    } else if (e.key === 'Escape') {
      // Cancel editing and revert to original values
      if (selectedAlbum) {
        setAlbumName(selectedAlbum.name);
        setAlbumDescription(selectedAlbum.description || '');
      }
      setIsEditingTitle(false);
    }
  };

  const handleRemovePhoto = async (photoId: string) => {
    if (selectedAlbum && confirm('Remove this photo from the album?')) {
      try {
        await removeFromAlbum(selectedAlbum.id, [photoId]);
      } catch (err) {
        console.error('Error removing photo from album:', err);
        alert('Failed to remove photo from album. Please try again.');
      }
    }
  };

  if (!selectedAlbum) {
    return (
      <div className="text-center py-20 text-gray-500">
        Select an album to view its contents
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <strong className="font-bold">Error!</strong>
        <span className="block sm:inline"> {error}</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Album header */}
      <div className="flex justify-between items-start">
        <div className="flex-1">
          {isEditingTitle ? (
            <div className="space-y-2">
              <input
                type="text"
                value={albumName}
                onChange={(e) => setAlbumName(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-xl font-medium"
                placeholder="Album Name"
                autoFocus
              />
              <textarea
                value={albumDescription}
                onChange={(e) => setAlbumDescription(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
                placeholder="Add a description (optional)"
                rows={2}
              />
              <div className="flex space-x-2">
                <button
                  onClick={handleSaveTitle}
                  className="px-3 py-1 text-sm text-white bg-blue-600 rounded hover:bg-blue-700"
                >
                  Save
                </button>
                <button
                  onClick={() => {
                    if (selectedAlbum) {
                      setAlbumName(selectedAlbum.name);
                      setAlbumDescription(selectedAlbum.description || '');
                    }
                    setIsEditingTitle(false);
                  }}
                  className="px-3 py-1 text-sm text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div
              className="cursor-pointer group"
              onClick={() => setIsEditingTitle(true)}
            >
              <h2 className="text-2xl font-medium text-gray-800 group-hover:text-blue-600 flex items-center">
                {selectedAlbum.name}
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 ml-2 opacity-0 group-hover:opacity-100 text-blue-600" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                </svg>
              </h2>
              {selectedAlbum.description && (
                <p className="text-gray-500 mt-1 group-hover:text-gray-700">{selectedAlbum.description}</p>
              )}
              <p className="text-sm text-gray-400 mt-1">
                {selectedAlbum.photo_count} {selectedAlbum.photo_count === 1 ? 'photo' : 'photos'}
              </p>
            </div>
          )}
        </div>

        <button
          onClick={() => setIsAddingPhotos(true)}
          className="flex items-center px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition-colors"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
          </svg>
          Add Photos
        </button>
      </div>

      {/* Photos grid */}
      {albumPhotos.length === 0 ? (
        <div className="text-center py-20">
          <div className="text-gray-500 flex flex-col items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <p className="text-xl font-medium">No photos in this album</p>
            <p className="text-sm mt-2">Add photos to get started</p>
            <button 
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
              onClick={() => setIsAddingPhotos(true)}
            >
              Add Photos
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-2">
          {albumPhotos.map((photo) => {
            const thumbnailUrl = getThumbnailUrl(photo);
            
            return (
              <div 
                key={photo.id} 
                className="aspect-square relative overflow-hidden rounded-sm cursor-pointer group"
              >
                <div 
                  className="w-full h-full"
                  onClick={() => onPhotoClick(photo)}
                >
                  <AuthenticatedImage 
                    src={thumbnailUrl} 
                    alt={photo.filename} 
                    className="w-full h-full object-cover"
                    onError={(err) => {
                      console.error(`Error loading image: ${thumbnailUrl}`, err);
                    }}
                  />
                </div>
                
                {/* Hover overlay with photo controls */}
                <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-40 transition-opacity flex justify-end items-start">
                  <button
                    className="m-2 p-1.5 bg-white rounded-full opacity-0 group-hover:opacity-100 transform translate-y-2 group-hover:translate-y-0 transition-all text-gray-600 hover:text-red-600"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleRemovePhoto(photo.id);
                    }}
                    title="Remove from album"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  </button>
                </div>

                {/* File info shown on hover */}
                <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black to-transparent opacity-0 group-hover:opacity-100 transition-opacity text-white z-10">
                  <p className="text-xs truncate">{photo.filename}</p>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Add Photos Modal */}
      <SelectPhotosModal 
        isOpen={isAddingPhotos}
        onClose={() => setIsAddingPhotos(false)}
        albumId={selectedAlbum.id}
      />
    </div>
  );
};

export default AlbumView;
