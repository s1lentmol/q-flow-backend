package grpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	queuev1 "github.com/s1lentmol/q-flow-backend/protos/gen/go/queue"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/domain/models"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/services/queue"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Queue interface {
	ListQueues(ctx context.Context, group string) ([]models.Queue, error)
	CreateQueue(ctx context.Context, title, description, group string, mode models.QueueMode, ownerID int64) (models.Queue, error)
	GetQueue(ctx context.Context, queueID int64, group string) (models.Queue, []models.Participant, error)
	JoinQueue(ctx context.Context, queueID, userID int64, fullName string, group string, slotTime string) (int32, error)
	LeaveQueue(ctx context.Context, queueID, userID int64, group string) error
	AdvanceQueue(ctx context.Context, queueID int64, actorID int64, group string) (models.Participant, error)
	RemoveParticipant(ctx context.Context, queueID int64, userID int64, actorID int64, group string) error
	ArchiveQueue(ctx context.Context, queueID int64, actorID int64, group string) error
	DeleteQueue(ctx context.Context, queueID int64, actorID int64, group string) error
	UpdateQueue(ctx context.Context, queueID int64, actorID int64, group string, title string, description string) (models.Queue, error)
	AddParticipant(ctx context.Context, queueID int64, userID int64, fullName string, actorID int64, group string, slotTime string) (int32, error)
}

type serverAPI struct {
	queuev1.UnimplementedQueueServer
	queue Queue
}

func Register(gRPC *grpc.Server, q Queue) {
	queuev1.RegisterQueueServer(gRPC, &serverAPI{queue: q})
}

