package grpcsvc

import (
	"context"

	hrv1 "erp/platform/genproto/erp/hr/v1"
	"erp/platform/hr-service/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HRServer struct {
	hrv1.UnimplementedHRServiceServer
	st *storage.Store
}

func NewHRServer(st *storage.Store) *HRServer {
	return &HRServer{st: st}
}

func (s *HRServer) CreateDepartment(ctx context.Context, in *hrv1.CreateDepartmentRequest) (*hrv1.Department, error) {
	id, ts, err := s.st.CreateDepartment(ctx, in.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &hrv1.Department{Id: id.String(), Name: in.GetName(), CreatedAtUnix: ts}, nil
}

func (s *HRServer) ListDepartments(ctx context.Context, in *hrv1.ListDepartmentsRequest) (*hrv1.ListDepartmentsResponse, error) {
	limit, offset := in.GetLimit(), in.GetOffset()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListDepartments(ctx, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*hrv1.Department
	for _, r := range rows {
		list = append(list, &hrv1.Department{Id: r.ID, Name: r.Name, CreatedAtUnix: r.Created})
	}
	return &hrv1.ListDepartmentsResponse{Departments: list}, nil
}

func (s *HRServer) CreateEmployee(ctx context.Context, in *hrv1.CreateEmployeeRequest) (*hrv1.Employee, error) {
	dept, err := uuid.Parse(in.GetDepartmentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "department_id")
	}
	var uid *uuid.UUID
	if in.GetUserId() != "" {
		u, err := uuid.Parse(in.GetUserId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "user_id")
		}
		uid = &u
	}
	id, ts, err := s.st.CreateEmployee(ctx, dept, uid, in.GetFullName(), in.GetJobTitle())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	e := &hrv1.Employee{
		Id: id.String(), DepartmentId: in.GetDepartmentId(), FullName: in.GetFullName(), JobTitle: in.GetJobTitle(), CreatedAtUnix: ts,
	}
	if uid != nil {
		e.UserId = uid.String()
	}
	return e, nil
}

func (s *HRServer) GetEmployee(ctx context.Context, in *hrv1.GetEmployeeRequest) (*hrv1.Employee, error) {
	id, err := uuid.Parse(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id")
	}
	deptID, userID, name, title, ts, err := s.st.GetEmployee(ctx, id)
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "employee")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &hrv1.Employee{
		Id: in.GetId(), DepartmentId: deptID, UserId: userID, FullName: name, JobTitle: title, CreatedAtUnix: ts,
	}, nil
}

func (s *HRServer) ListEmployees(ctx context.Context, in *hrv1.ListEmployeesRequest) (*hrv1.ListEmployeesResponse, error) {
	limit, offset := in.GetLimit(), in.GetOffset()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.st.ListEmployees(ctx, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var list []*hrv1.Employee
	for _, r := range rows {
		list = append(list, &hrv1.Employee{
			Id: r.ID, DepartmentId: r.DeptID, UserId: r.UserID, FullName: r.FullName, JobTitle: r.JobTitle, CreatedAtUnix: r.Created,
		})
	}
	return &hrv1.ListEmployeesResponse{Employees: list}, nil
}
