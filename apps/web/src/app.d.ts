// See https://kit.svelte.dev/docs/types#app
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user?: {
				id: number;
				organizationId: number;
				name: string;
				email: string;
			};
			accessToken?: string;
		}
		// interface PageData {}
		// interface Platform {}
	}
}

export {};
