/**
 * Interface for an album object
 */
export interface Album {
  id: string;
  name: string;
  description?: string;
  cover_photo_id?: string;
  created_at: string;
  updated_at: string;
  photo_count: number;
}

/**
 * Interface for album creation request
 */
export interface CreateAlbumRequest {
  name: string;
  description?: string;
  photo_ids?: string[];
}

/**
 * Interface for album update request
 */
export interface UpdateAlbumRequest {
  name?: string;
  description?: string;
  cover_photo_id?: string;
}

/**
 * Get all albums
 */
export const getAlbums = async (): Promise<Album[]> => {
  // In a real implementation, this would call the backend API
  // For now, we'll return mock data that's stored in localStorage
  
  const storedAlbums = localStorage.getItem('pixie_albums');
  if (storedAlbums) {
    return JSON.parse(storedAlbums);
  }
  
  return [];
};

/**
 * Get a single album by ID
 */
export const getAlbum = async (id: string): Promise<Album | null> => {
  const albums = await getAlbums();
  return albums.find(album => album.id === id) || null;
};

/**
 * Create a new album
 */
export const createAlbum = async (albumData: CreateAlbumRequest): Promise<Album> => {
  // In a real implementation, this would call the backend API
  // For now, we'll store the album in localStorage
  
  const albums = await getAlbums();
  
  const newAlbum: Album = {
    id: `album_${Date.now()}`,
    name: albumData.name,
    description: albumData.description || '',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    photo_count: albumData.photo_ids?.length || 0
  };
  
  albums.push(newAlbum);
  
  localStorage.setItem('pixie_albums', JSON.stringify(albums));
  
  // If we have photo IDs, add them to the album
  if (albumData.photo_ids && albumData.photo_ids.length > 0) {
    await addPhotosToAlbum(newAlbum.id, albumData.photo_ids);
  }
  
  return newAlbum;
};

/**
 * Update an existing album
 */
export const updateAlbum = async (id: string, albumData: UpdateAlbumRequest): Promise<Album | null> => {
  const albums = await getAlbums();
  const albumIndex = albums.findIndex(album => album.id === id);
  
  if (albumIndex === -1) {
    return null;
  }
  
  const updatedAlbum = {
    ...albums[albumIndex],
    ...albumData,
    updated_at: new Date().toISOString()
  };
  
  albums[albumIndex] = updatedAlbum;
  
  localStorage.setItem('pixie_albums', JSON.stringify(albums));
  
  return updatedAlbum;
};

/**
 * Delete an album
 */
export const deleteAlbum = async (id: string): Promise<boolean> => {
  const albums = await getAlbums();
  const filteredAlbums = albums.filter(album => album.id !== id);
  
  if (filteredAlbums.length === albums.length) {
    return false;
  }
  
  localStorage.setItem('pixie_albums', JSON.stringify(filteredAlbums));
  
  // Also remove album-photo relationships
  const albumPhotos = await getAlbumPhotos();
  delete albumPhotos[id];
  localStorage.setItem('pixie_album_photos', JSON.stringify(albumPhotos));
  
  return true;
};

/**
 * Get photos in an album
 */
export const getPhotosInAlbum = async (albumId: string): Promise<string[]> => {
  const albumPhotos = await getAlbumPhotos();
  return albumPhotos[albumId] || [];
};

/**
 * Add photos to an album
 */
export const addPhotosToAlbum = async (albumId: string, photoIds: string[]): Promise<boolean> => {
  const albumPhotos = await getAlbumPhotos();
  
  const existingPhotoIds = albumPhotos[albumId] || [];
  const uniquePhotoIds = Array.from(new Set([...existingPhotoIds, ...photoIds]));
  
  albumPhotos[albumId] = uniquePhotoIds;
  localStorage.setItem('pixie_album_photos', JSON.stringify(albumPhotos));
  
  // Update the photo count in the album
  const albums = await getAlbums();
  const albumIndex = albums.findIndex(album => album.id === albumId);
  
  if (albumIndex !== -1) {
    albums[albumIndex].photo_count = uniquePhotoIds.length;
    albums[albumIndex].updated_at = new Date().toISOString();
    localStorage.setItem('pixie_albums', JSON.stringify(albums));
  }
  
  return true;
};

/**
 * Remove photos from an album
 */
export const removePhotosFromAlbum = async (albumId: string, photoIds: string[]): Promise<boolean> => {
  const albumPhotos = await getAlbumPhotos();
  
  if (!albumPhotos[albumId]) {
    return false;
  }
  
  albumPhotos[albumId] = albumPhotos[albumId].filter(id => !photoIds.includes(id));
  localStorage.setItem('pixie_album_photos', JSON.stringify(albumPhotos));
  
  // Update the photo count in the album
  const albums = await getAlbums();
  const albumIndex = albums.findIndex(album => album.id === albumId);
  
  if (albumIndex !== -1) {
    albums[albumIndex].photo_count = albumPhotos[albumId].length;
    albums[albumIndex].updated_at = new Date().toISOString();
    localStorage.setItem('pixie_albums', JSON.stringify(albums));
  }
  
  return true;
};

/**
 * Get album-photo relationships
 */
const getAlbumPhotos = async (): Promise<Record<string, string[]>> => {
  const storedAlbumPhotos = localStorage.getItem('pixie_album_photos');
  
  if (storedAlbumPhotos) {
    return JSON.parse(storedAlbumPhotos);
  }
  
  return {};
};
