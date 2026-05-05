<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	let { data }: { data: PageData } = $props();

	const open = $derived(data.campaigns.filter((c) => c.status === 'open'));
	const closed = $derived(data.campaigns.filter((c) => c.status !== 'open'));
</script>

<div class="space-y-6">
	<header>
		<a href="/teacher" class="text-xs text-muted-foreground hover:underline">← Your classes</a>
		<h1 class="mt-1 text-2xl font-semibold">Class #{data.classID}</h1>
		<p class="mt-1 text-sm text-muted-foreground">Pick an open campaign to enter scores.</p>
	</header>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Open campaigns</Card.Title>
			<Card.Description>{open.length} open · {closed.length} closed</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if open.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">
					No open campaigns at this scope.
				</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Campaign</Table.Head>
							<Table.Head>Scale</Table.Head>
							<Table.Head>Progress</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each open as c}
							<Table.Row>
								<Table.Cell class="font-medium">{c.title}</Table.Cell>
								<Table.Cell><Badge variant="secondary">{c.scale_code}</Badge></Table.Cell>
								<Table.Cell class="text-muted-foreground">{c.scored} / {c.assigned}</Table.Cell>
								<Table.Cell>
									<Button href={`/teacher/classes/${data.classID}/campaigns/${c.id}`} size="sm">
										{c.scored > 0 ? 'Continue →' : 'Enter scores →'}
									</Button>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if closed.length > 0}
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Closed / draft campaigns</Card.Title>
			</Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Campaign</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Final</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each closed as c}
							<Table.Row>
								<Table.Cell class="font-medium">{c.title}</Table.Cell>
								<Table.Cell
									><Badge variant={c.status === 'closed' ? 'destructive' : 'secondary'}
										>{c.status}</Badge
									></Table.Cell
								>
								<Table.Cell class="text-muted-foreground">{c.scored} / {c.assigned}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
