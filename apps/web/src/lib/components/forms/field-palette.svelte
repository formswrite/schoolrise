<script lang="ts">
	import { FIELD_TYPE_GROUPS, FIELD_LABELS, RENDERER_PENDING } from '$lib/forms/field-types';
	import type { FieldTypeValue } from '$lib/forms/field-types';
	import { Badge } from '$lib/components/ui/badge';
	import { enhance } from '$app/forms';

	type Props = { actionUrl?: string };
	let { actionUrl = '?/addQuestion' }: Props = $props();
</script>

<aside
	class="hidden md:flex md:w-56 md:shrink-0 md:flex-col md:gap-3 md:overflow-y-auto md:border-r md:bg-muted/20 md:px-3 md:py-4"
>
	<header class="px-1">
		<h2 class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Add a field</h2>
		<p class="mt-1 text-[11px] text-muted-foreground">Click to append to the form.</p>
	</header>

	{#each FIELD_TYPE_GROUPS as group}
		<div class="space-y-1">
			<p class="px-1 text-[10px] font-semibold tracking-wider text-muted-foreground/70 uppercase">
				{group.label}
			</p>
			<ul class="space-y-0.5">
				{#each group.types as type (type)}
					{@const pending = RENDERER_PENDING.has(type)}
					<li>
						<form method="POST" action={actionUrl} use:enhance>
							<input type="hidden" name="type" value={type} />
							<button
								type="submit"
								class="group flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground"
								title={pending ? `${FIELD_LABELS[type]} — renderer pending` : FIELD_LABELS[type]}
							>
								<span class="truncate">{FIELD_LABELS[type]}</span>
								{#if pending}
									<Badge
										variant="outline"
										class="ml-1 shrink-0 border-amber-300 bg-amber-50 px-1 text-[9px] font-medium text-amber-900"
									>
										β
									</Badge>
								{/if}
							</button>
						</form>
					</li>
				{/each}
			</ul>
		</div>
	{/each}
</aside>
