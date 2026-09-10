import { useState } from 'react'
import { useT } from '../i18n'

/**
 * Képernyőképek az alkalmazásról.
 *
 * Miért fülekkel és nem egymás alatt: a négy kép ugyanannak a programnak négy
 * arca, nem négy külön funkció - egymás alatt görgetve úgy tűnne, mintha
 * sokfelé kellene kapkodni. Így viszont egy pillanat alatt átválthatók, és a
 * kép mérete is marad akkora, hogy tényleg lehessen látni rajta valamit.
 */
const SHOTS = ['search', 'editor', 'clips', 'library'] as const
type Shot = (typeof SHOTS)[number]

const FILES: Record<Shot, string> = {
  search: 'kereses',
  editor: 'szerkeszto',
  clips: 'snittjeim',
  library: 'videotar',
}

export default function Screenshots() {
  const t = useT()
  const [active, setActive] = useState<Shot>('search')

  return (
    <section className="section shots" id="shots">
      <div className="container">
        <span className="eyebrow">{t.shots.eyebrow}</span>
        <h2>{t.shots.title}</h2>
        <p className="lead">{t.shots.sub}</p>

        <div className="shot-tabs" role="tablist" aria-label={t.shots.eyebrow}>
          {SHOTS.map((key) => (
            <button
              key={key}
              type="button"
              role="tab"
              aria-selected={active === key}
              className={active === key ? 'shot-tab on' : 'shot-tab'}
              onClick={() => setActive(key)}
            >
              {t.shots.items[key].title}
            </button>
          ))}
        </div>

        <figure className="shot-figure">
          {/* Fix méret nincs: a képek vágása képernyőnként más, és egy hamis
              arány csak torzítaná őket. A helyet a keret min-height tartja,
              hogy fülváltáskor ne ugráljon az oldal. */}
          <img
            src={`${import.meta.env.BASE_URL}screenshots/${FILES[active]}.webp`}
            alt={t.shots.items[active].alt}
            loading="lazy"
            decoding="async"
          />
          <figcaption>{t.shots.items[active].body}</figcaption>
        </figure>
      </div>
    </section>
  )
}
