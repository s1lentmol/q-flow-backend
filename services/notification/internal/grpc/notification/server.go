package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	notificationv1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Notification interface {
	SetContact(ctx context.Context, userID int64, username, chatID string) error
	NotifyPositionSoon(ctx context.Context, userID int64, queueTitle string, position int32) error
}

type serverAPI struct {
	notificationv1.UnimplementedNotificationServer
	notif Notification
}

func Register(gRPC *grpc.Server, notif Notification) {
	notificationv1.RegisterNotificationServer(gRPC, &serverAPI{notif: notif})
}

func (s *serverAPI) NotifyPositionSoon(ctx context.Context, req *notificationv1.NotifyPositionSoonRequest) (*notificationv1.NotifyPositionSoonResponse, error) {
	input := struct {
		UserID     int64  `validate:"required,gt=0" json:"user_id"`
		QueueTitle string `validate:"required" json:"queue_title"`
		Position   int32  `validate:"required,gt=0" json:"position"`
	}{
		UserID:     req.GetUserId(),
		QueueTitle: req.GetQueueTitle(),
		Position:   req.GetPosition(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.notif.NotifyPositionSoon(ctx, req.GetUserId(), req.GetQueueTitle(), req.GetPosition()); err != nil {
		return nil, status.Error(codes.Internal, "failed to send notification")
	}

	return &notificationv1.NotifyPositionSoonResponse{}, nil
}

func (s *serverAPI) SetContact(ctx context.Context, req *notificationv1.SetContactRequest) (*notificationv1.SetContactResponse, error) {
	input := struct {
		UserID   int64  `validate:"required,gt=0" json:"user_id"`
		Username string `json:"telegram_username"`
	}{
		UserID:   req.GetUserId(),
		Username: req.GetTelegramUsername(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.notif.SetContact(ctx, req.GetUserId(), req.GetTelegramUsername(), req.GetChatId()); err != nil {
		return nil, status.Error(codes.Internal, "failed to set contact")
	}

	return &notificationv1.SetContactResponse{}, nil
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
