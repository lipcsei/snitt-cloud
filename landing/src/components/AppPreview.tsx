import { useEffect, useState } from 'react';

type Hit = {
  id: string;
  /** Film címe + évszám — sima szöveges forrásmegjelölés, sehol nincs plakát vagy logó. */
  title: string;
  /** Melyik átiratból jött a találat (felirat vagy Whisper, és milyen nyelven). */
  source: string;
  time: string;
  /** A sor idézőjelei nyelvfüggők: magyar „…”, angol "…". */
  quotes: 'hu' | 'en';
  before: string;
  hit: string;
  after: string;
  score: string;
  /** Jelentés szerinti (szemantikus) találat: a kérdés magyar volt, a sor angol. */
  semantic?: boolean;
  active?: boolean;
};

type Scene = {
  id: string;
  query: string;
  stats: string;
  hits: Hit[];
  cut: {
    range: string;
    file: string;
    /** A kijelölés helye a sávon, százalékban. */
    left: number;
    width: number;
    playhead: number;
  };
};

const SCENES: Scene[] = [
  {
    id: 'lotr',
    query: 'nem mehetsz át',
    stats: '2 találat · 0,3 mp',
    hits: [
      {
        id: 'lotr-hu',
        title: 'A Gyűrűk Ura: A Gyűrű Szövetsége (2001)',
        source: 'magyar felirat',
        time: '02:31:05',
        quotes: 'hu',
        before: '',
        hit: 'Nem mehetsz át!',
        after: '',
        score: '98%',
        active: true,
      },
      {
        id: 'lotr-en',
        title: 'A Gyűrűk Ura: A Gyűrű Szövetsége (2001)',
        source: 'whisper · angol',
        time: '02:31:04',
        quotes: 'en',
        before: 'You shall not ',
        hit: 'pass',
        after: '!',
        score: '91%',
        semantic: true,
      },
    ],
    cut: {
      range: '02:31:01 → 02:31:09 · 8,0 mp',
      file: 'gyuruk-ura_02-31-01.mp4',
      left: 30,
      width: 30,
      playhead: 44,
    },
  },
  {
    id: 'pulp',
    query: 'royale sajttal',
    stats: '2 találat · 0,2 mp',
    hits: [
      {
        id: 'pulp-hu',
        title: 'Ponyvaregény (1994)',
        source: 'magyar felirat',
        time: '00:12:34',
        quotes: 'hu',
        before: '',
        hit: 'Royale sajttal.',
        after: '',
        score: '97%',
        active: true,
      },
      {
        id: 'pulp-en',
        title: 'Ponyvaregény (1994)',
        source: 'whisper · angol',
        time: '00:12:31',
        quotes: 'en',
        before: 'They call it a ',
        hit: 'Royale with cheese',
        after: '.',
        score: '89%',
        semantic: true,
      },
    ],
    cut: {
      range: '00:12:29 → 00:12:37 · 8,0 mp',
      file: 'ponyvaregeny_00-12-29.mp4',
      left: 42,
      width: 24,
      playhead: 55,
    },
  },
  {
    id: 'esb',
    query: 'én vagyok az apád',
    stats: '2 találat · 0,3 mp',
    hits: [
      {
        id: 'esb-hu',
        title: 'Star Wars V. rész — A Birodalom visszavág (1980)',
        source: 'magyar felirat',
        time: '01:45:02',
        quotes: 'hu',
        before: '',
        hit: 'Én vagyok az apád.',
        after: '',
        score: '99%',
        active: true,
      },
      {
        id: 'esb-en',
        title: 'Star Wars V. rész — A Birodalom visszavág (1980)',
        source: 'whisper · angol',
        time: '01:45:01',
        quotes: 'en',
        before: 'I am your ',
        hit: 'father',
        after: '.',
        score: '92%',
        semantic: true,
      },
    ],
    cut: {
      range: '01:44:58 → 01:45:07 · 9,0 mp',
      file: 'birodalom-visszavag_01-44-58.mp4',
      left: 22,
      width: 33,
      playhead: 38,
    },
  },
];

const SCENE_MS = 4600;
const FADE_MS = 260;
const TYPE_MS = 55;

function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(() => matchReduced()?.matches ?? false);

  useEffect(() => {
    const mq = matchReduced();
    if (!mq) return;
    const onChange = () => setReduced(mq.matches);
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, []);

  return reduced;
}

function matchReduced(): MediaQueryList | null {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return null;
  return window.matchMedia('(prefers-reduced-motion: reduce)');
}

