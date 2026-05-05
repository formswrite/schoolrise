<script lang="ts">
	import type { Question } from '$lib/forms/field-types';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';

	type Props = { question: Question };
	let { question }: Props = $props();

	const opts = $derived(
		((question.options ?? []) as Array<string | { label?: string; value?: string }>).map((o, i) =>
			typeof o === 'string'
				? { label: o, value: o }
				: {
						label: o.label ?? o.value ?? `Option ${i + 1}`,
						value: o.value ?? o.label ?? `opt_${i}`
					}
		)
	);

	const extra = $derived((question.extra ?? {}) as Record<string, unknown>);

	function asStringArray(v: unknown, fallback: string[]): string[] {
		if (!Array.isArray(v)) return fallback;
		return v.filter((x): x is string => typeof x === 'string');
	}
</script>

<div class="pointer-events-none text-sm select-none">
	{#if question.type === 'SECTION'}
		<h2 class="mt-1 border-b border-border pb-1 text-base font-semibold text-foreground">
			{question.title || 'New section'}
		</h2>
	{:else if question.type === 'STATEMENT'}
		<p class="rounded-md bg-[#ffc8eb]/30 p-2 text-xs text-foreground">{question.title}</p>
	{:else if question.type === 'PARAGRAPH'}
		<textarea
			rows="2"
			disabled
			placeholder="Long answer…"
			class="w-full resize-none rounded-md border border-input bg-muted/30 px-2 py-1 text-xs"
		></textarea>
	{:else if question.type === 'ESSAY'}
		<textarea
			rows="4"
			disabled
			placeholder="Essay response…"
			class="w-full resize-none rounded-md border border-input bg-muted/30 px-2 py-1 text-xs"
		></textarea>
	{:else if question.type === 'NUMBER' || question.type === 'DECIMAL'}
		<Input type="number" disabled placeholder="0" class="h-8 text-xs" />
	{:else if question.type === 'EMAIL'}
		<Input type="email" disabled placeholder="name@example.com" class="h-8 text-xs" />
	{:else if question.type === 'PHONE' || question.type === 'HOME_NUMBER'}
		<Input type="tel" disabled placeholder="+224 …" class="h-8 text-xs" />
	{:else if question.type === 'DATE'}
		<Input type="date" disabled class="h-8 text-xs" />
	{:else if question.type === 'TIME'}
		<Input type="time" disabled class="h-8 text-xs" />
	{:else if question.type === 'LINEAR_SCALE' || question.type === 'RATING'}
		<div class="flex items-center gap-2 text-xs text-muted-foreground">
			<span>{question.scale_min ?? 1}</span>
			<input
				type="range"
				min={question.scale_min ?? 1}
				max={question.scale_max ?? (question.type === 'RATING' ? 5 : 10)}
				disabled
				class="flex-1 accent-[#6439B5]"
			/>
			<span>{question.scale_max ?? (question.type === 'RATING' ? 5 : 10)}</span>
		</div>
	{:else if question.type === 'YES_NO'}
		<div class="space-y-1 text-xs">
			<label class="flex items-center gap-2"><input type="radio" disabled /> Yes</label>
			<label class="flex items-center gap-2"><input type="radio" disabled /> No</label>
		</div>
	{:else if question.type === 'MULTIPLE_CHOICE' || question.type === 'RADIO'}
		<div class="space-y-1 text-xs">
			{#each opts as opt}
				<label class="flex items-center gap-2"><input type="radio" disabled /> {opt.label}</label>
			{/each}
			{#if opts.length === 0}<span class="text-muted-foreground">(no options yet)</span>{/if}
		</div>
	{:else if question.type === 'CHECKBOX'}
		<div class="space-y-1 text-xs">
			{#each opts as opt}
				<label class="flex items-center gap-2"><input type="checkbox" disabled /> {opt.label}</label
				>
			{/each}
			{#if opts.length === 0}<span class="text-muted-foreground">(no options yet)</span>{/if}
		</div>
	{:else if question.type === 'DROPDOWN'}
		<select disabled class="h-8 w-full rounded-md border border-input bg-muted/30 px-2 text-xs">
			<option>{opts[0]?.label ?? '— select —'}</option>
		</select>
	{:else if question.type === 'COUNTRY_REGION'}
		<div class="grid grid-cols-2 gap-2 text-xs">
			<select disabled class="h-8 rounded-md border border-input bg-muted/30 px-2"
				><option>Guinée</option></select
			>
			<select disabled class="h-8 rounded-md border border-input bg-muted/30 px-2"
				><option>— région —</option></select
			>
		</div>
	{:else if question.type === 'ADDRESS'}
		{@const fields = asStringArray(extra.fields, ['Quartier', 'Commune', 'Préfecture'])}
		<div class="space-y-1 text-xs">
			{#each fields as f}
				<input
					disabled
					placeholder={f}
					class="h-8 w-full rounded-md border border-input bg-muted/30 px-2"
				/>
			{/each}
		</div>
	{:else if question.type === 'TABLE'}
		{@const rows = asStringArray(extra.rows, ['CP1', 'CP2', 'CE1'])}
		{@const cols = asStringArray(extra.columns, ['Sept', 'Oct', 'Nov'])}
		<table class="w-full border text-xs">
			<thead>
				<tr>
					<th class="border bg-muted/30 px-1 py-0.5"></th>
					{#each cols as c}
						<th class="border bg-muted/30 px-1 py-0.5 font-medium">{c}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rows as r}
					<tr>
						<td class="border bg-muted/30 px-1 py-0.5 font-medium">{r}</td>
						{#each cols as _}
							<td class="border px-1 py-0.5"><div class="h-3 w-full rounded bg-muted/40"></div></td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	{:else if question.type === 'ORDERING'}
		<div class="space-y-1 text-xs">
			{#each opts.length > 0 ? opts : [{ label: 'Item A' }, { label: 'Item B' }, { label: 'Item C' }] as opt, i}
				<div class="flex items-center gap-2 rounded-md border bg-muted/20 px-2 py-1">
					<span class="text-muted-foreground">≡</span>
					<span class="text-muted-foreground">{i + 1}.</span>
					<span>{opt.label}</span>
				</div>
			{/each}
		</div>
	{:else if question.type === 'MATCHING'}
		{@const pairs = (extra.pairs as Array<{ left?: string; right?: string }>) ?? [
			{ left: 'Lion', right: 'Mammifère' },
			{ left: 'Aigle', right: 'Oiseau' },
			{ left: 'Tortue', right: 'Reptile' }
		]}
		<div class="grid grid-cols-2 gap-1 text-xs">
			<div class="space-y-1">
				{#each pairs as p}
					<div class="rounded-md border bg-muted/20 px-2 py-1">{p.left}</div>
				{/each}
			</div>
			<div class="space-y-1">
				{#each pairs as p}
					<select
						disabled
						class="h-7 w-full rounded-md border border-input bg-muted/30 px-2 text-xs"
					>
						<option>{p.right}</option>
					</select>
				{/each}
			</div>
		</div>
	{:else if question.type === 'FILL_IN_BLANK'}
		{@const text =
			(typeof extra.template === 'string' ? extra.template : question.title) ||
			'Le mot manquant est [[1]] et aussi [[2]].'}
		<p class="text-xs text-foreground">
			{#each text.split(/\[\[\d+\]\]/) as part, i}
				<span>{part}</span>
				{#if i < text.split(/\[\[\d+\]\]/).length - 1}
					<input
						disabled
						class="mx-1 inline-block w-20 border-b-2 border-foreground/40 bg-transparent text-center"
					/>
				{/if}
			{/each}
		</p>
	{:else if question.type === 'EQUATION'}
		<div class="space-y-1 text-xs">
			<div class="rounded-md border bg-muted/30 px-2 py-1 font-mono text-foreground">
				{(typeof extra.latex === 'string' ? extra.latex : '') || 'x^2 + 2x + 1 = ?'}
			</div>
			<input
				disabled
				placeholder="Réponse"
				class="h-7 w-full rounded-md border border-input bg-muted/30 px-2"
			/>
		</div>
	{:else if question.type === 'CODE_BLOCK'}
		<textarea
			rows="3"
			disabled
			placeholder={typeof extra.language === 'string' ? `// ${extra.language}` : '// code'}
			class="w-full resize-none rounded-md border border-input bg-muted/30 px-2 py-1 font-mono text-[11px]"
		></textarea>
	{:else if question.type === 'FILE_UPLOAD' || question.type === 'ATTACHMENT'}
		<div
			class="rounded-md border border-dashed bg-muted/20 px-2 py-3 text-center text-[11px] text-muted-foreground"
		>
			📎 Click to upload a file
		</div>
	{:else if question.type === 'IMAGE'}
		<div
			class="rounded-md border border-dashed bg-muted/20 px-2 py-4 text-center text-[11px] text-muted-foreground"
		>
			🖼 Image upload
		</div>
	{:else if question.type === 'SIGNATURE'}
		<div
			class="rounded-md border bg-muted/20 px-2 py-3 text-center text-[11px] text-muted-foreground"
		>
			✍ Sign here
		</div>
	{:else if question.type === 'HOTSPOT'}
		<div
			class="rounded-md border bg-muted/20 px-2 py-4 text-center text-[11px] text-muted-foreground"
		>
			🎯 Click on the image to mark a hotspot
		</div>
	{:else}
		<Input disabled placeholder={`(${question.type})`} class="h-8 text-xs" />
	{/if}
</div>
