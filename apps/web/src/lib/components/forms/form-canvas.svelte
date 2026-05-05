<script lang="ts">
	import type { Question } from '$lib/forms/field-types';
	import { dndzone, type DndEvent } from 'svelte-dnd-action';
	import * as Card from '$lib/components/ui/card';
	import FieldRow from './field-row.svelte';

	type Props = {
		questions: Question[];
		selectedId: number | null;
		onSelect: (id: number | null) => void;
		onReorder: (orderedIds: number[]) => void;
	};
	let { questions, selectedId, onSelect, onReorder }: Props = $props();

	let items = $state<Question[]>([]);
	$effect(() => {
		items = [...questions];
	});

	function handleConsider(e: CustomEvent<DndEvent<Question>>) {
		items = e.detail.items;
	}
	function handleFinalize(e: CustomEvent<DndEvent<Question>>) {
		items = e.detail.items;
		const ids = items.map((q) => q.id).filter((id): id is number => typeof id === 'number');
		onReorder(ids);
	}
</script>

<div class="flex-1 overflow-y-auto px-6 py-6">
	{#if items.length === 0}
		<Card.Root class="border-dashed">
			<Card.Content class="py-16 text-center text-sm text-muted-foreground">
				No questions yet. Pick a field type from the left to add one.
			</Card.Content>
		</Card.Root>
	{:else}
		<div
			class="space-y-2"
			use:dndzone={{
				items,
				flipDurationMs: 150,
				dragDisabled: false,
				dropTargetStyle: {}
			}}
			onconsider={handleConsider}
			onfinalize={handleFinalize}
		>
			{#each items as q, i (q.id ?? `tmp-${i}`)}
				<div>
					<FieldRow
						question={q}
						index={i}
						selected={selectedId === q.id}
						onSelect={() => onSelect(q.id ?? null)}
					/>
				</div>
			{/each}
		</div>
	{/if}
</div>
