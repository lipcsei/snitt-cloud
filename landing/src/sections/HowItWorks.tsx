const STEPS = [
  {
    n: '01',
    title: 'Behúzod: mappa vagy link',
    body: 'Kiválasztod a merevlemezeden azt a mappát, ahol a filmjeid vannak — a Snitt helyben indexeli őket, egyetlen fájlt sem másol át máshová. Ha valamit nem tárolsz helyben, elég bemásolni egy videó linkjét.',
    note: 'A fájljaid ott maradnak, ahol vannak.',
  },
  {
    n: '02',
    title: 'Rákeresel az idézetre',
    body: 'Beírod azt a mondatot, ami megmaradt — akár szó szerint, akár csak nagyjából. A Snitt végigmegy az összes átiraton, minden nyelven, és megmutatja a pontos időpontot, ahol elhangzik.',
    note: 'Szó szerinti és jelentés szerinti találat egyszerre.',
  },
  {
    n: '03',
    title: 'Kivágod fájlba',
    body: 'A találatra kattintva megnyílik a vágó. Az átirat alapján már be van állítva egy javasolt tartomány, a végét-elejét húzással igazítod. A végeredmény egy videófájl, oda mented, ahová akarod.',
    note: 'Sima helyi fájl, se vízjel, se feltöltés.',
  },
];

export default function HowItWorks() {
  return (
    <section id="hogyan" className="section">
      <div className="container">
        <header className="section-head">
          <span className="eyebrow">Hogyan működik</span>
          <h2>Három lépés a kérdéstől a kész klipig</h2>
          <p className="section-sub">
            Nincs beállítási maraton. Megmutatod, hol vannak a videóid, és onnantól kereshető a
            tartalmuk.
          </p>
        </header>

        <ol className="steps">
          {STEPS.map((s) => (
            <li key={s.n} className="step">
              <span className="step-n">{s.n}</span>
              <h3>{s.title}</h3>
              <p>{s.body}</p>
              <span className="step-note">{s.note}</span>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
