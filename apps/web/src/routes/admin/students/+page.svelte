<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import ScopePicker from '$lib/components/scope-picker.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="space-y-6">
	<header class="space-y-3">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold">Students</h1>
				{#if data.institution}
					<p class="mt-1 text-sm text-muted-foreground">
						at <a
							href="/admin/institutions?parent={data.institution.id}"
							class="font-medium text-foreground hover:underline">{data.institution.label}</a
						>
					</p>
				{/if}
			</div>
			{#if data.institutionId}
				<Button href="/admin/students/new?institution={data.institutionId}">+ New student</Button>
			{/if}
		</div>
		{#if data.institutionOptions && data.institutionOptions.length > 0 && data.institutionId}
			<ScopePicker
				options={data.institutionOptions}
				current={data.institutionId}
				paramName="institution"
			/>
		{/if}
	</header>

	{#if !data.institutionId}
		<Alert>
			<AlertDescription>
				No institutions exist yet. Create one first via <a
					href="/admin/institutions"
					class="font-medium underline">Institutions</a
				>.
			</AlertDescription>
		</Alert>
	{:else if data.students.length === 0}
		<Card.Root>
			<Card.Content class="py-12 text-center text-sm text-muted-foreground">
				No students at this institution yet.
			</Card.Content>
		</Card.Root>
	{:else}
		<Card.Root>
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>Code</Table.Head>
						<Table.Head>Enrolled</Table.Head>
						<Table.Head>Email</Table.Head>
						<Table.Head class="w-px"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.students as student}
						<Table.Row>
							<Table.Cell class="font-medium">{student.person.fullName}</Table.Cell>
							<Table.Cell>{student.studentCode || '—'}</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{student.enrollmentDate
									? new Date(student.enrollmentDate).toLocaleDateString()
									: '—'}
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">{student.person.email || '—'}</Table.Cell>
							<Table.Cell>
								<form method="POST" action="?/delete">
									<input type="hidden" name="id" value={student.id} />
									<Button
										type="submit"
										variant="ghost"
										size="sm"
										class="h-7 px-2 text-xs text-destructive hover:text-destructive"
										onclick={(e) => {
											if (!confirm('Delete this student?')) e.preventDefault();
										}}
									>
										Delete
									</Button>
								</form>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</Card.Root>

		{#if form?.error}
			<Alert variant="destructive">
				<AlertDescription>{form.error}</AlertDescription>
			</Alert>
		{/if}
	{/if}
</div>
