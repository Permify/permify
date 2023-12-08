// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Permify',
  url: 'https://docs.permify.co/',
  tagline: "Open Source Authorization Service Based on Google Zanzibar",
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'Permify', // Usually your GitHub org/user name.
  projectName: 'permify', // Usually your repo name.
  trailingSlash: true,

  onBrokenLinks: 'warn',

  plugins: [
    [
      require.resolve("@cmfcmf/docusaurus-search-local"),
      {
        indexDocs: true,
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
              label: '0.6.x',
            },
          },
          editUrl: ({ docPath }) => {
            return `https://holocron.so/github/pr/Permify/permify/master/editor/docs/docs/${docPath}`
          },
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        gtag: {
          trackingID: 'G-BSRXWHBYR1',
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
              href: 'https://discord.gg/n6KfzYxhPp',
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
                  to: '/',
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
                  to: 'https://www.permify.co/post',
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
