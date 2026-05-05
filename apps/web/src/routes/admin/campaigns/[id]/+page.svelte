<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	const assignedIDs = $derived(new Set(data.assignments.map((a) => a.student_id)));
	const submittedCount = $derived(data.assignments.filter((a) => a.submitted_at).length);
	const scoreByStudent = $derived(new Map(data.scores.map((s) => [s.student_id, s] as const)));

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
</script>

<div class="space-y-6">
	<FormToaster {form} />

	<header class="flex items-start justify-between">
		<div>
			<a
				href="/admin/campaigns?scope={data.campaign.scope_node_id}"
				class="text-xs text-muted-foreground hover:underline">← Campaigns</a
			>
			<h1 class="mt-1 text-2xl font-semibold">{data.campaign.title}</h1>
			<p class="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
				<Badge variant="secondary">{data.campaign.scale_code}</Badge>
				{#if data.campaign.status === 'open'}<Badge>Open</Badge>
				{:else if data.campaign.status === 'closed'}<Badge variant="destructive">Closed</Badge>
				{:else}<Badge variant="secondary">Draft</Badge>{/if}
				<span class="text-xs">Public ID: <code>{data.campaign.public_id}</code></span>
			</p>
		</div>
		<div class="flex gap-2">
			{#if data.campaign.status === 'draft'}
				<form method="POST" action="?/open" use:enhance>
					<Button type="submit" size="sm">Open campaign</Button>
				</form>
			{:else if data.campaign.status === 'open'}
				<form method="POST" action="?/close" use:enhance>
					<Button type="submit" size="sm" variant="destructive">Close campaign</Button>
				</form>
				<Button
					href="/admin/dashboard?scope={data.campaign.scope_node_id}&campaign={data.campaign
						.id}&period={data.campaign.period_id}"
					variant="outline"
					size="sm">View dashboard →</Button
				>
			{:else}
				<Button
					href="/admin/dashboard?scope={data.campaign.scope_node_id}&campaign={data.campaign
						.id}&period={data.campaign.period_id}"
					variant="outline"
					size="sm">View dashboard →</Button
				>
			{/if}
		</div>
	</header>

	<div class="grid gap-4 sm:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Assigned</Card.Description></Card.Header>
			<Card.Content><p class="text-3xl font-bold">{data.assignments.length}</p></Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Submitted</Card.Description></Card.Header>
			<Card.Content><p class="text-3xl font-bold text-[#6439B5]">{submittedCount}</p></Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Scored</Card.Description></Card.Header>
			<Card.Content><p class="text-3xl font-bold">{data.scores.length}</p></Card.Content>
		</Card.Root>
	</div>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Assign students</Card.Title>
			<Card.Description>
				{data.eligibleStudents.length} students at this institution.
				{#if data.campaign.status !== 'open'}<span class="text-destructive"
						>Campaign must be open to receive responses.</span
					>{/if}
			</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if data.eligibleStudents.length === 0}
				<Alert>
					<AlertDescription>
						No students at this institution. Add some via <a
							href="/admin/students?institution={data.campaign.scope_node_id}"
							class="underline">/admin/students</a
						>.
					</AlertDescription>
				</Alert>
			{:else}
				<form method="POST" action="?/assign" use:enhance class="space-y-3">
					<div class="flex items-center justify-between rounded-md border bg-muted/30 p-3">
						<div class="flex items-center gap-2">
							<Checkbox id="notify_by_email" name="notify_by_email" />
							<label for="notify_by_email" class="cursor-pointer text-sm"
								>Email each student a unique link (uses Resend)</label
							>
						</div>
						<Button type="submit" size="sm">Assign selected</Button>
					</div>
					<div class="max-h-72 overflow-y-auto rounded-md border">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head class="w-10"></Table.Head>
									<Table.Head>Name</Table.Head>
									<Table.Head>Code</Table.Head>
									<Table.Head>Email</Table.Head>
									<Table.Head>State</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each data.eligibleStudents as st}
									{@const isAssigned = assignedIDs.has(st.id)}
									{@const score = scoreByStudent.get(st.id)}
									<Table.Row>
										<Table.Cell>
											<input
												type="checkbox"
												name="student_id"
												value={st.id}
												disabled={isAssigned}
												class="size-4 rounded border-input"
											/>
										</Table.Cell>
										<Table.Cell class="font-medium"
											>{st.person?.fullName ?? `Student #${st.id}`}</Table.Cell
										>
										<Table.Cell class="font-mono text-xs">{st.studentCode || '—'}</Table.Cell>
										<Table.Cell class="text-xs text-muted-foreground"
											>{st.person?.email || '—'}</Table.Cell
										>
										<Table.Cell>
											{#if score}
												<div class="flex items-center gap-2">
													<span
														class="size-2 rounded-full {bandColor[score.band_code] ?? 'bg-muted'}"
													></span>
													<span class="text-xs">{score.band_code} ({score.raw_score})</span>
												</div>
											{:else if isAssigned}
												<Badge variant="secondary">Assigned</Badge>
											{:else}
												<Badge variant="outline">Eligible</Badge>
											{/if}
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</div>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if data.assignments.length > 0}
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Assignments ({data.assignments.length})</Card.Title>
				<Card.Description>Each student gets a unique signed access link.</Card.Description>
			</Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Student ID</Table.Head>
							<Table.Head>Token (copy & paste into URL)</Table.Head>
							<Table.Head>Submitted</Table.Head>
							<Table.Head>Score</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.assignments as a}
							{@const score = scoreByStudent.get(a.student_id)}
							<Table.Row>
								<Table.Cell class="font-mono text-xs">{a.student_id}</Table.Cell>
								<Table.Cell class="font-mono text-xs"><code>/r/{a.access_token}</code></Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{a.submitted_at ? new Date(a.submitted_at).toLocaleString() : '—'}
								</Table.Cell>
								<Table.Cell>
									{#if score}
										<div class="flex items-center gap-2">
											<span class="size-2 rounded-full {bandColor[score.band_code] ?? 'bg-muted'}"
											></span>
											<span class="text-xs">{score.band_code} · {score.raw_score}</span>
										</div>
									{:else}
										<span class="text-xs text-muted-foreground">—</span>
									{/if}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
