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
	import ScopePicker from '$lib/components/scope-picker.svelte';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let showEnroll = $state(false);
	let showTransfer = $state(false);

	function statusVariant(s: string) {
		switch (s) {
			case 'active': return 'default';
			case 'transferred': return 'secondary';
			case 'dropped': return 'destructive';
			default: return 'secondary';
		}
	}
</script>

<div class="space-y-6">
	<FormToaster {form} />
	<header class="space-y-3">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold">Enrollment</h1>
				{#if data.scope}
					<p class="mt-1 text-sm text-muted-foreground">
						at <a href="/admin/institutions?parent={data.scope.id}" class="font-medium text-foreground hover:underline">{data.scope.label}</a>
						{#if data.periodId}
							<span class="ml-2">— Period #{data.periodId}</span>
						{/if}
					</p>
				{/if}
			</div>
			{#if data.scopeNodeId && data.periodId}
				<div class="flex gap-2">
					<Button size="sm" variant="outline" onclick={() => (showTransfer = !showTransfer)}>
						Transfer student
					</Button>
					<Button size="sm" onclick={() => (showEnroll = !showEnroll)}>
						{showEnroll ? 'Cancel' : '+ Enroll student'}
					</Button>
				</div>
			{/if}
		</div>
		{#if data.scopeOptions && data.scopeOptions.length > 0 && data.scopeNodeId}
			<ScopePicker options={data.scopeOptions} current={data.scopeNodeId} extraParams={{ period: data.periodId ?? undefined }} />
		{/if}
	</header>

	{#if !data.scopeNodeId}
		<Alert>
			<AlertDescription>
				No scopes are configured yet. Set up your hierarchy first via <a href="/admin/institutions" class="font-medium underline">Institutions</a>.
			</AlertDescription>
		</Alert>
	{:else}
		{#if data.coverage}
			<div class="grid gap-4 sm:grid-cols-3 lg:grid-cols-5">
				<Card.Root>
					<Card.Header class="pb-2"><Card.Description>Total enrolled</Card.Description></Card.Header>
					<Card.Content><p class="text-3xl font-bold">{data.coverage.total_enrolled}</p></Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2"><Card.Description>Female</Card.Description></Card.Header>
					<Card.Content><p class="text-3xl font-bold text-[#6439B5]">{data.coverage.female}</p></Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2"><Card.Description>Male</Card.Description></Card.Header>
					<Card.Content><p class="text-3xl font-bold">{data.coverage.male}</p></Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2"><Card.Description>Other</Card.Description></Card.Header>
					<Card.Content><p class="text-3xl font-bold">{data.coverage.other}</p></Card.Content>
				</Card.Root>
				<Card.Root>
					<Card.Header class="pb-2"><Card.Description>Unknown</Card.Description></Card.Header>
					<Card.Content><p class="text-3xl font-bold text-muted-foreground">{data.coverage.unknown}</p></Card.Content>
				</Card.Root>
			</div>
		{/if}

		{#if showEnroll}
			<Card.Root>
				<Card.Header><Card.Title class="text-base">Enroll student</Card.Title></Card.Header>
				<Card.Content>
					<form method="POST" use:enhance action="?/enroll" class="space-y-4">
						<input type="hidden" name="institution_id" value={data.scopeNodeId} />
						<input type="hidden" name="period_id" value={data.periodId} />
						<div class="grid grid-cols-2 gap-4">
							<div class="space-y-2">
								<Label for="student_id">Student ID *</Label>
								<Input id="student_id" name="student_id" type="number" min="1" required />
							</div>
							<div class="space-y-2">
								<Label for="enrolled_on">Enrolled on *</Label>
								<Input id="enrolled_on" name="enrolled_on" type="date" required />
							</div>
						</div>
						<div class="space-y-2">
							<Label for="note">Note</Label>
							<Input id="note" name="note" placeholder="optional" />
						</div>
						{#if form?.error}<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>{/if}
						<div class="flex justify-end"><Button type="submit" size="sm">Enroll</Button></div>
					</form>
				</Card.Content>
			</Card.Root>
		{/if}

		{#if showTransfer}
			<Card.Root>
				<Card.Header><Card.Title class="text-base">Transfer student to another institution</Card.Title></Card.Header>
				<Card.Content>
					<form method="POST" use:enhance action="?/transfer" class="space-y-4">
						<input type="hidden" name="period_id" value={data.periodId} />
						<div class="grid grid-cols-3 gap-4">
							<div class="space-y-2">
								<Label for="t_student_id">Student ID *</Label>
								<Input id="t_student_id" name="student_id" type="number" min="1" required />
							</div>
							<div class="space-y-2">
								<Label for="to_institution_id">To institution ID *</Label>
								<Input id="to_institution_id" name="to_institution_id" type="number" min="1" required />
							</div>
							<div class="space-y-2">
								<Label for="effective_on">Effective on</Label>
								<Input id="effective_on" name="effective_on" type="date" />
							</div>
						</div>
						<div class="space-y-2">
							<Label for="t_note">Note</Label>
							<Input id="t_note" name="note" placeholder="optional" />
						</div>
						{#if form?.error}<Alert variant="destructive"><AlertDescription>{form.error}</AlertDescription></Alert>{/if}
						<div class="flex justify-end"><Button type="submit" size="sm">Transfer</Button></div>
					</form>
				</Card.Content>
			</Card.Root>
		{/if}

		{#if data.enrollments.length === 0}
			<Card.Root>
				<Card.Content class="py-12 text-center text-sm text-muted-foreground">
					No enrollments at this institution for this period.
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Student ID</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Enrolled</Table.Head>
							<Table.Head>Ended</Table.Head>
							<Table.Head class="w-px"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.enrollments as e}
							<Table.Row>
								<Table.Cell class="font-mono text-xs">{e.student_id}</Table.Cell>
								<Table.Cell><Badge variant={statusVariant(e.status)}>{e.status}</Badge></Table.Cell>
								<Table.Cell class="text-muted-foreground">{new Date(e.enrolled_on).toLocaleDateString()}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{e.ended_on ? new Date(e.ended_on).toLocaleDateString() : '—'}</Table.Cell>
								<Table.Cell>
									{#if e.status === 'active'}
										<form method="POST" use:enhance action="?/drop">
											<input type="hidden" name="enrollment_id" value={e.id} />
											<Button
												type="submit"
												variant="ghost"
												size="sm"
												class="h-7 px-2 text-xs text-destructive hover:text-destructive"
												onclick={(ev) => { if (!confirm('Drop this enrollment?')) ev.preventDefault(); }}
											>
												Drop
											</Button>
										</form>
									{/if}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Root>
		{/if}
	{/if}
</div>
