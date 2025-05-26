import { useState, useEffect } from 'react';
import { Photo, getTrash, restorePhoto, permanentlyDeletePhoto, emptyTrash, getThumbnailUrl } from '../../api';
import AuthenticatedImage from '../AuthenticatedImage';

interface TrashPageProps {
  onPhotoClick: (photo: Photo) => void;
}

const TrashPage = ({ onPhotoClick }: TrashPageProps) => {
  const [trashedPhotos, setTrashedPhotos] = useState<Photo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEmptyingTrash, setIsEmptyingTrash] = useState(false);
  const [dateGroups, setDateGroups] = useState<{[key: string]: Photo[]}>({});

  // Fetch all trashed photos
  const fetchTrash = async () => {
    try {
      setLoading(true);
      const photos = await getTrash();
      // Ensure we always set an array, even if API returns null
      setTrashedPhotos(photos || []);
      setError(null);
    } catch (err) {
      setTrashedPhotos([]);
      setError('Failed to load trash. Please check your connection and try again.');
    } finally {
      setLoading(false);
    }
  };

  // Load trashed photos on component mount
  useEffect(() => {
    fetchTrash();
  }, []);
  
  // Group photos by deletion date
  useEffect(() => {
    const groups: {[key: string]: Photo[]} = {};
    
    trashedPhotos.forEach(photo => {
      // Default to created_at if deleted_at is missing
      let date: Date;
      let dateKey: string;
      
      if (photo.deleted_at) {
        date = new Date(photo.deleted_at);
      } else {
        date = new Date(photo.created_at);
      }
      
      dateKey = date.toLocaleDateString('en-US', { 
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
  }, [trashedPhotos]);

  // Handle restoring a photo
  const handleRestore = async (photo: Photo, e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent opening the lightbox
    
    try {
      await restorePhoto(photo.id);
      // Remove photo from the list on success
      setTrashedPhotos(photos => photos.filter(p => p.id !== photo.id));
    } catch (err) {
      console.error('Failed to restore photo:', err);
      setError('Failed to restore photo. Please try again.');
    }
  };

  // Handle permanently deleting a photo
  const handleDelete = async (photo: Photo, e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent opening the lightbox
    
    if (window.confirm('Permanently delete this photo? This cannot be undone.')) {
      try {
        await permanentlyDeletePhoto(photo.id);
        // Remove photo from the list on success
        setTrashedPhotos(photos => photos.filter(p => p.id !== photo.id));
      } catch (err) {
        console.error('Failed to delete photo permanently:', err);
        setError('Failed to delete photo. Please try again.');
      }
    }
  };

  // Handle emptying the entire trash
  const handleEmptyTrash = async () => {
    if (window.confirm('Empty trash? This will permanently delete all photos in the trash and cannot be undone.')) {
      try {
        setIsEmptyingTrash(true);
        await emptyTrash();
        setTrashedPhotos([]);
      } catch (err) {
        console.error('Failed to empty trash:', err);
        setError('Failed to empty trash. Please try again.');
      } finally {
        setIsEmptyingTrash(false);
      }
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
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-medium text-gray-800">Trash</h2>
        {trashedPhotos.length > 0 && (
          <button
            onClick={handleEmptyTrash}
            disabled={isEmptyingTrash}
            className="bg-red-600 hover:bg-red-700 text-white py-2 px-4 rounded-md transition-colors disabled:opacity-70 disabled:cursor-not-allowed flex items-center"
          >
            {isEmptyingTrash ? (
              <>
                <svg className="animate-spin -ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Emptying...
              </>
            ) : (
              <>
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                Empty Trash
              </>
            )}
          </button>
        )}
      </div>
      
      {trashedPhotos.length === 0 ? (
        <div className="text-center py-20 bg-white rounded-lg shadow-sm">
          <div className="text-gray-500 flex flex-col items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <p className="text-xl font-medium">Trash is empty</p>
            <p className="text-sm mt-2">Deleted photos will appear here</p>
          </div>
        </div>
      ) : (
      
      <div className="space-y-8">
        {Object.entries(dateGroups || {}).map(([date, datePhotos]) => {
          if (!datePhotos || !Array.isArray(datePhotos) || datePhotos.length === 0) return null;
          
          return (
            <div key={date} className="mb-8">
              <h3 className="text-lg font-medium text-gray-800 sticky top-16 bg-gray-50 py-2 z-10 border-b mb-4">
                {date}
              </h3>
              
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-1 md:gap-2">
                {datePhotos.map((photo) => {
                  if (!photo || !photo.id) return null;
                  const thumbnailUrl = getThumbnailUrl(photo);
                  
                  return (
                    <div 
                      key={photo.id} 
                      className="aspect-square relative overflow-hidden rounded-sm cursor-pointer group"
                      onClick={() => onPhotoClick(photo)}
                    >
                      <div className="absolute inset-0 bg-black bg-opacity-40 opacity-0 group-hover:opacity-100 transition-opacity z-10 flex flex-col justify-end"></div>
                      
                      <AuthenticatedImage 
                        src={thumbnailUrl} 
                        alt={photo.filename} 
                        className="w-full h-full object-cover transition-transform duration-500 filter brightness-75"
                      />
                      
                      {/* Hover overlay with file info and actions */}
                      <div className="absolute inset-0 p-2 flex flex-col justify-between opacity-0 group-hover:opacity-100 transition-opacity text-white z-20">
                        <div className="flex justify-end space-x-2">
                          <button 
                            onClick={(e) => handleRestore(photo, e)}
                            className="p-2 bg-green-600 bg-opacity-80 rounded-full hover:bg-opacity-100 transition-opacity"
                            title="Restore photo"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                            </svg>
                          </button>
                          <button 
                            onClick={(e) => handleDelete(photo, e)}
                            className="p-2 bg-red-600 bg-opacity-80 rounded-full hover:bg-opacity-100 transition-opacity"
                            title="Delete permanently"
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                        <div>
                          <p className="text-sm truncate">{photo.filename}</p>
                          <p className="text-xs opacity-75">
                            Deleted {photo.deleted_at ? new Date(photo.deleted_at).toLocaleString() : 'Unknown date'}
                          </p>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
      )}
    </div>
  );
};

export default TrashPage;
