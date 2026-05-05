<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

<div class="min-h-screen bg-muted/30">
	<header class="border-b bg-background">
		<div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
			<a href="/teacher" class="flex items-center gap-3" aria-label="SchoolRise teacher home">
				<img src="/logo-lockup.svg" alt="SchoolRise" class="h-8 w-auto" />
				{#if data.teacherRole === 'admin-global'}
					<Badge variant="secondary" class="ml-1">Admin · acting as teacher</Badge>
				{:else}
					<Badge class="ml-1">Teacher</Badge>
				{/if}
			</a>
			<div class="flex items-center gap-3 text-sm">
				<span class="text-muted-foreground">{data.user.email}</span>
				{#if data.teacherRole === 'admin-global'}
					<a href="/admin/dashboard" class="text-xs text-[#6439B5] hover:underline">← back to admin</a>
				{/if}
				<form method="POST" action="/logout">
					<Button type="submit" variant="ghost" size="sm">Sign out</Button>
				</form>
			</div>
		</div>
	</header>

	<main class="mx-auto max-w-5xl px-4 py-6">
		{@render children()}
	</main>
</div>
