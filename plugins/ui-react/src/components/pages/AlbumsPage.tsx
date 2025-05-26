import { useState, useEffect } from 'react';
import { useAlbums } from '../../context/AlbumsContext';
import AlbumsList from '../albums/AlbumsList';
import AlbumView from '../albums/AlbumView';
import { Photo } from '../../api';

interface AlbumsPageProps {
  onPhotoClick: (photo: Photo) => void;
}

const AlbumsPage = ({ onPhotoClick }: AlbumsPageProps) => {
  const { selectedAlbum, setSelectedAlbum } = useAlbums();
  const [viewMode, setViewMode] = useState<'list' | 'detail'>('list');

  // Update view mode when selected album changes
  useEffect(() => {
    if (selectedAlbum) {
      setViewMode('detail');
    } else {
      setViewMode('list');
    }
  }, [selectedAlbum]);

  // Go back to albums list
  const handleBackToAlbums = () => {
    setSelectedAlbum(null);
    setViewMode('list');
  };

  return (
    <div className="space-y-6">
      {viewMode === 'detail' && selectedAlbum && (
        <div className="mb-4">
          <button
            onClick={handleBackToAlbums}
            className="flex items-center text-blue-600 hover:text-blue-800 transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clipRule="evenodd" />
            </svg>
            Back to albums
          </button>
        </div>
      )}

      {viewMode === 'list' ? (
        <AlbumsList />
      ) : (
        <AlbumView onPhotoClick={onPhotoClick} />
      )}
    </div>
  );
};

export default AlbumsPage;
