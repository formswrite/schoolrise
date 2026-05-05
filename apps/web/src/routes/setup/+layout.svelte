<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
	let pct = $derived(((data.stepIndex + 1) / data.stepCount) * 100);
</script>

<div class="min-h-screen bg-brand-hero">
	<header class="border-b-2 border-[#060419] bg-white/70 backdrop-blur">
		<div class="mx-auto flex max-w-3xl items-center justify-between px-4 py-5">
			<div class="flex items-center gap-3">
				<img src="/logo-lockup.svg" alt="SchoolRise" class="h-8 w-auto" />
				<span class="font-display text-[16px] font-bold uppercase tracking-[0.18em] text-[#6439B5]">installer</span>
			</div>
			<span class="text-sm font-semibold text-[#51545d]">
				Step {data.stepIndex + 1} of {data.stepCount}
			</span>
		</div>
		<div class="mx-auto max-w-3xl px-4 pb-4">
			<div class="h-2 w-full overflow-hidden rounded-full border border-[#0b0d2a] bg-white">
				<div class="h-full bg-[#6439B5] transition-all" style="width: {pct}%"></div>
			</div>
		</div>
	</header>

	<main class="mx-auto max-w-3xl px-4 py-10">
		{@render children()}
	</main>
</div>
