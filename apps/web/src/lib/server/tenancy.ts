import { encoreApiUrl } from './encore';

export type Level = {
	code: string;
	label: string;
	parentLevel: string;
	depth: number;
	sortOrder: number;
};

export type Node = {
	id: number;
	parentId: number | null;
	level: string;
	code: string;
	label: string;
	createdAt: string;
	updatedAt: string;
};

type FetchOpts = {
	token: string;
};

async function authedFetch(path: string, token: string, init?: RequestInit) {
	return fetch(`${encoreApiUrl}${path}`, {
		...init,
		headers: {
			...(init?.headers ?? {}),
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	});
}

export async function listLevels({ token }: FetchOpts): Promise<Level[]> {
	const res = await authedFetch('/v1/tenancy/levels', token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.levels ?? []) as Level[];
}

export async function listNodes({ token }: FetchOpts, parentId: number | null, level?: string): Promise<Node[]> {
	const params = new URLSearchParams();
	if (parentId !== null) params.set('parentId', String(parentId));
	if (level) params.set('level', level);

	const qs = params.toString();
	const path = qs ? `/v1/tenancy/nodes?${qs}` : '/v1/tenancy/nodes';

	const res = await authedFetch(path, token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.nodes ?? []) as Node[];
}

export async function getNode({ token }: FetchOpts, id: number): Promise<Node | null> {
	const res = await authedFetch(`/v1/tenancy/nodes/${id}`, token);
	if (!res.ok) return null;

	const data = await res.json();
	return (data.node ?? null) as Node | null;
}

export type CreateNodeResult =
	| { ok: true; node: Node }
	| { ok: false; status: number; message: string };

export async function createNode(
	{ token }: FetchOpts,
	parentId: number | null,
	level: string,
	code: string,
	label: string
): Promise<CreateNodeResult> {
	const body: Record<string, unknown> = { Level: level, Code: code, Label: label };
	if (parentId !== null) body.ParentId = parentId;

	const res = await authedFetch('/v1/tenancy/nodes', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if (res.ok) {
		const data = await res.json();
		return { ok: true, node: data.node as Node };
	}

	let message = `Request failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {
	}

	return { ok: false, status: res.status, message };
}

export async function deleteNode({ token }: FetchOpts, id: number): Promise<{ ok: boolean; message?: string }> {
	const res = await authedFetch(`/v1/tenancy/nodes/${id}`, token, { method: 'DELETE' });
	if (res.ok) return { ok: true };

	let message = `Delete failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {
	}

	return { ok: false, message };
}
