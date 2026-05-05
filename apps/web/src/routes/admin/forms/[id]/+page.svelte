<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';

	import FieldPalette from '$lib/components/forms/field-palette.svelte';
	import FormCanvas from '$lib/components/forms/form-canvas.svelte';
	import FieldSettingsDrawer from '$lib/components/forms/field-settings-drawer.svelte';
	import type { Question } from '$lib/forms/field-types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let selectedId = $state<number | null>(null);
	const questions = $derived(data.questions as unknown as Question[]);
	const selectedQuestion = $derived(
		questions.find((q) => q.id === selectedId) ?? null
	);

	let reorderForm: HTMLFormElement | undefined = $state();
	let pendingOrder = $state<string>('');

	function submitReorder(orderedIds: number[]) {
		pendingOrder = orderedIds.join(',');
		queueMicrotask(() => reorderForm?.requestSubmit());
	}
</script>

<div class="flex h-[calc(100vh-3.5rem)] flex-col">
	<FormToaster {form} />

	<header class="flex shrink-0 items-start justify-between gap-4 border-b bg-background px-6 py-3">
		<div class="min-w-0">
			<a href="/admin/forms" class="text-xs text-muted-foreground hover:underline"
				>← Forms</a
			>
			<h1 class="mt-0.5 truncate text-xl font-semibold">{data.form.title}</h1>
			<p class="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
				<code>{data.form.public_id}</code>
				{#if data.form.status === 'published'}<Badge>Published</Badge>
				{:else if data.form.status === 'closed'}<Badge variant="destructive">Closed</Badge>
				{:else}<Badge variant="secondary">Draft</Badge>{/if}
				<span>· {questions.length} question{questions.length === 1 ? '' : 's'}</span>
			</p>
		</div>

		<div class="flex shrink-0 gap-2">
			<form method="POST" use:enhance action="?/publish">
				<Button type="submit" size="sm" disabled={questions.length === 0}>
					Publish version
				</Button>
			</form>
		</div>
	</header>

	{#if form?.success && form?.version}
		<div class="border-b bg-emerald-50 px-6 py-2">
			<Alert class="border-emerald-200 bg-transparent text-emerald-900">
				<AlertDescription>
					Published as version <strong>{form.version.version_num}</strong>. ID:
					<code>{form.version.id}</code>
				</AlertDescription>
			</Alert>
		</div>
	{/if}

	<div class="flex flex-1 overflow-hidden">
		<FieldPalette />

		<FormCanvas
			questions={questions}
			selectedId={selectedId}
			onSelect={(id) => (selectedId = id)}
			onReorder={submitReorder}
		/>

		<FieldSettingsDrawer
			question={selectedQuestion}
			allQuestions={questions}
			logicRules={data.logicRules}
			onClose={() => (selectedId = null)}
		/>
	</div>

	<form
		bind:this={reorderForm}
		method="POST"
		action="?/reorderQuestion"
		class="hidden"
		use:enhance={() => {
			return async ({ result, update }) => {
				await update();
				if (result.type === 'success') await invalidateAll();
			};
		}}
	>
		<input type="hidden" name="ordered_ids" value={pendingOrder} />
	</form>
</div>
