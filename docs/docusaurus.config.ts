import {themes as prismThemes} from 'prism-react-renderer'
import type {Config} from '@docusaurus/types'

const config: Config = {
  title: 'Velane Docs',
  tagline: 'Feature-first docs for Velane',

  url: 'https://velane.dev',
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
  plugins: [
    [
      '@easyops-cn/docusaurus-search-local',
      {
        hashed: true,
        docsDir: '.',
        docsRouteBasePath: '/',
        indexDocs: true,
        indexBlog: false,
        indexPages: false,
        explicitSearchResultPath: false
      }
    ]
  ],

  themeConfig: {
    navbar: {
      title: 'Velane Docs',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs'
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
    }
  }
}

export default config
