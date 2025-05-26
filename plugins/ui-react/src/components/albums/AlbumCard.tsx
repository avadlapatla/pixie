import { Album } from '../../api/albums';
import { getPhotoUrl } from '../../api';
import AuthenticatedImage from '../AuthenticatedImage';

interface AlbumCardProps {
  album: Album;
  onClick: () => void;
  onDelete: () => void;
  isDeleting: boolean;
}

const AlbumCard = ({ album, onClick, onDelete, isDeleting }: AlbumCardProps) => {
  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete();
  };

  return (
    <div
      className="bg-white rounded-lg shadow-md overflow-hidden cursor-pointer group transition-transform hover:shadow-lg hover:scale-[1.02]"
      onClick={onClick}
    >
      <div className="aspect-square relative w-full bg-gray-100">
        {album.cover_photo_id ? (
          <AuthenticatedImage
            src={getPhotoUrl(album.cover_photo_id)}
            alt={album.name}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gray-200">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        )}

        <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-10 transition-opacity"></div>
        
        {/* Album count overlay */}
        <div className="absolute bottom-0 right-0 m-2 px-2 py-1 bg-black bg-opacity-70 rounded text-white text-xs">
          {album.photo_count} {album.photo_count === 1 ? 'photo' : 'photos'}
        </div>
      </div>
      
      <div className="p-3 flex justify-between items-center">
        <div className="overflow-hidden">
          <h3 className="font-medium text-gray-800 truncate">{album.name}</h3>
          {album.description && (
            <p className="text-xs text-gray-500 truncate">{album.description}</p>
          )}
        </div>
        
        <button 
          className="text-gray-500 hover:text-red-600 transition-colors p-1 rounded-full hover:bg-red-50 opacity-0 group-hover:opacity-100"
          onClick={handleDelete}
          disabled={isDeleting}
          aria-label="Delete album"
        >
          {isDeleting ? (
            <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          )}
        </button>
      </div>
    </div>
  );
};

export default AlbumCard;
