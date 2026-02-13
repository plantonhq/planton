'use client';

import { FC, useState, useEffect, useMemo } from 'react';
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
      const provider = pathParts[1];
      const component = pathParts[2];
      const componentIconPath = `/images/providers/${provider}/${component}/logo.svg`;
      
      return (
        <Image 
          src={componentIconPath} 
          alt={component} 
          width={20}
          height={20}
          className="w-5 h-5 object-contain" 
        />
      );
    }
    
    // Check if this is a provider directory under catalog/
    const isProvider = item.path.startsWith('catalog/') && item.type === 'directory' && pathParts.length === 2;
    
    if (isProvider) {
      // Extract provider name from path (e.g., catalog/aws -> aws)
      const provider = item.path.split('/')[1];
      const providerIconMap: Record<string, string> = {
        'aws': '/images/providers/aws.svg',
        'gcp': '/images/providers/gcp.svg',
        'azure': '/images/providers/azure.svg',
        'auth0': '/images/providers/auth0.svg',
        'cloudflare': '/images/providers/cloudflare.svg',
        'civo': '/images/providers/civo.svg',
        'digitalocean': '/images/providers/digital-ocean.svg',
        'atlas': '/images/providers/mongodb-atlas.svg',
        'confluent': '/images/providers/confluent.svg',
        'kubernetes': '/images/providers/kubernetes.svg',
        'snowflake': '/images/providers/snowflake.svg',
      };
      
      const iconPath = providerIconMap[provider];
      if (iconPath) {
        return (
          <Image 
            src={iconPath} 
            alt={provider.toUpperCase()} 
            width={20}
            height={20}
            className="w-5 h-5 object-contain" 
          />
        );
      }
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

  const currentDocPath = useMemo(() => {
    // Convert pathname like /docs/platform/getting-started to platform/getting-started
    const prefix = '/docs/';
    return pathname.startsWith(prefix) ? pathname.slice(prefix.length) : '';
  }, [pathname]);

  useEffect(() => {
    const loadStructure = async () => {
      try {
        const response = await fetch('/docs-structure.json');
        if (response.ok) {
          const data = await response.json();
          setStructure(data);
          // Initialize expanded paths: only expand ancestors of current path
          const initial = new Set<string>();
          if (currentDocPath) {
            const segments = currentDocPath.split('/').filter(Boolean);
            let acc = '';
            for (const segment of segments) {
              acc = acc ? `${acc}/${segment}` : segment;
              initial.add(acc);
            }
          }
          setExpandedPaths(initial);
        }
      } catch (error) {
        console.error('Failed to load documentation structure:', error);
      } finally {
        setLoading(false);
      }
    };

    loadStructure();
  }, [currentDocPath]);

  // Ensure ancestors of the active page are expanded on route change
  useEffect(() => {
    if (!currentDocPath) return;
    setExpandedPaths(() => {
      // Only expand ancestors of the current path
      const next = new Set<string>();
      
      // Add ancestors of the current path
      const segments = currentDocPath.split('/').filter(Boolean);
      let acc = '';
      for (const segment of segments) {
        acc = acc ? `${acc}/${segment}` : segment;
        next.add(acc);
      }
      return next;
    });
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

