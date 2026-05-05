<script lang="ts">
	import type { ScopeOption } from '$lib/server/default-scope';

	type Props = {
		options: ScopeOption[];
		current: number;
		paramName?: string;
		extraParams?: Record<string, string | number | undefined | null>;
	};

	let { options, current, paramName = 'scope', extraParams = {} }: Props = $props();

	function buildHref(id: number): string {
		const params = new URLSearchParams();
		params.set(paramName, String(id));
		for (const [k, v] of Object.entries(extraParams)) {
			if (v !== undefined && v !== null && v !== '') params.set(k, String(v));
		}
		return `?${params.toString()}`;
	}

	function levelBadge(level: string): string {
		const colors: Record<string, string> = {
			country: 'bg-[#6439B5] text-white',
			region: 'bg-[#f0ecff] text-[#6439B5]',
			district: 'bg-blue-100 text-blue-700',
			institution: 'bg-emerald-100 text-emerald-700'
		};
		return colors[level] ?? 'bg-muted text-muted-foreground';
	}
</script>

<div class="flex flex-wrap items-center gap-1.5">
	<span class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Scope:</span>
	{#each options as opt}
		{@const isActive = opt.id === current}
		<a
			href={buildHref(opt.id)}
			class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs transition-colors {isActive
				? 'bg-[#6439B5] font-semibold text-white'
				: 'border border-border hover:bg-accent'}"
		>
			{#if opt.depth > 0}<span class="text-muted-foreground/60">↳</span>{/if}
			<span>{opt.label}</span>
			<span
				class="rounded px-1 py-0.5 text-[10px] {isActive ? 'bg-white/20' : levelBadge(opt.level)}"
			>
				{opt.level}
			</span>
		</a>
	{/each}
</div>
