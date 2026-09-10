// A nyelvi táblákat tölti be Node alól: a TypeScript forrást esbuilddel
// fordítjuk egyetlen ESM modullá, hogy a build-szkriptek (prerender, OG-kép)
// ugyanazokat a szövegeket lássák, mint az alkalmazás.

import { build } from 'esbuild';
import { mkdir, rm, writeFile } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { pathToFileURL } from 'node:url';

export async function loadStrings(langs) {
  const outfile = join('node_modules', '.cache', 'snitt-i18n.mjs');
  await mkdir(dirname(outfile), { recursive: true });

  // Több belépési pont nem fordítható egy outfile-ba, ezért egy köztes modult
  // írunk, ami mindegyik nyelvet újraexportálja.
  const shim = join('node_modules', '.cache', 'snitt-i18n-entry.ts');
  const reexports = langs
    .map(
      (lang) =>
        `export { ${lang} } from '${pathToFileURL(join(process.cwd(), `src/i18n/${lang}.ts`)).pathname}';`,
    )
    .join('\n');
  await writeFile(shim, `${reexports}\n`);
  await build({ entryPoints: [shim], bundle: true, format: 'esm', platform: 'node', outfile, logLevel: 'silent' });
  await rm(shim, { force: true });

  // Cache-törés: az import URL-hez időbélyeget fűzünk, hogy egy futáson belüli
  // újratöltés se adjon vissza régi modult.
  return import(`${pathToFileURL(join(process.cwd(), outfile)).href}?t=${Date.now()}`);
}
