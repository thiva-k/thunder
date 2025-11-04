# @thunder/i18n

Internationalization (i18n) package for Thunder applications using react-i18next with full TypeScript support and tree-shaking capabilities.

## Features

- **Tree-shakable translations** - Import only the languages you need
- **Type-safe translations** - Full TypeScript support with autocomplete
- **React hooks** - Easy-to-use hooks for language switching
- **Language persistence** - Automatically saves user's language preference
- **Browser detection** - Detects user's preferred language from browser settings
- **Formatting utilities** - Built-in date, number, and currency formatting
- **Multiple apps support** - Shared translations for develop, gate, and other Thunder apps

## Installation

Since this is a workspace package, install dependencies from the root:

```bash
pnpm install
```

## Usage

### 1. Initialize i18n in your app

In your app's entry point (e.g., `main.tsx`):

```tsx
import { initI18n } from '@thunder/i18n';
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';

// Initialize i18n before rendering your app
await initI18n({
  translations: { en, si },
  options: {
    defaultLanguage: 'en',
    debug: process.env.NODE_ENV === 'development',
  },
});

// Then render your app
root.render(<App />);
```

### 2. Use translations in components

```tsx
import { useTranslation } from '@thunder/i18n';

function MyComponent() {
  const { t } = useTranslation();

  return (
    <div>
      <h1>{t('common.navigation.home')}</h1>
      <p>{t('common.messages.welcomeMessage')}</p>
      <button>{t('common.actions.save')}</button>
    </div>
  );
}
```

### 3. Language switching

```tsx
import { useLanguage } from '@thunder/i18n';

function LanguageSwitcher() {
  const { currentLanguage, availableLanguages, setLanguage } = useLanguage();

  return (
    <select
      value={currentLanguage}
      onChange={(e) => setLanguage(e.target.value as 'en' | 'si')}
    >
      {availableLanguages.map((lang) => (
        <option key={lang.code} value={lang.code}>
          {lang.nativeName}
        </option>
      ))}
    </select>
  );
}
```

### 4. Using translation namespaces

The translations are organized by app and feature:

```tsx
// Common translations (shared across all apps)
t('common.actions.save')
t('common.status.loading')
t('common.form.email')

// Thunder Develop app
t('develop.users.title')
t('develop.applications.addApplication')
t('develop.dashboard.welcomeMessage')

// Thunder Gate app
t('gate.auth.signIn')
t('gate.mfa.setupMfa')
t('gate.consent.title')
```

### 5. Formatting utilities

```tsx
import { formatDate, formatNumber, formatCurrency } from '@thunder/i18n';

// Format dates
const date = formatDate(new Date(), { dateStyle: 'long' });

// Format numbers
const number = formatNumber(1234567.89);

// Format currency
const price = formatCurrency(99.99, 'USD');
```

### 6. Advanced: Using Trans component for complex translations

```tsx
import { Trans } from '@thunder/i18n';

function Component() {
  return (
    <Trans i18nKey="common.messages.welcome">
      Welcome <strong>{{ name: user.name }}</strong>!
    </Trans>
  );
}
```

## Translation Structure

Translations are organized into **namespaces** for better modularity and tree-shaking:

### Namespace Organization

```
src/locales/
├── en/
│   ├── index.ts                  # Main export (combines all namespaces)
│   └── namespaces/
│       ├── common.ts             # Shared across all apps
│       ├── develop.ts            # Thunder Develop app
│       └── gate.ts               # Thunder Gate app
└── si/
    ├── index.ts                  # Main export (combines all namespaces)
    └── namespaces/
        ├── common.ts             # Shared across all apps
        ├── develop.ts            # Thunder Develop app
        └── gate.ts               # Thunder Gate app
```

### Namespace Details

**`common` namespace** - Shared across all Thunder applications:
- `actions` - Action buttons (save, cancel, etc.)
- `status` - Status messages (loading, error, etc.)
- `form` - Form labels and validation
- `messages` - Common messages
- `navigation` - Navigation items

**`develop` namespace** - Thunder Develop application:
- `pages` - Page titles
- `users` - User management
- `userTypes` - User type management
- `integrations` - Integration management
- `applications` - Application management
- `dashboard` - Dashboard content

**`gate` namespace** - Thunder Gate application:
- `auth` - Authentication flows
- `mfa` - Multi-factor authentication
- `social` - Social login
- `consent` - Consent management
- `errors` - Error messages

### Selective Namespace Imports

You can import individual namespaces for better tree-shaking:

