import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'pgxcli',
  tagline: 'An interactive PostgreSQL client for your terminal.',
  favicon: 'img/favicon.ico',

  headTags: [
    {
      tagName: 'script',
      attributes: {
        type: 'application/ld+json',
      },
      innerHTML: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        'name': 'pgxcli',
        'operatingSystem': 'Linux, macOS, Windows',
        'applicationCategory': 'DeveloperApplication',
        'offers': {
          '@type': 'Offer',
          'price': '0.00',
          'priceCurrency': 'USD'
        },
        'description': 'An interactive PostgreSQL command-line client written in Go with syntax highlighting and autocompletion.',
        'url': 'https://pgxcli.vercel.app'
      }),
    },
  ],

  future: {
    v4: true,
  },

  url: 'https://pgxcli.vercel.app',
  baseUrl: '/',
  clientModules: ['./src/theme/vercel-analytics.ts'],

  onBrokenLinks: 'warn',
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  stylesheets: [
    {
      href: 'https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,400;0,500;0,600;0,700;0,800;1,400;1,600&display=swap',
      type: 'text/css',
    },
  ],

  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        hashed: true,
        docsRouteBasePath: '/docs',
        indexBlog: false,
        highlightSearchTermsOnTargetPage: true,
        searchResultLimits: 8,
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/app-screenshot.png',
    metadata: [
      {name: 'keywords', content: 'postgres, postgresql, psql, cli, terminal, database, go, golang, database client, sql'},
      {name: 'twitter:card', content: 'summary_large_image'},
    ],
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: false,
    },
    navbar: {
      title: 'pgxcli',
      items: [
        {
          type: 'html',
          position: 'right',
          value: `<span class="navbar-quote"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" style="vertical-align:-2px;margin-right:5px;color:#f97316"><path d="M17.66 11.2c-.23-.3-.51-.56-.77-.82-.67-.6-1.43-1.03-2.07-1.66C13.33 7.26 13 4.85 13.95 3c-.95.23-1.78.75-2.49 1.32-2.59 2.08-3.61 5.75-2.39 8.9.04.1.08.2.08.33 0 .22-.15.42-.35.5-.23.1-.47.04-.66-.12a.58.58 0 0 1-.14-.17c-1.13-1.43-1.31-3.48-.55-5.12C5.78 10 4.87 12.3 5 14.47c.06.5.12 1 .29 1.5.14.6.41 1.2.71 1.73 1.08 1.73 2.95 2.97 4.96 3.22 2.14.27 4.43-.12 6.07-1.6 1.83-1.66 2.47-4.32 1.53-6.6l-.13-.28c-.2-.47-.43-.8-.77-1.04z"/></svg>Set your heart ablaze</span>`,
        },
        {
          to: '/support',
          label: 'Support',
          position: 'right',
        },
        {
          type: 'custom-Github',
          href: 'https://github.com/balaji01-4d/pgxcli',
          position: 'right',
        },
      ],
    },
    prism: {
      // vsDark gives clean neutral syntax colours that don't clash with purple UI
      theme: prismThemes.nightOwl,
      darkTheme: prismThemes.nightOwl,
      additionalLanguages: ['bash', 'toml', 'sql', 'powershell'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
