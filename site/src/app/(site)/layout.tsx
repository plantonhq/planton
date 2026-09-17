import { WebsiteShell } from '@planton/website-shell';

export default function RootGroupLayout({ children }: { children: React.ReactNode }) {
  return <WebsiteShell>{children}</WebsiteShell>;
}
