import { encoreApiUrl } from './encore';

export type Email = {
	id: number;
	kind: string;
	to_email: string;
	to_name?: string;
	subject: string;
	status: 'pending' | 'sending' | 'sent' | 'failed' | 'dropped';
	attempts: number;
	last_error?: string;
	provider_id?: string;
	created_at: string;
	sent_at?: string;
};

export type ProviderStatus = {
	provider: string;
	from: string;
};

type Opts = { token: string };

async function req(path: string, token: string, init?: RequestInit) {
	return fetch(`${encoreApiUrl}${path}`, {
		...init,
		headers: {
			...(init?.headers ?? {}),
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	});
}

export async function listEmails({ token }: Opts, limit = 50): Promise<Email[]> {
	const res = await req(`/v1/notifications/outbox?limit=${limit}`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.emails ?? [];
}

export async function getProviderStatus({ token }: Opts): Promise<ProviderStatus | null> {
	const res = await req('/v1/notifications/provider', token);
	if (!res.ok) return null;
	return await res.json();
}

export async function sendTestEmail(
	{ token }: Opts,
	body: { to: string; subject?: string; body?: string }
) {
	const res = await req('/v1/notifications/test', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function processOutbox({ token }: Opts) {
	const res = await req('/v1/notifications/process', token, { method: 'POST' });
	return { ok: res.ok, status: res.status } as const;
}
