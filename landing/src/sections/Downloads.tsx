import { Link } from 'react-router-dom';
import { RELEASES_URL } from '../constants';

const PLATFORMS = [
  {
    id: 'windows',
    name: 'Windows',
    detail: 'Windows 10 vagy újabb · 64 bites',
    file: '.exe telepítő',
  },
  {
    id: 'macos',
    name: 'macOS',
    detail: 'macOS 12 vagy újabb · Apple Silicon és Intel',
    file: '.dmg lemezkép',
  },
  {
    id: 'linux',
    name: 'Linux',
    detail: 'x86_64 · GTK/WebKit2GTK környezet',
    file: '.AppImage / .deb',
  },
];

function PlatformMark({ id }: { id: string }) {
  if (id === 'windows') {
    return (
      <svg viewBox="0 0 24 24" width="26" height="26" focusable="false" aria-hidden="true">
        <path
          fill="currentColor"
          d="M3 5.6 10.4 4.5v7.1H3zM11.6 4.3 21 3v8.6h-9.4zM3 12.6h7.4v7L3 18.5zM11.6 12.6H21V21l-9.4-1.3z"
        />
      </svg>
    );
  }
  if (id === 'macos') {
    return (
      <svg viewBox="0 0 24 24" width="26" height="26" focusable="false" aria-hidden="true">
        <path
          fill="currentColor"
          d="M16.3 12.6c0-2.2 1.8-3.3 1.9-3.3-1-1.5-2.6-1.7-3.2-1.7-1.4-.1-2.7.8-3.3.8-.7 0-1.7-.8-2.8-.8-1.5 0-2.8.8-3.6 2.2-1.5 2.7-.4 6.6 1.1 8.8.7 1 1.6 2.2 2.7 2.2 1.1 0 1.5-.7 2.8-.7s1.6.7 2.8.7c1.2 0 1.9-1.1 2.6-2.1.8-1.2 1.2-2.4 1.2-2.4s-2.2-.9-2.2-3.7M14.2 5.9c.6-.7 1-1.7.9-2.7-.9 0-2 .6-2.6 1.3-.6.7-1.1 1.7-.9 2.7 1 .1 2-.5 2.6-1.3"
        />
      </svg>
    );
  }
  return (
    <svg viewBox="0 0 24 24" width="26" height="26" focusable="false" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 2c2.2 0 3.4 1.7 3.4 4 0 1.5-.3 2.4.4 3.6.9 1.5 2.5 3.2 2.9 5.3.3 1.5-.1 2.4.6 3.5.6 1 .3 2.2-.9 2.5-1.2.3-2.5-.3-3.6.3-1.7.9-3.9.9-5.6 0-1.1-.6-2.4 0-3.6-.3-1.2-.3-1.5-1.5-.9-2.5.7-1.1.3-2 .6-3.5.4-2.1 2-3.8 2.9-5.3.7-1.2.4-2.1.4-3.6 0-2.3 1.2-4 3.4-4m-1.5 4.2c-.5 0-.9.5-.9 1.1s.4 1.1.9 1.1.9-.5.9-1.1-.4-1.1-.9-1.1m3 0c-.5 0-.9.5-.9 1.1s.4 1.1.9 1.1.9-.5.9-1.1-.4-1.1-.9-1.1"
      />
    </svg>
  );
}

export default function Downloads() {
  return (
    <section id="letoltes" className="section">
      <div className="container">
        <header className="section-head">
          <span className="eyebrow">Letöltés</span>
          <h2>Töltsd le, és mutasd meg neki a videóid mappáját</h2>
          <p className="section-sub">
            A kiadások a GitHubon érhetők el. Válaszd ki a rendszeredhez tartozó csomagot.
          </p>
        </header>

        <div className="platforms">
          {PLATFORMS.map((p) => (
            <article key={p.id} className="platform">
              <span className="platform-mark" aria-hidden="true">
                <PlatformMark id={p.id} />
              </span>
              <h3>{p.name}</h3>
              <p className="platform-detail">{p.detail}</p>
              <span className="platform-file">{p.file}</span>
              <a
                className="btn btn-primary btn-block"
                href={RELEASES_URL}
                target="_blank"
                rel="noreferrer noopener"
              >
                Letöltés — {p.name}
              </a>
            </article>
          ))}
        </div>

        <aside className="notice">
          <h3>Mire van szükség a gépeden?</h3>
          <p>
            A Snitt három külső eszközre támaszkodik, és ezeket <strong>nem</strong> csomagolja
            magába — külön kell telepítened őket:
          </p>
          <ul className="notice-list">
            <li>
              <code>ffmpeg</code> — a videók vágásához és a hangsáv kinyeréséhez.
            </li>
            <li>
              <code>yt-dlp</code> — ha linkről szeretnél videót behozni.
            </li>
            <li>
              <code>faster-whisper</code> — ha beszédfelismeréssel is szeretnél átiratot készíteni.
            </li>
          </ul>
          <p className="notice-foot">
            Ha csak meglévő feliratokban keresel a saját fájljaid között, elég az <code>ffmpeg</code>
            . A többi akkor kell, amikor tényleg használod őket.
          </p>
          <p className="notice-cta">
            <Link className="btn btn-outline" to="/telepites">
              Telepítési útmutató rendszerenként →
            </Link>
          </p>
        </aside>
      </div>
    </section>
  );
}
