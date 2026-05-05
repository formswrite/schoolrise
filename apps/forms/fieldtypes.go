package forms

const (
	TypeShortAnswer    = "SHORT_ANSWER"
	TypeParagraph      = "PARAGRAPH"
	TypeEmail          = "EMAIL"
	TypePhone          = "PHONE"
	TypeHomeNumber     = "HOME_NUMBER"
	TypeNumber         = "NUMBER"
	TypeDecimal        = "DECIMAL"

	TypeMultipleChoice = "MULTIPLE_CHOICE"
	TypeCheckbox       = "CHECKBOX"
	TypeDropdown       = "DROPDOWN"
	TypeRadio          = "RADIO"
	TypeYesNo          = "YES_NO"
	TypeCountryRegion  = "COUNTRY_REGION"

	TypeLinearScale    = "LINEAR_SCALE"
	TypeRating         = "RATING"

	TypeDate           = "DATE"
	TypeTime           = "TIME"

	TypeFileUpload     = "FILE_UPLOAD"
	TypeAttachment     = "ATTACHMENT"
	TypeImage          = "IMAGE"
	TypeSignature      = "SIGNATURE"

	TypeAddress        = "ADDRESS"
	TypeTable          = "TABLE"

	TypeOrdering       = "ORDERING"
	TypeMatching       = "MATCHING"
	TypeFillInBlank    = "FILL_IN_BLANK"
	TypeEquation       = "EQUATION"
	TypeEssay          = "ESSAY"
	TypeHotspot        = "HOTSPOT"
	TypeCodeBlock      = "CODE_BLOCK"

	TypeSection        = "SECTION"
	TypeStatement      = "STATEMENT"
)

var validTypes = map[string]struct{}{
	TypeShortAnswer: {}, TypeParagraph: {}, TypeEmail: {}, TypePhone: {},
	TypeHomeNumber: {}, TypeNumber: {}, TypeDecimal: {},
	TypeMultipleChoice: {}, TypeCheckbox: {}, TypeDropdown: {}, TypeRadio: {},
	TypeYesNo: {}, TypeCountryRegion: {},
	TypeLinearScale: {}, TypeRating: {},
	TypeDate: {}, TypeTime: {},
	TypeFileUpload: {}, TypeAttachment: {}, TypeImage: {}, TypeSignature: {},
	TypeAddress: {}, TypeTable: {},
	TypeOrdering: {}, TypeMatching: {}, TypeFillInBlank: {}, TypeEquation: {},
	TypeEssay: {}, TypeHotspot: {}, TypeCodeBlock: {},
	TypeSection: {}, TypeStatement: {},
}

func IsValidType(t string) bool {
	_, ok := validTypes[t]
	return ok
}

func AllTypes() []string {
	out := make([]string, 0, len(validTypes))
	for t := range validTypes {
		out = append(out, t)
	}
	return out
}

var nonSubmittable = map[string]struct{}{
	TypeSection:   {},
	TypeStatement: {},
}

func IsSubmittable(t string) bool {
	_, layout := nonSubmittable[t]
	return !layout
}

var gradableTypes = map[string]struct{}{
	TypeMultipleChoice: {}, TypeCheckbox: {}, TypeDropdown: {}, TypeRadio: {},
	TypeYesNo: {}, TypeShortAnswer: {}, TypeNumber: {}, TypeDecimal: {},
	TypeOrdering: {}, TypeMatching: {}, TypeFillInBlank: {}, TypeEquation: {},
	TypeEssay: {},
}

func IsGradable(t string) bool {
	_, ok := gradableTypes[t]
	return ok
}
