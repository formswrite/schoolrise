<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import ScopePicker from '$lib/components/scope-picker.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

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
	<header class="space-y-3">
		<div>
			<h1 class="text-2xl font-semibold">Progression dashboard</h1>
			{#if data.scope}
				<p class="mt-1 text-sm text-muted-foreground">
					Showing <strong>{data.scope.label}</strong> ({data.scope.level})
				</p>
			{/if}
		</div>
		{#if data.scopeOptions && data.scopeOptions.length > 0 && data.scopeNodeId}
			<ScopePicker
				options={data.scopeOptions}
				current={data.scopeNodeId}
				extraParams={{ campaign: data.campaignId ?? undefined, period: data.periodId ?? undefined }}
			/>
		{/if}
	</header>

	{#if !data.scopeNodeId}
		<Alert>
			<AlertDescription>
				No scopes are configured yet. Set up your hierarchy first via <a
					href="/admin/institutions"
					class="font-medium underline">Institutions</a
				>.
			</AlertDescription>
		</Alert>
	{:else if data.campaigns.length === 0}
		<Alert>
			<AlertDescription>
				No campaigns at this scope yet. Create one via the <a href="/admin/forms" class="underline"
					>Forms</a
				> screen first.
			</AlertDescription>
		</Alert>
	{:else}
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Pick a campaign</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-wrap gap-2">
					{#each data.campaigns as c}
						{@const isActive = c.id === data.campaignId}
						<a
							href="/admin/dashboard?scope={data.scopeNodeId}&campaign={c.id}&period={c.period_id}"
							class="rounded-md border px-3 py-1.5 text-sm transition-colors {isActive
								? 'border-[#6439B5] bg-[#f0ecff] font-semibold text-[#6439B5]'
								: 'hover:bg-accent'}"
						>
							{c.title}
							<Badge variant="secondary" class="ml-2">{c.scale_code}</Badge>
							<span class="ml-1 text-xs text-muted-foreground">{c.status}</span>
						</a>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>

		{#if data.progression}
			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between">
					<div>
						<Card.Title class="text-base">Band distribution at this scope</Card.Title>
						<Card.Description>{data.progression.total_scored} students scored</Card.Description>
					</div>
					<form method="POST" use:enhance action="?/refresh">
						<Button type="submit" size="sm" variant="outline">Refresh snapshot</Button>
					</form>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if data.progression.total_scored === 0}
						<p class="text-sm text-muted-foreground">No scored responses yet.</p>
					{:else}
						<div class="flex h-8 overflow-hidden rounded-md border">
							{#each data.progression.bands as b}
								{#if b.percentage > 0}
									<div
										class="{bandColor[b.band_code] ??
											'bg-muted'} flex items-center justify-center text-xs font-semibold text-white"
										style="width: {b.percentage}%"
										title="{b.band_label}: {b.student_count} ({b.percentage}%)"
									>
										{#if b.percentage >= 8}{b.percentage}%{/if}
									</div>
								{/if}
							{/each}
						</div>
						<div class="grid gap-2 sm:grid-cols-5">
							{#each data.progression.bands as b}
								<div class="rounded-md border p-3">
									<div class="flex items-center gap-2">
										<span class="size-3 rounded-full {bandColor[b.band_code] ?? 'bg-muted'}"></span>
										<p class="text-sm font-semibold">{b.band_label}</p>
									</div>
									<p class="mt-1 text-2xl font-bold">{b.student_count}</p>
									<p class="text-xs text-muted-foreground">{b.percentage}%</p>
								</div>
							{/each}
						</div>
					{/if}
					{#if form?.error}<Alert variant="destructive"
							><AlertDescription>{form.error}</AlertDescription></Alert
						>{/if}
				</Card.Content>
			</Card.Root>
		{/if}

		{#if data.drilldown && data.drilldown.children.length > 0}
			<Card.Root>
				<Card.Header><Card.Title class="text-base">Children scopes</Card.Title></Card.Header>
				<Card.Content>
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Node</Table.Head>
								<Table.Head>Level</Table.Head>
								<Table.Head>Total</Table.Head>
								<Table.Head>Distribution</Table.Head>
								<Table.Head class="w-px"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each data.drilldown.children as child}
								<Table.Row>
									<Table.Cell class="font-medium">{child.label}</Table.Cell>
									<Table.Cell><Badge variant="secondary">{child.level}</Badge></Table.Cell>
									<Table.Cell>{child.total}</Table.Cell>
									<Table.Cell>
										<div class="flex h-3 w-48 overflow-hidden rounded">
											{#each child.bands as b}
												{#if b.percentage > 0}
													<div
														class={bandColor[b.band_code] ?? 'bg-muted'}
														style="width: {b.percentage}%"
														title="{b.band_label}: {b.percentage}%"
													></div>
												{/if}
											{/each}
										</div>
									</Table.Cell>
									<Table.Cell>
										<a
											href="/admin/dashboard?scope={child.node_id}&campaign={data.campaignId}&period={data.periodId}"
											class="text-xs text-[#6439B5] hover:underline">Drill in →</a
										>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>
