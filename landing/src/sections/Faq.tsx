import { useT } from '../i18n';
import Rich from '../i18n/Rich';

export default function Faq() {
  const t = useT();

  return (
    <section id={t.sections.faq} className="section">
      <div className="container container-narrow">
        <header className="section-head">
          <span className="eyebrow">{t.faq.eyebrow}</span>
          <h2>{t.faq.title}</h2>
        </header>

        <div className="faq">
          {t.faq.items.map((item) => (
            <details key={item.q} className="faq-item">
              <summary>
                <span>{item.q}</span>
                <span className="faq-mark" aria-hidden="true" />
              </summary>
              <div className="faq-answer">
                {item.a.map((paragraph) => (
                  <p key={paragraph}>
                    <Rich text={paragraph} />
                  </p>
                ))}
              </div>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
