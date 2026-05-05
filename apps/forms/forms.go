package forms

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/forms/dbforms"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusClosed    = "closed"
)

var (
	ErrFormNotFound       = errors.New("forms: form not found")
	ErrQuestionNotFound   = errors.New("forms: question not found")
	ErrInvalidFormInput   = errors.New("forms: invalid form input")
	ErrInvalidQuestionInput = errors.New("forms: invalid question input")
	ErrInvalidFieldType   = errors.New("forms: invalid field type")
	ErrFormNotPublishable = errors.New("forms: form has no submittable questions")
	ErrPublicIDCollision  = errors.New("forms: public id collision")
)

type Form struct {
	ID            int64
	PublicID      string
	OwnerID       int64
	Title         string
	Description   string
	Status        string
	Settings      map[string]any
	ResponseCount int
	ViewCount     int
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Question struct {
	ID          int64
	FormID      int64
	ClientID    string
	SortOrder   int32
	Title       string
	Description string
	Type        string
	Required    bool
	Options     []any
	ScaleMin    *int32
	ScaleMax    *int32
	ScaleLabels map[string]any
	Validation  map[string]any
	Grading     map[string]any
	Extra       map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FormVersion struct {
	ID          int64
	FormID      int64
	VersionNum  int32
	Title       string
	Description string
	Snapshot    map[string]any
	PublishedAt time.Time
}

type CreateFormParams struct {
	OwnerID     int64
	Title       string
	Description string
	Settings    map[string]any
}

func CreateForm(ctx context.Context, p CreateFormParams) (*Form, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" || p.OwnerID <= 0 {
		return nil, ErrInvalidFormInput
	}

	settings, err := jsonOrDefault(p.Settings, "{}")
	if err != nil {
		return nil, err
	}

	publicID, err := newPublicID(12)
	if err != nil {
		return nil, err
	}

	row, err := queries.CreateForm(ctx, dbforms.CreateFormParams{
		PublicID:    publicID,
		OwnerID:     p.OwnerID,
		Title:       title,
		Description: strings.TrimSpace(p.Description),
		Settings:    settings,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrPublicIDCollision
		}
		return nil, err
	}
	return formFromRow(row), nil
}

func GetFormByID(ctx context.Context, id int64) (*Form, error) {
	row, err := queries.GetFormByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFormNotFound
	}
	if err != nil {
		return nil, err
	}
	return formFromRow(row), nil
}

func GetFormByPublicID(ctx context.Context, publicID string) (*Form, error) {
	row, err := queries.GetFormByPublicID(ctx, publicID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFormNotFound
	}
	if err != nil {
		return nil, err
	}
	return formFromRow(row), nil
}

func ListFormsByOwner(ctx context.Context, ownerID int64) ([]*Form, error) {
	rows, err := queries.ListFormsByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]*Form, 0, len(rows))
	for _, r := range rows {
		out = append(out, formFromRow(r))
	}
	return out, nil
}

type UpdateFormParams struct {
	ID          int64
	Title       string
	Description string
	Settings    map[string]any
}

func UpdateFormMeta(ctx context.Context, p UpdateFormParams) (*Form, error) {
	if _, err := GetFormByID(ctx, p.ID); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidFormInput
	}

	settings, err := jsonOrDefault(p.Settings, "{}")
	if err != nil {
		return nil, err
	}

	row, err := queries.UpdateFormMeta(ctx, dbforms.UpdateFormMetaParams{
		ID: p.ID, Title: title, Description: strings.TrimSpace(p.Description), Settings: settings,
	})
	if err != nil {
		return nil, err
	}
	return formFromRow(row), nil
}

func DeleteForm(ctx context.Context, id int64) error {
	if _, err := GetFormByID(ctx, id); err != nil {
		return err
	}
	return queries.SoftDeleteForm(ctx, id)
}

type CreateQuestionParams struct {
	FormID      int64
	ClientID    string
	SortOrder   int32
	Title       string
	Description string
	Type        string
	Required    bool
	Options     []any
	ScaleMin    *int32
	ScaleMax    *int32
	ScaleLabels map[string]any
	Validation  map[string]any
	Grading     map[string]any
	Extra       map[string]any
}

