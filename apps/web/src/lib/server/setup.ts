import { encoreApiUrl } from './encore';

export type SetupStatus = {
	setupComplete: boolean;
	installTokenSet: boolean;
	tokenConsumed: boolean;
	failedAttempts: number;
};

let statusCache: { value: SetupStatus; expiresAt: number } | null = null;

const cacheTTLms = 5_000;

export async function fetchSetupStatus(): Promise<SetupStatus> {
	const now = Date.now();

	if (statusCache && statusCache.expiresAt > now) {
		return statusCache.value;
	}

	try {
		const res = await fetch(`${encoreApiUrl}/v1/setup/status`);

		if (!res.ok) {
			return {
				setupComplete: false,
				installTokenSet: false,
				tokenConsumed: false,
				failedAttempts: 0
			};
		}

		const value = (await res.json()) as SetupStatus;
		statusCache = { value, expiresAt: now + cacheTTLms };
		return value;
	} catch {
		return {
			setupComplete: false,
			installTokenSet: false,
			tokenConsumed: false,
			failedAttempts: 0
		};
	}
}

export function invalidateSetupStatusCache(): void {
	statusCache = null;
}

export type UnlockResult =
	| { ok: true; sessionToken: string; expiresAt: string }
	| { ok: false; status: number; message: string };

export async function setupUnlock(installToken: string): Promise<UnlockResult> {
	try {
		const res = await fetch(`${encoreApiUrl}/v1/setup/unlock`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ InstallToken: installToken })
		});

		if (res.ok) {
			const data = await res.json();
			return { ok: true, sessionToken: data.SessionToken, expiresAt: data.ExpiresAt };
		}

		const message =
			res.status === 401
				? 'Invalid install token.'
				: res.status === 403
					? 'Install token already consumed or setup is locked.'
					: res.status === 429
						? 'Too many failed attempts. Restart the app to issue a new token.'
						: 'Unlock failed.';

		return { ok: false, status: res.status, message };
	} catch {
		return { ok: false, status: 502, message: 'Setup service unreachable.' };
	}
}

async function postSetup(
	path: string,
	body: Record<string, unknown>
): Promise<{ ok: boolean; status: number; message?: string; data?: unknown }> {
	try {
		const res = await fetch(`${encoreApiUrl}${path}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});

		if (res.ok) {
			const text = await res.text();
			if (!text) return { ok: true, status: res.status };

			try {
				return { ok: true, status: res.status, data: JSON.parse(text) };
			} catch {
				return { ok: true, status: res.status };
			}
		}

		let message = `Request failed (${res.status})`;
		try {
			const data = await res.json();
			if (data?.message) message = String(data.message);
		} catch {}

		return { ok: false, status: res.status, message };
	} catch {
		return { ok: false, status: 502, message: 'Setup service unreachable.' };
	}
}

export async function setupCreateAdmin(
	sessionToken: string,
	email: string,
	fullName: string,
	password: string
) {
	return postSetup('/v1/setup/admin', {
		SessionToken: sessionToken,
		Email: email,
		FullName: fullName,
		Password: password
	});
}

export async function setupSaveSystem(
	sessionToken: string,
	instanceName: string,
	defaultLocale: string,
	baseURL: string,
	timeZone: string
) {
	return postSetup('/v1/setup/system', {
		SessionToken: sessionToken,
		InstanceName: instanceName,
		DefaultLocale: defaultLocale,
		BaseURL: baseURL,
		TimeZone: timeZone
	});
}

export type LevelInput = {
	code: string;
	label: string;
	parent: string;
	depth: number;
	sort: number;
};

export async function setupSetLevels(sessionToken: string, levels: LevelInput[]) {
	return postSetup('/v1/setup/levels', {
		SessionToken: sessionToken,
		Levels: levels.map((l) => ({
			Code: l.code,
			Label: l.label,
			Parent: l.parent,
			Depth: l.depth,
			Sort: l.sort
		}))
	});
}

export async function setupImportSchools(sessionToken: string, csv: string) {
	return postSetup('/v1/setup/schools/import', { SessionToken: sessionToken, CSV: csv });
}

export async function setupSkipSchools(sessionToken: string) {
	return postSetup('/v1/setup/schools/skip', { SessionToken: sessionToken });
}

export async function setupSaveIntegrations(sessionToken: string, fields: Record<string, string>) {
	return postSetup('/v1/setup/integrations', { SessionToken: sessionToken, ...fields });
}

export async function setupSkipIntegrations(sessionToken: string) {
	return postSetup('/v1/setup/integrations/skip', { SessionToken: sessionToken });
}

export async function setupSaveSMTP(
	sessionToken: string,
	host: string,
	port: number,
	username: string,
	password: string,
	useTLS: boolean,
	fromAddress: string
) {
	return postSetup('/v1/setup/smtp', {
		SessionToken: sessionToken,
		Host: host,
		Port: port,
		Username: username,
		Password: password,
		UseTLS: useTLS,
		FromAddress: fromAddress
	});
}

export async function setupSkipSMTP(sessionToken: string) {
	return postSetup('/v1/setup/smtp/skip', { SessionToken: sessionToken });
}

export async function setupFinalize(sessionToken: string) {
	return postSetup('/v1/setup/finalize', { SessionToken: sessionToken });
}