func (s *serverAPI) ListQueues(ctx context.Context, req *queuev1.ListQueuesRequest) (*queuev1.ListQueuesResponse, error) {
	input := struct {
		GroupCode string `validate:"required" json:"group_code"`
	}{
		GroupCode: req.GetGroupCode(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	queues, err := s.queue.ListQueues(ctx, req.GetGroupCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list queues")
	}

	resp := &queuev1.ListQueuesResponse{}
	for _, q := range queues {
		resp.Queues = append(resp.Queues, toQueueDTO(q))
	}
	return resp, nil
}

func (s *serverAPI) CreateQueue(ctx context.Context, req *queuev1.CreateQueueRequest) (*queuev1.CreateQueueResponse, error) {
	input := struct {
		Title       string            `validate:"required" json:"title"`
		Description string            `json:"description"`
		Mode        queuev1.QueueMode `validate:"required,gt=0" json:"mode"`
		GroupCode   string            `validate:"required" json:"group_code"`
		OwnerID     int64             `validate:"required,gt=0" json:"owner_id"`
	}{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Mode:        req.GetMode(),
		GroupCode:   req.GetGroupCode(),
		OwnerID:     req.GetOwnerId(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	queueModel, err := s.queue.CreateQueue(ctx, req.GetTitle(), req.GetDescription(), req.GetGroupCode(), toMode(req.GetMode()), req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create queue")
	}

	return &queuev1.CreateQueueResponse{Queue: toQueueDTO(queueModel)}, nil
}

func (s *serverAPI) GetQueue(ctx context.Context, req *queuev1.GetQueueRequest) (*queuev1.GetQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		GroupCode string `validate:"required" json:"group_code"`
	}{
		QueueID:   req.GetQueueId(),
		GroupCode: req.GetGroupCode(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	q, participants, err := s.queue.GetQueue(ctx, req.GetQueueId(), req.GetGroupCode())
	if err != nil {
		return nil, mapErr(err, "failed to get queue")
	}

	resp := &queuev1.GetQueueResponse{Queue: toQueueDTO(q)}
	for _, p := range participants {
		resp.Participants = append(resp.Participants, toParticipantDTO(p))
	}
	return resp, nil
}

func (s *serverAPI) JoinQueue(ctx context.Context, req *queuev1.JoinQueueRequest) (*queuev1.JoinQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		UserID    int64  `validate:"required,gt=0" json:"user_id"`
		GroupCode string `validate:"required" json:"group_code"`
		UserName  string `validate:"required" json:"user_name"`
	}{
		QueueID:   req.GetQueueId(),
		UserID:    req.GetUserId(),
		GroupCode: req.GetGroupCode(),
		UserName:  req.GetUserName(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	position, err := s.queue.JoinQueue(ctx, req.GetQueueId(), req.GetUserId(), req.GetUserName(), req.GetGroupCode(), req.GetSlotTime())
	if err != nil {
		return nil, mapErr(err, "failed to join queue")
	}

	return &queuev1.JoinQueueResponse{Position: position}, nil
}

func (s *serverAPI) LeaveQueue(ctx context.Context, req *queuev1.LeaveQueueRequest) (*queuev1.LeaveQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		UserID    int64  `validate:"required,gt=0" json:"user_id"`
		GroupCode string `validate:"required" json:"group_code"`
	}{
		QueueID:   req.GetQueueId(),
		UserID:    req.GetUserId(),
		GroupCode: req.GetGroupCode(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.queue.LeaveQueue(ctx, req.GetQueueId(), req.GetUserId(), req.GetGroupCode()); err != nil {
		return nil, mapErr(err, "failed to leave queue")
	}

	return &queuev1.LeaveQueueResponse{}, nil
}

func (s *serverAPI) AdvanceQueue(ctx context.Context, req *queuev1.AdvanceQueueRequest) (*queuev1.AdvanceQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		GroupCode string `validate:"required" json:"group_code"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
	}{
		QueueID:   req.GetQueueId(),
		GroupCode: req.GetGroupCode(),
		ActorID:   req.GetActorId(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	removed, err := s.queue.AdvanceQueue(ctx, req.GetQueueId(), req.GetActorId(), req.GetGroupCode())
	if err != nil {
		return nil, mapErr(err, "failed to advance queue")
	}

	return &queuev1.AdvanceQueueResponse{Removed: toParticipantDTO(removed)}, nil
}

func (s *serverAPI) RemoveParticipant(ctx context.Context, req *queuev1.RemoveParticipantRequest) (*queuev1.RemoveParticipantResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		UserID    int64  `validate:"required,gt=0" json:"user_id"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
		GroupCode string `validate:"required" json:"group_code"`
	}{
		QueueID:   req.GetQueueId(),
		UserID:    req.GetUserId(),
		ActorID:   req.GetActorId(),
		GroupCode: req.GetGroupCode(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.queue.RemoveParticipant(ctx, req.GetQueueId(), req.GetUserId(), req.GetActorId(), req.GetGroupCode()); err != nil {
		return nil, mapErr(err, "failed to remove participant")
	}

	return &queuev1.RemoveParticipantResponse{}, nil
}

func (s *serverAPI) ArchiveQueue(ctx context.Context, req *queuev1.ArchiveQueueRequest) (*queuev1.ArchiveQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		GroupCode string `validate:"required" json:"group_code"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
	}{
		QueueID:   req.GetQueueId(),
		GroupCode: req.GetGroupCode(),
		ActorID:   req.GetActorId(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.queue.ArchiveQueue(ctx, req.GetQueueId(), req.GetActorId(), req.GetGroupCode()); err != nil {
		return nil, mapErr(err, "failed to archive queue")
	}

	return &queuev1.ArchiveQueueResponse{}, nil
}

func (s *serverAPI) DeleteQueue(ctx context.Context, req *queuev1.DeleteQueueRequest) (*queuev1.DeleteQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		GroupCode string `validate:"required" json:"group_code"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
	}{
		QueueID:   req.GetQueueId(),
		GroupCode: req.GetGroupCode(),
		ActorID:   req.GetActorId(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	if err := s.queue.DeleteQueue(ctx, req.GetQueueId(), req.GetActorId(), req.GetGroupCode()); err != nil {
		return nil, mapErr(err, "failed to delete queue")
	}

	return &queuev1.DeleteQueueResponse{}, nil
}

func (s *serverAPI) UpdateQueue(ctx context.Context, req *queuev1.UpdateQueueRequest) (*queuev1.UpdateQueueResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		GroupCode string `validate:"required" json:"group_code"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
	}{
		QueueID:   req.GetQueueId(),
		GroupCode: req.GetGroupCode(),
		ActorID:   req.GetActorId(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	q, err := s.queue.UpdateQueue(ctx, req.GetQueueId(), req.GetActorId(), req.GetGroupCode(), req.GetTitle(), req.GetDescription())
	if err != nil {
		return nil, mapErr(err, "failed to update queue")
	}

	return &queuev1.UpdateQueueResponse{Queue: toQueueDTO(q)}, nil
}

func (s *serverAPI) AddParticipant(ctx context.Context, req *queuev1.AddParticipantRequest) (*queuev1.AddParticipantResponse, error) {
	input := struct {
		QueueID   int64  `validate:"required,gt=0" json:"queue_id"`
		UserID    int64  `validate:"required,gt=0" json:"user_id"`
		ActorID   int64  `validate:"required,gt=0" json:"actor_id"`
		GroupCode string `validate:"required" json:"group_code"`
		UserName  string `json:"user_name"`
	}{
		QueueID:   req.GetQueueId(),
		UserID:    req.GetUserId(),
		ActorID:   req.GetActorId(),
		GroupCode: req.GetGroupCode(),
		UserName:  req.GetUserName(),
	}
	if err := validate.Struct(input); err != nil {
		return nil, status.Error(codes.InvalidArgument, formatValidationError(err))
	}

	position, err := s.queue.AddParticipant(ctx, req.GetQueueId(), req.GetUserId(), req.GetUserName(), req.GetActorId(), req.GetGroupCode(), req.GetSlotTime())
	if err != nil {
		return nil, mapErr(err, "failed to add participant")
	}

	return &queuev1.AddParticipantResponse{Position: position}, nil
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

func toQueueDTO(q models.Queue) *queuev1.QueueDTO {
	return &queuev1.QueueDTO{
		Id:          q.ID,
		Title:       q.Title,
		Description: q.Description,
		Mode:        toProtoMode(q.Mode),
		Status:      toProtoStatus(q.Status),
		GroupCode:   q.GroupCode,
		OwnerId:     q.OwnerID,
		CreatedAt:   q.CreatedAt.Unix(),
		UpdatedAt:   q.UpdatedAt.Unix(),
	}
}

func toParticipantDTO(p models.Participant) *queuev1.ParticipantDTO {
	dto := &queuev1.ParticipantDTO{
		Id:        p.ID,
		QueueId:   p.QueueID,
		UserId:    p.UserID,
		Position:  p.Position,
		FullName:  p.FullName,
		CreatedAt: p.CreatedAt.Unix(),
	}
	if p.SlotTime != nil {
		dto.SlotTime = p.SlotTime.UTC().Format(time.RFC3339)
	}
	return dto
}

func toProtoMode(mode models.QueueMode) queuev1.QueueMode {
	switch mode {
	case models.ModeLive:
		return queuev1.QueueMode_QUEUE_MODE_LIVE
	case models.ModeManaged:
		return queuev1.QueueMode_QUEUE_MODE_MANAGED
	case models.ModeRandom:
		return queuev1.QueueMode_QUEUE_MODE_RANDOM
	case models.ModeSlots:
		return queuev1.QueueMode_QUEUE_MODE_SLOTS
	default:
		return queuev1.QueueMode_QUEUE_MODE_UNSPECIFIED
	}
}

func toMode(mode queuev1.QueueMode) models.QueueMode {
	switch mode {
	case queuev1.QueueMode_QUEUE_MODE_LIVE:
		return models.ModeLive
	case queuev1.QueueMode_QUEUE_MODE_MANAGED:
		return models.ModeManaged
	case queuev1.QueueMode_QUEUE_MODE_RANDOM:
		return models.ModeRandom
	case queuev1.QueueMode_QUEUE_MODE_SLOTS:
		return models.ModeSlots
	default:
		return models.ModeLive
	}
}

func toProtoStatus(status models.QueueStatus) queuev1.QueueStatus {
	switch status {
	case models.StatusActive:
		return queuev1.QueueStatus_QUEUE_STATUS_ACTIVE
	case models.StatusArchived:
		return queuev1.QueueStatus_QUEUE_STATUS_ARCHIVED
	default:
		return queuev1.QueueStatus_QUEUE_STATUS_UNSPECIFIED
	}
}

func mapErr(err error, fallback string) error {
	switch {
	case errors.Is(err, storage.ErrQueueNotFound):
		return status.Error(codes.NotFound, "queue not found")
	case errors.Is(err, storage.ErrParticipantExists):
		return status.Error(codes.AlreadyExists, "participant already exists")
	case errors.Is(err, storage.ErrParticipantMissing):
		return status.Error(codes.NotFound, "participant not found")
	case errors.Is(err, queue.ErrGroupMismatch):
		return status.Error(codes.PermissionDenied, "queue not in your group")
	case errors.Is(err, queue.ErrForbidden):
		return status.Error(codes.PermissionDenied, "not allowed")
	case errors.Is(err, queue.ErrSlotRequired):
		return status.Error(codes.InvalidArgument, "slot_time is required for slots")
	case errors.Is(err, queue.ErrQueueInactive):
		return status.Error(codes.FailedPrecondition, "queue is not active")
	default:
		return status.Error(codes.Internal, fallback)
	}
}
