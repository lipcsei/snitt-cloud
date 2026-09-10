import { useEffect, useRef } from 'react';
import '../styles.notyet.css';

type Props = {
  /** Melyik gomb nyitotta meg - a szöveg ehhez igazodik. */
  kind: 'download' | 'register';
  onClose: () => void;
};

/**
 * A letöltés és a regisztráció még nem elérhető (a felhő oldal egyelőre csak
 * helyben fut), ezért egy türelemre intő üzenetet kap a látogató.
 *
 * Az illusztráció szándékosan SAJÁT rajz, nem filmkocka: a rövid idézet
 * forrásmegjelöléssel belefér az idézés kivételébe, egy filmből vett kép egy
 * nyilvános marketingoldalon viszont nem.
 */
export default function NotYetModal({ kind, onClose }: Props) {
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    closeRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div className="notyet-overlay" onClick={onClose} role="presentation">
      <div
        className="notyet-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="notyet-title"
        onClick={(e) => e.stopPropagation()}
      >
        <EntIllustration />

        <blockquote className="notyet-quote" id="notyet-title">
          „Ne legyetek hamariak!"
          <cite>Szirszakáll — A Gyűrűk Ura: A két torony</cite>
        </blockquote>

        <p className="notyet-body">
          {kind === 'download'
            ? 'A Snitt még készül: a telepítők hamarosan letölthetők lesznek. Addig is a forráskód nyilvános, és magadnak is lefordíthatod.'
            : 'A regisztráció még nem él - a fiókok a későbbi Pro funkciókhoz kellenek majd. Az alkalmazás fiók nélkül is teljes értékű lesz.'}
        </p>

        <button ref={closeRef} type="button" className="notyet-close" onClick={onClose}>
          Türelmes leszek
        </button>
      </div>
    </div>
  );
}

/** Saját rajz: egy fás-szakállas figura, nem a filmbeli karakter másolata. */
function EntIllustration() {
  return (
    <svg className="notyet-art" viewBox="0 0 200 150" role="img" aria-label="Öreg fa illusztráció">
      <defs>
        <linearGradient id="notyet-bark" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#6b5842" />
          <stop offset="100%" stopColor="#3b3026" />
        </linearGradient>
        <linearGradient id="notyet-sky" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#16213a" />
          <stop offset="100%" stopColor="#0d1119" />
        </linearGradient>
      </defs>

      <rect width="200" height="150" fill="url(#notyet-sky)" />

      {/* lombkorona */}
      <g fill="#2f4a35">
        <ellipse cx="100" cy="38" rx="52" ry="28" />
        <ellipse cx="66" cy="48" rx="30" ry="20" />
        <ellipse cx="134" cy="48" rx="30" ry="20" />
      </g>
      <g fill="#3d5f43" opacity="0.75">
        <ellipse cx="88" cy="32" rx="26" ry="14" />
        <ellipse cx="120" cy="40" rx="22" ry="12" />
      </g>

      {/* törzs */}
      <path d="M84 150 C86 110 82 92 88 66 L112 66 C118 92 114 110 116 150 Z" fill="url(#notyet-bark)" />
      {/* kéreg-barázdák */}
      <g stroke="#2a2119" strokeWidth="1.4" strokeLinecap="round" opacity="0.8">
        <path d="M94 74 C92 96 95 118 93 144" fill="none" />
        <path d="M104 72 C107 98 103 120 106 146" fill="none" />
      </g>

      {/* ágkarok */}
      <g stroke="url(#notyet-bark)" strokeWidth="7" strokeLinecap="round" fill="none">
        <path d="M86 84 C68 88 58 78 50 66" />
        <path d="M114 88 C132 92 142 84 150 72" />
      </g>

      {/* szem-párok a kéregben */}
      <g>
        <ellipse cx="94" cy="86" rx="4.5" ry="3.2" fill="#0f1116" />
        <ellipse cx="106" cy="86" rx="4.5" ry="3.2" fill="#0f1116" />
        <circle cx="94.8" cy="85.6" r="1.5" fill="#d7c48a" />
        <circle cx="106.8" cy="85.6" r="1.5" fill="#d7c48a" />
      </g>

      {/* szakáll-gyökerek */}
      <g stroke="#5a4a37" strokeWidth="2" strokeLinecap="round" fill="none" opacity="0.9">
        <path d="M92 96 C90 106 92 114 89 122" />
        <path d="M100 97 C100 108 99 116 101 126" />
        <path d="M108 96 C110 106 108 114 111 122" />
      </g>

      {/* talaj */}
      <path d="M0 150 C40 138 70 144 100 141 C130 138 160 144 200 150 Z" fill="#131a14" />
    </svg>
  );
}
