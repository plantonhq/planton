import { Metadata } from 'next';
import StartupFoundersPage from '@/components/product/solutions/startup-founders';

export const metadata: Metadata = {
  title: 'For Startup Founders | Planton',
  description:
    "Your CTO's first infrastructure decision. Production in an afternoon, free tier, no lock-in.",
};

export default function Page() {
  return <StartupFoundersPage />;
}
