import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupSaveIntegrations, setupSkipIntegrations } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	save: async ({ request, cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		const data = await request.formData();
		const fields: Record<string, string> = {
			ResendAPIKey: String(data.get('resend_api_key') ?? ''),
			EmailFrom: String(data.get('email_from') ?? ''),
			EmailFromName: String(data.get('email_from_name') ?? ''),
			OpenAIAPIKey: String(data.get('openai_api_key') ?? ''),
			OpenAIModel: String(data.get('openai_model') ?? ''),
			AnthropicAPIKey: String(data.get('anthropic_api_key') ?? ''),
			S3Endpoint: String(data.get('s3_endpoint') ?? ''),
			S3Bucket: String(data.get('s3_bucket') ?? ''),
			S3AccessKey: String(data.get('s3_access_key') ?? ''),
			S3SecretKey: String(data.get('s3_secret_key') ?? ''),
			S3Region: String(data.get('s3_region') ?? '')
		};

		const result = await setupSaveIntegrations(session, fields);
		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Could not save integrations.' });
		}

		throw redirect(303, '/setup/smtp');
	},
	skip: async ({ cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		await setupSkipIntegrations(session);

		throw redirect(303, '/setup/smtp');
	}
};
