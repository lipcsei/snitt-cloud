import { Fragment, type ReactNode } from 'react';
import { Link } from 'react-router-dom';

/**
 * Pici jelölésértelmező, hogy a formázott mondatok is a nyelvi táblákban
 * maradhassanak, ne szakadjanak szét a komponensekben:
 *
 *   `kód`  **félkövér**  *kiemelt*  [szöveg](/utvonal)  \n = sortörés
 *
 * Szándékosan nem markdown: csak ez az öt eset kell, és ez így nulla függőség.
 */
const TOKEN = /(`[^`]+`|\*\*[^*]+\*\*|\*[^*\n]+\*|\[[^\]]+\]\([^)]+\)|\n)/g;
const LINK = /^\[([^\]]+)\]\(([^)]+)\)$/;

function render(text: string): ReactNode[] {
  return text.split(TOKEN).map((part, index) => {
    if (!part) return null;
    const key = `${index}-${part.slice(0, 8)}`;

    if (part === '\n') return <br key={key} />;

    if (part.startsWith('`') && part.endsWith('`')) {
      return <code key={key}>{part.slice(1, -1)}</code>;
    }

    if (part.startsWith('**') && part.endsWith('**')) {
      return <strong key={key}>{render(part.slice(2, -2))}</strong>;
    }

    if (part.startsWith('*') && part.endsWith('*')) {
      return <em key={key}>{render(part.slice(1, -1))}</em>;
    }

    const link = LINK.exec(part);
    if (link) {
      const [, label, href] = link;
      if (href.startsWith('/')) {
        return (
          <Link key={key} to={href}>
            {label}
          </Link>
        );
      }
      return (
        <a key={key} href={href} target="_blank" rel="noreferrer noopener">
          {label}
        </a>
      );
    }

    return <Fragment key={key}>{part}</Fragment>;
  });
}

export default function Rich({ text }: { text: string }) {
  return <>{render(text)}</>;
}