func CreateQuestion(ctx context.Context, p CreateQuestionParams) (*Question, error) {
	if p.FormID <= 0 || strings.TrimSpace(p.ClientID) == "" {
		return nil, ErrInvalidQuestionInput
	}
	if !IsValidType(p.Type) {
		return nil, ErrInvalidFieldType
	}
	if _, err := GetFormByID(ctx, p.FormID); err != nil {
		return nil, err
	}

	options, err := jsonOrDefault(p.Options, "[]")
	if err != nil {
		return nil, err
	}
	scaleLabels, err := jsonOrDefault(p.ScaleLabels, "{}")
	if err != nil {
		return nil, err
	}
	validation, err := jsonOrDefault(p.Validation, "{}")
	if err != nil {
		return nil, err
	}
	grading, err := jsonOrDefault(p.Grading, "{}")
	if err != nil {
		return nil, err
	}
	extra, err := jsonOrDefault(p.Extra, "{}")
	if err != nil {
		return nil, err
	}

	row, err := queries.CreateQuestion(ctx, dbforms.CreateQuestionParams{
		FormID:      p.FormID,
		ClientID:    p.ClientID,
		SortOrder:   p.SortOrder,
		Title:       p.Title,
		Description: p.Description,
		Type:        p.Type,
		Required:    p.Required,
		Options:     options,
		ScaleMin:    nullableInt32Ptr(p.ScaleMin),
		ScaleMax:    nullableInt32Ptr(p.ScaleMax),
		ScaleLabels: scaleLabels,
		Validation:  validation,
		Grading:     grading,
		Extra:       extra,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrInvalidQuestionInput
		}
		return nil, err
	}
	return questionFromRow(row), nil
}

func ListQuestionsByForm(ctx context.Context, formID int64) ([]*Question, error) {
	rows, err := queries.ListQuestionsByForm(ctx, formID)
	if err != nil {
		return nil, err
	}
	out := make([]*Question, 0, len(rows))
	for _, r := range rows {
		out = append(out, questionFromRow(r))
	}
	return out, nil
}

func GetQuestionByID(ctx context.Context, id int64) (*Question, error) {
	row, err := queries.GetQuestionByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrQuestionNotFound
	}
	if err != nil {
		return nil, err
	}
	return questionFromRow(row), nil
}

type UpdateQuestionParams struct {
	ID          int64
	Title       string
	Description string
	Type        string
	Required    bool
	SortOrder   int32
	Options     []any
	ScaleMin    *int32
	ScaleMax    *int32
	ScaleLabels map[string]any
	Validation  map[string]any
	Grading     map[string]any
	Extra       map[string]any
}

func UpdateQuestion(ctx context.Context, p UpdateQuestionParams) (*Question, error) {
	if _, err := GetQuestionByID(ctx, p.ID); err != nil {
		return nil, err
	}
	if !IsValidType(p.Type) {
		return nil, ErrInvalidFieldType
	}
	options, _ := jsonOrDefault(p.Options, "[]")
	scaleLabels, _ := jsonOrDefault(p.ScaleLabels, "{}")
	validation, _ := jsonOrDefault(p.Validation, "{}")
	grading, _ := jsonOrDefault(p.Grading, "{}")
	extra, _ := jsonOrDefault(p.Extra, "{}")

	row, err := queries.UpdateQuestion(ctx, dbforms.UpdateQuestionParams{
		ID: p.ID, Title: p.Title, Description: p.Description, Type: p.Type, Required: p.Required,
		SortOrder: p.SortOrder, Options: options,
		ScaleMin: nullableInt32Ptr(p.ScaleMin), ScaleMax: nullableInt32Ptr(p.ScaleMax),
		ScaleLabels: scaleLabels, Validation: validation, Grading: grading, Extra: extra,
	})
	if err != nil {
		return nil, err
	}
	return questionFromRow(row), nil
}

func DeleteQuestion(ctx context.Context, id int64) error {
	if _, err := GetQuestionByID(ctx, id); err != nil {
		return err
	}
	return queries.SoftDeleteQuestion(ctx, id)
}

