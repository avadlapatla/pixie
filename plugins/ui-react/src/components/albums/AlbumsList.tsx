import { useState } from 'react';
import { useAlbums } from '../../context/AlbumsContext';
import { Album } from '../../api/albums';
import CreateAlbumModal from './CreateAlbumModal';
import AlbumCard from './AlbumCard';

const AlbumsList = () => {
  const { albums, loading, error, selectAlbum, removeAlbum } = useAlbums();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [albumToDelete, setAlbumToDelete] = useState<Album | null>(null);

  const handleCreateAlbum = () => {
    setIsCreateModalOpen(true);
  };

  const handleDeleteAlbum = async (album: Album) => {
    setAlbumToDelete(album);
    
    if (confirm(`Are you sure you want to delete the album "${album.name}"? This won't delete the photos inside.`)) {
      try {
        await removeAlbum(album.id);
      } catch (err) {
        console.error('Error deleting album:', err);
        alert('Failed to delete album. Please try again.');
      }
      
      setAlbumToDelete(null);
    } else {
      setAlbumToDelete(null);
    }
  };

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
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-medium text-gray-800">Albums</h2>
        <button
          onClick={handleCreateAlbum}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors shadow-sm"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
          </svg>
          Create Album
        </button>
      </div>

      {albums.length === 0 ? (
        <div className="text-center py-20">
          <div className="text-gray-500 flex flex-col items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <p className="text-xl font-medium">No albums yet</p>
            <p className="text-sm mt-2">Create your first album to organize your photos</p>
            <button 
              className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
              onClick={handleCreateAlbum}
            >
              Create Album
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
          {albums.map((album) => (
            <AlbumCard 
              key={album.id} 
              album={album}
              onClick={() => selectAlbum(album.id)}
              onDelete={() => handleDeleteAlbum(album)}
              isDeleting={albumToDelete?.id === album.id}
            />
          ))}
        </div>
      )}

      <CreateAlbumModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </div>
  );
};

export default AlbumsList;