```tsx
// Import only the common namespace for English
import { common as commonEn } from '@thunder/i18n/locales/en/namespaces/common';
import { common as commonSi } from '@thunder/i18n/locales/si/namespaces/common';

// Import only develop namespace
import { develop as developEn } from '@thunder/i18n/locales/en/namespaces/develop';
import { develop as developSi } from '@thunder/i18n/locales/si/namespaces/develop';

// Use them in initialization
await initI18n({
  translations: {
    en: { common: commonEn, develop: developEn },
    si: { common: commonSi, develop: developSi },
  },
});
```

This approach allows apps to only bundle the translations they actually use!

## Supported Languages

- **English (en)** - Default language
- **Sinhala (si)** - සිංහල

## Configuration Options

```tsx
interface I18nOptions {
  defaultLanguage?: 'en' | 'si';        // Default: 'en'
  fallbackLanguage?: 'en' | 'si';       // Default: 'en'
  debug?: boolean;                       // Default: false
  namespace?: string;                    // Default: 'translation'
  detectLanguage?: boolean;              // Default: true
  storageKey?: string;                   // Default: 'thunder-language'
}
```

## TypeScript Support

This package is fully typed. You'll get autocomplete for all translation keys:

```tsx
// ✅ TypeScript will autocomplete these
t('common.actions.save')
t('develop.users.title')
t('gate.auth.signIn')

// ❌ TypeScript will error on invalid keys
t('common.invalid.key')  // Error: Key does not exist
```

## Tree Shaking

Import only the languages you need to reduce bundle size:

```tsx
// Import only English
import en from '@thunder/i18n/locales/en';

await initI18n({
  translations: { en },
});

// Or import specific languages as needed
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';

await initI18n({
  translations: { en, si },
});
```

## Adding New Translations

Translations are now organized by namespaces. To add new translations:

### 1. Choose the appropriate namespace

- Add to `common` namespace if the translation is shared across multiple apps
- Add to `develop` namespace for Thunder Develop-specific features
- Add to `gate` namespace for Thunder Gate-specific features

### 2. Add translations to both languages

Edit the namespace files for both English and Sinhala:

```tsx
// src/locales/en/namespaces/common.ts
export const common = {
  // ... existing translations
  newFeature: {
    title: 'New Feature',
    description: 'This is a new feature',
  },
} as const;

// src/locales/si/namespaces/common.ts
export const common = {
  // ... existing translations
  newFeature: {
    title: 'නව විශේෂාංගය',
    description: 'මෙය නව විශේෂාංගයකි',
  },
} as const;
```

### 3. Rebuild the package

```bash
cd packages/thunder-i18n
pnpm build
pnpm tsc -p tsconfig.lib.json --emitDeclarationOnly
```

TypeScript will automatically pick up the new keys and provide autocomplete!

### 4. Use the new translation

```tsx
const { t } = useTranslation();
<h1>{t('common.newFeature.title')}</h1>
```

## Development

```bash
# Build the package
pnpm build

# Watch mode for development
pnpm dev

# Type checking
pnpm typecheck

# Linting
pnpm lint
```

## Integration with Thunder Apps

### Thunder Develop App

Example integration in `apps/thunder-develop/src/main.tsx`:

```tsx
import { initI18n } from '@thunder/i18n';
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';

await initI18n({
  translations: { en, si },
  options: {
    defaultLanguage: 'en',
    debug: import.meta.env.DEV,
  },
});
```

### Thunder Gate App

Example integration in `apps/thunder-gate/src/main.tsx`:

```tsx
import { initI18n } from '@thunder/i18n';
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';

await initI18n({
  translations: { en, si },
  options: {
    defaultLanguage: 'en',
    detectLanguage: true,
  },
});
```

## API Reference

### Functions

- `initI18n(config)` - Initialize i18n with translations
- `changeLanguage(lang)` - Change current language
- `getCurrentLanguage()` - Get current language
- `getAvailableLanguages()` - Get all available languages
- `formatDate(date, options)` - Format date according to current locale
- `formatNumber(num, options)` - Format number according to current locale
- `formatCurrency(num, currency, options)` - Format currency

### Hooks

- `useLanguage()` - Hook for language management
- `useTranslation()` - React-i18next hook for translations

### Types

- `SupportedLanguage` - Union type of supported languages
- `TranslationKey` - All available translation keys
- `I18nOptions` - Configuration options
- `LanguageConfig` - Language metadata

## License

Apache-2.0
