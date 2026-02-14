'use client';

import { FC, useState, useEffect, useMemo, useRef } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { usePathname } from 'next/navigation';
import { Box, Typography, Chip, IconButton } from '@mui/material';
import {
  Folder as FolderIcon,
  Description as FileIcon,
  OpenInNew as ExternalLinkIcon,
  KeyboardArrowRight as CollapseIcon,
  KeyboardArrowDown as ExpandIcon
} from '@mui/icons-material';
import { DocItem } from '@/app/docs/utils/fileSystem';
import { ProviderIcon } from '@/app/docs/components/ProviderIcon';

interface DocsSidebarProps {
  onNavigate?: () => void;
}

interface SidebarItemProps {
  item: DocItem;
  level?: number;
  onNavigate?: () => void;
  expandedPaths: Set<string>;
  onToggle: (path: string) => void;
}

const SidebarItem: FC<SidebarItemProps> = ({
  item,
  level = 0,
  onNavigate,
  expandedPaths,
  onToggle
}) => {
  const pathname = usePathname();
  const isActive = pathname === `/docs/${item.path}`;
  const [iconError, setIconError] = useState(false);

  const handleNavigate = () => {
    if (onNavigate) {
      onNavigate();
    }
  };

  // Render icon based on item type and metadata
  const renderIcon = () => {
    // Check if this is a component page under catalog/{provider}/{component}
    const pathParts = item.path.split('/');
    if (pathParts.length === 3 && pathParts[0] === 'catalog' && item.type === 'file') {
      // Use componentName from item data for icon path (survives URL slug changes)
      const provider = pathParts[1];
      const componentName = (item as DocItem & { componentName?: string }).componentName || pathParts[2];
      const componentIconPath = `/images/providers/${provider}/${componentName}/logo.svg`;

      // If icon failed to load, show a letter placeholder
      if (iconError) {
        const label = (item.title || item.name || '?').charAt(0).toUpperCase();
        return (
          <span className="w-5 h-5 flex items-center justify-center rounded bg-slate-700 text-[10px] font-bold text-gray-300 flex-shrink-0">
            {label}
          </span>
        );
      }

      return (
        <Image
          src={componentIconPath}
          alt={componentName}
          width={20}
          height={20}
          className="w-5 h-5 object-contain"
          onError={() => setIconError(true)}
        />
      );
    }

    // Check if this is a provider directory under catalog/
    const isProvider = item.path.startsWith('catalog/') && item.type === 'directory' && pathParts.length === 2;

    if (isProvider) {
      const provider = pathParts[1];
      return <ProviderIcon provider={provider} size={20} className="w-5 h-5" />;
    }

    if (item.icon) {
      return (
        <span className="text-lg" role="img" aria-label={item.title || item.name}>
          {item.icon}
        </span>
      );
    }

    if (item.type === 'directory') {
      return <FolderIcon className="text-purple-400" fontSize="small" />;
    }

    return <FileIcon className="text-gray-400" fontSize="small" />;
  };

  // Render badge if present
  const renderBadge = () => {
    if (!item.badge) return null;

    const badgeColors: Record<string, string> = {
      'Popular': 'bg-green-100 text-green-800',
      'Beta': 'bg-blue-100 text-blue-800',
      'New': 'bg-purple-100 text-purple-800',
      'Deprecated': 'bg-red-100 text-red-800',
      'Experimental': 'bg-yellow-100 text-yellow-800'
    };

    const colorClass = badgeColors[item.badge] || 'bg-gray-100 text-gray-800';

    return (
      <Chip
        label={item.badge}
        size="small"
        className={`ml-2 text-xs ${colorClass}`}
      />
    );
  };

  if (item.type === 'directory') {
    const isExpanded = expandedPaths.has(item.path);
    return (
      <Box>
        <Box
          className="flex items-center justify-between px-4 py-2 hover:bg-purple-900/20 cursor-pointer"
        >
          <Box className="flex items-center gap-2 flex-1">
            {renderIcon()}
            {item.hasIndex ? (
              <Link
                href={`/docs/${item.path}`}
                onClick={handleNavigate}
                className="flex-1"
              >
                <Typography className="text-gray-300 text-sm font-medium hover:text-purple-400">
                  {item.title || formatName(item.name)}
                </Typography>
              </Link>
            ) : (
              <Typography className="text-gray-300 text-sm font-medium">
                {item.title || formatName(item.name)}
              </Typography>
            )}
            {renderBadge()}
          </Box>
          <IconButton
            size="small"
            aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
            aria-expanded={isExpanded}
            onClick={() => onToggle(item.path)}
            className="text-gray-300"
          >
            {isExpanded ? <ExpandIcon fontSize="small" /> : <CollapseIcon fontSize="small" />}
          </IconButton>
        </Box>
        {isExpanded && (
          <Box className="ml-4">
            {item.children?.map((child, index) => (
              <SidebarItem
                key={index}
                item={child}
                level={level + 1}
                onNavigate={onNavigate}
                expandedPaths={expandedPaths}
                onToggle={onToggle}
              />
            ))}
          </Box>
        )}
      </Box>
    );
  }

  // Handle external links
  if (item.isExternal && item.externalUrl) {
    return (
      <a
        href={item.externalUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="block"
      >
        <Box className="flex items-center gap-2 px-4 py-2 hover:bg-purple-900/20 cursor-pointer text-gray-300">
          {renderIcon()}
          <Typography className="text-sm flex-1">
            {item.title || formatName(item.name)}
          </Typography>
          <ExternalLinkIcon className="text-gray-400" fontSize="small" />
          {renderBadge()}
        </Box>
      </a>
    );
  }

  return (
    <Link href={`/docs/${item.path}`} onClick={handleNavigate}>
      <Box
        data-active={isActive || undefined}
        className={`flex items-center gap-2 px-4 py-2 hover:bg-purple-900/20 cursor-pointer ${
          isActive ? 'bg-purple-600 text-white' : 'text-gray-300'
        }`}
      >
        {renderIcon()}
        <Typography className="text-sm flex-1">
          {item.title || formatName(item.name)}
        </Typography>
        {renderBadge()}
      </Box>
    </Link>
  );
};

function formatName(name: string): string {
  // Convert kebab-case or snake_case to Title Case
  return name
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, l => l.toUpperCase())
    .replace(/\s+/g, ' ')
    .trim();
}

