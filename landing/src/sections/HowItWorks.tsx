import { useT } from '../i18n';

const NUMBERS = ['01', '02', '03'];

export default function HowItWorks() {
  const t = useT();

  return (
    <section id={t.sections.how} className="section">
      <div className="container">
        <header className="section-head">
          <span className="eyebrow">{t.howItWorks.eyebrow}</span>
          <h2>{t.howItWorks.title}</h2>
          <p className="section-sub">{t.howItWorks.sub}</p>
        </header>

        <ol className="steps">
          {t.howItWorks.steps.map((s, i) => (
            <li key={s.title} className="step">
              <span className="step-n">{NUMBERS[i]}</span>
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
