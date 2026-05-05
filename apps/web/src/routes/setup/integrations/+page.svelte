<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let { form }: { form: ActionData } = $props();
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="text-2xl">Integrations (optional)</Card.Title>
		<Card.Description
			>All fields are optional and can be set later in admin settings.</Card.Description
		>
	</Card.Header>
	<Card.Content>
		<form method="POST" action="?/save" class="space-y-6">
			<section class="space-y-3">
				<h3 class="text-sm font-medium">Email (Resend)</h3>
				<div class="space-y-2">
					<Label for="resend_api_key">API key</Label>
					<Input id="resend_api_key" name="resend_api_key" placeholder="re_..." />
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-2">
						<Label for="email_from">From address</Label>
						<Input id="email_from" name="email_from" placeholder="schoolrise@example.gov" />
					</div>
					<div class="space-y-2">
						<Label for="email_from_name">From name</Label>
						<Input id="email_from_name" name="email_from_name" placeholder="SchoolRise" />
					</div>
				</div>
			</section>

			<Separator />

			<section class="space-y-3">
				<h3 class="text-sm font-medium">AI</h3>
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-2">
						<Label for="openai_api_key">OpenAI API key</Label>
						<Input id="openai_api_key" name="openai_api_key" placeholder="sk-..." />
					</div>
					<div class="space-y-2">
						<Label for="openai_model">Model</Label>
						<Input id="openai_model" name="openai_model" placeholder="gpt-4o-mini" />
					</div>
				</div>
				<div class="space-y-2">
					<Label for="anthropic_api_key">Anthropic API key (optional)</Label>
					<Input id="anthropic_api_key" name="anthropic_api_key" placeholder="sk-ant-..." />
				</div>
			</section>

			<Separator />

			<section class="space-y-3">
				<h3 class="text-sm font-medium">File storage (S3-compatible)</h3>
				<div class="grid grid-cols-2 gap-3">
					<div class="space-y-2">
						<Label for="s3_endpoint">Endpoint</Label>
						<Input id="s3_endpoint" name="s3_endpoint" />
					</div>
					<div class="space-y-2">
						<Label for="s3_bucket">Bucket</Label>
						<Input id="s3_bucket" name="s3_bucket" />
					</div>
					<div class="space-y-2">
						<Label for="s3_access_key">Access key</Label>
						<Input id="s3_access_key" name="s3_access_key" />
					</div>
					<div class="space-y-2">
						<Label for="s3_secret_key">Secret key</Label>
						<Input id="s3_secret_key" name="s3_secret_key" type="password" />
					</div>
					<div class="space-y-2">
						<Label for="s3_region">Region</Label>
						<Input id="s3_region" name="s3_region" placeholder="auto" />
					</div>
				</div>
			</section>

			{#if form?.error}
				<Alert variant="destructive">
					<AlertDescription>{form.error}</AlertDescription>
				</Alert>
			{/if}

			<div class="flex justify-between">
				<Button variant="ghost" href="/setup/schools">Back</Button>
				<div class="flex gap-2">
					<Button type="submit" formaction="?/skip" variant="outline">Skip</Button>
					<Button type="submit">Save</Button>
				</div>
			</div>
		</form>
	</Card.Content>
</Card.Root>
