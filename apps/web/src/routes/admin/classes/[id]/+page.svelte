<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import FormToaster from '$lib/components/form-toaster.svelte';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	function staffRoleFor(staffID: number): string {
		return data.staffRoles.find((m) => m.staff_id === staffID)?.role ?? 'teacher';
	}
</script>

<div class="space-y-6">
	<FormToaster {form} />

	<header>
		<a href="/admin/classes?institution={data.cls.institution_id}" class="text-xs text-muted-foreground hover:underline">← Classes</a>
		<h1 class="mt-1 text-2xl font-semibold">{data.cls.code}</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			{data.cls.label}
			{#if data.institution}
				· at <a href="/admin/institutions?parent={data.institution.id}" class="font-medium text-foreground hover:underline">{data.institution.label}</a>
			{/if}
			{#if data.niveau}
				· <Badge variant="secondary" class="ml-1">{data.niveau.label}</Badge>
			{/if}
			{#if data.period}
				· {data.period.label}{data.period.is_current ? ' (current)' : ''}
			{/if}
			{#if data.cls.capacity}
				· capacity {data.cls.capacity}
			{/if}
		</p>
	</header>

	<div class="grid gap-4 sm:grid-cols-3">
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Students</Card.Description></Card.Header>
			<Card.Content><p class="text-3xl font-bold">{data.inClass.length}</p></Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Teachers / staff</Card.Description></Card.Header>
			<Card.Content><p class="text-3xl font-bold text-[#6439B5]">{data.onStaff.length}</p></Card.Content>
		</Card.Root>
		<Card.Root>
			<Card.Header class="pb-2"><Card.Description>Capacity used</Card.Description></Card.Header>
			<Card.Content>
				<p class="text-3xl font-bold">
					{data.cls.capacity ? `${data.inClass.length} / ${data.cls.capacity}` : '—'}
				</p>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Staff on this class</Card.Title>
			<Card.Description>Teachers can enter scores via <a href="/teacher" class="underline">/teacher</a> for any class they're on.</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			{#if data.onStaff.length === 0}
				<p class="py-4 text-center text-sm text-muted-foreground">No staff on this class yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Name</Table.Head>
							<Table.Head>Position</Table.Head>
							<Table.Head>Role on class</Table.Head>
							<Table.Head>Email</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.onStaff as s}
							{@const role = staffRoleFor(s.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{s.person?.fullName ?? `Staff #${s.id}`}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{s.position || '—'}</Table.Cell>
								<Table.Cell><Badge>{role}</Badge></Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{s.person?.email || '—'}</Table.Cell>
								<Table.Cell>
									<form method="POST" action="?/removeStaff" use:enhance>
										<input type="hidden" name="staff_id" value={s.id} />
										<input type="hidden" name="role" value={role} />
										<Button
											type="submit"
											variant="ghost"
											size="sm"
											class="h-7 px-2 text-xs text-destructive hover:text-destructive"
											onclick={(e) => { if (!confirm('Remove from class?')) e.preventDefault(); }}
										>
											Remove
										</Button>
									</form>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}

			{#if data.eligibleStaff.length > 0}
				<form method="POST" action="?/addStaff" use:enhance class="rounded-md border bg-muted/30 p-3">
					<p class="mb-2 text-sm font-medium">Add staff to this class</p>
					<div class="flex items-center gap-2">
						<select
							name="staff_id"
							required
							class="flex h-9 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs"
						>
							<option value=""></option>
							{#each data.eligibleStaff as s}
								<option value={s.id}>{s.person?.fullName ?? `Staff #${s.id}`}{s.position ? ` · ${s.position}` : ''}</option>
							{/each}
						</select>
						<select
							name="role"
							class="flex h-9 w-32 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs"
						>
							<option value="teacher" selected>teacher</option>
							<option value="assistant">assistant</option>
							<option value="head">head</option>
						</select>
						<Button type="submit" size="sm">Add</Button>
					</div>
				</form>
			{:else if data.onStaff.length > 0}
				<p class="text-xs text-muted-foreground">
					All staff at this institution are already assigned. Add more via <a href="/admin/staff?scope={data.cls.institution_id}" class="underline">/admin/staff</a>.
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">Students in this class</Card.Title>
			<Card.Description>{data.inClass.length} students enrolled</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			{#if data.inClass.length === 0}
				<p class="py-4 text-center text-sm text-muted-foreground">No students yet. Add some below.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Name</Table.Head>
							<Table.Head>Code</Table.Head>
							<Table.Head>Email</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.inClass as s}
							<Table.Row>
								<Table.Cell class="font-medium">{s.person?.fullName ?? `Student #${s.id}`}</Table.Cell>
								<Table.Cell class="font-mono text-xs">{s.studentCode || '—'}</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{s.person?.email || '—'}</Table.Cell>
								<Table.Cell>
									<form method="POST" action="?/removeStudent" use:enhance>
										<input type="hidden" name="student_id" value={s.id} />
										<Button
											type="submit"
											variant="ghost"
											size="sm"
											class="h-7 px-2 text-xs text-destructive hover:text-destructive"
											onclick={(e) => { if (!confirm('Remove from class?')) e.preventDefault(); }}
										>
											Remove
										</Button>
									</form>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}

			{#if data.eligible.length > 0}
				<form method="POST" action="?/addStudent" use:enhance class="rounded-md border bg-muted/30 p-3">
					<p class="mb-2 text-sm font-medium">Add a student to this class</p>
					<div class="flex items-center gap-2">
						<select
							name="student_id"
							required
							class="flex h-9 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs"
						>
							<option value=""></option>
							{#each data.eligible as s}
								<option value={s.id}>{s.person?.fullName ?? `Student #${s.id}`}{s.studentCode ? ` · ${s.studentCode}` : ''}</option>
							{/each}
						</select>
						<Button type="submit" size="sm">Add</Button>
					</div>
				</form>
			{:else if data.inClass.length > 0}
				<Alert>
					<AlertDescription>
						All students at this institution are in this class. Add more via <a href="/admin/students?institution={data.cls.institution_id}" class="underline">/admin/students</a>.
					</AlertDescription>
				</Alert>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
