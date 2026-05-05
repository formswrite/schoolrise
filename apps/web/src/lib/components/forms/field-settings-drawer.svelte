<script lang="ts">
	import type { Question } from '$lib/forms/field-types';
	import {
		FIELD_LABELS,
		FIELD_TYPE_GROUPS,
		TYPES_WITH_OPTIONS,
		SCALE_TYPES,
		NON_SUBMITTABLE_TYPES,
		RENDERER_PENDING
	} from '$lib/forms/field-types';
	import type { LogicRule } from '$lib/forms/logic';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { X } from '@lucide/svelte';
	import { enhance } from '$app/forms';
	import LogicPanel from './logic-panel.svelte';

	type Props = {
		question: Question | null;
		onClose: () => void;
		allQuestions?: Question[];
		logicRules?: LogicRule[];
	};
	let { question, onClose, allQuestions = [], logicRules = [] }: Props = $props();

	function optsToText(q: Question | null): string {
		if (!q?.options) return '';
		return (q.options as Array<string | { label?: string; value?: string }>)
			.map((o) => (typeof o === 'string' ? o : (o.label ?? o.value ?? '')))
			.join('\n');
	}

	const optionsText = $derived(optsToText(question));
	const showOptions = $derived(
		question ? TYPES_WITH_OPTIONS.has(question.type) || question.type === 'ORDERING' : false
	);
	const showScale = $derived(question ? SCALE_TYPES.has(question.type) : false);
	const isLayout = $derived(question ? NON_SUBMITTABLE_TYPES.has(question.type) : false);
	const isPending = $derived(question ? RENDERER_PENDING.has(question.type) : false);

	const TEXT_TYPES = new Set([
		'SHORT_ANSWER',
		'PARAGRAPH',
		'EMAIL',
		'PHONE',
		'HOME_NUMBER',
		'ESSAY',
		'CODE_BLOCK'
	]);
	const NUMERIC_TYPES = new Set(['NUMBER', 'DECIMAL']);
	const GRADABLE_TYPES = new Set([
		'MULTIPLE_CHOICE',
		'CHECKBOX',
		'DROPDOWN',
		'RADIO',
		'YES_NO',
		'SHORT_ANSWER',
		'NUMBER',
		'DECIMAL',
		'ORDERING',
		'MATCHING',
		'FILL_IN_BLANK',
		'EQUATION',
		'ESSAY'
	]);

	const showTextValidation = $derived(question ? TEXT_TYPES.has(question.type) : false);
	const showNumericValidation = $derived(question ? NUMERIC_TYPES.has(question.type) : false);
	const showGrading = $derived(question ? GRADABLE_TYPES.has(question.type) : false);

	const validation = $derived((question?.validation ?? {}) as Record<string, unknown>);
	const grading = $derived((question?.grading ?? {}) as Record<string, unknown>);
	const extra = $derived((question?.extra ?? {}) as Record<string, unknown>);

	function asString(v: unknown, fallback = ''): string {
		return typeof v === 'string' ? v : typeof v === 'number' ? String(v) : fallback;
	}
	function asArray(v: unknown): string[] {
		return Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : [];
	}
</script>

