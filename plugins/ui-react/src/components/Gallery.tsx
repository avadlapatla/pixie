import { useState, useEffect } from 'react';
import { Photo, getPhotos, getThumbnailUrl } from '../api';
import AuthenticatedImage from './AuthenticatedImage';

interface GalleryProps {
  onPhotoClick: (photo: Photo) => void;
  newPhoto: Photo | null;
  searchQuery?: string;
  onDeletePhoto?: (photoId: string) => void;
}

const Gallery = ({ onPhotoClick, newPhoto, searchQuery = "", onDeletePhoto }: GalleryProps) => {
  const [photos, setPhotos] = useState<Photo[]>([]);
  const [filteredPhotos, setFilteredPhotos] = useState<Photo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dateGroups, setDateGroups] = useState<{[key: string]: Photo[]}>({});

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

  // Filter photos when the search query changes or when photos array changes
  useEffect(() => {
    if (!searchQuery.trim()) {
      setFilteredPhotos(photos);
      return;
    }

    const query = searchQuery.toLowerCase();
    const filtered = photos.filter(photo => {
      // Filter by filename
      if (photo.filename.toLowerCase().includes(query)) {
        return true;
      }
      
      // Filter by date
      const createdAt = new Date(photo.created_at).toLocaleDateString();
      if (createdAt.toLowerCase().includes(query)) {
        return true;
      }
      
      return false;
    });

    setFilteredPhotos(filtered);
  }, [photos, searchQuery]);
  
  // Group photos by date
  useEffect(() => {
    const photosToGroup = searchQuery ? filteredPhotos : photos;
    const groups: {[key: string]: Photo[]} = {};
    
    photosToGroup.forEach(photo => {
      const date = new Date(photo.created_at);
      const dateKey = date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'long',
        day: 'numeric'
      });
      
      if (!groups[dateKey]) {
        groups[dateKey] = [];
      }
      
      groups[dateKey].push(photo);
    });
    
    setDateGroups(groups);
  }, [photos, filteredPhotos, searchQuery]);

  // Pass photo deletion to parent when needed
  useEffect(() => {
    if (onDeletePhoto) {
      // This effect is just to prevent the TypeScript warning
      // The actual deletion is handled by the parent component
    }
  }, [onDeletePhoto]);

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
      <div className="text-center py-20">
        <div className="text-gray-500 flex flex-col items-center">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <p className="text-xl font-medium">No photos yet</p>
          <p className="text-sm mt-2">Upload your first photo to get started</p>
        </div>
      </div>
    );
  }

  if (filteredPhotos.length === 0 && searchQuery.trim()) {
    return (
      <div className="text-center py-20">
        <div className="text-gray-500 flex flex-col items-center">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <p className="text-xl font-medium">No results found</p>
          <p className="text-sm mt-2">No photos match your search: "{searchQuery}"</p>
          <button 
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
            onClick={() => window.location.reload()}
          >
            Clear search
          </button>
        </div>
      </div>
    );
  }
  
  return (
    <div className="space-y-8">
      {Object.entries(dateGroups).map(([date, datePhotos]) => (
        <div key={date} className="mb-8">
          <h3 className="text-lg font-medium text-gray-800 sticky top-16 bg-gray-50 py-2 z-10 border-b mb-4">
            {date}
          </h3>
          
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-1 md:gap-2">
            {datePhotos.map((photo) => {
              const thumbnailUrl = getThumbnailUrl(photo);
              
              return (
                <div 
                  key={photo.id} 
                  className="aspect-square relative overflow-hidden rounded-sm cursor-pointer group"
                  onClick={() => onPhotoClick(photo)}
                >
                  <div className="absolute inset-0 bg-black opacity-0 group-hover:opacity-20 transition-opacity z-10"></div>
                  
                  <AuthenticatedImage 
                    src={thumbnailUrl} 
                    alt={photo.filename} 
                    className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                    onError={(err) => {
                      console.error(`Error loading image: ${thumbnailUrl}`, err);
                    }}
                  />
                  
                  {/* Hover overlay with file info */}
                  <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black to-transparent opacity-0 group-hover:opacity-100 transition-opacity text-white z-10">
                    <p className="text-xs truncate">{photo.filename}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
};

export default Gallery;
