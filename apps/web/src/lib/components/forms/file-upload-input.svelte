<script lang="ts">
	import { Upload, CheckCircle2, X } from '@lucide/svelte';

	type Props = {
		name: string;
		required?: boolean;
		accept?: string;
		label?: string;
	};
	let { name, required = false, accept, label = 'Choose file' }: Props = $props();

	type Uploaded = {
		key: string;
		url: string;
		content_type: string;
		size: number;
		original_name: string;
	};

	let status: 'idle' | 'uploading' | 'done' | 'error' = $state('idle');
	let uploaded: Uploaded | null = $state(null);
	let errorMsg: string | null = $state(null);
	let progress = $state(0);
	let fileInput: HTMLInputElement | undefined = $state(undefined);

	async function handleFile(file: File) {
		errorMsg = null;
		status = 'uploading';
		progress = 0;

		const fd = new FormData();
		fd.append('file', file);

		try {
			const res = await fetch('/api/uploads', { method: 'POST', body: fd });
			if (!res.ok) {
				const text = await res.text();
				throw new Error(text || `upload failed (${res.status})`);
			}
			uploaded = await res.json();
			status = 'done';
			progress = 100;
		} catch (e) {
			errorMsg = (e as Error).message;
			status = 'error';
		}
	}

	function onChange(e: Event) {
		const t = e.target as HTMLInputElement;
		if (t.files && t.files[0]) handleFile(t.files[0]);
	}

	function reset() {
		uploaded = null;
		status = 'idle';
		errorMsg = null;
		if (fileInput) fileInput.value = '';
	}

	const isImage = $derived(
		(uploaded as Uploaded | null)?.content_type?.startsWith('image/') ?? false
	);
</script>

<div class="space-y-2">
	{#if status === 'idle'}
		<label
			class="flex w-full cursor-pointer items-center justify-center gap-2 rounded-md border-2 border-dashed border-input bg-muted/20 px-4 py-6 text-sm text-muted-foreground hover:border-[#6439B5] hover:bg-muted/40"
		>
			<Upload class="size-4" />
			<span>{label}</span>
			<input
				bind:this={fileInput}
				type="file"
				class="hidden"
				{accept}
				{required}
				onchange={onChange}
			/>
		</label>
	{:else if status === 'uploading'}
		<div class="rounded-md border bg-muted/20 px-4 py-3 text-sm">
			<p class="text-muted-foreground">Uploading…</p>
			<div class="mt-1 h-1 w-full overflow-hidden rounded-full bg-muted">
				<div class="h-1 bg-[#6439B5] transition-all" style="width:{progress}%"></div>
			</div>
		</div>
	{:else if status === 'done' && uploaded}
		<div class="flex items-start gap-3 rounded-md border bg-emerald-50/50 px-3 py-2 text-sm">
			{#if isImage}
				<img
					src={`/api/uploads/${uploaded.key}`}
					alt={uploaded.original_name}
					class="size-12 rounded border object-cover"
				/>
			{:else}
				<CheckCircle2 class="mt-0.5 size-5 shrink-0 text-emerald-600" />
			{/if}
			<div class="min-w-0 flex-1">
				<p class="truncate font-medium">{uploaded.original_name}</p>
				<p class="text-xs text-muted-foreground">
					{uploaded.content_type} · {(uploaded.size / 1024).toFixed(1)} KB
				</p>
			</div>
			<button
				type="button"
				onclick={reset}
				class="rounded-md p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
				aria-label="Replace file"
			>
				<X class="size-4" />
			</button>
		</div>
		<input type="hidden" {name} value={uploaded.key} />
	{:else if status === 'error'}
		<div
			class="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
		>
			Upload failed: {errorMsg}
			<button type="button" onclick={reset} class="ml-2 underline">Try again</button>
		</div>
	{/if}
</div>
