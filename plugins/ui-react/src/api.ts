/**
 * API client for the Pixie backend
 */

// Base URL for API requests
const API_BASE = import.meta.env.VITE_API_BASE || '';

/**
 * Interface for API response with photos
 */
export interface PhotosResponse {
  photos: Photo[];
}

/**
 * Interface for a photo object
 */
export interface Photo {
  id: string;
  filename: string;
  mime: string;
  created_at: string;
  deleted_at?: string;
  status?: string;
  meta?: {
    thumbnails?: {
      [size: string]: string;
    };
  };
}

/**
 * Get the JWT token from localStorage
 */
export const getToken = (): string | null => {
  return localStorage.getItem('token');
};

/**
 * Set the JWT token in localStorage
 */
export const setToken = (token: string): void => {
  localStorage.setItem('token', token);
};

/**
 * Clear the JWT token from localStorage
 */
export const clearToken = (): void => {
  localStorage.removeItem('token');
};

/**
 * Check if the user is authenticated
 */
export const isAuthenticated = (): boolean => {
  return !!getToken();
};

/**
 * Fetch wrapper that adds the Authorization header with the JWT token
 */
export const fetchWithAuth = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const token = getToken();
  const headers = {
    ...options.headers,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  return fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  });
};

/**
 * Get all photos
 */
export const getPhotos = async (): Promise<Photo[]> => {
  const response = await fetchWithAuth('/api/photos');
  
  if (!response.ok) {
    throw new Error(`Failed to fetch photos: ${response.statusText}`);
  }
  
  const data: PhotosResponse = await response.json();
  return data.photos;
};

/**
 * Upload a photo
 */
export const uploadPhoto = async (file: File): Promise<Photo> => {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetchWithAuth('/api/upload', {
    method: 'POST',
    body: formData,
  });
  
  if (!response.ok) {
    throw new Error(`Failed to upload photo: ${response.statusText}`);
  }
  
  return response.json();
};

/**
 * Move a photo to trash
 */
export const trashPhoto = async (id: string): Promise<void> => {
  const response = await fetchWithAuth(`/api/photos/trash/${id}`, {
    method: 'PUT',
  });
  
  if (!response.ok) {
    throw new Error(`Failed to move photo to trash: ${response.statusText}`);
  }
};

/**
 * Restore a photo from trash
 */
export const restorePhoto = async (id: string): Promise<void> => {
  const response = await fetchWithAuth(`/api/photos/trash/${id}/restore`, {
    method: 'PUT',
  });
  
  if (!response.ok) {
    throw new Error(`Failed to restore photo from trash: ${response.statusText}`);
  }
};

/**
 * Permanently delete a photo from trash
 */
export const permanentlyDeletePhoto = async (id: string): Promise<void> => {
  const response = await fetchWithAuth(`/api/photos/trash/${id}`, {
    method: 'DELETE',
  });
  
  if (!response.ok) {
    throw new Error(`Failed to delete photo permanently: ${response.statusText}`);
  }
};

/**
 * Get all photos in trash
 */
export const getTrash = async (): Promise<Photo[]> => {
  const response = await fetchWithAuth('/api/photos/trash');
  
  if (!response.ok) {
    throw new Error(`Failed to fetch trash: ${response.statusText}`);
  }
  
  const data = await response.json();
  // Ensure we always return an array, even if the API returns null
  return data.photos || [];
};

/**
 * Empty the trash (delete all trashed photos permanently)
 */
export const emptyTrash = async (): Promise<{count: number}> => {
  const response = await fetchWithAuth('/api/photos/trash', {
    method: 'DELETE',
  });
  
  if (!response.ok) {
    throw new Error(`Failed to empty trash: ${response.statusText}`);
  }
  
  return response.json();
};

/**
 * Delete a photo (this is kept for backward compatibility)
 */
export const deletePhoto = async (id: string): Promise<void> => {
  const response = await fetchWithAuth(`/api/photo/${id}`, {
    method: 'DELETE',
  });
  
  if (!response.ok) {
    throw new Error(`Failed to delete photo: ${response.statusText}`);
  }
};

/**
 * Get the URL for a photo
 */
export const getPhotoUrl = (id: string): string => {
  return `${API_BASE}/api/photo/${id}`;
};

/**
 * Get the URL for a photo thumbnail
 * @param photo The photo object
 * @param size The thumbnail size (default: 512)
 * @returns The thumbnail URL or the original photo URL if no thumbnail is available
 */
export const getThumbnailUrl = (photo: Photo, size: number = 512): string => {
  try {
    // Safely check if the photo has thumbnails for the requested size
    if (photo && photo.meta && photo.meta.thumbnails && photo.meta.thumbnails[size.toString()]) {
      // Return the thumbnail URL - auth header will be added by the fetch interceptor
      return `${API_BASE}/api/photo/${photo.id}?thumbnail=${size}`;
    }
  } catch (error) {
    console.error("Error generating thumbnail URL:", error);
  }
  
  // Fall back to the original photo URL - auth header will be added by the fetch interceptor
  return getPhotoUrl(photo?.id || '');
};
