import { createContext, useContext, useState, ReactNode } from 'react';
import { Photo } from '../api';

interface SearchContextType {
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  filterPhotos: (photos: Photo[]) => Photo[];
}

const SearchContext = createContext<SearchContextType | undefined>(undefined);

export const SearchProvider = ({ children }: { children: ReactNode }) => {
  const [searchQuery, setSearchQuery] = useState('');

  const filterPhotos = (photos: Photo[]): Photo[] => {
    if (!searchQuery.trim()) {
      return photos;
    }

    const query = searchQuery.toLowerCase();
    return photos.filter(photo => {
      // Filter by filename
      if (photo.filename.toLowerCase().includes(query)) {
        return true;
      }
      
      // Filter by date if the search query looks like a date
      const createdAt = new Date(photo.created_at).toLocaleDateString();
      if (createdAt.toLowerCase().includes(query)) {
        return true;
      }
      
      return false;
    });
  };

  const value = {
    searchQuery,
    setSearchQuery,
    filterPhotos
  };

  return (
    <SearchContext.Provider value={value}>
      {children}
    </SearchContext.Provider>
  );
};

export const useSearch = (): SearchContextType => {
  const context = useContext(SearchContext);
  if (context === undefined) {
    throw new Error('useSearch must be used within a SearchProvider');
  }
  return context;
};
