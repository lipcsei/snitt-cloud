import type { MouseEvent, ReactNode } from 'react';
import { useHref, useLocation, useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';

type SectionLinkProps = {
  to: string;
  className?: string;
  children: ReactNode;
};

/**
 * Szekcióra ugró link. A landing oldalon simán görget, más útvonalról
 * előbb visszavisz a főoldalra, és az ott futó effekt görget oda. A főoldal
 * és a horgony azonosítója is nyelvfüggő.
 */
export default function SectionLink({ to, className, children }: SectionLinkProps) {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const { path } = useI18n();
  const home = path('home');
  const href = useHref({ pathname: home, hash: `#${to}` });

  const onClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.button !== 0) return;
    event.preventDefault();

    if (pathname !== home) {
      navigate(home, { state: { scrollTo: to } });
      return;
    }
    document.getElementById(to)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    history.replaceState(null, '', `#${to}`);
  };

  return (
    <a href={href} className={className} onClick={onClick}>
      {children}
    </a>
  );
}
