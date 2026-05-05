package forms

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type FormDTO struct {
	ID          int64           `json:"id"`
	PublicID    string          `json:"public_id"`
	OwnerID     int64           `json:"owner_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Settings    json.RawMessage `json:"settings"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type QuestionDTO struct {
	ID          int64           `json:"id"`
	FormID      int64           `json:"form_id"`
	ClientID    string          `json:"client_id"`
	SortOrder   int32           `json:"sort_order"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Required    bool            `json:"required"`
	Options     json.RawMessage `json:"options"`
	ScaleMin    *int32          `json:"scale_min,omitempty"`
	ScaleMax    *int32          `json:"scale_max,omitempty"`
	ScaleLabels json.RawMessage `json:"scale_labels"`
	Validation  json.RawMessage `json:"validation"`
	Grading     json.RawMessage `json:"grading"`
	Extra       json.RawMessage `json:"extra"`
}

type FormVersionDTO struct {
	ID          int64           `json:"id"`
	FormID      int64           `json:"form_id"`
	VersionNum  int32           `json:"version_num"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Snapshot    json.RawMessage `json:"snapshot"`
	PublishedAt time.Time       `json:"published_at"`
}

func toRaw(v any) json.RawMessage {
	if v == nil {
		return json.RawMessage("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return b
}

func formToDTO(f *Form) FormDTO {
	return FormDTO{
		ID: f.ID, PublicID: f.PublicID, OwnerID: f.OwnerID,
		Title: f.Title, Description: f.Description, Status: f.Status,
		Settings: toRaw(f.Settings), PublishedAt: f.PublishedAt,
		CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt,
	}
}

func questionToDTO(q *Question) QuestionDTO {
	return QuestionDTO{
		ID: q.ID, FormID: q.FormID, ClientID: q.ClientID, SortOrder: q.SortOrder,
		Title: q.Title, Description: q.Description, Type: q.Type, Required: q.Required,
		Options: toRaw(q.Options), ScaleMin: q.ScaleMin, ScaleMax: q.ScaleMax,
		ScaleLabels: toRaw(q.ScaleLabels), Validation: toRaw(q.Validation),
		Grading: toRaw(q.Grading), Extra: toRaw(q.Extra),
	}
}

func versionToDTO(v *FormVersion) FormVersionDTO {
	return FormVersionDTO{
		ID: v.ID, FormID: v.FormID, VersionNum: v.VersionNum,
		Title: v.Title, Description: v.Description, Snapshot: toRaw(v.Snapshot), PublishedAt: v.PublishedAt,
	}
}

type ListFormsResponse struct {
	Forms []FormDTO `json:"forms"`
}

//encore:api auth method=GET path=/v1/forms
func (s *Service) ListFormsAPI(ctx context.Context) (*ListFormsResponse, error) {
	uid, _ := encauth.UserID()
	ownerID, _ := strconv.ParseInt(string(uid), 10, 64)
	if ownerID == 0 {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing user"}
	}
	rows, err := ListFormsByOwner(ctx, ownerID)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]FormDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, formToDTO(r))
	}
	return &ListFormsResponse{Forms: out}, nil
}

type CreateFormAPIRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

