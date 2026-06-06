import {themes as prismThemes} from 'prism-react-renderer'
import type {Config} from '@docusaurus/types'

const algoliaAppId = process.env.ALGOLIA_APP_ID
const algoliaApiKey = process.env.ALGOLIA_API_KEY
const algoliaIndexName = process.env.ALGOLIA_INDEX_NAME
const useAlgolia = Boolean(algoliaAppId && algoliaApiKey && algoliaIndexName)

const config: Config = {
  title: 'Velane Docs',
  tagline: 'Feature-first docs for Velane',

  url: 'https://docs.velane.sh',
  baseUrl: '/',

  onBrokenLinks: 'throw',
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn'
    }
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en']
  },

  presets: [
    [
      'classic',
      {
        docs: {
          path: '.',
          routeBasePath: '/',
          sidebarPath: './sidebars.ts',
          exclude: ['**/node_modules/**', '**/build/**', '**/.docusaurus/**']
        },
        blog: false,
        pages: false,
        theme: {
          customCss: './src/css/custom.css'
        }
      }
    ]
  ],
  plugins: [],

  themeConfig: {
    navbar: {
      title: 'Velane Docs',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs'
        },
        {
          href: 'https://github.com/abskrj/velane',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
          position: 'right'
        }
      ]
    },
    footer: {
      style: 'dark',
      links: [],
      copyright: `Copyright ${new Date().getFullYear()} Velane`
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula
    },
    ...(useAlgolia
      ? {
          algolia: {
            appId: algoliaAppId!,
            apiKey: algoliaApiKey!,
            indexName: algoliaIndexName!,
            contextualSearch: true
          }
        }
      : {})
  }
}

export default config
