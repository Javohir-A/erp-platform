package grpcsvc

import (
	"context"
	"time"

	"erp/platform/auth-service/config"
	"erp/platform/auth-service/internal/jwtutil"
	"erp/platform/auth-service/storage"
	authv1 "erp/platform/genproto/erp/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	cfg   config.Config
	store *storage.Store
}

func NewAuthServer(cfg config.Config, store *storage.Store) *AuthServer {
	return &AuthServer{cfg: cfg, store: store}
}

func (s *AuthServer) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	u, err := s.store.GetByEmail(ctx, in.GetEmail())
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !storage.CheckPassword(u.PasswordHash, in.GetPassword()) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	tok, err := jwtutil.Sign(u.ID, u.Role, s.cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.LoginResponse{
		AccessToken: tok,
		UserId:      u.ID.String(),
		Role:        u.Role,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, in *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := jwtutil.Parse(in.GetToken(), s.cfg.JWTSecret)
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Role:   claims.Role,
	}, nil
}

func (s *AuthServer) Register(ctx context.Context, in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if _, err := s.store.GetByEmail(ctx, in.GetEmail()); err == nil {
		return nil, status.Error(codes.AlreadyExists, "email taken")
	} else if err != storage.ErrNotFound {
		return nil, status.Error(codes.Internal, err.Error())
	}
	role := in.GetRole()
	if role == "" {
		role = "user"
	}
	id, err := s.store.CreateUser(ctx, in.GetEmail(), in.GetPassword(), role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.RegisterResponse{UserId: id.String()}, nil
}

// BootstrapAdmin ensures one admin exists (for first deploy).
func BootstrapAdmin(ctx context.Context, store *storage.Store, email, password string) error {
	_, err := store.GetByEmail(ctx, email)
	if err == nil {
		return nil
	}
	if err != storage.ErrNotFound {
		return err
	}
	_, err = store.CreateUser(ctx, email, password, "admin")
	return err
}
