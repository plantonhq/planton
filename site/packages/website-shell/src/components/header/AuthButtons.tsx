'use client';

import type { FC } from 'react';
import Link from 'next/link';
import { Button } from '@mui/material';
import { useLoggedIn } from '../../hooks/useLoggedIn';
import { SIGN_IN, START_FREE } from '../../data/navigation';
import { scopedTokens as tokens } from '../../theme/tokens';

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

// The header's one filled button is Start Free, the same primary door as every primary button on the site.
const primaryStyle: React.CSSProperties = {
  backgroundColor: tokens.cta.background,
  color: tokens.cta.text,
};
const quietLinkStyle: React.CSSProperties = { color: tokens.text.secondary };

type ActionProps = {
  /** Called when a visitor clicks Start Free, for the site's own analytics. */
  onSelfServiceAction?: () => void;
};

/** Sign In, quiet, then Start Free; a signed-in visitor sees only the way back to the console. */
export const DesktopAuthButtons: FC<ActionProps> = ({ onSelfServiceAction }) => {
  const loggedIn = useLoggedIn();

  if (loggedIn) {
    return (
      <Button LinkComponent={Link} href="/dashboard" style={primaryStyle} sx={ctaSx}>
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
      <Button
        LinkComponent={Link}
        href={START_FREE.href}
        onClick={onSelfServiceAction}
        style={primaryStyle}
        sx={ctaSx}
      >
        {START_FREE.label}
      </Button>
    </>
  );
};

const ctaFullWidthSx = {
  ...ctaSx,
  height: 44,
  fontSize: '1rem',
  width: '100%',
  justifyContent: 'center',
} as const;

/** The same two doors, full width, at the foot of the mobile drawer. */
export const MobileAuthButtons: FC<ActionProps> = ({ onSelfServiceAction }) => {
  const loggedIn = useLoggedIn();

  if (loggedIn) {
    return (
      <Button LinkComponent={Link} href="/dashboard" style={primaryStyle} sx={ctaFullWidthSx}>
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
      <Button
        LinkComponent={Link}
        href={START_FREE.href}
        onClick={onSelfServiceAction}
        style={primaryStyle}
        sx={ctaFullWidthSx}
      >
        {START_FREE.label}
      </Button>
    </>
  );
};
