package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	authv1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/auth"
	authsvc "github.com/s1lentmol/q-flow-backend/services/auth/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(ctx context.Context,
		userID int64,
	) (bool, error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {

	input := struct {
		Email    string `validate:"required,email" json:"email"`
		Password string `validate:"required" json:"password"`
		AppID    int32  `validate:"required,gt=0" json:"app_id"`
	}{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		AppID:    req.GetAppId(),
	}

	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, mapAuthErr(err, "failed to login user")
	}

	return &authv1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	input := struct {
		Email    string `validate:"required,email" json:"email"`
		Password string `validate:"required" json:"password"`
	}{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthErr(err, "failed to register user")
	}

	return &authv1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *authv1.IsAdminRequest,
) (*authv1.IsAdminResponse, error) {
	input := struct {
		UserID int64 `validate:"required,gt=0" json:"user_id"`
	}{
		UserID: req.GetUserId(),
	}

	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())

	if err != nil {
		return nil, mapAuthErr(err, "failed to check admin status")
	}

	return &authv1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" {
			return fld.Name
		}
		return name
	})
	return v
}()

func formatValidationError(err error) string {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		msgs := make([]string, 0, len(verrs))
		for _, ve := range verrs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", ve.Field(), ve.Tag()))
		}
		return strings.Join(msgs, "; ")
	}
	return err.Error()
}

func mapAuthErr(err error, fallback string) error {
	switch {
	case errors.Is(err, authsvc.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, authsvc.ErrInvalidAppID):
		return status.Error(codes.InvalidArgument, "invalid app id")
	case errors.Is(err, authsvc.ErrUserExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	default:
		return status.Error(codes.Internal, fallback)
	}
}
