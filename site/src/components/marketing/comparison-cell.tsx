/**
 * One cell of a comparison table: a status glyph and an optional note. The
 * compare page describes categories, never named vendors; this component
 * renders whatever it is handed and has no opinion about the columns.
 */
import { Box } from '@mui/material';
import type { FC } from 'react';
import { CheckIcon, WarningIcon, XIcon } from './icons';

type ComparisonStatus = 'yes' | 'no' | 'partial' | 'na';

interface ComparisonCellProps {
  status: ComparisonStatus;
  text?: string;
  className?: string;
}

export const ComparisonCell: FC<ComparisonCellProps> = ({
  status,
  text,
  className = '',
}) => {
  const getStatusDisplay = () => {
    switch (status) {
      case 'yes':
        return <CheckIcon />;
      case 'no':
        return <XIcon />;
      case 'partial':
        return <WarningIcon />;
      case 'na':
        return <span className="text-fg-muted">N/A</span>;
    }
  };

  return (
    <Box className={`flex items-center gap-2 ${className}`}>
      {getStatusDisplay()}
      {text && <span className="text-sm text-fg-body">{text}</span>}
    </Box>
  );
};