//encore:api auth method=POST path=/v1/forms
func (s *Service) CreateFormAPI(ctx context.Context, req *CreateFormAPIRequest) (*FormDTO, error) {
	uid, _ := encauth.UserID()
	ownerID, _ := strconv.ParseInt(string(uid), 10, 64)
	if ownerID == 0 {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing user"}
	}
	f, err := CreateForm(ctx, CreateFormParams{
		OwnerID: ownerID, Title: req.Title, Description: req.Description, Settings: rawToMap(req.Settings),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := formToDTO(f)
	return &out, nil
}

type GetFormResponse struct {
	Form      FormDTO        `json:"form"`
	Questions []QuestionDTO  `json:"questions"`
}

//encore:api auth method=GET path=/v1/forms/items/:id
func (s *Service) GetFormAPI(ctx context.Context, id int64) (*GetFormResponse, error) {
	f, err := GetFormByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	qs, err := ListQuestionsByForm(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]QuestionDTO, 0, len(qs))
	for _, q := range qs {
		out = append(out, questionToDTO(q))
	}
	return &GetFormResponse{Form: formToDTO(f), Questions: out}, nil
}

type UpdateFormAPIRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

//encore:api auth method=PUT path=/v1/forms/items/:id
func (s *Service) UpdateFormAPI(ctx context.Context, id int64, req *UpdateFormAPIRequest) (*FormDTO, error) {
	f, err := UpdateFormMeta(ctx, UpdateFormParams{
		ID: id, Title: req.Title, Description: req.Description, Settings: rawToMap(req.Settings),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := formToDTO(f)
	return &out, nil
}

//encore:api auth method=DELETE path=/v1/forms/items/:id
func (s *Service) DeleteFormAPI(ctx context.Context, id int64) error {
	return mapErr(DeleteForm(ctx, id))
}

type CreateQuestionAPIRequest struct {
	ClientID    string          `json:"client_id"`
	SortOrder   int32           `json:"sort_order"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Required    bool            `json:"required"`
	Options     json.RawMessage `json:"options,omitempty"`
	ScaleMin    *int32          `json:"scale_min,omitempty"`
	ScaleMax    *int32          `json:"scale_max,omitempty"`
	ScaleLabels json.RawMessage `json:"scale_labels,omitempty"`
	Validation  json.RawMessage `json:"validation,omitempty"`
	Grading     json.RawMessage `json:"grading,omitempty"`
	Extra       json.RawMessage `json:"extra,omitempty"`
}

//encore:api auth method=POST path=/v1/forms/items/:id/questions
func (s *Service) AddQuestionAPI(ctx context.Context, id int64, req *CreateQuestionAPIRequest) (*QuestionDTO, error) {
	q, err := CreateQuestion(ctx, CreateQuestionParams{
		FormID: id, ClientID: req.ClientID, SortOrder: req.SortOrder,
		Title: req.Title, Description: req.Description, Type: req.Type, Required: req.Required,
		Options: rawToSlice(req.Options), ScaleMin: req.ScaleMin, ScaleMax: req.ScaleMax,
		ScaleLabels: rawToMap(req.ScaleLabels), Validation: rawToMap(req.Validation),
		Grading: rawToMap(req.Grading), Extra: rawToMap(req.Extra),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := questionToDTO(q)
	return &out, nil
}

type UpdateQuestionAPIRequest struct {
	ClientID    string          `json:"client_id"`
	SortOrder   int32           `json:"sort_order"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Required    bool            `json:"required"`
	Options     json.RawMessage `json:"options,omitempty"`
	ScaleMin    *int32          `json:"scale_min,omitempty"`
	ScaleMax    *int32          `json:"scale_max,omitempty"`
	ScaleLabels json.RawMessage `json:"scale_labels,omitempty"`
	Validation  json.RawMessage `json:"validation,omitempty"`
	Grading     json.RawMessage `json:"grading,omitempty"`
	Extra       json.RawMessage `json:"extra,omitempty"`
}

//encore:api auth method=PUT path=/v1/forms/questions/:qid
func (s *Service) UpdateQuestionAPI(ctx context.Context, qid int64, req *UpdateQuestionAPIRequest) (*QuestionDTO, error) {
	q, err := UpdateQuestion(ctx, UpdateQuestionParams{
		ID: qid, Title: req.Title, Description: req.Description, Type: req.Type, Required: req.Required,
		SortOrder: req.SortOrder, Options: rawToSlice(req.Options), ScaleMin: req.ScaleMin, ScaleMax: req.ScaleMax,
		ScaleLabels: rawToMap(req.ScaleLabels), Validation: rawToMap(req.Validation),
		Grading: rawToMap(req.Grading), Extra: rawToMap(req.Extra),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := questionToDTO(q)
	return &out, nil
}

//encore:api auth method=DELETE path=/v1/forms/questions/:qid
func (s *Service) DeleteQuestionAPI(ctx context.Context, qid int64) error {
	return mapErr(DeleteQuestion(ctx, qid))
}

type PublishResponse struct {
	Form    FormDTO        `json:"form"`
	Version FormVersionDTO `json:"version"`
}

//encore:api auth method=POST path=/v1/forms/items/:id/publish
func (s *Service) PublishFormAPI(ctx context.Context, id int64) (*PublishResponse, error) {
	v, f, err := PublishForm(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &PublishResponse{Form: formToDTO(f), Version: versionToDTO(v)}, nil
}

type ListVersionsResponse struct {
	Versions []FormVersionDTO `json:"versions"`
}

//encore:api auth method=GET path=/v1/forms/items/:id/versions
func (s *Service) ListVersionsAPI(ctx context.Context, id int64) (*ListVersionsResponse, error) {
	rows, err := ListVersions(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]FormVersionDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, versionToDTO(r))
	}
	return &ListVersionsResponse{Versions: out}, nil
}

//encore:api public method=GET path=/v1/forms/public-versions/:id
func (s *Service) GetPublicVersionAPI(ctx context.Context, id int64) (*FormVersionDTO, error) {
	v, err := GetFormVersion(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := versionToDTO(v)
	return &out, nil
}

type FieldTypesResponse struct {
	Types []string `json:"types"`
}

//encore:api auth method=GET path=/v1/forms/field-types
func (s *Service) ListFieldTypesAPI(ctx context.Context) (*FieldTypesResponse, error) {
	return &FieldTypesResponse{Types: AllTypes()}, nil
}

func rawToMap(r json.RawMessage) map[string]any {
	if len(r) == 0 || string(r) == "null" {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(r, &m)
	return m
}

func rawToSlice(r json.RawMessage) []any {
	if len(r) == 0 || string(r) == "null" {
		return nil
	}
	var s []any
	_ = json.Unmarshal(r, &s)
	return s
}

func internal(err error) error { return apierr.WrapInternal("forms", err) }

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrFormNotFound), errors.Is(err, ErrQuestionNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	case errors.Is(err, ErrInvalidFormInput), errors.Is(err, ErrInvalidQuestionInput),
		errors.Is(err, ErrInvalidFieldType), errors.Is(err, ErrFormNotPublishable):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	}
	return internal(err)
}
