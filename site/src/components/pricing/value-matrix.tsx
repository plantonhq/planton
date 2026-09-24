'use client';

import { FC, Fragment } from 'react';
import { Box, Stack, Tooltip, Typography } from '@mui/material';
import {
  Badge,
  CheckIcon,
  SectionSubtitle,
  SectionTitle,
} from '@/components/marketing';
import {
  MatrixCell,
  VALUE_MATRIX,
  VALUE_MATRIX_COLUMNS,
} from '@/data/value-matrix';

/**
 * The value matrix: what each plan includes, granularly, from one data
 * module. "Coming Soon" cells are decided packaging whose feature is still
 * shipping -- announced honestly, never claimed as live. This is the
 * below-the-fold reference; the plan cards above show its headlines.
 */

const CellContent: FC<{ cell: MatrixCell }> = ({ cell }) => {
  switch (cell.kind) {
    case 'included':
      return (
        <Box className="flex justify-center" aria-label="Included">
          <CheckIcon />
        </Box>
      );
    case 'not_included':
      return (
        <Typography className="text-[#4a4a4a] text-sm leading-none" aria-label="Not included">
          —
        </Typography>
      );
    case 'coming_soon':
      return (
        <Badge className="!px-2 !py-0.5 !text-[10px] whitespace-nowrap">Coming Soon</Badge>
      );
    case 'text':
      return (
        <Typography className="text-xs text-[#c0c0c0] leading-snug">{cell.label}</Typography>
      );
  }
};

export const ValueMatrix: FC = () => {
  return (
    <Box className="w-full px-4 md:px-8 py-12 bg-[#0a0a0a]" id="compare">
      <Box className="max-w-7xl mx-auto">
        <Stack className="items-center text-center mb-8">
          <SectionTitle>What Each Plan Includes</SectionTitle>
          <SectionSubtitle className="mx-auto">
            Almost everything Planton does is included everywhere — the paid
            tiers add organizational scale and enterprise identity.
          </SectionSubtitle>
        </Stack>
        <Box className="overflow-x-auto rounded-xl border border-[#2a2a2a] bg-[#151515]">
          <table className="w-full border-collapse min-w-[880px]">
            <thead>
              <tr className="sticky top-0 z-10">
                <th className="text-left p-4 sticky left-0 bg-[#151515] min-w-[240px]" />
                {VALUE_MATRIX_COLUMNS.map((column) => (
                  <th key={column.id} className="p-3.5 text-center min-w-[124px] bg-[#151515]">
                    <Stack className="items-center gap-0.5">
                      <Typography className="text-sm font-semibold text-white">
                        {column.label}
                      </Typography>
                      {column.sublabel && (
                        <Typography className="text-[11px] text-[#8a8a8a]">
                          {column.sublabel}
                        </Typography>
                      )}
                    </Stack>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {VALUE_MATRIX.map((category) => (
                <Fragment key={category.category}>
                  <tr>
                    <td
                      colSpan={VALUE_MATRIX_COLUMNS.length + 1}
                      className="px-4 pt-5 pb-1.5 border-t border-[#2a2a2a]"
                    >
                      <Typography className="text-[11px] font-semibold uppercase tracking-wider text-[#8a8a8a]">
                        {category.category}
                      </Typography>
                    </td>
                  </tr>
                  {category.rows.map((row) => (
                    <tr
                      key={row.feature}
                      className="border-t border-[#1e1e1e] hover:bg-white/[0.02] transition-colors duration-150 group"
                    >
                      <td className="px-4 py-2.5 sticky left-0 bg-[#151515] group-hover:bg-[#181818]">
                        {row.description ? (
                          <Tooltip title={row.description} placement="top-start">
                            <Typography className="text-sm text-[#e0e0e0] w-fit cursor-default border-b border-dotted border-[#4a4a4a]">
                              {row.feature}
                            </Typography>
                          </Tooltip>
                        ) : (
                          <Typography className="text-sm text-[#e0e0e0]">{row.feature}</Typography>
                        )}
                      </td>
                      {VALUE_MATRIX_COLUMNS.map((column) => (
                        <td key={column.id} className="px-3 py-2.5 text-center align-middle">
                          <Box className="flex justify-center">
                            <CellContent cell={row.cells[column.id]} />
                          </Box>
                        </td>
                      ))}
                    </tr>
                  ))}
                </Fragment>
              ))}
            </tbody>
          </table>
        </Box>
      </Box>
    </Box>
  );
};
