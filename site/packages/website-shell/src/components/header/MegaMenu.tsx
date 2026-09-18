'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Menu, Paper, Stack, Typography } from '@mui/material';
import { NavigateNext, KeyboardArrowDown } from '@mui/icons-material';
import { MegaMenuItem } from './MegaMenuItem';
import type { MenuSection, MenuItem } from '../../data/navigation';
import { tokens } from '../../theme/tokens';

interface MegaMenuProps {
  title: string;
  leftMenu: MenuSection[];
  rightMenu?: MenuSection[];
  footerMenu?: MenuItem;
  leftWidth?: number;
}

export function MegaMenu({
  title,
  leftMenu,
  rightMenu,
  footerMenu,
  leftWidth = 270,
}: MegaMenuProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <>
      <Stack
        aria-controls={open ? 'mega-menu' : undefined}
        aria-expanded={open ? 'true' : undefined}
        aria-haspopup="true"
        direction="row"
        onClick={handleClick}
        sx={{
          cursor: 'pointer',
          alignItems: 'center',
          '&:hover': { color: tokens.text.primary },
        }}
      >
        <Typography
          sx={{
            fontSize: '0.875rem',
            fontWeight: 500,
            color: 'grey.100',
            transition: 'color 150ms ease',
          }}
        >
          {title}
        </Typography>
        <KeyboardArrowDown
          sx={{
            fontSize: '20px !important',
            fontVariationSettings: "'FILL' 1, 'wght' 500, 'GRAD' 200, 'opsz' 48",
          }}
        />
      </Stack>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{ sx: { padding: 0 } }}
        slotProps={{
          paper: {
            sx: {
              mt: 1.5,
              backgroundColor: 'transparent',
              backgroundImage: 'none',
              boxShadow: '0 0 0 1px rgba(255,255,255,0.06), 0 8px 30px rgba(0,0,0,0.5)',
            },
          },
        }}
      >
        <Stack
          component={Paper}
          onClick={handleClose}
          sx={{
            gap: 2,
            justifyContent: 'space-between',
            borderRadius: 3,
            bgcolor: tokens.surface.raised,
            border: `1px solid ${tokens.edge.hover}`,
          }}
        >
          <Stack direction="row" sx={{ p: 2.5 }}>
            <Stack sx={{ width: leftWidth, mr: 2 }}>
              {leftMenu.map((section, index) => (
                <Stack key={index} sx={{ mb: 3, '&:last-child': { mb: 0 } }}>
                  {section.title && (
                    <Typography
                      sx={{
                        fontWeight: 600,
                        mb: 1.5,
                        color: tokens.text.muted,
                        fontSize: '0.75rem',
                        letterSpacing: '0.02em',
                      }}
                    >
                      {section.title}
                    </Typography>
                  )}
                  <Stack gap={section.title ? 0.5 : 1}>
                    {section.items.map((item) => (
                      <MegaMenuItem key={item.label} {...item} />
                    ))}
                  </Stack>
                </Stack>
              ))}
            </Stack>

            {rightMenu && (
              <Stack sx={{ width: 170, pl: 2, borderLeft: `1px solid ${tokens.edge.hover}` }}>
                {rightMenu.map((section, index) => (
                  <Stack key={index} sx={{ mb: 3, '&:last-child': { mb: 0 } }}>
                    {section.title && (
                      <Typography
                        sx={{
                          fontWeight: 600,
                          mb: 1.5,
                          color: tokens.text.muted,
                          fontSize: '0.75rem',
                          letterSpacing: '0.02em',
                        }}
                      >
                        {section.title}
                      </Typography>
                    )}
                    <Stack gap={section.title ? 0.5 : 1}>
                      {section.items.map((item) => (
                        <MegaMenuItem key={item.label} {...item} />
                      ))}
                    </Stack>
                  </Stack>
                ))}
              </Stack>
            )}
          </Stack>

          {footerMenu && (
            <Link href={footerMenu.href} style={{ textDecoration: 'none', color: 'inherit' }}>
              <Stack
                direction="row"
                sx={{
                  py: 1.25,
                  px: 2.5,
                  gap: 2,
                  justifyContent: 'space-between',
                  cursor: 'pointer',
                  bgcolor: 'rgba(255,255,255,0.05)',
                  '&:hover': { bgcolor: 'rgba(255,255,255,0.1)' },
                  transition: 'background-color 150ms ease',
                  borderRadius: '0 0 12px 12px',
                  borderTop: `1px solid ${tokens.edge.hover}`,
                }}
              >
                <Typography sx={{ fontSize: '0.875rem', fontWeight: 500, color: tokens.text.secondary }}>
                  {footerMenu.label}
                </Typography>
                <NavigateNext sx={{ color: tokens.text.muted }} />
              </Stack>
            </Link>
          )}
        </Stack>
      </Menu>
    </>
  );
}
