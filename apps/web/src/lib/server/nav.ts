import type { RoleSummary } from './encore';

export type NavGroup = 'Insights' | 'Setup' | 'People' | 'Assessment' | 'Operations' | 'System';

export type NavItem = {
	href: string;
	label: string;
	icon: string;
	group: NavGroup;
	roles: Array<keyof RoleSummary>;
};

export const ADMIN_NAV_ALL: NavItem[] = [
	{ href: '/admin/dashboard',     label: 'Dashboard',     icon: 'BarChart3',     group: 'Insights',   roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/institutions',  label: 'Institutions',  icon: 'Building2',     group: 'Setup',      roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/periods',       label: 'Periods',       icon: 'CalendarRange', group: 'Setup',      roles: ['isGlobalAdmin'] },
	{ href: '/admin/niveaux',       label: 'Niveaux',       icon: 'Layers',        group: 'Setup',      roles: ['isGlobalAdmin'] },
	{ href: '/admin/classes',       label: 'Classes',       icon: 'Users2',        group: 'Setup',      roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/students',      label: 'Students',      icon: 'GraduationCap', group: 'People',     roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/staff',         label: 'Staff',         icon: 'IdCard',        group: 'People',     roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/enrollment',    label: 'Enrollment',    icon: 'UserCheck',     group: 'People',     roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/forms',         label: 'Forms',         icon: 'FileText',      group: 'Assessment', roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/campaigns',     label: 'Campaigns',     icon: 'Megaphone',     group: 'Assessment', roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/ai',            label: 'AI',            icon: 'Sparkles',      group: 'Assessment', roles: ['isGlobalAdmin', 'isInspector'] },
	{ href: '/admin/imports',       label: 'Imports',       icon: 'Upload',        group: 'Operations', roles: ['isGlobalAdmin'] },
	{ href: '/admin/notifications', label: 'Notifications', icon: 'Mail',          group: 'Operations', roles: ['isGlobalAdmin'] },
	{ href: '/admin/users',         label: 'Users',         icon: 'Shield',        group: 'System',     roles: ['isGlobalAdmin'] }
];

export function visibleNav(summary: RoleSummary): NavItem[] {
	return ADMIN_NAV_ALL.filter((item) => item.roles.some((r) => summary[r]));
}

export function canAccessAdminRoute(pathname: string, summary: RoleSummary): boolean {
	if (summary.isGlobalAdmin) return true;
	for (const item of ADMIN_NAV_ALL) {
		if (pathname === item.href || pathname.startsWith(item.href + '/')) {
			return item.roles.some((r) => summary[r]);
		}
	}
	return summary.isGlobalAdmin || summary.isInspector;
}
