# Integration Guide

This guide shows how to integrate `@thunder/i18n` into Thunder apps.

## Step 1: Add Package Dependency

Add the package to your app's `package.json`:

```json
{
  "dependencies": {
    "@thunder/i18n": "workspace:^"
  }
}
```

Then run `pnpm install` from the root.

## Step 2: Initialize i18n in Your App

### For Thunder Develop App

Edit `frontend/apps/thunder-develop/src/main.tsx`:

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { initI18n } from '@thunder/i18n';
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';
import App from './App';
import './index.css';

// Initialize i18n before rendering
(async () => {
  await initI18n({
    translations: { en, si },
    options: {
      defaultLanguage: 'en',
      debug: import.meta.env.DEV,
      storageKey: 'thunder-develop-language',
    },
  });

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>
  );
})();
```

### For Thunder Gate App

Edit `frontend/apps/thunder-gate/src/main.tsx`:

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { initI18n } from '@thunder/i18n';
import en from '@thunder/i18n/locales/en';
import si from '@thunder/i18n/locales/si';
import App from './App';
import './index.css';

// Initialize i18n before rendering
(async () => {
  await initI18n({
    translations: { en, si },
    options: {
      defaultLanguage: 'en',
      detectLanguage: true,
      storageKey: 'thunder-gate-language',
    },
  });

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>
  );
})();
```

## Step 3: Create a Language Switcher Component

Create a reusable language switcher component:

```tsx
// components/LanguageSwitcher.tsx
import { useLanguage } from '@thunder/i18n';
import { IconButton, Menu, MenuItem } from '@mui/material';
import { Language as LanguageIcon } from '@mui/icons-material';
import { useState } from 'react';

export function LanguageSwitcher() {
  const { currentLanguage, availableLanguages, setLanguage } = useLanguage();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageChange = (langCode: 'en' | 'si') => {
    setLanguage(langCode);
    handleClose();
  };

  return (
    <>
      <IconButton onClick={handleClick} color="inherit">
        <LanguageIcon />
      </IconButton>
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleClose}>
        {availableLanguages.map((lang) => (
          <MenuItem
            key={lang.code}
            selected={lang.code === currentLanguage}
            onClick={() => handleLanguageChange(lang.code)}
          >
            {lang.nativeName}
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
```

## Step 4: Use Translations in Components

### Example: Users Page (Thunder Develop)

```tsx
// pages/Users/Users.tsx
import { useTranslation } from '@thunder/i18n';
import { Button, Typography } from '@mui/material';

export function UsersPage() {
  const { t } = useTranslation();

  return (
    <div>
      <Typography variant="h4">{t('develop.users.title')}</Typography>
      <Button variant="contained">{t('develop.users.addUser')}</Button>

      {/* Table with translated headers */}
      <table>
        <thead>
          <tr>
            <th>{t('develop.users.firstName')}</th>
            <th>{t('develop.users.lastName')}</th>
            <th>{t('develop.users.email')}</th>
            <th>{t('develop.users.role')}</th>
            <th>{t('develop.users.status')}</th>
            <th>{t('develop.users.actions')}</th>
          </tr>
        </thead>
      </table>
    </div>
  );
}
```

### Example: Sign In Page (Thunder Gate)

```tsx
// pages/SignIn/SignIn.tsx
import { useTranslation } from '@thunder/i18n';
import { Button, TextField, Typography } from '@mui/material';

export function SignInPage() {
  const { t } = useTranslation();

  return (
    <div>
      <Typography variant="h4">{t('gate.auth.welcomeBack')}</Typography>
      <Typography variant="h5">{t('gate.auth.signIn')}</Typography>

      <TextField
        label={t('common.form.email')}
        placeholder={t('gate.auth.enterEmail')}
        required
      />
      <TextField
        type="password"
        label={t('common.form.password')}
        placeholder={t('gate.auth.enterPassword')}
        required
      />

      <Button variant="contained">{t('gate.auth.signIn')}</Button>
      <Button variant="text">{t('gate.auth.forgotPassword')}</Button>
    </div>
  );
}
```

## Step 5: Add Language Switcher to App Bar

