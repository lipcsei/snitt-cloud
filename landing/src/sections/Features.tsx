import type { ReactNode } from 'react';
import { useT } from '../i18n';

const S = { fill: 'none', stroke: 'currentColor', strokeWidth: 1.6, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const };

const icons: Record<string, ReactNode> = {
  quote: (
    <>
      <circle cx="10.5" cy="10.5" r="6.5" {...S} />
      <path d="M15.5 15.5 21 21" {...S} />
      <path d="M8 9.5h5M8 12h3" {...S} />
    </>
  ),
  subs: (
    <>
      <rect x="3" y="5" width="18" height="14" rx="3" {...S} />
      <path d="M6.5 14.5h4M13 14.5h4.5" {...S} />
      <path d="M6.5 10.5h11" {...S} opacity=".5" />
    </>
  ),
  mic: (
    <>
      <rect x="9" y="3" width="6" height="10" rx="3" {...S} />
      <path d="M5.5 11.5a6.5 6.5 0 0 0 13 0M12 18v3" {...S} />
    </>
  ),
  brain: (
    <>
      <circle cx="7" cy="8" r="3" {...S} />
      <circle cx="17" cy="8" r="3" {...S} />
      <circle cx="12" cy="17" r="3" {...S} />
      <path d="M9.4 9.8 10.9 14M14.6 9.8 13.1 14M10 8h4" {...S} />
    </>
  ),
  cut: (
    <>
      <rect x="3" y="7" width="18" height="10" rx="2.5" {...S} />
      <path d="M9 7v10M15 7v10" {...S} opacity=".45" />
      <path d="M9 4.5v2M15 17.5v2" {...S} />
    </>
  ),
  library: (
    <>
      <rect x="3" y="4" width="4" height="16" rx="1.5" {...S} />
      <rect x="9" y="4" width="4" height="16" rx="1.5" {...S} />
      <path d="m16 6 4.5 14.5" {...S} />
    </>
  ),
  move: (
    <>
      <path d="M12 3v12" {...S} />
      <path d="m8 11 4 4 4-4" {...S} />
      <path d="M4 17v2.5A1.5 1.5 0 0 0 5.5 21h13a1.5 1.5 0 0 0 1.5-1.5V17" {...S} />
    </>
  ),
  spell: (
    <>
      <path d="M4 17.5 8 6.5l4 11M5.4 13.5h5.2" {...S} />
      <path d="m14.5 15 2.5 2.5 4.5-6" {...S} />
    </>
  ),
  storage: (
    <>
      <ellipse cx="12" cy="6" rx="7.5" ry="3" {...S} />
      <path d="M4.5 6v6c0 1.7 3.4 3 7.5 3s7.5-1.3 7.5-3V6" {...S} />
      <path d="M12 16.5v4m-2-2 2 2 2-2" {...S} />
    </>
  ),
  missing: (
    <>
      <path d="M12 4 3 19.5h18L12 4Z" {...S} />
      <path d="M12 10v4.5M12 17.2v.1" {...S} />
    </>
  ),
  retry: (
    <>
      <path d="M20 12a8 8 0 1 1-2.6-5.9" {...S} />
      <path d="M20 4.5V9h-4.5" {...S} />
    </>
  ),
  lock: (
    <>
      <rect x="4" y="10" width="16" height="10" rx="3" {...S} />
      <path d="M8 10V7.5a4 4 0 0 1 8 0V10M12 14v2.5" {...S} />
    </>
  ),
};

/** A kártyák sorrendje kötött: az ikonok a nyelvi tábla elemeihez tartoznak. */
const ICON_ORDER = ['quote', 'subs', 'mic', 'brain', 'cut', 'library', 'move', 'lock', 'spell', 'storage', 'missing', 'retry'];

export default function Features() {
  const t = useT();

  return (
    <section id={t.sections.features} className="section section-alt">
      <div className="container">
        <header className="section-head">
          <span className="eyebrow">{t.features.eyebrow}</span>
          <h2>{t.features.title}</h2>
          <p className="section-sub">{t.features.sub}</p>
        </header>

        <div className="cards">
          {t.features.items.map((f, i) => (
            <article key={f.title} className="card">
              <span className="card-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" width="22" height="22" focusable="false">
                  {icons[ICON_ORDER[i]]}
                </svg>
              </span>
              <h3>{f.title}</h3>
              <p>{f.body}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
