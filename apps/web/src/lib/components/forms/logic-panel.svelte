<script lang="ts">
	import type { Question } from '$lib/forms/field-types';
	import { TYPES_WITH_OPTIONS } from '$lib/forms/field-types';
	import type { LogicRule, ConditionOperator, RuleOperator } from '$lib/forms/logic';
	import { OPERATOR_LABELS } from '$lib/forms/logic';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Trash2 } from '@lucide/svelte';
	import { enhance } from '$app/forms';

	type Props = {
		target: Question;
		allQuestions: Question[];
		rules: LogicRule[];
	};
	let { target, allQuestions, rules }: Props = $props();

	const targetRules = $derived(
		rules.filter((r) => r.target_question_client_id === target.client_id)
	);
	const otherQuestions = $derived(
		allQuestions.filter(
			(q) => q.client_id !== target.client_id && q.type !== 'SECTION' && q.type !== 'STATEMENT'
		)
	);

	let showAdd = $state(false);
	let selectedSource = $state<string>('');
	let selectedOp = $state<ConditionOperator>('equals');
	let selectedRuleOp = $state<RuleOperator>('show_if');
	let selectedValue = $state<string>('');

	const sourceQuestion = $derived(
		otherQuestions.find((q) => q.client_id === selectedSource) ?? null
	);
	const sourceHasOptions = $derived(
		sourceQuestion ? TYPES_WITH_OPTIONS.has(sourceQuestion.type) : false
	);
	const sourceOptions = $derived(
		((sourceQuestion?.options ?? []) as Array<string | { label?: string; value?: string }>).map(
			(o, i) =>
				typeof o === 'string'
					? { label: o, value: o }
					: {
							label: o.label ?? o.value ?? `Option ${i + 1}`,
							value: o.value ?? o.label ?? `opt_${i}`
						}
		)
	);

	function labelFor(clientId: string): string {
		return allQuestions.find((q) => q.client_id === clientId)?.title ?? '(deleted)';
	}
</script>

<section class="space-y-3 border-t pt-4">
	<header class="flex items-center justify-between">
		<div>
			<h3 class="text-sm font-semibold">Visibility logic</h3>
			<p class="text-[11px] text-muted-foreground">
				Show or hide this question based on answers to others.
			</p>
		</div>
		{#if otherQuestions.length > 0}
			<button
				type="button"
				onclick={() => (showAdd = !showAdd)}
				class="rounded-md border px-2 py-1 text-xs hover:bg-accent"
			>
				{showAdd ? 'Cancel' : '+ Add rule'}
			</button>
		{/if}
	</header>

	{#if otherQuestions.length === 0}
		<p class="rounded-md border border-dashed bg-muted/30 px-3 py-3 text-xs text-muted-foreground">
			Add at least one other question first to define a rule.
		</p>
	{/if}

	{#each targetRules as rule (rule.id)}
		<div class="flex items-start gap-2 rounded-md border bg-muted/20 px-3 py-2 text-xs">
			<div class="flex-1 space-y-1">
				<div class="flex items-center gap-1.5">
					<Badge
						variant={rule.operator === 'show_if' ? 'default' : 'destructive'}
						class="text-[10px]"
					>
						{rule.operator === 'show_if' ? 'Show if' : 'Hide if'}
					</Badge>
				</div>
				{#each rule.conditions as c}
					<div class="text-foreground">
						<span class="font-medium">{labelFor(c.source_question_client_id)}</span>
						<span class="text-muted-foreground"> {OPERATOR_LABELS[c.op]} </span>
						<span class="font-mono">"{c.value}"</span>
					</div>
				{/each}
			</div>
			<form method="POST" action="?/deleteLogicRule" use:enhance>
				<input type="hidden" name="rule_id" value={rule.id} />
				<button
					type="submit"
					class="rounded-md p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
					aria-label="Delete rule"
					onclick={(e) => {
						if (!confirm('Delete this rule?')) e.preventDefault();
					}}
				>
					<Trash2 class="size-3.5" />
				</button>
			</form>
		</div>
	{/each}

	{#if showAdd && otherQuestions.length > 0}
		<form
			method="POST"
			action="?/createLogicRule"
			use:enhance
			class="space-y-2 rounded-md border bg-background p-3 text-xs"
		>
			<input type="hidden" name="target_client_id" value={target.client_id} />

			<div class="space-y-1">
				<Label for="rule_operator" class="text-[11px]">Behavior</Label>
				<select
					id="rule_operator"
					name="rule_operator"
					bind:value={selectedRuleOp}
					class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-xs"
				>
					<option value="show_if">Show this question if…</option>
					<option value="hide_if">Hide this question if…</option>
				</select>
			</div>

			<div class="space-y-1">
				<Label for="source_client_id" class="text-[11px]">Source question</Label>
				<select
					id="source_client_id"
					name="source_client_id"
					bind:value={selectedSource}
					required
					class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-xs"
				>
					<option value="">— select —</option>
					{#each otherQuestions as q}
						<option value={q.client_id}>{q.title || '(untitled)'}</option>
					{/each}
				</select>
			</div>

			<div class="grid grid-cols-2 gap-2">
				<div class="space-y-1">
					<Label for="cond_op" class="text-[11px]">Operator</Label>
					<select
						id="cond_op"
						name="cond_op"
						bind:value={selectedOp}
						class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-xs"
					>
						<option value="equals">equals</option>
						<option value="not_equals">does not equal</option>
						<option value="contains">contains</option>
						<option value="gt">greater than</option>
						<option value="lt">less than</option>
						<option value="gte">at least</option>
						<option value="lte">at most</option>
					</select>
				</div>
				<div class="space-y-1">
					<Label for="cond_value" class="text-[11px]">Value</Label>
					{#if sourceHasOptions}
						<select
							id="cond_value"
							name="cond_value"
							bind:value={selectedValue}
							required
							class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-xs"
						>
							<option value="">— select —</option>
							{#each sourceOptions as opt}
								<option value={opt.value}>{opt.label}</option>
							{/each}
						</select>
					{:else if sourceQuestion?.type === 'YES_NO'}
						<select
							id="cond_value"
							name="cond_value"
							bind:value={selectedValue}
							required
							class="flex h-8 w-full rounded-md border border-input bg-transparent px-2 text-xs"
						>
							<option value="">—</option>
							<option value="yes">yes</option>
							<option value="no">no</option>
						</select>
					{:else}
						<Input
							id="cond_value"
							name="cond_value"
							bind:value={selectedValue}
							placeholder="value"
							class="h-8 text-xs"
							required
						/>
					{/if}
				</div>
			</div>

			<div class="flex justify-end pt-1">
				<Button type="submit" size="sm">Save rule</Button>
			</div>
		</form>
	{/if}
</section>
