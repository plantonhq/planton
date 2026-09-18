'use client';

import type { FC } from 'react';
import Link from 'next/link';
import { Button } from '@mui/material';
import { useLoggedIn } from '../../hooks/useLoggedIn';
import { SIGN_IN, START_FREE } from '../../data/navigation';
import { tokens } from '../../theme/tokens';

const ctaSx = {
  height: 32,
  px: { xs: 1, sm: 1.5 },
  py: 0.5,
  fontSize: { xs: '0.75rem', sm: '0.875rem' },
  fontWeight: 500,
  whiteSpace: 'nowrap',
  borderRadius: '10px',
  textTransform: 'none',
} as const;

// The one true-white fill in the header: the primary door, the same as every primary button on the site.
const whiteButtonStyle: React.CSSProperties = {
  backgroundColor: tokens.cta.background,
  color: tokens.cta.text,
};
const quietLinkStyle: React.CSSProperties = { color: tokens.text.secondary };

export const DesktopAuthButtons: FC = () => {
  const loggedIn = useLoggedIn();

  if (loggedIn) {
    return (
      <Button LinkComponent={Link} href="/dashboard" style={whiteButtonStyle} sx={ctaSx}>
        Dashboard
      </Button>
    );
  }

  return (
    <>
      <Button
        LinkComponent={Link}
        href={SIGN_IN.href}
        style={quietLinkStyle}
        sx={{
          display: { xs: 'none', sm: 'inline-flex' },
          fontSize: '0.875rem',
          fontWeight: 500,
          textTransform: 'none',
          borderRadius: '10px',
        }}
      >
        {SIGN_IN.label}
      </Button>
      <Button LinkComponent={Link} href={START_FREE.href} style={whiteButtonStyle} sx={ctaSx}>
        {START_FREE.label}
      </Button>
    </>
  );
};

const ctaFullWidthSx = {
  ...ctaSx,
  height: 40,
  width: '100%',
  justifyContent: 'center',
} as const;

export const MobileAuthButtons: FC = () => {
  const loggedIn = useLoggedIn();

  if (loggedIn) {
    return (
      <Button LinkComponent={Link} href="/dashboard" style={whiteButtonStyle} sx={ctaFullWidthSx}>
        Dashboard
      </Button>
    );
  }

  return (
    <>
      <Button
        LinkComponent={Link}
        href={SIGN_IN.href}
        style={quietLinkStyle}
        sx={{
          width: '100%',
          justifyContent: 'center',
          textTransform: 'none',
          borderRadius: '10px',
          fontWeight: 500,
        }}
      >
        {SIGN_IN.label}
      </Button>
      <Button LinkComponent={Link} href={START_FREE.href} style={whiteButtonStyle} sx={ctaFullWidthSx}>
        {START_FREE.label}
      </Button>
    </>
  );
};
