<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';
	import AdminSidebar from '$lib/components/admin-sidebar.svelte';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
	const showTeacherLink = $derived(data.roleSummary.isGlobalAdmin || data.roleSummary.isTeacher);
	const roleLabel = $derived(
		data.roleSummary.isGlobalAdmin ? 'Admin'
			: data.roleSummary.isInspector ? 'Inspector'
			: data.roleSummary.isTeacher ? 'Teacher' : 'User'
	);
</script>

<div class="min-h-screen bg-muted/30">
	<AdminSidebar
		items={data.navItems}
		userEmail={data.user.email}
		{roleLabel}
		{showTeacherLink}
	/>

	<main class="md:pl-60">
		<div class="mx-auto max-w-6xl px-6 py-8">
			{@render children()}
		</div>
	</main>
</div>
