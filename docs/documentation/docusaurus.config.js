// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const redirectJson = require('./redirects.json');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Permify',
  url: 'https://permify.co',
  tagline: "Open Source Authorization Service Based on Google Zanzibar",
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'Permify', // Usually your GitHub org/user name.
  projectName: 'permify', // Usually your repo name.
  trailingSlash: false,

  plugins: [
      [
        "@docusaurus/plugin-client-redirects",
        {
            redirects: redirectJson.redirects,
        },
      ],
  ],
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          lastVersion: 'current',
          versions: {
            current: {
              label: '0.3.x',
            },
          },
        },
        blog: {
          path: 'blog',
          editLocalizedFiles: false,
          blogTitle: 'Blog',
          blogDescription: 'Blog',
          blogSidebarCount: 50,
          blogSidebarTitle: 'Recent posts',
          routeBasePath: 'blog',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        gtag: {
          trackingID: 'GTM-5RNMNVR',
          anonymizeIP: true,
        },
      },
    ],
  ],

  themeConfig:
      {
        navbar: {
          title: 'Permify',
          logo: {
            alt: 'Permify Logo',
            src: 'img/logo.svg',
          },
          items: [
            {
              type: 'doc',
              docId: 'permify-overview/intro',
              position: 'left',
              label: 'Docs',
            },
            {to: 'blog', label: 'Blog', position: 'left'},
            {
              type: 'dropdown',
              label: 'API Reference',
              position: 'left',
              items: [
                {
                  label: 'gRPC API Reference',
                  href: 'https://buf.build/permify/permify/docs/main:base.v1',
                },
                {
                  label: 'REST API Reference',
                  href: 'https://app.swaggerhub.com/apis-docs/permify/permify/latest',
                },
              ],
            },
            {
              label: 'Playground',
              href: 'https://play.permify.co',
              position: 'left',
              className: 'header-playground-link'
            },
            {
              type: 'docsVersionDropdown',
              position: 'right',
            },
            {
              href: 'https://github.com/Permify/permify',
              position: 'right',
              className: 'header-github-link'
            },
            {
              href: 'https://discord.gg/MJbUjwskdH',
              position: 'right',
              className: 'header-discord-link'
            },
            {
              href: 'https://twitter.com/getPermify',
              position: 'right',
              className: 'header-twitter-link'
            }
          ],
        },
        /* algolia: {
          // The application ID provided by Algolia
          appId: 'PV97F452C4',
    
          // Public API key: it is safe to commit it
          apiKey: '3952e8f4f172e50b44abb743181536e8',
    
          indexName: 'permify',
        }, */
        metadata: [
          {
              name: "keywords",
              content:
                  "google zanzibar, authorization, permissions, rbac, rebac, abac, access control, fine grained",
          },
        ],
        footer: {
          style: 'dark',
          links: [
            {
              title: 'Docs',
              items: [
                {
                  label: 'Docs',
                  to: '/docs/',
                },
              ],
            },
            {
              title: 'Community',
              items: [
                {
                  label: 'Twitter',
                  href: 'https://twitter.com/getPermify',
                },
              ],
            },
            {
              title: 'More',
              items: [
                {
                  label: 'Blog',
                  to: 'https://www.permify.co/blog',
                },
                {
                  label: 'GitHub',
                  href: 'https://github.com/Permify',
                },
              ],
            },
          ],
          copyright: `Copyright © ${new Date().getFullYear()} Permify.`,
        },
        colorMode: {
          disableSwitch: false,
          respectPrefersColorScheme: true,
        },
        prism: {
          theme: darkCodeTheme,
          darkTheme: darkCodeTheme,
          additionalLanguages: ['php'],
        },
      },
};

module.exports = config;
