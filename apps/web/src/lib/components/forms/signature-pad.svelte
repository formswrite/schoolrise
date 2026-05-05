<script lang="ts">
	import { Eraser, CheckCircle2 } from '@lucide/svelte';
	import { onMount } from 'svelte';

	type Props = { name: string; required?: boolean };
	let { name, required = false }: Props = $props();

	let canvas: HTMLCanvasElement | undefined = $state(undefined);
	let ctx: CanvasRenderingContext2D | null = null;
	let drawing = false;
	let hasInk = $state(false);
	let uploadedKey: string | null = $state(null);
	let status: 'idle' | 'uploading' | 'done' | 'error' = $state('idle');
	let errorMsg: string | null = $state(null);

	function getPoint(e: PointerEvent): { x: number; y: number } {
		const rect = canvas!.getBoundingClientRect();
		return {
			x: ((e.clientX - rect.left) / rect.width) * canvas!.width,
			y: ((e.clientY - rect.top) / rect.height) * canvas!.height
		};
	}

	function startDraw(e: PointerEvent) {
		if (!ctx) return;
		drawing = true;
		hasInk = true;
		uploadedKey = null;
		status = 'idle';
		const { x, y } = getPoint(e);
		ctx.beginPath();
		ctx.moveTo(x, y);
	}

	function moveDraw(e: PointerEvent) {
		if (!drawing || !ctx) return;
		const { x, y } = getPoint(e);
		ctx.lineTo(x, y);
		ctx.stroke();
	}

	function endDraw() {
		drawing = false;
	}

	function clear() {
		if (!ctx || !canvas) return;
		ctx.clearRect(0, 0, canvas.width, canvas.height);
		hasInk = false;
		uploadedKey = null;
		status = 'idle';
	}

	async function commit() {
		if (!canvas || !hasInk) return;
		status = 'uploading';
		errorMsg = null;
		canvas.toBlob(
			async (blob) => {
				if (!blob) {
					status = 'error';
					errorMsg = 'could not capture signature';
					return;
				}
				const file = new File([blob], 'signature.png', { type: 'image/png' });
				const fd = new FormData();
				fd.append('file', file);
				try {
					const res = await fetch('/api/uploads', { method: 'POST', body: fd });
					if (!res.ok) throw new Error(await res.text());
					const out = await res.json();
					uploadedKey = out.key;
					status = 'done';
				} catch (e) {
					errorMsg = (e as Error).message;
					status = 'error';
				}
			},
			'image/png',
			0.9
		);
	}

	onMount(() => {
		if (!canvas) return;
		ctx = canvas.getContext('2d');
		if (!ctx) return;
		ctx.lineWidth = 2;
		ctx.lineCap = 'round';
		ctx.strokeStyle = '#0b0d2a';
	});
</script>

<div class="space-y-2">
	<canvas
		bind:this={canvas}
		width="600"
		height="160"
		class="block w-full cursor-crosshair touch-none rounded-md border-2 border-dashed border-input bg-white"
		onpointerdown={startDraw}
		onpointermove={moveDraw}
		onpointerup={endDraw}
		onpointerleave={endDraw}
	></canvas>
	<div class="flex items-center justify-between gap-2 text-xs">
		<button
			type="button"
			onclick={clear}
			class="inline-flex items-center gap-1 rounded-md border px-2 py-1 text-muted-foreground hover:bg-accent"
		>
			<Eraser class="size-3" /> Clear
		</button>
		{#if status === 'done'}
			<span class="inline-flex items-center gap-1 text-emerald-700">
				<CheckCircle2 class="size-3" /> Signature saved
			</span>
		{:else if status === 'uploading'}
			<span class="text-muted-foreground">Uploading…</span>
		{:else if status === 'error'}
			<span class="text-destructive">Error: {errorMsg}</span>
		{:else}
			<button
				type="button"
				onclick={commit}
				disabled={!hasInk}
				class="rounded-md border bg-[#6439B5] px-2 py-1 text-white disabled:opacity-50"
			>
				Save signature
			</button>
		{/if}
	</div>
	{#if required}
		<input type="hidden" {name} value={uploadedKey ?? ''} required />
	{:else}
		<input type="hidden" {name} value={uploadedKey ?? ''} />
	{/if}
</div>
