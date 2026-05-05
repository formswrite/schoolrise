<script lang="ts">
	import type { Question } from '$lib/forms/field-types';
	import { FIELD_LABELS, RENDERER_PENDING } from '$lib/forms/field-types';
	import { Badge } from '$lib/components/ui/badge';
	import { GripVertical, Trash2 } from '@lucide/svelte';
	import { enhance } from '$app/forms';
	import FieldPreview from './field-preview.svelte';

	type Props = {
		question: Question;
		index: number;
		selected: boolean;
		onSelect: () => void;
	};
	let { question, index, selected, onSelect }: Props = $props();

	const pending = $derived(RENDERER_PENDING.has(question.type));
</script>

<div
	role="button"
	tabindex="0"
	onclick={onSelect}
	onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && onSelect()}
	class="group flex gap-2 rounded-lg border bg-background px-3 py-3 transition-colors {selected
		? 'border-[#6439B5] ring-1 ring-[#6439B5]/40 shadow-sm'
		: 'border-border hover:border-[#6439B5]/30 hover:bg-accent/30'}"
>
	<div
		class="dnd-handle flex shrink-0 cursor-grab items-start pt-1 text-muted-foreground hover:text-foreground active:cursor-grabbing"
		aria-label="Drag to reorder"
	>
		<GripVertical class="size-4" />
	</div>

	<div class="min-w-0 flex-1 space-y-2">
		<div class="flex items-start justify-between gap-2">
			<div class="min-w-0 flex-1">
				<p class="text-sm font-medium text-foreground">
					<span class="text-muted-foreground">{index + 1}.</span>
					{question.title || '(untitled)'}
					{#if question.required}<span class="text-destructive">*</span>{/if}
				</p>
				<div class="mt-0.5 flex items-center gap-1.5">
					<Badge variant="secondary" class="text-[10px]">{FIELD_LABELS[question.type]}</Badge>
					{#if pending}
						<Badge variant="outline" class="border-amber-300 bg-amber-50 text-[10px] text-amber-900">
							renderer pending
						</Badge>
					{/if}
				</div>
			</div>

			<form method="POST" action="?/deleteQuestion" use:enhance class="shrink-0">
				<input type="hidden" name="question_id" value={question.id} />
				<button
					type="submit"
					onclick={(e) => {
						e.stopPropagation();
						if (!confirm('Delete this question?')) e.preventDefault();
					}}
					class="rounded-md p-1.5 text-muted-foreground opacity-0 transition-opacity hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100"
					aria-label="Delete question"
				>
					<Trash2 class="size-3.5" />
				</button>
			</form>
		</div>

		<FieldPreview {question} />
	</div>
</div>
