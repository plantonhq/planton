'use client';

import { ComponentProps, FC } from 'react';
import { Button, ButtonProps, SvgIcon, SvgIconProps, Typography, TypographyProps } from '@mui/material';
import { styled } from '@mui/material/styles';
import DiscordIcon from 'public/_site/images/discord.svg';

/** An icon that grows from 16px to 24px past the tablet breakpoint; used by the Discord button below. */
const SvgIconSizeResponsive = styled(SvgIcon)<SvgIconProps>(({ theme }) => ({
  fontSize: 16,
  [theme.breakpoints.up(767)]: {
    fontSize: 24,
  },
}));

export const TypoH1: FC<TypographyProps> = ({ className, ...props }) => {
  return (
    <Typography className={`text-[48px] leading-[1.1] font-semibold tracking-tight ${className}`} {...props} />
  );
};

export const TypoH2: FC<TypographyProps> = ({ className, ...props }) => {
  return (
    <Typography
      className={`text-[24px] md:text-[40px] leading-[1.15] font-semibold tracking-tight ${className}`}
      {...props}
    />
  );
};

export const TypoH3: FC<TypographyProps> = ({ className, ...props }) => {
  return (
    <Typography
      className={`text-sm md:text-[28px] leading-[1.2] md:leading-[1.3] font-semibold ${className}`}
      {...props}
    />
  );
};

export const TypoP1: FC<TypographyProps> = ({ className, ...props }) => {
  return (
    <Typography
      className={`text-xs md:text-base md:leading-[1.5] font-normal md:font-medium ${className}`}
      {...props}
    />
  );
};

export const TypoB2Medium: FC<TypographyProps> = ({ className, ...props }) => {
  return (
    <Typography
      className={`text-xs md:text-[16px] font-medium leading-[16px] ${className}`}
      {...props}
    />
  );
};

export const TypoB2Regular: FC<TypographyProps> = ({ className, ...props }) => {
  return <Typography className={`text-base font-normal leading-[25px] ${className}`} {...props} />;
};

export const Btn: FC<ButtonProps & ComponentProps<'a'>> = ({ className, ...props }) => {
  return (
    <Button
      className={`px-3 md:px-5 py-2 md:py-3 h-8 md:h-10 text-xs md:text-base leading-[2] md:leading-[1] font-medium rounded-[10px] w-fit ${className}`}
      {...props}
    />
  );
};

export const PrimaryBtn: FC<ButtonProps & ComponentProps<'a'>> = ({ className, ...props }) => {
  return <Btn className={`bg-[#ffffff] ${className}`} {...props} />;
};

export const GetStartedBtn: FC<ButtonProps & ComponentProps<'a'>> = ({ className, ...props }) => {
  return (
    <PrimaryBtn
      href="/signup"
      className={`hover:bg-gray-200 transition-colors ${className}`}
      {...props}
    />
  );
};

export const BookDemoBtn: FC<ButtonProps & ComponentProps<'a'>> = ({ className, ...props }) => {
  return (
    <Btn
      href="/book-demo"
      className={`hover:bg-white/10 transition-colors ${className}`}
      {...props}
    />
  );
};

export const JoinDiscordBtn: FC<ButtonProps & ComponentProps<'a'>> = ({ className, ...props }) => {
  return (
    <Btn
      startIcon={<SvgIconSizeResponsive viewBox="0 0 14 14" component={DiscordIcon} />}
      href="https://discord.gg/pwcSapdQAp"
      target="_blank"
      className={` ${className}`}
      {...props}
    />
  );
};
