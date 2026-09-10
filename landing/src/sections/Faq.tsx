import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

type Item = { q: string; a: ReactNode };

const ITEMS: Item[] = [
  {
    q: 'Feltölti valahová a videóimat?',
    a: (
      <>
        <p>
          Nem. A Snitt asztali alkalmazás: a fájljaid ott maradnak, ahol vannak, az indexelés
          helyben történik, az adatbázis pedig egyetlen SQLite fájl a gépeden. Nincs mögötte szerver,
          amire feltöltene bármit.
        </p>
        <p>
          Internet két esetben kell: ha linkről hozol be videót, illetve ha első alkalommal töltesz
          le egy Whisper modellt.
        </p>
      </>
    ),
  },
  {
    q: 'Mi történik, ha egy filmhez több felirat is tartozik?',
    a: (
      <p>
        Mindegyik megmarad. Egy filmhez tartozhat több fordítás és több nyelv, és ezek nem
        ugyanazokat a mondatokat tartalmazzák. A Snitt külön átiratként kezeli őket, a keresés
        pedig egyszerre fut mindegyiken — így akkor is találsz, ha épp az angol változatra emlékszel,
        de a magyar felirat van meg.
      </p>
    ),
  },
  {
    q: 'Akkor is működik, ha egyáltalán nincs feliratom?',
    a: (
      <>
        <p>
          Igen, ilyenkor jön a Whisper. A beszédfelismerés a videó hangsávjából készít átiratot,
          tehát pont azt kapod, ami elhangzik — nem a fordító megoldását. Ehhez a{' '}
          <code>faster-whisper</code> telepítése szükséges, és a hosszabb filmeknél ez a lépés időbe
          telik; utána viszont a keresés már azonnali.
        </p>
        <p>
          A telepítése rendszerenként pár parancs: végigvezet rajta a{' '}
          <Link to="/telepites">telepítési útmutató</Link>.
        </p>
      </>
    ),
  },
  {
    q: 'Muszáj pontosan idéznem?',
    a: (
      <p>
        Nem. A keresés hibrid: a szó szerinti egyezés mellett jelentés alapján is keres, így a
        körülírásra és az emlékezetből felidézett, kicsit pontatlan mondatra is hoz találatot. A
        listában látod, melyik átiratból és hányadik másodpercből származik az adott sor.
      </p>
    ),
  },
  {
    q: 'Kell fiók a használatához?',
    a: (
      <p>
        Nem. Az asztali alkalmazás önállóan, fiók nélkül működik. A weboldali regisztráció a készülő
        extra funkciókhoz tartozik, ezek nélkül is teljes értékű marad az alkalmazás.
      </p>
    ),
  },
  {
    q: 'Legális ez? Mit kezdhetek a kivágott klippel?',
    a: (
      <>
        <p>
          A Snitt a <strong>saját videótáradat</strong> indexeli, és a kivágott jelenetet helyi
          fájlként menti a gépedre. Az alkalmazás nem szerez be tartalmat és nem tesz közzé semmit —
          hogy mihez van jogod a saját másolataiddal, az a te felelősséged, és attól függ, hogyan
          jutottál hozzájuk, illetve hol élsz.
        </p>
        <p>
          Egy rövid részlet idézése — elemzéshez, kritikához, kommentárhoz, forrásmegjelöléssel — a
          legtöbb jogrendszerben az idézés szabályai alá esik. Egy egész film megosztása nyilván nem.
          A kettő között a hossz, a cél és a kontextus számít. Ez nem jogi tanács; ha konkrét ügyben
          bizonytalan vagy, kérdezz szakértőt.
        </p>
      </>
    ),
  },
];

export default function Faq() {
  return (
    <section id="gyik" className="section">
      <div className="container container-narrow">
        <header className="section-head">
          <span className="eyebrow">GYIK</span>
          <h2>Gyakori kérdések</h2>
        </header>

        <div className="faq">
          {ITEMS.map((item) => (
            <details key={item.q} className="faq-item">
              <summary>
                <span>{item.q}</span>
                <span className="faq-mark" aria-hidden="true" />
              </summary>
              <div className="faq-answer">{item.a}</div>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
