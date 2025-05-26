import { useState } from 'react';
import { useAlbums } from '../../context/AlbumsContext';

interface CreateAlbumModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialPhotoIds?: string[];
}

const CreateAlbumModal = ({ 
  isOpen, 
  onClose, 
  initialPhotoIds = [] 
}: CreateAlbumModalProps) => {
  const { createNewAlbum } = useAlbums();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError('Album name is required');
      return;
    }
    
    setIsCreating(true);
    setError(null);
    
    try {
      await createNewAlbum(
        name.trim(), 
        description.trim() || undefined,
        initialPhotoIds.length > 0 ? initialPhotoIds : undefined
      );
      
      // Reset form
      setName('');
      setDescription('');
      setIsCreating(false);
      
      // Close modal
      onClose();
    } catch (err) {
      console.error('Error creating album:', err);
      setError('Failed to create album. Please try again.');
      setIsCreating(false);
    }
  };

  const handleCancel = () => {
    setName('');
    setDescription('');
    setError(null);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Create New Album</h2>
        
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="albumName" className="block text-sm font-medium text-gray-700 mb-1">
              Album Name<span className="text-red-500">*</span>
            </label>
            <input
              id="albumName"
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="My Album"
              required
              disabled={isCreating}
              autoFocus
            />
          </div>
          
          <div className="mb-4">
            <label htmlFor="albumDescription" className="block text-sm font-medium text-gray-700 mb-1">
              Description (optional)
            </label>
            <textarea
              id="albumDescription"
              value={description}
              onChange={e => setDescription(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="A description of your album"
              rows={3}
              disabled={isCreating}
            />
          </div>
          
          {initialPhotoIds.length > 0 && (
            <div className="mb-4 bg-blue-50 p-3 rounded-md">
              <p className="text-sm text-blue-800">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 inline-block mr-1" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5z" />
                  <path d="M11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
                {initialPhotoIds.length} {initialPhotoIds.length === 1 ? 'photo' : 'photos'} will be added to this album.
              </p>
            </div>
          )}
          
          {error && (
            <div className="mb-4 bg-red-50 border border-red-200 text-red-700 p-3 rounded-md">
              {error}
            </div>
          )}
          
          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={handleCancel}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              disabled={isCreating}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              disabled={isCreating}
            >
              {isCreating ? (
                <span className="flex items-center">
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Creating...
                </span>
              ) : 'Create Album'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateAlbumModal;
