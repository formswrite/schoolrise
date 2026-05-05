<script lang="ts">
	import { page } from '$app/stores';
	import type { NavItem, NavGroup } from '$lib/server/nav';
	import {
		BarChart3, Building2, CalendarRange, Layers, Users2,
		GraduationCap, IdCard, UserCheck, FileText, Megaphone, Sparkles, Upload, Mail, Shield,
		LogOut, Presentation
	} from '@lucide/svelte';
	import { cn } from '$lib/utils';

	type Props = {
		items: NavItem[];
		userEmail: string;
		roleLabel: string;
		showTeacherLink: boolean;
	};

	let { items, userEmail, roleLabel, showTeacherLink }: Props = $props();

	const ICONS = {
		BarChart3, Building2, CalendarRange, Layers, Users2,
		GraduationCap, IdCard, UserCheck, FileText, Megaphone, Sparkles, Upload, Mail, Shield
	} as const;

	const GROUP_ORDER: NavGroup[] = ['Insights', 'Setup', 'People', 'Assessment', 'Operations', 'System'];

	const grouped = $derived(
		GROUP_ORDER.map((g) => ({
			group: g,
			items: items.filter((i) => i.group === g)
		})).filter((g) => g.items.length > 0)
	);

	function isActive(href: string, pathname: string): boolean {
		if (href === pathname) return true;
		if (href !== '/' && pathname.startsWith(href + '/')) return true;
		return false;
	}
</script>

<aside class="hidden md:flex md:w-60 md:flex-col md:fixed md:inset-y-0 border-r bg-background z-30">
	<div class="flex h-14 shrink-0 items-center gap-2 border-b px-4">
		<a href="/admin/dashboard" class="flex items-center gap-2" aria-label="SchoolRise home">
			<img src="/logo-mark.svg" alt="" class="h-7 w-7" aria-hidden="true" />
			<span class="font-display text-[15px] font-bold tracking-tight text-[#060419]">SchoolRise</span>
		</a>
	</div>

	<nav class="flex-1 overflow-y-auto px-2 py-3">
		{#each grouped as section}
			<div class="mb-3">
				<p class="mb-1 px-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">{section.group}</p>
				<ul class="space-y-0.5">
					{#each section.items as item}
						{@const Icon = ICONS[item.icon as keyof typeof ICONS]}
						{@const active = isActive(item.href, $page.url.pathname)}
						<li>
							<a
								href={item.href}
								class={cn(
									'group flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm transition-colors',
									active
										? 'bg-primary text-primary-foreground font-semibold'
										: 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
								)}
							>
								<Icon class={cn('size-4 shrink-0', active ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-accent-foreground')} />
								<span class="truncate">{item.label}</span>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>

	<div class="border-t px-3 py-3">
		{#if showTeacherLink}
			<a href="/teacher" class="mb-2 flex items-center gap-2 rounded-md px-2 py-1.5 text-xs text-[#6439B5] hover:bg-accent">
				<Presentation class="size-3.5" /> Teacher view
			</a>
		{/if}
		<div class="flex items-center gap-2 px-2 py-1.5">
			<div class="flex size-8 items-center justify-center rounded-full bg-secondary text-xs font-semibold text-secondary-foreground">
				{userEmail.charAt(0).toUpperCase()}
			</div>
			<div class="min-w-0 flex-1">
				<p class="truncate text-xs font-semibold">{userEmail}</p>
				<p class="text-[10px] text-muted-foreground">{roleLabel}</p>
			</div>
			<form method="POST" action="/logout">
				<button type="submit" class="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground" aria-label="Sign out">
					<LogOut class="size-4" />
				</button>
			</form>
		</div>
	</div>
</aside>

<div class="flex h-14 items-center justify-between border-b bg-background px-4 md:hidden">
	<a href="/admin/dashboard" class="flex items-center gap-2">
		<img src="/logo-mark.svg" alt="" class="h-7 w-7" aria-hidden="true" />
		<span class="font-display text-sm font-bold">SchoolRise</span>
	</a>
	<div class="flex items-center gap-2 text-xs text-muted-foreground">
		<span>{roleLabel}</span>
		<form method="POST" action="/logout">
			<button type="submit" class="flex size-8 items-center justify-center rounded-md hover:bg-accent" aria-label="Sign out">
				<LogOut class="size-4" />
			</button>
		</form>
	</div>
</div>
