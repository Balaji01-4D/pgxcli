// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://Balaji01-4D.github.io',
	base: '/pgxcli/',
	integrations: [
		starlight({
			title: 'Pgxcli',
			customCss: ['./src/styles/custom.css'],
			social: [
				{
					label: 'GitHub',
					href: 'https://github.com/balaji01-4d/pgxcli',
					icon: 'github',
				},
			],
			sidebar: [
				{
					label: 'Introduction',
					items: [
						{ label: 'Getting Started', slug: 'guides/getting-started' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Reference', slug: 'reference/cli-reference' },
					],
				},
			],
		}),
	],
});
