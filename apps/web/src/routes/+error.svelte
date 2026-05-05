<script lang="ts">
	import { page } from '$app/stores';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
</script>

<div class="mx-auto max-w-xl py-16">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">
				{$page.status === 403 ? '403 — Access denied' : `${$page.status} — Something went wrong`}
			</Card.Title>
			<Card.Description>
				{$page.error?.message ?? 'You do not have access to this section.'}
			</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-3 text-sm text-muted-foreground">
			{#if $page.status === 403}
				<p>
					Your account does not include this area. If you think this is a mistake, ask your
					administrator to update your role assignment.
				</p>
			{/if}
			<div class="flex gap-2 pt-2">
				<Button href="/admin/dashboard" size="sm">Go to dashboard</Button>
				<Button href="/login" variant="ghost" size="sm">Switch user</Button>
			</div>
		</Card.Content>
	</Card.Root>
</div>
