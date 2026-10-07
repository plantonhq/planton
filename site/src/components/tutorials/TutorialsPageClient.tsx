'use client';

import React, { useState, useMemo } from 'react';
import TutorialListRow from './TutorialListRow';
import TutorialsSidebar from './TutorialsSidebar';
import { IconButton, Typography, Stack, Drawer, Menu, MenuItem, Button } from '@mui/material';
import { Close as CloseIcon, Sort as SortIcon, KeyboardArrowDown as KeyboardArrowDownIcon } from '@mui/icons-material';
import { Tutorial } from '@/lib/tutorials';

interface TutorialsPageClientProps {
  tutorials: Tutorial[];
  categories: string[];
}

export default function TutorialsPageClient({ tutorials, categories }: TutorialsPageClientProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedTag, setSelectedTag] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sortAnchorEl, setSortAnchorEl] = useState<null | HTMLElement>(null);
  const [sortBy, setSortBy] = useState<'date-desc' | 'date-asc' | 'title-asc' | 'title-desc'>('date-desc');

  const ITEMS_PER_PAGE = 20;

  const filteredTutorials = useMemo(() => {
    let filtered = [...tutorials];

    if (selectedCategory && selectedCategory !== 'all') {
      filtered = filtered.filter(tutorial => tutorial.category === selectedCategory);
    }

    if (selectedTag) {
      filtered = filtered.filter(tutorial => tutorial.tags.includes(selectedTag));
    }

    switch (sortBy) {
      case 'date-desc':
        filtered.sort((a, b) => new Date(b.date || '1970-01-01').getTime() - new Date(a.date || '1970-01-01').getTime());
        break;
      case 'date-asc':
        filtered.sort((a, b) => new Date(a.date || '1970-01-01').getTime() - new Date(b.date || '1970-01-01').getTime());
        break;
      case 'title-asc':
        filtered.sort((a, b) => (a.title || '').localeCompare(b.title || ''));
        break;
      case 'title-desc':
        filtered.sort((a, b) => (b.title || '').localeCompare(a.title || ''));
        break;
    }

    return filtered;
  }, [tutorials, selectedCategory, selectedTag, sortBy]);

  const displayedTutorials = useMemo(
    () => filteredTutorials.slice(0, currentPage * ITEMS_PER_PAGE),
    [filteredTutorials, currentPage],
  );
  const hasMore = displayedTutorials.length < filteredTutorials.length;

  const loadMore = () => {
    setCurrentPage(prev => prev + 1);
  };

  const handleCategoryChange = (category: string) => {
    setSelectedCategory(category);
    setSelectedTag('');
    setCurrentPage(1);
  };

  const handleTagClick = (tag: string) => {
    setSelectedTag(tag);
    setSelectedCategory('all');
    setCurrentPage(1);
  };

  const handleSidebarToggle = () => {
    setSidebarOpen(!sidebarOpen);
  };

  const handleSortClick = (event: React.MouseEvent<HTMLElement>) => {
    setSortAnchorEl(event.currentTarget);
  };

  const handleSortClose = () => {
    setSortAnchorEl(null);
  };

  const handleSortChange = (newSortBy: 'date-desc' | 'date-asc' | 'title-asc' | 'title-desc') => {
    setSortBy(newSortBy);
    setSortAnchorEl(null);
    setCurrentPage(1); // Reset to first page when sorting changes
  };

  return (
    <div className="min-h-screen font-inter antialiased">


      <div className="flex">
        {/* Left Sidebar - Sticky */}
        <div className="hidden md:block sticky top-16 h-screen w-80 flex-shrink-0">
          <div className="h-full border-r border-[#c7c7c4] overflow-y-auto">
            <TutorialsSidebar
              categories={categories}
              selectedCategory={selectedCategory}
              onCategoryChange={handleCategoryChange}
              onTagClick={handleTagClick}
            />
          </div>
        </div>

        {/* Mobile Sidebar */}
        <Drawer
          anchor="left"
          open={sidebarOpen}
          onClose={handleSidebarToggle}
          className="md:hidden"
          PaperProps={{
            className: 'w-80 bg-[#f6f6f3]',
          }}
        >
          <Stack
            direction="row"
            className="items-center justify-between p-4 border-b border-[#c7c7c4]"
          >
            <Typography variant="h6" className="text-[#171717] font-semibold">
              Tutorials
            </Typography>
            <IconButton onClick={handleSidebarToggle} className="text-[#171717]">
              <CloseIcon />
            </IconButton>
          </Stack>
          <TutorialsSidebar
            categories={categories}
            selectedCategory={selectedCategory}
            onCategoryChange={handleCategoryChange}
            onTagClick={handleTagClick}
            onNavigate={() => setSidebarOpen(false)}
          />
        </Drawer>

        {/* Main Content Area */}
        <div className="flex-1 min-h-screen">
          <div className="px-12 py-8 max-w-4xl mx-auto">
            {/* Header with Title and Sorting */}
            <div className="flex items-center justify-between mb-8">
              <h1 className="text-3xl font-bold text-[#171717]">Guides & tutorials</h1>
              <div className="flex items-center gap-2">
                <Button
                  variant="outlined"
                  onClick={handleSortClick}
                  endIcon={<KeyboardArrowDownIcon />}
                  startIcon={<SortIcon />}
                  className={`text-[#171717] border-[#999996] hover:border-[#999996] hover:text-[#171717] transition-all duration-200`}
                  size="small"
                >
                  Sort by {sortBy === 'date-desc' ? 'Newest' : sortBy === 'date-asc' ? 'Oldest' : sortBy === 'title-asc' ? 'A-Z' : 'Z-A'}
                </Button>
                <Menu
                  anchorEl={sortAnchorEl}
                  open={Boolean(sortAnchorEl)}
                  onClose={handleSortClose}
                  className="mt-2"
                  PaperProps={{
                    className: 'bg-[#eeeeeb] border border-[#c7c7c4]',
                  }}
                >
                  <MenuItem 
                    onClick={() => handleSortChange('date-desc')}
                    className={`text-[#171717] hover:bg-black/5 ${sortBy === 'date-desc' ? 'bg-white text-black' : ''}`}
                  >
                    Newest first
                  </MenuItem>
                  <MenuItem 
                    onClick={() => handleSortChange('date-asc')}
                    className={`text-[#171717] hover:bg-black/5 ${sortBy === 'date-asc' ? 'bg-white text-black' : ''}`}
                  >
                    Oldest first
                  </MenuItem>
                  <MenuItem 
                    onClick={() => handleSortChange('title-asc')}
                    className={`text-[#171717] hover:bg-black/5 ${sortBy === 'title-asc' ? 'bg-white text-black' : ''}`}
                  >
                    A-Z
                  </MenuItem>
                  <MenuItem 
                    onClick={() => handleSortChange('title-desc')}
                    className={`text-[#171717] hover:bg-black/5 ${sortBy === 'title-desc' ? 'bg-white text-black' : ''}`}
                  >
                    Z-A
                  </MenuItem>
                </Menu>
              </div>
            </div>
            
            {/* Tutorials List */}
            <div className="space-y-0">
              {displayedTutorials.map((tutorial) => (
                <TutorialListRow key={tutorial.slug} tutorial={tutorial} />
              ))}
            </div>

            {/* Load More Button */}
            {hasMore && (
              <div className="mt-8 text-center">
                <button
                  onClick={loadMore}
                  className="bg-[#fff] text-black hover:bg-white/80 px-6 py-3 rounded-lg font-medium transition-colors duration-200"
                >
                  Load More Tutorials
                </button>
              </div>
            )}

            {/* No Results */}
            {filteredTutorials.length === 0 && (
              <div className="text-center py-12">
                <h3 className="text-xl font-semibold text-[#171717] mb-2">
                  No tutorials found
                </h3>
                <p className="text-[#616161]">
                  Try adjusting your filters or browse all tutorials.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
