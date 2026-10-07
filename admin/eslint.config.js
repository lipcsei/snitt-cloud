// eslint - forrás: lipcsei/commons ops/static (a breath-ben kiérlelt beállítás). A frontend
// könyvtárába másolandó; kell hozzá: eslint @eslint/js typescript-eslint globals
// eslint-plugin-react-hooks eslint-plugin-react-refresh
import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist', 'src/api/schema.d.ts'] },
  {
    // recommendedTypeChecked, nem a sima recommended: a tipizálatlan változat nem látja a
    // típusokat, így a hibaosztályok fele kiesik (pl. el nem kapott Promise egy onClick-ben).
    extends: [js.configs.recommended, ...tseslint.configs.recommendedTypeChecked],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
      // A típus-tudatos szabályokhoz a tsconfig-projekt kell.
      parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname },
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
)
