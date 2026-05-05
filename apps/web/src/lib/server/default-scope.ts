import type { Session } from './encore';
import { listNodes, getNode, type Node } from './tenancy';

export type ScopeOption = {
	id: number;
	label: string;
	level: string;
	depth: number;
};

export type ResolvedScope = {
	scopeNodeId: number;
	scope: Node;
	options: ScopeOption[];
};

export async function resolveDefaultScope(
	token: string,
	session: Session,
	requested: number | null
): Promise<ResolvedScope | null> {
	if (requested) {
		const scope = await getNode({ token }, requested);
		if (scope) {
			return { scopeNodeId: requested, scope, options: await scopeOptions(token, session) };
		}
	}

	for (const a of session.assignments) {
		if (a.role === 'inspector' && a.scopeNodeId !== null) {
			const scope = await getNode({ token }, a.scopeNodeId);
			if (scope) {
				return { scopeNodeId: a.scopeNodeId, scope, options: await scopeOptions(token, session) };
			}
		}
	}

	const roots = await listNodes({ token }, null);
	if (roots.length > 0) {
		return { scopeNodeId: roots[0].id, scope: roots[0], options: await scopeOptions(token, session) };
	}

	return null;
}

export async function scopeOptions(token: string, session: Session): Promise<ScopeOption[]> {
	const isGlobalAdmin = session.assignments.some((a) => a.role === 'admin' && a.scopeNodeId === null);
	const out: ScopeOption[] = [];

	if (isGlobalAdmin) {
		const roots = await listNodes({ token }, null);
		for (const r of roots) {
			out.push({ id: r.id, label: r.label, level: r.level, depth: 0 });
			const children = await listNodes({ token }, r.id);
			for (const c of children) {
				out.push({ id: c.id, label: c.label, level: c.level, depth: 1 });
			}
		}
		return out;
	}

	for (const a of session.assignments) {
		if (a.scopeNodeId !== null) {
			const node = await getNode({ token }, a.scopeNodeId);
			if (node) out.push({ id: node.id, label: node.label, level: node.level, depth: 0 });
		}
	}
	return out;
}

export type InstitutionPick = {
	institutionId: number;
	institution: Node;
	options: ScopeOption[];
};

export async function resolveDefaultInstitution(
	token: string,
	session: Session,
	requested: number | null
): Promise<InstitutionPick | null> {
	if (requested) {
		const node = await getNode({ token }, requested);
		if (node) return { institutionId: requested, institution: node, options: await institutionOptions(token, session) };
	}

	const opts = await institutionOptions(token, session);
	if (opts.length > 0) {
		const node = await getNode({ token }, opts[0].id);
		if (node) return { institutionId: opts[0].id, institution: node, options: opts };
	}
	return null;
}

export async function institutionOptions(token: string, session: Session): Promise<ScopeOption[]> {
	const isGlobalAdmin = session.assignments.some((a) => a.role === 'admin' && a.scopeNodeId === null);
	const institutions = await listNodes({ token }, null, 'institution').catch(() => [] as Node[]);

	const flat: ScopeOption[] = institutions.map((n) => ({
		id: n.id,
		label: n.label,
		level: n.level,
		depth: 0
	}));

	if (isGlobalAdmin) return flat.slice(0, 100);

	const allowedRoots = new Set<number>();
	for (const a of session.assignments) {
		if (a.scopeNodeId !== null) allowedRoots.add(a.scopeNodeId);
	}
	if (allowedRoots.size === 0) return [];

	const filtered: ScopeOption[] = [];
	for (const opt of flat) {
		const node = await getNode({ token }, opt.id);
		if (!node) continue;
		let cur: Node | null = node;
		let isUnder = false;
		while (cur) {
			if (allowedRoots.has(cur.id)) {
				isUnder = true;
				break;
			}
			if (!cur.parentId) break;
			cur = await getNode({ token }, cur.parentId);
		}
		if (isUnder) filtered.push(opt);
		if (filtered.length >= 100) break;
	}
	return filtered;
}
