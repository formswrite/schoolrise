import type { Question } from './field-types';

export type ConditionOperator =
	| 'equals'
	| 'not_equals'
	| 'gt'
	| 'lt'
	| 'gte'
	| 'lte'
	| 'contains'
	| 'either';

export type Condition = {
	source_question_client_id: string;
	op: ConditionOperator;
	value: string | number | string[];
};

export type RuleOperator = 'show_if' | 'hide_if';

export type LogicRule = {
	id: string;
	target_question_client_id: string;
	operator: RuleOperator;
	conditions: Condition[];
};

export type Answers = Record<string, string | number | string[] | undefined>;

export const OPERATOR_LABELS: Record<ConditionOperator, string> = {
	equals: 'equals',
	not_equals: 'does not equal',
	gt: 'greater than',
	lt: 'less than',
	gte: 'at least',
	lte: 'at most',
	contains: 'contains',
	either: 'is one of'
};

const NUMERIC_OPS: ReadonlySet<ConditionOperator> = new Set(['gt', 'lt', 'gte', 'lte']);

function asNumber(v: unknown): number | null {
	if (typeof v === 'number') return v;
	if (typeof v === 'string' && v.trim() !== '') {
		const n = Number(v);
		return Number.isFinite(n) ? n : null;
	}
	return null;
}

export function evaluateCondition(c: Condition, answers: Answers): boolean {
	const ans = answers[c.source_question_client_id];

	if (NUMERIC_OPS.has(c.op)) {
		const a = asNumber(ans);
		const b = asNumber(c.value);
		if (a === null || b === null) return false;
		switch (c.op) {
			case 'gt':
				return a > b;
			case 'lt':
				return a < b;
			case 'gte':
				return a >= b;
			case 'lte':
				return a <= b;
		}
	}

	switch (c.op) {
		case 'equals':
			return String(ans ?? '') === String(c.value ?? '');
		case 'not_equals':
			return String(ans ?? '') !== String(c.value ?? '');
		case 'contains':
			if (Array.isArray(ans)) return ans.includes(String(c.value));
			return String(ans ?? '').includes(String(c.value ?? ''));
		case 'either': {
			const set = Array.isArray(c.value) ? c.value : [String(c.value ?? '')];
			if (Array.isArray(ans)) return ans.some((v) => set.includes(String(v)));
			return set.includes(String(ans ?? ''));
		}
	}
	return false;
}

export function evaluateRule(rule: LogicRule, answers: Answers): boolean {
	if (rule.conditions.length === 0) return rule.operator === 'show_if';
	return rule.conditions.every((c) => evaluateCondition(c, answers));
}

export function computeVisibleQuestions(
	questions: Question[],
	rules: LogicRule[],
	answers: Answers
): Question[] {
	if (rules.length === 0) return questions;

	const rulesByTarget = new Map<string, LogicRule[]>();
	for (const r of rules) {
		const list = rulesByTarget.get(r.target_question_client_id) ?? [];
		list.push(r);
		rulesByTarget.set(r.target_question_client_id, list);
	}

	return questions.filter((q) => {
		const rs = rulesByTarget.get(q.client_id);
		if (!rs || rs.length === 0) return true;

		for (const r of rs) {
			const matched = evaluateRule(r, answers);
			if (r.operator === 'show_if' && !matched) return false;
			if (r.operator === 'hide_if' && matched) return false;
		}
		return true;
	});
}

export function newRuleId(): string {
	const chars = 'abcdefghijkmnpqrstuvwxyz23456789';
	let s = 'r_';
	for (let i = 0; i < 12; i++) s += chars[Math.floor(Math.random() * chars.length)];
	return s;
}