func PublishForm(ctx context.Context, formID int64) (*FormVersion, *Form, error) {
	form, err := GetFormByID(ctx, formID)
	if err != nil {
		return nil, nil, err
	}

	questions, err := ListQuestionsByForm(ctx, formID)
	if err != nil {
		return nil, nil, err
	}

	hasSubmittable := false
	for _, q := range questions {
		if IsSubmittable(q.Type) {
			hasSubmittable = true
			break
		}
	}
	if !hasSubmittable {
		return nil, nil, ErrFormNotPublishable
	}

	prev, err := queries.GetLatestFormVersion(ctx, formID)
	var nextVersion int32 = 1
	if err == nil {
		nextVersion = prev.VersionNum + 1
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	snapshotPayload := map[string]any{
		"title":       form.Title,
		"description": form.Description,
		"settings":    form.Settings,
		"questions":   serializeQuestions(questions),
	}
	snapshotJSON, err := json.Marshal(snapshotPayload)
	if err != nil {
		return nil, nil, err
	}

	versionRow, err := queries.CreateFormVersion(ctx, dbforms.CreateFormVersionParams{
		FormID:      formID,
		VersionNum:  nextVersion,
		Title:       form.Title,
		Description: form.Description,
		Snapshot:    snapshotJSON,
	})
	if err != nil {
		return nil, nil, err
	}

	updated, err := queries.UpdateFormStatus(ctx, dbforms.UpdateFormStatusParams{
		ID: formID, Status: StatusPublished,
	})
	if err != nil {
		return nil, nil, err
	}

	return versionFromRow(versionRow), formFromRow(updated), nil
}

func GetFormVersion(ctx context.Context, id int64) (*FormVersion, error) {
	row, err := queries.GetFormVersion(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFormNotFound
	}
	if err != nil {
		return nil, err
	}
	return versionFromRow(row), nil
}

func ListVersions(ctx context.Context, formID int64) ([]*FormVersion, error) {
	rows, err := queries.ListFormVersions(ctx, formID)
	if err != nil {
		return nil, err
	}
	out := make([]*FormVersion, 0, len(rows))
	for _, r := range rows {
		out = append(out, versionFromRow(r))
	}
	return out, nil
}

func CloseForm(ctx context.Context, formID int64) (*Form, error) {
	if _, err := GetFormByID(ctx, formID); err != nil {
		return nil, err
	}
	row, err := queries.UpdateFormStatus(ctx, dbforms.UpdateFormStatusParams{ID: formID, Status: StatusClosed})
	if err != nil {
		return nil, err
	}
	return formFromRow(row), nil
}

func serializeQuestions(qs []*Question) []map[string]any {
	out := make([]map[string]any, 0, len(qs))
	for _, q := range qs {
		out = append(out, map[string]any{
			"id":           q.ID,
			"client_id":    q.ClientID,
			"sort_order":   q.SortOrder,
			"title":        q.Title,
			"description":  q.Description,
			"type":         q.Type,
			"required":     q.Required,
			"options":      q.Options,
			"scale_min":    q.ScaleMin,
			"scale_max":    q.ScaleMax,
			"scale_labels": q.ScaleLabels,
			"validation":   q.Validation,
			"grading":      q.Grading,
			"extra":        q.Extra,
		})
	}
	return out
}

func formFromRow(r dbforms.Form) *Form {
	f := &Form{
		ID:            r.ID,
		PublicID:      r.PublicID,
		OwnerID:       r.OwnerID,
		Title:         r.Title,
		Description:   r.Description,
		Status:        r.Status,
		ResponseCount: int(r.ResponseCount),
		ViewCount:     int(r.ViewCount),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	_ = json.Unmarshal(r.Settings, &f.Settings)
	if f.Settings == nil {
		f.Settings = map[string]any{}
	}
	if r.PublishedAt.Valid {
		t := r.PublishedAt.Time
		f.PublishedAt = &t
	}
	return f
}

func questionFromRow(r dbforms.Question) *Question {
	q := &Question{
		ID:          r.ID,
		FormID:      r.FormID,
		ClientID:    r.ClientID,
		SortOrder:   r.SortOrder,
		Title:       r.Title,
		Description: r.Description,
		Type:        r.Type,
		Required:    r.Required,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	_ = json.Unmarshal(r.Options, &q.Options)
	_ = json.Unmarshal(r.ScaleLabels, &q.ScaleLabels)
	_ = json.Unmarshal(r.Validation, &q.Validation)
	_ = json.Unmarshal(r.Grading, &q.Grading)
	_ = json.Unmarshal(r.Extra, &q.Extra)
	if q.Options == nil {
		q.Options = []any{}
	}
	if q.ScaleLabels == nil {
		q.ScaleLabels = map[string]any{}
	}
	if q.Validation == nil {
		q.Validation = map[string]any{}
	}
	if q.Grading == nil {
		q.Grading = map[string]any{}
	}
	if q.Extra == nil {
		q.Extra = map[string]any{}
	}
	if r.ScaleMin.Valid {
		v := r.ScaleMin.Int32
		q.ScaleMin = &v
	}
	if r.ScaleMax.Valid {
		v := r.ScaleMax.Int32
		q.ScaleMax = &v
	}
	return q
}

func versionFromRow(r dbforms.FormVersion) *FormVersion {
	v := &FormVersion{
		ID:          r.ID,
		FormID:      r.FormID,
		VersionNum:  r.VersionNum,
		Title:       r.Title,
		Description: r.Description,
		PublishedAt: r.PublishedAt,
	}
	_ = json.Unmarshal(r.Snapshot, &v.Snapshot)
	return v
}

func jsonOrDefault(v any, def string) ([]byte, error) {
	if v == nil {
		return []byte(def), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if string(b) == "null" {
		return []byte(def), nil
	}
	return b, nil
}

func nullableInt32Ptr(p *int32) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *p, Valid: true}
}

const publicIDAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"

func newPublicID(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i, b := range buf {
		out[i] = publicIDAlphabet[int(b)%len(publicIDAlphabet)]
	}
	return string(out), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