export default function AppPreview() {
  const reduced = usePrefersReducedMotion();
  const [index, setIndex] = useState(0);
  const [fading, setFading] = useState(false);

  const scene = SCENES[index] ?? SCENES[0];

  // Csökkentett mozgás esetén nincs forgatás: az első jelenet marad állóképként.
  useEffect(() => {
    if (reduced) {
      setIndex(0);
      setFading(false);
      return;
    }

    let swap = 0;
    const tick = window.setInterval(() => {
      setFading(true);
      swap = window.setTimeout(() => {
        setIndex((i) => (i + 1) % SCENES.length);
        setFading(false);
      }, FADE_MS);
    }, SCENE_MS);

    return () => {
      window.clearInterval(tick);
      window.clearTimeout(swap);
    };
  }, [reduced]);

  const typed = useTypedQuery(scene.query, reduced);

  return (
    <div className="preview" aria-hidden="true">
      <div className="preview-chrome">
        <span className="dot dot-r" />
        <span className="dot dot-y" />
        <span className="dot dot-g" />
        <span className="preview-title">Snitt — Könyvtár: 412 videó · 1 908 átirat</span>
      </div>

      <div className="preview-body">
        <div className="preview-search">
          <svg viewBox="0 0 24 24" className="preview-search-icon" focusable="false">
            <circle cx="11" cy="11" r="7" fill="none" stroke="currentColor" strokeWidth="2" />
            <path d="M16.5 16.5 21 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <span className="preview-query">
            {typed}
            <span className="caret" />
          </span>
          <span className="preview-kbd">Enter</span>
        </div>

        <div className={fading ? 'preview-scene is-fading' : 'preview-scene'}>
          <div className="preview-meta">
            <span className="chip chip-on">Teljes szöveg + jelentés</span>
            <span className="chip">Minden nyelv · minden átirat</span>
            <span className="chip">{scene.stats}</span>
          </div>

          <ul className="preview-results">
            {scene.hits.map((r) => (
              <li key={r.id} className={r.active ? 'result result-active' : 'result'}>
                <span className="result-thumb" />
                <span className="result-main">
                  <span className="result-head">
                    <span className="result-title">{r.title}</span>
                    <span className="result-lang">{r.source}</span>
                  </span>
                  <span className="result-body">
                    <span className="result-line">
                      {r.quotes === 'hu' ? '„' : '"'}
                      {r.before}
                      <mark>{r.hit}</mark>
                      {r.after}
                      {r.quotes === 'hu' ? '”' : '"'}
                    </span>
                    {r.semantic && <span className="result-badge">jelentés szerint</span>}
                  </span>
                </span>
                <span className="result-side">
                  <span className="timecode">{r.time}</span>
                  <span className="score">{r.score}</span>
                </span>
              </li>
            ))}
          </ul>

          <div className="preview-timeline">
            <div className="timeline-head">
              <span className="timeline-label">Kivágás</span>
              <span className="timeline-range">{scene.cut.range}</span>
            </div>
            <div className="timeline-track">
              <div className="timeline-wave">
                {WAVE.map((h, i) => (
                  <span key={i} style={{ height: `${h}%` }} />
                ))}
              </div>
              <div
                className="timeline-selection"
                style={{ left: `${scene.cut.left}%`, width: `${scene.cut.width}%` }}
              >
                <span className="handle handle-l" />
                <span className="handle handle-r" />
              </div>
              <div className="timeline-playhead" style={{ left: `${scene.cut.playhead}%` }} />
            </div>
            <div className="timeline-actions">
              <span className="mini-btn mini-btn-primary">Kivágás fájlba</span>
              <span className="mini-btn mini-btn-ghost">Előnézet</span>
              <span className="timeline-out">{scene.cut.file}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

/** A keresőmező úgy néz ki, mintha épp begépelnék a kérdést. */
function useTypedQuery(query: string, reduced: boolean): string {
  const [typed, setTyped] = useState(query);

  useEffect(() => {
    if (reduced) {
      setTyped(query);
      return;
    }

    setTyped('');
    let n = 0;
    const id = window.setInterval(() => {
      n += 1;
      setTyped(query.slice(0, n));
      if (n >= query.length) window.clearInterval(id);
    }, TYPE_MS);

    return () => window.clearInterval(id);
  }, [query, reduced]);

  return typed;
}

/** A hullámforma csak dekoráció, fix értékekkel, hogy ne ugráljon renderenként. */
const WAVE = [
  22, 41, 30, 58, 44, 71, 35, 52, 28, 63, 48, 80, 55, 38, 66, 44, 74, 51, 33, 60, 45, 82, 57, 40,
  68, 36, 54, 47, 72, 30, 61, 43, 78, 50, 34, 65, 42, 56, 29, 70, 46, 37, 59, 32, 67, 49, 25, 53,
  39, 62,
];
