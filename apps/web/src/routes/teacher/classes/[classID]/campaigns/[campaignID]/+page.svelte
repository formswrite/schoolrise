<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	const initialScores = $derived(
		Object.fromEntries(data.roster.rows.map((r) => [r.student_id, r.raw_score ?? null]))
	);
	let scores = $state<Record<number, number | null>>({});
	$effect(() => {
		scores = { ...initialScores };
	});

	function normalize(v: unknown): string {
		if (v === null || v === undefined) return '';
		return String(v).trim();
	}

	const dirty = $derived(
		data.roster.rows.filter((r) => {
			const cur = normalize(scores[r.student_id]);
			const orig = normalize(initialScores[r.student_id]);
			return cur !== '' && cur !== orig;
		})
	);

	const totalEntered = $derived(
		Object.values(scores).filter((v) => normalize(v) !== '').length
	);

	const bandColor: Record<string, string> = {
		debutant: 'bg-[#dc2626]',
		lettres: 'bg-[#f59e0b]',
		mots: 'bg-[#eab308]',
		paragraphe: 'bg-[#22c55e]',
		histoire: 'bg-[#6439B5]',
		un_chiffre: 'bg-[#f59e0b]',
		deux_chiffres: 'bg-[#eab308]',
		soustraction: 'bg-[#22c55e]',
		division: 'bg-[#6439B5]'
	};

	function previewBand(score: number | null): string {
		if (score === null || score === undefined || isNaN(Number(score))) return '';
		const n = Number(score);
		if (n < 0 || n > 100) return '⚠ out of range';
		if (n <= 19) return 'Débutant';
		if (n <= 39) return 'Lettres / 1 chiffre';
		if (n <= 59) return 'Mots / 2 chiffres';
		if (n <= 79) return 'Paragraphe / Soustraction';
		return 'Histoire / Division';
	}
</script>

<div class="space-y-6">
	<header>
		<a href="/teacher/classes/{data.classID}" class="text-xs text-muted-foreground hover:underline">← Class</a>
		<h1 class="mt-1 text-2xl font-semibold">{data.roster.campaign.title}</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			Class #{data.classID} ·
			<Badge variant="secondary" class="ml-1">{data.roster.campaign.scale_code}</Badge>
			<Badge variant={data.roster.campaign.status === 'open' ? 'default' : 'destructive'} class="ml-1">{data.roster.campaign.status}</Badge>
		</p>
	</header>

	<FormToaster {form} />

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Roster ({data.roster.rows.length} students)</Card.Title>
			<Card.Description>
				Type each student's raw score (0–100). Leave blank to skip. The system computes the band on submit.
			</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if data.roster.rows.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">
					No students in this class roster yet.
				</p>
			{:else}
				<form method="POST" action="?/submit" use:enhance class="space-y-4">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head class="w-10">#</Table.Head>
								<Table.Head>Student</Table.Head>
								<Table.Head>Code</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="w-32">Score (0-100)</Table.Head>
								<Table.Head>Band</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each data.roster.rows as r, i}
								<Table.Row>
									<Table.Cell class="text-muted-foreground">{i + 1}</Table.Cell>
									<Table.Cell class="font-medium">{r.full_name || `Student #${r.student_id}`}</Table.Cell>
									<Table.Cell class="font-mono text-xs">{r.student_code || '—'}</Table.Cell>
									<Table.Cell>
										{#if r.has_score}
											<Badge>Saved</Badge>
										{:else}
											<Badge variant="secondary">New</Badge>
										{/if}
									</Table.Cell>
									<Table.Cell>
										<Input
											type="number"
											name="score_{r.student_id}"
											min="0"
											max="100"
											step="1"
											bind:value={scores[r.student_id]}
											class="h-9 w-24"
										/>
									</Table.Cell>
									<Table.Cell>
										{#if r.has_score && scores[r.student_id] === r.raw_score}
											<div class="flex items-center gap-2">
												<span class="size-3 rounded-full {bandColor[r.band_code ?? ''] ?? 'bg-muted'}"></span>
												<span class="text-xs">{r.band_code}</span>
											</div>
										{:else if scores[r.student_id] !== null && scores[r.student_id] !== undefined && String(scores[r.student_id]) !== ''}
											<span class="text-xs text-muted-foreground">→ {previewBand(scores[r.student_id])}</span>
										{:else}
											<span class="text-xs text-muted-foreground">—</span>
										{/if}
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>

					<div class="flex items-center justify-between border-t pt-4">
						<p class="text-sm text-muted-foreground">
							{totalEntered} entered · {dirty.length} unsaved
						</p>
						<div class="flex gap-2">
							<Button type="button" variant="ghost" size="sm" onclick={() => location.reload()}>Cancel</Button>
							<Button type="submit" size="sm">
								Submit batch
							</Button>
						</div>
					</div>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if form?.errors && form.errors.length > 0}
		<Card.Root>
			<Card.Header><Card.Title class="text-base text-destructive">Errors</Card.Title></Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Student ID</Table.Head>
							<Table.Head>Error</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each form.errors as e}
							<Table.Row>
								<Table.Cell class="font-mono text-xs">{e.student_id}</Table.Cell>
								<Table.Cell class="text-destructive">{e.message}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
