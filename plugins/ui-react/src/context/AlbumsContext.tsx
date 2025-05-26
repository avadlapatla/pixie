import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { Album, getAlbums, createAlbum, updateAlbum, deleteAlbum, getPhotosInAlbum, addPhotosToAlbum, removePhotosFromAlbum } from '../api/albums';
import { Photo, getPhotos } from '../api';

interface AlbumsContextType {
  albums: Album[];
  loading: boolean;
  error: string | null;
  selectedAlbum: Album | null;
  albumPhotos: Photo[];
  fetchAlbums: () => Promise<void>;
  createNewAlbum: (name: string, description?: string, photoIds?: string[]) => Promise<Album>;
  updateExistingAlbum: (id: string, name?: string, description?: string, coverPhotoId?: string) => Promise<Album | null>;
  removeAlbum: (id: string) => Promise<boolean>;
  selectAlbum: (id: string) => Promise<void>;
  addToAlbum: (albumId: string, photoIds: string[]) => Promise<boolean>;
  removeFromAlbum: (albumId: string, photoIds: string[]) => Promise<boolean>;
  setSelectedAlbum: (album: Album | null) => void;
  refreshAlbumPhotos: (albumId: string) => Promise<void>;
}

const AlbumsContext = createContext<AlbumsContextType | undefined>(undefined);

export const AlbumsProvider = ({ children }: { children: ReactNode }) => {
  const [albums, setAlbums] = useState<Album[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedAlbum, setSelectedAlbum] = useState<Album | null>(null);
  const [albumPhotos, setAlbumPhotos] = useState<Photo[]>([]);

  const fetchAlbums = async () => {
    setLoading(true);
    try {
      const albumsData = await getAlbums();
      setAlbums(albumsData);
      setError(null);
    } catch (err) {
      console.error('Error fetching albums:', err);
      setError('Failed to fetch albums. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const createNewAlbum = async (name: string, description?: string, photoIds?: string[]) => {
    try {
      const newAlbum = await createAlbum({ 
        name, 
        description, 
        photo_ids: photoIds 
      });
      
      setAlbums(prevAlbums => [...prevAlbums, newAlbum]);
      return newAlbum;
    } catch (err) {
      console.error('Error creating album:', err);
      throw new Error('Failed to create album');
    }
  };

  const updateExistingAlbum = async (id: string, name?: string, description?: string, coverPhotoId?: string) => {
    try {
      const updatedAlbum = await updateAlbum(id, {
        name,
        description,
        cover_photo_id: coverPhotoId
      });

      if (updatedAlbum) {
        setAlbums(prevAlbums => 
          prevAlbums.map(album => 
            album.id === id ? updatedAlbum : album
          )
        );

        if (selectedAlbum && selectedAlbum.id === id) {
          setSelectedAlbum(updatedAlbum);
        }
      }

      return updatedAlbum;
    } catch (err) {
      console.error('Error updating album:', err);
      throw new Error('Failed to update album');
    }
  };

  const removeAlbum = async (id: string) => {
    try {
      const success = await deleteAlbum(id);
      
      if (success) {
        setAlbums(prevAlbums => prevAlbums.filter(album => album.id !== id));
        
        if (selectedAlbum && selectedAlbum.id === id) {
          setSelectedAlbum(null);
          setAlbumPhotos([]);
        }
      }
      
      return success;
    } catch (err) {
      console.error('Error deleting album:', err);
      throw new Error('Failed to delete album');
    }
  };

  const selectAlbum = async (id: string) => {
    const album = albums.find(a => a.id === id);
    if (!album) return;
    
    setSelectedAlbum(album);
    await refreshAlbumPhotos(id);
  };

  const refreshAlbumPhotos = async (albumId: string) => {
    setLoading(true);
    try {
      const photoIds = await getPhotosInAlbum(albumId);
      const allPhotos = await getPhotos();
      const filteredPhotos = allPhotos.filter(photo => photoIds.includes(photo.id));
      setAlbumPhotos(filteredPhotos);
      setError(null);
    } catch (err) {
      console.error('Error fetching album photos:', err);
      setError('Failed to load photos in this album');
      setAlbumPhotos([]);
    } finally {
      setLoading(false);
    }
  };

  const addToAlbum = async (albumId: string, photoIds: string[]) => {
    try {
      const success = await addPhotosToAlbum(albumId, photoIds);
      
      if (success && selectedAlbum && selectedAlbum.id === albumId) {
        await refreshAlbumPhotos(albumId);
        
        // Update the album in the list
        const updatedAlbums = [...albums];
        const albumIndex = updatedAlbums.findIndex(a => a.id === albumId);
        
        if (albumIndex !== -1) {
          updatedAlbums[albumIndex] = {
            ...updatedAlbums[albumIndex],
            photo_count: updatedAlbums[albumIndex].photo_count + photoIds.length
          };
          
          setAlbums(updatedAlbums);
          setSelectedAlbum(updatedAlbums[albumIndex]);
        }
      }
      
      return success;
    } catch (err) {
      console.error('Error adding photos to album:', err);
      throw new Error('Failed to add photos to album');
    }
  };

  const removeFromAlbum = async (albumId: string, photoIds: string[]) => {
    try {
      const success = await removePhotosFromAlbum(albumId, photoIds);
      
      if (success && selectedAlbum && selectedAlbum.id === albumId) {
        await refreshAlbumPhotos(albumId);
        
        // Update the album in the list
        const updatedAlbums = [...albums];
        const albumIndex = updatedAlbums.findIndex(a => a.id === albumId);
        
        if (albumIndex !== -1) {
          updatedAlbums[albumIndex] = {
            ...updatedAlbums[albumIndex],
            photo_count: Math.max(0, updatedAlbums[albumIndex].photo_count - photoIds.length)
          };
          
          setAlbums(updatedAlbums);
          setSelectedAlbum(updatedAlbums[albumIndex]);
        }
      }
      
      return success;
    } catch (err) {
      console.error('Error removing photos from album:', err);
      throw new Error('Failed to remove photos from album');
    }
  };

  // Fetch albums on mount
  useEffect(() => {
    fetchAlbums();
  }, []);

  const value = {
    albums,
    loading,
    error,
    selectedAlbum,
    albumPhotos,
    fetchAlbums,
    createNewAlbum,
    updateExistingAlbum,
    removeAlbum,
    selectAlbum,
    addToAlbum,
    removeFromAlbum,
    setSelectedAlbum,
    refreshAlbumPhotos
  };

  return (
    <AlbumsContext.Provider value={value}>
      {children}
    </AlbumsContext.Provider>
  );
};

export const useAlbums = (): AlbumsContextType => {
  const context = useContext(AlbumsContext);
  if (context === undefined) {
    throw new Error('useAlbums must be used within an AlbumsProvider');
  }
  return context;
};
