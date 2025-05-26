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
 * Delete a photo
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
  // For debugging
  console.log('Photo:', photo);
  console.log('Meta:', photo.meta);
  
  // Check if the photo has thumbnails and the requested size
  if (photo.meta?.thumbnails?.[size.toString()]) {
    const thumbnailPath = photo.meta.thumbnails[size.toString()];
    console.log('Thumbnail path:', thumbnailPath);
    
    // Return the thumbnail URL - auth header will be added by the fetch interceptor
    return `${API_BASE}/api/photo/${photo.id}?thumbnail=${size}`;
  }
  
  // Fall back to the original photo URL - auth header will be added by the fetch interceptor
  return getPhotoUrl(photo.id);
};
