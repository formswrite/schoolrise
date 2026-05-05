package academics

import (
	"context"

	"encore.app/apps/academics/dbacademics"
)

func AddStudentToClass(ctx context.Context, classID, studentID int64) error {
	if _, err := GetClassByID(ctx, classID); err != nil {
		return err
	}
	return queries.AddStudentToClass(ctx, dbacademics.AddStudentToClassParams{
		ClassID:   classID,
		StudentID: studentID,
	})
}

func RemoveStudentFromClass(ctx context.Context, classID, studentID int64) error {
	return queries.RemoveStudentFromClass(ctx, dbacademics.RemoveStudentFromClassParams{
		ClassID:   classID,
		StudentID: studentID,
	})
}

func ListClassStudents(ctx context.Context, classID int64) ([]int64, error) {
	rows, err := queries.ListClassStudents(ctx, classID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func ListClassesForStudent(ctx context.Context, studentID int64) ([]*Class, error) {
	rows, err := queries.ListClassesForStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	out := make([]*Class, 0, len(rows))
	for _, r := range rows {
		out = append(out, classFromRow(r))
	}
	return out, nil
}

type ClassStaffMember struct {
	StaffID int64
	Role    string
}

func AddStaffToClass(ctx context.Context, classID, staffID int64, role string) error {
	if role == "" {
		role = "teacher"
	}
	if _, err := GetClassByID(ctx, classID); err != nil {
		return err
	}
	return queries.AddStaffToClass(ctx, dbacademics.AddStaffToClassParams{
		ClassID: classID,
		StaffID: staffID,
		Role:    role,
	})
}

func RemoveStaffFromClass(ctx context.Context, classID, staffID int64, role string) error {
	if role == "" {
		role = "teacher"
	}
	return queries.RemoveStaffFromClass(ctx, dbacademics.RemoveStaffFromClassParams{
		ClassID: classID,
		StaffID: staffID,
		Role:    role,
	})
}

func ListClassStaff(ctx context.Context, classID int64) ([]ClassStaffMember, error) {
	rows, err := queries.ListClassStaff(ctx, classID)
	if err != nil {
		return nil, err
	}
	out := make([]ClassStaffMember, 0, len(rows))
	for _, r := range rows {
		out = append(out, ClassStaffMember{StaffID: r.StaffID, Role: r.Role})
	}
	return out, nil
}

func ListClassesForAnyStaff(ctx context.Context, staffIDs []int64) ([]*Class, error) {
	if len(staffIDs) == 0 {
		return []*Class{}, nil
	}
	rows, err := queries.ListClassesForAnyStaff(ctx, staffIDs)
	if err != nil {
		return nil, err
	}
	out := make([]*Class, 0, len(rows))
	for _, r := range rows {
		out = append(out, classFromRow(r))
	}
	return out, nil
}

func IsClassStaff(ctx context.Context, classID int64, staffIDs []int64) (bool, error) {
	if len(staffIDs) == 0 {
		return false, nil
	}
	return queries.IsClassStaff(ctx, dbacademics.IsClassStaffParams{
		ClassID:  classID,
		StaffIds: staffIDs,
	})
}

func ListClassesForStaff(ctx context.Context, staffID int64) ([]*Class, error) {
	rows, err := queries.ListClassesForStaff(ctx, staffID)
	if err != nil {
		return nil, err
	}
	out := make([]*Class, 0, len(rows))
	for _, r := range rows {
		out = append(out, classFromRow(r))
	}
	return out, nil
}
