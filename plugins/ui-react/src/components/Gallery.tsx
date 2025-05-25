import { useState, useEffect } from 'react';
import { Photo, getPhotos, getPhotoUrl } from '../api';

interface GalleryProps {
  onPhotoClick: (photo: Photo) => void;
  newPhoto: Photo | null;
}

const Gallery = ({ onPhotoClick, newPhoto }: GalleryProps) => {
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPhotos = async () => {
      try {
        setLoading(true);
        const fetchedPhotos = await getPhotos();
        setPhotos(fetchedPhotos);
        setError(null);
      } catch (err) {
        setError('Failed to load photos. Please check your connection and try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchPhotos();
  }, []);

  // Add new photo to the gallery when it's uploaded
  useEffect(() => {
    if (newPhoto) {
      setPhotos(prevPhotos => [newPhoto, ...prevPhotos]);
    }
  }, [newPhoto]);

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

  if (photos.length === 0) {
    return (
      <div className="text-center py-10">
        <p className="text-gray-500">No photos yet. Upload your first photo!</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      {photos.map((photo) => (
        <div 
          key={photo.id} 
          className="bg-white rounded-lg shadow-md overflow-hidden cursor-pointer transform transition-transform hover:scale-105"
          onClick={() => onPhotoClick(photo)}
        >
          <img 
            src={getPhotoUrl(photo.id)} 
            alt={photo.filename} 
            className="w-full h-48 object-cover"
            loading="lazy"
          />
          <div className="p-3">
            <p className="text-sm text-gray-700 truncate">{photo.filename}</p>
            <p className="text-xs text-gray-500">{new Date(photo.created_at).toLocaleDateString()}</p>
          </div>
        </div>
      ))}
    </div>
  );
};

export default Gallery;
