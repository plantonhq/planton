import { Box, Typography, Button } from '@mui/material';
import Link from 'next/link';

export default function DocsNotFound() {
  return (
    <Box className="text-center py-12">
      <Typography variant="h2" className="text-white mb-4">
        404 - Page Not Found
      </Typography>
      <Typography className="text-gray-300 mb-8 text-lg">
        The documentation page you&apos;re looking for doesn&apos;t exist.
      </Typography>
      <Link href="/docs">
        <Button variant="contained" className="bg-white hover:bg-white/80">
          Back to Documentation
        </Button>
      </Link>
    </Box>
  );
}
