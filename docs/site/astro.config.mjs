import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'KeePass Vault CLI',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/frostyeti/kpv' },
			],
			sidebar: [
				{
					label: 'Documentation',
					items: [
						{ label: 'Home', link: '/' },
					],
				}
			],
		}),
	],
});