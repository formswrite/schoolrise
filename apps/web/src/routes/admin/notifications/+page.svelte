<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	function statusVariant(s: string) {
		switch (s) {
			case 'sent':
				return 'default';
			case 'pending':
			case 'sending':
				return 'secondary';
			case 'failed':
			case 'dropped':
				return 'destructive';
			default:
				return 'secondary';
		}
	}
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header>
		<h1 class="text-2xl font-semibold">Notifications</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			Email outbox + provider status. Assignment emails go out automatically when the campaign is
			opened with <code>notify_by_email</code>.
		</p>
	</header>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Provider</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if data.provider}
				<dl class="grid grid-cols-2 gap-y-2 text-sm">
					<dt class="text-muted-foreground">Provider</dt>
					<dd>
						{#if data.provider.provider === 'resend'}
							<Badge>Resend (live)</Badge>
						{:else if data.provider.provider === 'log'}
							<Badge variant="secondary">Log (dev)</Badge>
							<span class="ml-2 text-xs text-muted-foreground"
								>No <code>RESEND_API_KEY</code> set; emails are logged to stdout.</span
							>
						{:else}
							<Badge variant="destructive">{data.provider.provider}</Badge>
						{/if}
					</dd>
					<dt class="text-muted-foreground">From</dt>
					<dd class="font-mono text-xs">{data.provider.from}</dd>
				</dl>
			{:else}
				<p class="text-sm text-muted-foreground">Provider info unavailable.</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Send a test email</Card.Title>
			<Card.Description>Verifies the provider is reachable end-to-end.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form method="POST" use:enhance action="?/test" class="flex items-end gap-3">
				<div class="flex-1 space-y-2">
					<Label for="to">Recipient</Label>
					<Input id="to" name="to" type="email" required placeholder="you@example.com" />
				</div>
				<Button type="submit" size="sm">Send test</Button>
			</form>
			{#if form?.success && form?.sentTo}
				<Alert class="mt-3"
					><AlertDescription>Test sent to <strong>{form.sentTo}</strong>.</AlertDescription></Alert
				>
			{:else if form?.error}
				<Alert variant="destructive" class="mt-3"
					><AlertDescription>{form.error}</AlertDescription></Alert
				>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between">
			<Card.Title class="text-base">Recent outbox</Card.Title>
			<form method="POST" use:enhance action="?/process">
				<Button type="submit" size="sm" variant="outline">Process pending</Button>
			</form>
		</Card.Header>
		<Card.Content>
			{#if data.emails.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">No emails yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>When</Table.Head>
							<Table.Head>Kind</Table.Head>
							<Table.Head>To</Table.Head>
							<Table.Head>Subject</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Tries</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.emails as e}
							<Table.Row>
								<Table.Cell class="text-xs text-muted-foreground"
									>{new Date(e.created_at).toLocaleString()}</Table.Cell
								>
								<Table.Cell><Badge variant="secondary">{e.kind}</Badge></Table.Cell>
								<Table.Cell class="text-xs">{e.to_email}</Table.Cell>
								<Table.Cell class="font-medium">{e.subject}</Table.Cell>
								<Table.Cell><Badge variant={statusVariant(e.status)}>{e.status}</Badge></Table.Cell>
								<Table.Cell class="text-muted-foreground">{e.attempts}</Table.Cell>
							</Table.Row>
							{#if e.last_error}
								<Table.Row>
									<Table.Cell></Table.Cell>
									<Table.Cell colspan={5} class="pt-0 text-xs text-destructive"
										>⚠ {e.last_error}</Table.Cell
									>
								</Table.Row>
							{/if}
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
