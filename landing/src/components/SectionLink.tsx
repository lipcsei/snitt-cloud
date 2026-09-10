import type { MouseEvent, ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

type SectionLinkProps = {
  to: string;
  className?: string;
  children: ReactNode;
};

/**
 * Szekcióra ugró link. A landing oldalon simán görget, más útvonalról
 * előbb visszavisz a főoldalra, és az ott futó effekt görget oda.
 */
export default function SectionLink({ to, className, children }: SectionLinkProps) {
  const navigate = useNavigate();
  const { pathname } = useLocation();

  const onClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.button !== 0) return;
    event.preventDefault();

    if (pathname !== '/') {
      navigate('/', { state: { scrollTo: to } });
      return;
    }
    document.getElementById(to)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    history.replaceState(null, '', `#${to}`);
  };

  return (
    <a href={`/#${to}`} className={className} onClick={onClick}>
      {children}
    </a>
  );
}
