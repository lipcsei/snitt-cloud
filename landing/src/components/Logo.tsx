import { useId } from 'react';
import { useT } from '../i18n';

type Props = { size?: number; className?: string };

/**
 * Snitt márkajelzés: egy elvágott filmszalag, aminek a két fele elcsúszott -
 * a "snitt" maga a vágás. A clipPath/gradiens azonosítók useId-vel egyediek,
 * különben több logó egy oldalon összeakadna.
 */
export default function Logo({ size = 32, className }: Props) {
  const t = useT();
  const id = useId().replace(/:/g, '');
  const grad = `snitt-grad-${id}`;
  const top = `snitt-top-${id}`;
  const bottom = `snitt-bottom-${id}`;

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      className={className}
      role="img"
      aria-label={t.header.logoAlt}
    >
      <defs>
        <linearGradient id={grad} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#5b8dff" />
          <stop offset="100%" stopColor="#2a5fe0" />
        </linearGradient>
        {/* A vágás vonala: a felső fél eddig tart, az alsó innen indul. */}
        <clipPath id={top}>
          <path d="M0 0h64v26L0 38z" />
        </clipPath>
        <clipPath id={bottom}>
          <path d="M0 42l64-12v34H0z" />
        </clipPath>
      </defs>

      <rect width="64" height="64" rx="15" fill={`url(#${grad})`} />

      <g clipPath={`url(#${top})`}>
        <g transform="translate(5,0)">
          <rect x="19" y="5" width="26" height="54" rx="3.5" fill="#fff" />
          <rect x="22" y="9" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="38.4" y="9" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="22" y="19" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="38.4" y="19" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
        </g>
      </g>
      <g clipPath={`url(#${bottom})`}>
        <g transform="translate(-5,0)">
          <rect x="19" y="5" width="26" height="54" rx="3.5" fill="#fff" />
          <rect x="22" y="41" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="38.4" y="41" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="22" y="51" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
          <rect x="38.4" y="51" width="3.6" height="5" rx="1.1" fill="#2a5fe0" />
        </g>
      </g>
    </svg>
  );
}
