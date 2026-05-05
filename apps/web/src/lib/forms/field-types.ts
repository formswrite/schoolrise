export const FieldType = {
	SHORT_ANSWER: 'SHORT_ANSWER',
	PARAGRAPH: 'PARAGRAPH',
	EMAIL: 'EMAIL',
	PHONE: 'PHONE',
	HOME_NUMBER: 'HOME_NUMBER',
	NUMBER: 'NUMBER',
	DECIMAL: 'DECIMAL',

	MULTIPLE_CHOICE: 'MULTIPLE_CHOICE',
	CHECKBOX: 'CHECKBOX',
	DROPDOWN: 'DROPDOWN',
	RADIO: 'RADIO',
	YES_NO: 'YES_NO',
	COUNTRY_REGION: 'COUNTRY_REGION',

	LINEAR_SCALE: 'LINEAR_SCALE',
	RATING: 'RATING',

	DATE: 'DATE',
	TIME: 'TIME',

	FILE_UPLOAD: 'FILE_UPLOAD',
	ATTACHMENT: 'ATTACHMENT',
	IMAGE: 'IMAGE',
	SIGNATURE: 'SIGNATURE',

	ADDRESS: 'ADDRESS',
	TABLE: 'TABLE',

	ORDERING: 'ORDERING',
	MATCHING: 'MATCHING',
	FILL_IN_BLANK: 'FILL_IN_BLANK',
	EQUATION: 'EQUATION',
	ESSAY: 'ESSAY',
	HOTSPOT: 'HOTSPOT',
	CODE_BLOCK: 'CODE_BLOCK',

	SECTION: 'SECTION',
	STATEMENT: 'STATEMENT'
} as const;

export type FieldTypeValue = (typeof FieldType)[keyof typeof FieldType];

export const FIELD_LABELS: Record<FieldTypeValue, string> = {
	SHORT_ANSWER: 'Short answer',
	PARAGRAPH: 'Paragraph',
	EMAIL: 'Email',
	PHONE: 'Phone',
	HOME_NUMBER: 'Home number',
	NUMBER: 'Number',
	DECIMAL: 'Decimal',

	MULTIPLE_CHOICE: 'Multiple choice',
	CHECKBOX: 'Checkbox',
	DROPDOWN: 'Dropdown',
	RADIO: 'Radio',
	YES_NO: 'Yes / No',
	COUNTRY_REGION: 'Country / Region',

	LINEAR_SCALE: 'Linear scale',
	RATING: 'Rating',

	DATE: 'Date',
	TIME: 'Time',

	FILE_UPLOAD: 'File upload',
	ATTACHMENT: 'Attachment',
	IMAGE: 'Image',
	SIGNATURE: 'Signature',

	ADDRESS: 'Address',
	TABLE: 'Table',

	ORDERING: 'Ordering',
	MATCHING: 'Matching',
	FILL_IN_BLANK: 'Fill in the blank',
	EQUATION: 'Equation',
	ESSAY: 'Essay',
	HOTSPOT: 'Hotspot',
	CODE_BLOCK: 'Code block',

	SECTION: 'Section heading',
	STATEMENT: 'Statement / info'
};

export type FieldTypeGroup = {
	label: string;
	types: FieldTypeValue[];
};

export const FIELD_TYPE_GROUPS: FieldTypeGroup[] = [
	{
		label: 'Text',
		types: ['SHORT_ANSWER', 'PARAGRAPH', 'EMAIL', 'PHONE', 'HOME_NUMBER', 'NUMBER', 'DECIMAL']
	},
	{
		label: 'Choice',
		types: ['MULTIPLE_CHOICE', 'CHECKBOX', 'DROPDOWN', 'RADIO', 'YES_NO', 'COUNTRY_REGION']
	},
	{ label: 'Scale & Rating', types: ['LINEAR_SCALE', 'RATING'] },
	{ label: 'Date & Time', types: ['DATE', 'TIME'] },
	{ label: 'Media', types: ['FILE_UPLOAD', 'ATTACHMENT', 'IMAGE', 'SIGNATURE'] },
	{ label: 'Compound', types: ['ADDRESS', 'TABLE'] },
	{
		label: 'Assessment',
		types: ['ORDERING', 'MATCHING', 'FILL_IN_BLANK', 'EQUATION', 'ESSAY', 'HOTSPOT', 'CODE_BLOCK']
	},
	{ label: 'Layout', types: ['SECTION', 'STATEMENT'] }
];

export const TYPES_WITH_OPTIONS: ReadonlySet<FieldTypeValue> = new Set([
	'MULTIPLE_CHOICE',
	'CHECKBOX',
	'DROPDOWN',
	'RADIO'
]);

export const SCALE_TYPES: ReadonlySet<FieldTypeValue> = new Set(['LINEAR_SCALE', 'RATING']);

export const NON_SUBMITTABLE_TYPES: ReadonlySet<FieldTypeValue> = new Set(['SECTION', 'STATEMENT']);

export const RENDERER_PENDING: ReadonlySet<FieldTypeValue> = new Set(['HOTSPOT']);

export type Question = {
	id?: number;
	client_id: string;
	sort_order: number;
	type: FieldTypeValue;
	title: string;
	description?: string;
	required: boolean;
	options?: Array<{ label?: string; value?: string } | string>;
	scale_min?: number;
	scale_max?: number;
	scale_labels?: Record<string, string>;
	validation?: Record<string, unknown>;
	grading?: Record<string, unknown>;
	extra?: Record<string, unknown>;
};

export function getDefaultQuestion(type: FieldTypeValue, sortOrder: number): Partial<Question> {
	const base: Partial<Question> = {
		type,
		sort_order: sortOrder,
		title: FIELD_LABELS[type],
		required: false
	};

	if (TYPES_WITH_OPTIONS.has(type)) {
		base.options = [
			{ label: 'Option A', value: 'A' },
			{ label: 'Option B', value: 'B' }
		];
	}
	if (SCALE_TYPES.has(type)) {
		base.scale_min = 1;
		base.scale_max = type === 'RATING' ? 5 : 10;
	}
	if (type === 'YES_NO') {
		base.required = true;
	}
	if (type === 'SECTION') {
		base.title = 'New section';
	}
	if (type === 'STATEMENT') {
		base.title = 'Information for the respondent';
	}
	return base;
}
