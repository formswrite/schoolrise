import type { LayoutServerLoad } from './$types';

const stepOrder = [
	'welcome',
	'unlock',
	'admin',
	'system',
	'levels',
	'schools',
	'integrations',
	'smtp',
	'review'
];

export const load: LayoutServerLoad = async ({ url }) => {
	const segment = url.pathname.split('/')[2] ?? 'welcome';
	const stepIndex = Math.max(0, stepOrder.indexOf(segment));

	return {
		stepOrder,
		currentStep: segment,
		stepIndex,
		stepCount: stepOrder.length
	};
};
