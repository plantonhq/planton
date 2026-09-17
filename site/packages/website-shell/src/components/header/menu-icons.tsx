import {
  Article as BlogIcon,
  Assignment as CatalogIcon,
  Cloud as HostedIcon,
  Code as OpenSourceIcon,
  Dns as SelfHostedIcon,
  Explore as TourIcon,
  Hub as InfraHubIcon,
  Input as ImportIcon,
  Laptop as DesktopAppIcon,
  MenuBook as DocsIcon,
  NewReleases as ChangelogIcon,
  PlayCircle as DemoIcon,
  RocketLaunch as ServiceHubIcon,
  School as TutorialsIcon,
  SmartToy as CodingAgentsIcon,
  Terminal as CliIcon,
} from '@mui/icons-material';
import type { MenuItem } from '../../data/navigation';

/**
 * The icons the header menus draw beside their items, keyed by the menu label
 * exactly as navigation.ts spells it. One map serves the desktop and the
 * mobile header so the two can never disagree; a key that drifts from the
 * label renders the item without its icon.
 */
const iconSx = { fontSize: { xs: 16, md: 24 } } as const;

export const productIcons: Record<string, React.ReactNode> = {
  'Infra Hub': <InfraHubIcon sx={iconSx} />,
  'Service Hub': <ServiceHubIcon sx={iconSx} />,
  'Coding Agents': <CodingAgentsIcon sx={iconSx} />,
  CLI: <CliIcon sx={iconSx} />,
  Catalog: <CatalogIcon sx={iconSx} />,
  Import: <ImportIcon sx={iconSx} />,
  'Open Source': <OpenSourceIcon sx={iconSx} />,
};

export const distributionIcons: Record<string, React.ReactNode> = {
  Hosted: <HostedIcon sx={iconSx} />,
  'Self-Hosted': <SelfHostedIcon sx={iconSx} />,
  Desktop: <DesktopAppIcon sx={iconSx} />,
};

export const resourceIcons: Record<string, React.ReactNode> = {
  Docs: <DocsIcon sx={iconSx} />,
  Tutorials: <TutorialsIcon sx={iconSx} />,
  Blog: <BlogIcon sx={iconSx} />,
  Changelog: <ChangelogIcon sx={iconSx} />,
  Tour: <TourIcon sx={iconSx} />,
  Demo: <DemoIcon sx={iconSx} />,
};

/** Pair each menu item with its icon; an item without one renders without. */
export const withIcons = (items: MenuItem[], icons: Record<string, React.ReactNode>): MenuItem[] =>
  items.map((item) => ({ ...item, icon: icons[item.label] }));
