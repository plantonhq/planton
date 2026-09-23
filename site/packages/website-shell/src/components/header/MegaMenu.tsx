'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Menu, Stack, Typography } from '@mui/material';
import { NavigateNext, KeyboardArrowDown } from '@mui/icons-material';
import { MegaMenuItem } from './MegaMenuItem';
import type { MenuSection, MenuItem } from '../../data/navigation';
import { scopedTokens as tokens } from '../../theme/tokens';

interface MegaMenuProps {
  title: string;
  leftMenu: MenuSection[];
  rightMenu?: MenuSection[];
  footerMenu?: MenuItem;
  /** Column widths in px. A column with sub-labels beside icons needs about 240 for a sub-label to hold to two lines. */
  leftWidth?: number;
  rightWidth?: number;
}

export function MegaMenu({
  title,
  leftMenu,
  rightMenu,
  footerMenu,
  leftWidth = 270,
  rightWidth = 170,
}: MegaMenuProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };
  const rightHasMarks = Boolean(rightMenu?.some((section) => section.items.some((item) => item.icon)));

  return (
    <>
      <Stack
        component="button"
        type="button"
        aria-controls={open ? 'mega-menu' : undefined}
        aria-expanded={open ? 'true' : undefined}
        aria-haspopup="true"
        direction="row"
        onClick={handleClick}
        sx={{
          cursor: 'pointer',
          background: 'transparent', border: 0, padding: 0, color: 'inherit',
          '&:focus-visible': { outline: '2px solid currentColor', outlineOffset: 4 },
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
            // The open menu's trigger is the one whose caret points up.
            transform: open ? 'rotate(180deg)' : 'none',
          }}
        />
      </Stack>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{ sx: { padding: 0 } }}
        slotProps={{
          // One frame: the menu's own paper carries the fill, the border, the
          // radius, and the shadow, so there is one edge and one corner. The
          // border is an inline style, not sx: the website's Tailwind preflight
          // zeroes every border with an unlayered rule that beats MUI's layered
          // styles, and an inline style is the one declaration that beats both
          // (the same reason the shell's colors are inline where a host's CSS
          // would otherwise win).
          paper: {
            style: { border: `1px solid ${tokens.edge.hover}` },
            sx: {
              mt: 1.5,
              backgroundColor: tokens.surface.raised,
              backgroundImage: 'none',
              borderRadius: 3,
              boxShadow: '0 8px 30px rgba(0,0,0,0.5)',
            },
          },
        }}
      >
        <Stack onClick={handleClose} sx={{ gap: 2, justifyContent: 'space-between' }}>
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
              <Stack sx={{ width: rightWidth, pl: 2 }} style={{ borderLeft: `1px solid ${tokens.edge.hover}` }}>
                {/* A column is one list of links; an entry without a mark keeps the mark's width so every label shares one edge. */}
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
                        <MegaMenuItem key={item.label} {...item} alignWithMarks={rightHasMarks} />
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
                }}
                style={{ borderTop: `1px solid ${tokens.edge.hover}` }}
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
