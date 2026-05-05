import type { Session } from '$lib/server/encore';

declare global {
	namespace App {
		interface Locals {
			user: Session | null;
			setupComplete: boolean;
		}
	}
}

export {};
