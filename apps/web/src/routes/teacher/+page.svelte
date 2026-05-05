<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { data }: { data: PageData } = $props();
</script>

<div class="space-y-6">
	<header>
		<h1 class="text-2xl font-semibold">Your classes</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			Pick a class to see its open assessment campaigns and enter scores.
		</p>
	</header>

	{#if data.teacherRole === 'admin-global'}
		<Alert>
			<AlertDescription>
				You're an admin browsing the teacher view. Admins can enter scores for any class.
				Use <a href="/admin/dashboard" class="font-medium underline">/admin/dashboard</a> for read-only analytics.
			</AlertDescription>
		</Alert>
	{/if}

	{#if data.classes.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				{#if data.teacherRole === 'admin-global'}
					No classes are linked to your staff record. Use <a href="/admin/classes" class="underline">Admin → Classes</a> to add yourself, or impersonate via the API.
				{:else}
					No classes are assigned to you yet. Ask an administrator to add you to a class via <code>/admin/classes</code>.
				{/if}
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.classes as c}
				<a href="/teacher/classes/{c.id}" class="block">
					<Card.Root class="transition hover:shadow-md hover:border-[#6439B5]">
						<Card.Header class="pb-2">
							<div class="flex items-baseline justify-between">
								<Card.Title class="text-base">{c.code}</Card.Title>
								<Badge variant="secondary">Niveau {c.niveau_id}</Badge>
							</div>
							<Card.Description class="text-xs">{c.label}</Card.Description>
						</Card.Header>
						<Card.Content class="text-xs text-muted-foreground">
							Period #{c.period_id} · Institution #{c.institution_id}
						</Card.Content>
					</Card.Root>
				</a>
			{/each}
		</div>
	{/if}
</div>
