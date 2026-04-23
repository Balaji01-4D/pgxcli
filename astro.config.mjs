import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import icon from 'astro-icon';

// https://astro.build/config
export default defineConfig({
	site: 'https://balaji01-4d.github.io',
	base: '/pgxcli/',
	integrations: [
		starlight({
			title: 'Pgxcli',
			logo: {
				src: './src/assets/logo.png',
			},
			customCss: ['./src/styles/custom.css'],
			social: [
				{
					label: 'GitHub',
					href: 'https://github.com/balaji01-4d/pgxcli',
					icon: 'github',
				},
			],
			sidebar: [
				{ label: 'Getting Started', slug: 'guides/getting-started' },
				{ label: 'pgxcli vs pgcli', slug: 'guides/comparison-with-pgcli' },
				{
					label: 'Usage Guides',
					items: [
						{ label: 'Connecting', slug: 'guides/connecting' },
						{ label: 'Configuration', slug: 'guides/configuration' },
						{ label: 'Special Commands', slug: 'guides/special-commands' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Flags', slug: 'reference/cli-reference' },
						{ label: 'Environment Variables', slug: 'reference/environment-variables' },
					],
				},
			],
		}),
		icon(),
	],
});
