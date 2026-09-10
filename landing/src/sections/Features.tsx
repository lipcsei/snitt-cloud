import type { ReactNode } from 'react';

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
  lock: (
    <>
      <rect x="4" y="10" width="16" height="10" rx="3" {...S} />
      <path d="M8 10V7.5a4 4 0 0 1 8 0V10M12 14v2.5" {...S} />
    </>
  ),
};

const FEATURES = [
  {
    icon: 'quote',
    title: 'Keresés idézetre',
    body: 'Nem a fájlnévre, nem a címkékre: arra keresel, ami elhangzik. Beírod a mondatot, és megkapod a filmet meg a másodpercet, ahol elhangzik.',
  },
  {
    icon: 'subs',
    title: 'Több felirat egy filmhez',
    body: 'Egy filmhez gyakran több felirat is tartozik — más fordítások, más nyelvek, és nem ugyanazt írják. A Snitt mindet megtartja egymás mellett, és mindegyikben keres.',
  },
  {
    icon: 'mic',
    title: 'Whisper átirat a hangból',
    body: 'Ha nincs felirat, vagy nem bízol benne, a Whisper beszédfelismerés legenerálja az átiratot magából a hangsávból. Így azt kapod, ami tényleg elhangzott, nem a fordító változatát.',
  },
  {
    icon: 'brain',
    title: 'Jelentés szerinti keresés',
    body: 'Ritkán emlékszünk pontosan. A hibrid keresés a teljes szöveges találatok mellé a jelentés alapján hasonló mondatokat is behozza, így a körülírás is elég.',
  },
  {
    icon: 'cut',
    title: 'Vizuális klipvágó',
    body: 'A találatból az átirat alapján rögtön javasol egy tartományt, amit az idővonalon húzással igazítasz. A kimenet egy videófájl, amit oda mentesz, ahová akarsz.',
  },
  {
    icon: 'lock',
    title: 'Minden helyben marad',
    body: 'Nincs szerver, nincs feltöltés, nincs fiókkényszer. Az index egyetlen SQLite fájl a gépeden, a videóid pedig ott maradnak, ahol eddig is voltak.',
  },
];

export default function Features() {
  return (
    <section id="funkciok" className="section section-alt">
      <div className="container">
        <header className="section-head">
          <span className="eyebrow">Funkciók</span>
          <h2>Amit egy videótár keresője tud, ha komolyan gondolják</h2>
          <p className="section-sub">
            A Snitt a saját gyűjteményedhez készült: sok fájl, sok nyelv, kevés rendszerezés.
          </p>
        </header>

        <div className="cards">
          {FEATURES.map((f) => (
            <article key={f.title} className="card">
              <span className="card-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" width="22" height="22" focusable="false">
                  {icons[f.icon]}
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
