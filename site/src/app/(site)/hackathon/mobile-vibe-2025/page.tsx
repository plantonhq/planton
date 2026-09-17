import { Metadata } from 'next';
import { Box } from '@mui/material';
import {
  HeroSection,
  EventDetails,
  EventAgenda,
  GoalsSection,
  AppIdeas,
  ParticipationCTA,
  FAQ,
} from '@/components/hackathon';

export const metadata: Metadata = {
  title: 'MobileVibe Hackathon 2025 | Planton',
  description:
    'Join Planton\'s first-ever mobile app hackathon. Build real apps, establish workflows, and normalize mobile development with vibe coding. December 22, 2025.',
  keywords:
    'mobile app hackathon, mobile development, Android, iOS, React Native, Flutter, vibe coding, Planton, hackathon 2025',
  openGraph: {
    title: 'MobileVibe Hackathon 2025',
    description:
      'Build real mobile apps, establish workflows, and normalize mobile development at Planton. Join us December 22, 2025!',
    type: 'website',
    url: 'https://planton.ai/hackathon/mobile-vibe-2025',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'MobileVibe Hackathon 2025',
    description:
      'Join Planton\'s first mobile app hackathon. December 22, 2025 - Hybrid event!',
  },
};

export default function MobileVibeHackathon() {
  return (
    <Box>
      <HeroSection />
      <EventDetails />
      <EventAgenda />
      <GoalsSection />
      <AppIdeas />
      <ParticipationCTA />
      <FAQ />
    </Box>
  );
}










