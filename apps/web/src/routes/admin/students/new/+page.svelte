<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="space-y-6">
	<header>
		<a href="/admin/students?institution={data.institutionId}" class="text-xs text-muted-foreground hover:underline">← Students</a>
		<h1 class="mt-1 text-2xl font-semibold">New student</h1>
		{#if data.institution}
			<p class="text-sm text-muted-foreground">at <strong>{data.institution.label}</strong></p>
		{/if}
	</header>

	<Card.Root>
		<Card.Content class="pt-6">
			<form method="POST" class="space-y-4">
				<div class="space-y-2">
					<Label for="full_name">Full name *</Label>
					<Input id="full_name" name="full_name" required />
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-2">
						<Label for="given_name">Given name</Label>
						<Input id="given_name" name="given_name" />
					</div>
					<div class="space-y-2">
						<Label for="family_name">Family name</Label>
						<Input id="family_name" name="family_name" />
					</div>
				</div>
				<div class="grid grid-cols-3 gap-4">
					<div class="space-y-2">
						<Label for="student_code">Student code</Label>
						<Input id="student_code" name="student_code" placeholder="e.g. STU-001" />
					</div>
					<div class="space-y-2">
						<Label for="enrollment_date">Enrolled</Label>
						<Input id="enrollment_date" name="enrollment_date" type="date" />
					</div>
					<div class="space-y-2">
						<Label for="gender">Gender</Label>
						<select
							id="gender"
							name="gender"
							class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-1 focus:ring-ring"
						>
							<option value=""></option>
							<option value="female">Female</option>
							<option value="male">Male</option>
							<option value="other">Other</option>
						</select>
					</div>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-2">
						<Label for="email">Email</Label>
						<Input id="email" name="email" type="email" />
					</div>
					<div class="space-y-2">
						<Label for="phone">Phone</Label>
						<Input id="phone" name="phone" type="tel" />
					</div>
				</div>

				{#if form?.error}
					<Alert variant="destructive">
						<AlertDescription>{form.error}</AlertDescription>
					</Alert>
				{/if}

				<div class="flex justify-between">
					<Button variant="ghost" href="/admin/students?institution={data.institutionId}">Cancel</Button>
					<Button type="submit">Create student</Button>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