```tsx
// components/AppBar/AppBar.tsx
import { AppBar, Toolbar, Typography } from '@mui/material';
import { LanguageSwitcher } from '../LanguageSwitcher';
import { useTranslation } from '@thunder/i18n';

export function AppBarComponent() {
  const { t } = useTranslation();

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" sx={{ flexGrow: 1 }}>
          {t('common.navigation.dashboard')}
        </Typography>
        <LanguageSwitcher />
      </Toolbar>
    </AppBar>
  );
}
```

## Step 6: Use Formatting Utilities

```tsx
// components/Dashboard/Dashboard.tsx
import { useTranslation } from '@thunder/i18n';
import { formatDate, formatNumber } from '@thunder/i18n';

export function Dashboard() {
  const { t } = useTranslation();

  return (
    <div>
      <h1>{t('develop.dashboard.welcomeMessage')}</h1>

      <div>
        <h3>{t('develop.dashboard.totalUsers')}</h3>
        <p>{formatNumber(12345)}</p>
      </div>

      <div>
        <h3>{t('develop.dashboard.recentActivity')}</h3>
        <p>{formatDate(new Date(), { dateStyle: 'long', timeStyle: 'short' })}</p>
      </div>
    </div>
  );
}
```

## Step 7: Handle Form Validation Messages

```tsx
// components/UserForm/UserForm.tsx
import { useTranslation } from '@thunder/i18n';
import { useForm } from 'react-hook-form';

export function UserForm() {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm();

  return (
    <form>
      <input
        {...register('email', {
          required: t('common.form.requiredField'),
          pattern: {
            value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
            message: t('common.form.invalidEmail'),
          },
        })}
      />
      {errors.email && <span>{errors.email.message}</span>}

      <button type="submit">{t('common.actions.submit')}</button>
    </form>
  );
}
```

## Step 8: Add TypeScript Support

Create a `tsconfig.json` reference in your app if not already present:

```json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "types": ["@thunder/i18n"]
  },
  "references": [
    { "path": "../../packages/thunder-i18n" }
  ]
}
```

## Tips and Best Practices

### 1. Loading State

Show a loading state while i18n initializes:

```tsx
// App.tsx
import { Suspense } from 'react';

function App() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <YourApp />
    </Suspense>
  );
}
```

### 2. Lazy Loading Languages

For better performance, lazy load languages:

```tsx
async function loadLanguage(lang: 'en' | 'si') {
  if (lang === 'en') {
    return import('@thunder/i18n/locales/en');
  }
  return import('@thunder/i18n/locales/si');
}

const detectedLang = detectLanguage('thunder-language');
const translation = await loadLanguage(detectedLang);

await initI18n({
  translations: { [detectedLang]: translation.default },
});
```

### 3. Pluralization

```tsx
// Add to translations
{
  "users": {
    "count_one": "{{count}} user",
    "count_other": "{{count}} users"
  }
}

// Use in component
<p>{t('users.count', { count: userCount })}</p>
```

### 4. Interpolation

```tsx
// Translation
{
  "welcome": "Welcome, {{name}}!"
}

// Usage
<p>{t('common.welcome', { name: user.name })}</p>
```

### 5. Date and Time Formats

```tsx
import { formatDate } from '@thunder/i18n';

// Different formats
const shortDate = formatDate(date, { dateStyle: 'short' });
const longDate = formatDate(date, { dateStyle: 'long' });
const dateTime = formatDate(date, {
  dateStyle: 'medium',
  timeStyle: 'short'
});
```

## Troubleshooting

### Translations Not Loading

1. Make sure i18n is initialized before rendering the app
2. Check that the language files are properly imported
3. Verify the translation keys exist in both language files

### TypeScript Errors

1. Rebuild the i18n package: `pnpm build --filter @thunder/i18n`
2. Restart your TypeScript server in VSCode
3. Ensure the package is listed in dependencies

### Language Not Persisting

1. Check browser localStorage for the storage key
2. Verify the `storageKey` option is set correctly
3. Ensure the app has permission to use localStorage

## Next Steps

- Add more languages by creating new files in `src/locales/`
- Extend translations for your specific features
- Create shared translation components
- Set up translation management workflow
