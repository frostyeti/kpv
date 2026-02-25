// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'kpv',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/frostyeti/kpv' }],
			sidebar: [
				{
					label: 'Guides',
					items: [
						{ label: 'Getting Started', slug: 'guides/getting-started' },
					],
				},
				{
					label: 'Commands',
					autogenerate: { directory: 'commands' },
				},
			],
		}),
	],
});