export const DocsSidebar: FC<DocsSidebarProps> = ({ onNavigate }) => {
  const [structure, setStructure] = useState<DocItem[]>([]);
  const [loading, setLoading] = useState(true);
  const pathname = usePathname();
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set());
  const structureLoadedRef = useRef(false);

  const currentDocPath = useMemo(() => {
    // Convert pathname like /docs/platform/getting-started to platform/getting-started
    const prefix = '/docs/';
    return pathname.startsWith(prefix) ? pathname.slice(prefix.length) : '';
  }, [pathname]);

  // Load structure once on mount
  useEffect(() => {
    if (structureLoadedRef.current) return;

    const loadStructure = async () => {
      try {
        const response = await fetch('/docs-structure.json');
        if (response.ok) {
          const data = await response.json();
          setStructure(data);
          structureLoadedRef.current = true;

          // Initialize expanded paths with ancestors of current path
          if (currentDocPath) {
            const initial = new Set<string>();
            const segments = currentDocPath.split('/').filter(Boolean);
            let acc = '';
            for (const segment of segments) {
              acc = acc ? `${acc}/${segment}` : segment;
              initial.add(acc);
            }
            setExpandedPaths(initial);
          }
        }
      } catch (error) {
        console.error('Failed to load documentation structure:', error);
      } finally {
        setLoading(false);
      }
    };

    loadStructure();
  }, []);

  // On route change: merge ancestors of active page into expanded set (don't replace)
  useEffect(() => {
    if (!currentDocPath || !structureLoadedRef.current) return;
    setExpandedPaths((prev) => {
      const next = new Set(prev); // Keep existing expansions
      const segments = currentDocPath.split('/').filter(Boolean);
      let acc = '';
      for (const segment of segments) {
        acc = acc ? `${acc}/${segment}` : segment;
        next.add(acc);
      }
      return next;
    });
  }, [currentDocPath]);

  // Scroll active item into view after route change
  useEffect(() => {
    if (!currentDocPath) return;
    const timer = setTimeout(() => {
      const el = document.querySelector('[data-active="true"]');
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }, 150);
    return () => clearTimeout(timer);
  }, [currentDocPath]);

  const handleToggle = (path: string) => {
    setExpandedPaths((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  if (loading) {
    return (
      <Box className="p-4">
        <Typography className="text-gray-400">Loading...</Typography>
      </Box>
    );
  }

  return (
    <Box className="h-full overflow-y-auto">
      <Box className="py-2">
        {structure.map((item, index) => (
          <SidebarItem
            key={index}
            item={item}
            onNavigate={onNavigate}
            expandedPaths={expandedPaths}
            onToggle={handleToggle}
          />
        ))}
      </Box>
    </Box>
  );
};
