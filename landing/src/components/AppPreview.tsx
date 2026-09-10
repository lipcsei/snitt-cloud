import { useEffect, useState } from 'react';
import { useT } from '../i18n';
import type { Lang } from '../i18n';

/** Nyelvenkénti idézőjelpár: a német záró jele a magyartól is eltér. */
const QUOTES: Record<Lang, readonly [string, string]> = {
  hu: ['„', '”'],
  en: ['"', '"'],
  de: ['„', '“'],
};

/**
 * A jelenetek szövege a nyelvi táblákban van, a geometria (kijelölés helye a
 * sávon) viszont nyelvfüggetlen, ezért marad itt - így nem kell kétszer
 * karbantartani ugyanazokat a számokat.
 */
const SCENE_GEOMETRY = [
  { left: 30, width: 30, playhead: 44 },
  { left: 42, width: 24, playhead: 55 },
  { left: 22, width: 33, playhead: 38 },
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
  const t = useT();
  const reduced = usePrefersReducedMotion();
  const [index, setIndex] = useState(0);
  const [fading, setFading] = useState(false);

  const scenes = t.preview.scenes;
  const scene = scenes[index] ?? scenes[0];
  const geometry = SCENE_GEOMETRY[index] ?? SCENE_GEOMETRY[0];

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
        setIndex((i) => (i + 1) % scenes.length);
        setFading(false);
      }, FADE_MS);
    }, SCENE_MS);

    return () => {
      window.clearInterval(tick);
      window.clearTimeout(swap);
    };
  }, [reduced, scenes.length]);

  const typed = useTypedQuery(scene.query, reduced);

  return (
    <div className="preview" aria-hidden="true">
      <div className="preview-chrome">
        <span className="dot dot-r" />
        <span className="dot dot-y" />
        <span className="dot dot-g" />
        <span className="preview-title">{t.preview.windowTitle}</span>
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
            <span className="chip chip-on">{t.preview.chipHybrid}</span>
            <span className="chip">{t.preview.chipAll}</span>
            <span className="chip">{scene.stats}</span>
          </div>

          <ul className="preview-results">
            {scene.hits.map((r, i) => (
              <li key={r.time} className={i === 0 ? 'result result-active' : 'result'}>
                <span className="result-thumb" />
                <span className="result-main">
                  <span className="result-head">
                    <span className="result-title">{r.title}</span>
                    <span className="result-lang">{r.source}</span>
                  </span>
                  <span className="result-body">
                    <span className="result-line">
                      {QUOTES[r.quotes][0]}
                      {r.before}
                      <mark>{r.hit}</mark>
                      {r.after}
                      {QUOTES[r.quotes][1]}
                    </span>
                    {/* A második találat mindig a jelentés szerinti, másik nyelvű sor. */}
                    {i === 1 && <span className="result-badge">{t.preview.semanticBadge}</span>}
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
              <span className="timeline-label">{t.preview.cutLabel}</span>
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
                style={{ left: `${geometry.left}%`, width: `${geometry.width}%` }}
              >
                <span className="handle handle-l" />
                <span className="handle handle-r" />
              </div>
              <div className="timeline-playhead" style={{ left: `${geometry.playhead}%` }} />
            </div>
            <div className="timeline-actions">
              <span className="mini-btn mini-btn-primary">{t.preview.cutAction}</span>
              <span className="mini-btn mini-btn-ghost">{t.preview.previewAction}</span>
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