{#if question}
	<aside
		class="hidden lg:flex lg:w-80 lg:shrink-0 lg:flex-col lg:overflow-y-auto lg:border-l lg:bg-background"
	>
		<header class="flex items-center justify-between border-b px-4 py-3">
			<div class="min-w-0">
				<h2 class="text-sm font-semibold">Question settings</h2>
				<p class="mt-0.5 truncate text-xs text-muted-foreground">{FIELD_LABELS[question.type]}</p>
			</div>
			<button
				type="button"
				onclick={onClose}
				class="rounded-md p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
				aria-label="Close settings"
			>
				<X class="size-4" />
			</button>
		</header>

		<form
			method="POST"
			action="?/updateQuestion"
			use:enhance
			class="flex-1 space-y-4 px-4 py-4 text-sm"
		>
			<input type="hidden" name="question_id" value={question.id} />

			{#if isPending}
				<div
					class="rounded-md border border-amber-300 bg-amber-50/60 px-3 py-2 text-xs text-amber-900"
				>
					<Badge
						variant="outline"
						class="mb-1 border-amber-300 bg-amber-100 text-[10px] text-amber-900"
					>
						Renderer pending
					</Badge>
					<p>Authoring is supported. The public renderer for this type ships in Phase 3.</p>
				</div>
			{/if}

			<div class="space-y-1.5">
				<Label for="title">Question title</Label>
				<Input id="title" name="title" value={question.title} required />
			</div>

			<div class="space-y-1.5">
				<Label for="description">Description (optional)</Label>
				<textarea
					id="description"
					name="description"
					rows="2"
					class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
					>{question.description ?? ''}</textarea
				>
			</div>

			<div class="space-y-1.5">
				<Label for="type">Type</Label>
				<select
					id="type"
					name="type"
					class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
				>
					{#each FIELD_TYPE_GROUPS as group}
						<optgroup label={group.label}>
							{#each group.types as t}
								<option value={t} selected={question.type === t}>
									{FIELD_LABELS[t]}{RENDERER_PENDING.has(t) ? ' (β)' : ''}
								</option>
							{/each}
						</optgroup>
					{/each}
				</select>
			</div>

			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1.5">
					<Label for="sort_order">Sort order</Label>
					<Input
						id="sort_order"
						name="sort_order"
						type="number"
						min="0"
						value={question.sort_order}
					/>
				</div>
				{#if !isLayout}
					<div class="flex items-end gap-2 pb-1.5">
						<Checkbox id="required" name="required" checked={question.required} />
						<Label for="required" class="cursor-pointer">Required</Label>
					</div>
				{/if}
			</div>

			{#if showOptions}
				<div class="space-y-1.5">
					<Label for="options">Options (one per line)</Label>
					<textarea
						id="options"
						name="options"
						rows="5"
						placeholder={'Option A\nOption B'}
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-xs shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
						>{optionsText}</textarea
					>
				</div>
			{/if}

			{#if showScale}
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-1.5">
						<Label for="scale_min">Scale min</Label>
						<Input id="scale_min" name="scale_min" type="number" value={question.scale_min ?? 1} />
					</div>
					<div class="space-y-1.5">
						<Label for="scale_max">Scale max</Label>
						<Input
							id="scale_max"
							name="scale_max"
							type="number"
							value={question.scale_max ?? (question.type === 'RATING' ? 5 : 10)}
						/>
					</div>
				</div>
			{/if}

			{#if question.type === 'EQUATION'}
				<div class="space-y-1.5">
					<Label for="extra_latex">LaTeX equation (KaTeX)</Label>
					<Input
						id="extra_latex"
						name="extra_latex"
						value={asString(extra.latex)}
						placeholder="x^2 + 2x + 1"
					/>
				</div>
			{/if}

			{#if question.type === 'CODE_BLOCK'}
				<div class="space-y-1.5">
					<Label for="extra_language">Language</Label>
					<Input
						id="extra_language"
						name="extra_language"
						value={asString(extra.language)}
						placeholder="javascript"
					/>
				</div>
			{/if}

			{#if question.type === 'FILL_IN_BLANK'}
				<div class="space-y-1.5">
					<Label for="extra_template">Template (use [[1]], [[2]] for blanks)</Label>
					<textarea
						id="extra_template"
						name="extra_template"
						rows="3"
						placeholder={'Le mot manquant est [[1]] et [[2]] aussi.'}
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
						>{asString(extra.template)}</textarea
					>
				</div>
			{/if}

			{#if question.type === 'ADDRESS'}
				<div class="space-y-1.5">
					<Label for="extra_fields">Sub-fields (one per line)</Label>
					<textarea
						id="extra_fields"
						name="extra_fields"
						rows="4"
						placeholder={'Quartier\nCommune\nPréfecture'}
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-xs shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
						>{asArray(extra.fields).join('\n')}</textarea
					>
				</div>
			{/if}

			{#if question.type === 'TABLE'}
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-1.5">
						<Label for="extra_rows">Rows (one per line)</Label>
						<textarea
							id="extra_rows"
							name="extra_rows"
							rows="4"
							placeholder={'CP1\nCP2\nCE1'}
							class="flex w-full rounded-md border border-input bg-transparent px-2 py-2 font-mono text-xs"
							>{asArray(extra.rows).join('\n')}</textarea
						>
					</div>
					<div class="space-y-1.5">
						<Label for="extra_columns">Columns (one per line)</Label>
						<textarea
							id="extra_columns"
							name="extra_columns"
							rows="4"
							placeholder={'Sept\nOct\nNov'}
							class="flex w-full rounded-md border border-input bg-transparent px-2 py-2 font-mono text-xs"
							>{asArray(extra.columns).join('\n')}</textarea
						>
					</div>
				</div>
			{/if}

			{#if question.type === 'MATCHING'}
				<div class="space-y-1.5">
					<Label for="extra_pairs"
						>Matching pairs (left → right, one per line, separated by →)</Label
					>
					<textarea
						id="extra_pairs"
						name="extra_pairs"
						rows="4"
						placeholder={'Lion → Mammifère\nAigle → Oiseau\nTortue → Reptile'}
						class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 font-mono text-xs shadow-xs focus:ring-1 focus:ring-ring focus:outline-none"
						>{((extra.pairs ?? []) as Array<{ left?: string; right?: string }>)
							.map((p) => `${p.left ?? ''} → ${p.right ?? ''}`)
							.join('\n')}</textarea
					>
				</div>
			{/if}

			{#if showTextValidation}
				<details class="rounded-md border bg-muted/10 px-3 py-2">
					<summary class="cursor-pointer text-xs font-semibold text-muted-foreground">
						Validation
					</summary>
					<div class="mt-2 grid grid-cols-2 gap-2">
						<div class="space-y-1">
							<Label for="val_min_length" class="text-[11px]">Min length</Label>
							<Input
								id="val_min_length"
								name="val_min_length"
								type="number"
								min="0"
								value={asString(validation.min_length)}
							/>
						</div>
						<div class="space-y-1">
							<Label for="val_max_length" class="text-[11px]">Max length</Label>
							<Input
								id="val_max_length"
								name="val_max_length"
								type="number"
								min="0"
								value={asString(validation.max_length)}
							/>
						</div>
						<div class="col-span-2 space-y-1">
							<Label for="val_pattern" class="text-[11px]">Regex pattern</Label>
							<Input
								id="val_pattern"
								name="val_pattern"
								value={asString(validation.pattern)}
								placeholder="^[A-Za-z]+$"
								class="font-mono text-xs"
							/>
						</div>
					</div>
				</details>
			{/if}

			{#if showNumericValidation}
				<details class="rounded-md border bg-muted/10 px-3 py-2">
					<summary class="cursor-pointer text-xs font-semibold text-muted-foreground">
						Validation
					</summary>
					<div class="mt-2 grid grid-cols-2 gap-2">
						<div class="space-y-1">
							<Label for="val_min" class="text-[11px]">Min value</Label>
							<Input id="val_min" name="val_min" type="number" value={asString(validation.min)} />
						</div>
						<div class="space-y-1">
							<Label for="val_max" class="text-[11px]">Max value</Label>
							<Input id="val_max" name="val_max" type="number" value={asString(validation.max)} />
						</div>
					</div>
				</details>
			{/if}

			{#if showGrading}
				<details class="rounded-md border bg-muted/10 px-3 py-2">
					<summary class="cursor-pointer text-xs font-semibold text-muted-foreground">
						Grading
					</summary>
					<div class="mt-2 space-y-2">
						<div class="space-y-1">
							<Label for="grading_points_max" class="text-[11px]">Points (max)</Label>
							<Input
								id="grading_points_max"
								name="grading_points_max"
								type="number"
								min="0"
								value={asString(grading.points_max, '1')}
							/>
						</div>
						{#if question.type === 'SHORT_ANSWER' || question.type === 'NUMBER' || question.type === 'DECIMAL' || question.type === 'EQUATION'}
							<div class="space-y-1">
								<Label for="grading_correct_value" class="text-[11px]">Correct answer</Label>
								<Input
									id="grading_correct_value"
									name="grading_correct_value"
									value={asString(grading.correct_value)}
								/>
							</div>
						{/if}
						{#if TYPES_WITH_OPTIONS.has(question.type) || question.type === 'YES_NO'}
							<div class="space-y-1">
								<Label for="grading_correct_value" class="text-[11px]">Correct option (value)</Label
								>
								<Input
									id="grading_correct_value"
									name="grading_correct_value"
									value={asString(grading.correct_value)}
								/>
							</div>
						{/if}
						{#if question.type === 'FILL_IN_BLANK'}
							<div class="space-y-1">
								<Label for="grading_answers" class="text-[11px]"
									>Answers (one per line, in order)</Label
								>
								<textarea
									id="grading_answers"
									name="grading_answers"
									rows="3"
									class="flex w-full rounded-md border border-input bg-transparent px-2 py-2 font-mono text-xs"
									>{asArray(grading.answers).join('\n')}</textarea
								>
							</div>
						{/if}
						{#if question.type === 'ESSAY'}
							<div class="space-y-1">
								<Label for="grading_rubric_url" class="text-[11px]">Rubric URL</Label>
								<Input
									id="grading_rubric_url"
									name="grading_rubric_url"
									value={asString(grading.rubric_url)}
									placeholder="https://…"
								/>
							</div>
							<p class="text-[10px] text-muted-foreground">
								Essays are not auto-graded; rubric is for human reviewers.
							</p>
						{/if}
					</div>
				</details>
			{/if}

			<div class="flex justify-end pt-2">
				<Button type="submit" size="sm">Save</Button>
			</div>
		</form>

		{#if !isLayout}
			<div class="px-4 pb-6">
				<LogicPanel target={question} {allQuestions} rules={logicRules} />
			</div>
		{/if}
	</aside>
{/if}
