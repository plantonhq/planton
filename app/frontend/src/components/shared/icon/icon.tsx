'use client';
import React, { FC, useContext, useMemo } from 'react';
import { SvgIcon, SvgIconProps } from '@mui/material';
import { AppContext } from '@/contexts';

// Import SVG icons as React components
// Note: Requires @svgr/webpack to be installed and configured in next.config.js
import PlantonLogoIcon from '../../../../public/images/planton-logo.svg';
import SunIcon from '../../../../public/images/sun.svg';
import MoonIcon from '../../../../public/images/moon.svg';
import NavIcon from '../../../../public/images/nav.svg';
import DashboardIcon from '../../../../public/images/leftnav-icons/4square.svg';
import InfraHubIcon from '../../../../public/images/leftnav-icons/3square.svg';
import ConnectionsIcon from '../../../../public/images/connections.svg';
import EditIcon from '../../../../public/images/edit-icon.svg';
import DeleteIcon from '../../../../public/images/delete.svg';
import GcpIcon from '../../../../public/images/resources/gcp.svg';
import AwsIcon from '../../../../public/images/resources/aws.svg';
import AzureIcon from '../../../../public/images/resources/azure.svg';

export enum ICON_NAMES {
  PLANTON_LOGO = 'PLANTON_LOGO',
  SUN = 'SUN',
  MOON = 'MOON',
  NAV = 'NAV',
  DASHBOARD = 'DASHBOARD',
  INFRA_HUB = 'INFRA_HUB',
  CONNECTIONS = 'CONNECTIONS',
  EDIT = 'EDIT',
  DELETE = 'DELETE',
  GCP = 'GCP',
  AWS = 'AWS',
  AZURE = 'AZURE',
}

interface IconConfig {
  icon: FC<React.SVGProps<SVGSVGElement>>;
  viewBox?: string;
  color?: string;
  fill?: string;
}

interface IconRegistry {
  [key: string]: IconConfig;
}

// Light mode icons registry
const icons: IconRegistry = {
  [ICON_NAMES.PLANTON_LOGO]: {
    icon: PlantonLogoIcon,
    viewBox: '0 0 738 750',
  },
  [ICON_NAMES.SUN]: {
    icon: SunIcon,
    viewBox: '0 0 16 17',
  },
  [ICON_NAMES.MOON]: {
    icon: MoonIcon,
    viewBox: '0 0 16 17',
  },
  [ICON_NAMES.NAV]: {
    icon: NavIcon,
    viewBox: '0 0 20 20',
  },
  [ICON_NAMES.DASHBOARD]: {
    icon: DashboardIcon,
    viewBox: '0 0 16 16',
  },
  [ICON_NAMES.INFRA_HUB]: {
    icon: InfraHubIcon,
    viewBox: '0 0 16 16',
  },
  [ICON_NAMES.CONNECTIONS]: {
    icon: ConnectionsIcon,
    viewBox: '0 0 16 16',
  },
  [ICON_NAMES.EDIT]: {
    icon: EditIcon,
    color: '#6C7A8D',
    fill: 'grey.100',
    viewBox: '0 0 20 20',
  },
  [ICON_NAMES.DELETE]: {
    icon: DeleteIcon,
    viewBox: '0 0 14 14',
  },
  [ICON_NAMES.GCP]: {
    icon: GcpIcon,
    viewBox: '0 0 41 40',
  },
  [ICON_NAMES.AWS]: {
    icon: AwsIcon,
    viewBox: '0 0 16 9',
  },
  [ICON_NAMES.AZURE]: {
    icon: AzureIcon,
    viewBox: '0 0 100 100',
  },
};

// Dark mode icons registry (overrides light mode icons when available)
const iconsDark: IconRegistry = {};

export interface IconProps extends Omit<SvgIconProps, 'component'> {
  name: ICON_NAMES | string;
  alt?: string;
}

/**
 * Icon component that renders SVG icons with theme-aware support.
 * Automatically switches between light and dark variants based on current theme.
 *
 * @example
 * ```tsx
 * <Icon name={ICON_NAMES.PLANTON_LOGO} sx={{ width: 32, height: 32 }} />
 * ```
 */
export const Icon: FC<IconProps> = ({ name, sx, ...props }) => {
  const { theme } = useContext(AppContext);

  const iconConfig = useMemo(() => {
    // Try dark mode first, fallback to light mode
    const config = theme.mode === 'dark' ? iconsDark[name] || icons[name] : icons[name];

    if (!config) {
      console.warn(`Icon "${name}" not found in registry`);
      return null;
    }

    return config;
  }, [name, theme.mode]);

  if (!iconConfig) {
    return null;
  }

  const { icon: IconComponent, color, fill = 'none', viewBox = '0 0 20 20' } = iconConfig;

  return (
    <SvgIcon
      component={IconComponent}
      sx={{
        fontSize: 16,
        color,
        fill,
        cursor: props.onClick ? 'pointer' : 'inherit',
        ...sx,
      }}
      viewBox={viewBox}
      {...props}
    />
  );
};
