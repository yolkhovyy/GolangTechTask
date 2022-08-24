package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/guregu/dynamo"

	"github.com/google/uuid"
	"github.com/yolkhovyy/golang-grpc-demo/api"
)

const tableName = "voteables"

//nolint:gochecknoglobals
var tracer = otel.Tracer("yolkhovyy/GolangTectTask/server")

type VotingServiceServer struct {
	api.UnimplementedVotingServiceServer
	db     *dynamo.DB
	logger *zerolog.Logger
}

func NewVotingServiceServer(ctx context.Context, logger *zerolog.Logger, db *dynamo.DB) *VotingServiceServer {
	ctx, span := tracer.Start(ctx, "server.NewVotingServiceServer")
	defer span.End()

	if err := db.CreateTable(tableName, Voteable{}).RunWithContext(ctx); err != nil {
		if err.(awserr.Error).Code() == "ResourceInUseException" {
			logger.Info().Msgf("table %s exists", tableName)
		} else {
			span.RecordError(err)
			logger.Fatal().Err(err).Msgf("create %s table failed", tableName)
		}
	} else {
		logger.Debug().Msgf("created %s table", tableName)
	}

	return &VotingServiceServer{db: db, logger: logger}
}

func (s *VotingServiceServer) CreateVoteable(ctx context.Context, req *api.CreateVoteableRequest) (*api.CreateVoteableResponse, error) {
	ctx, span := tracer.Start(ctx, "server.CreateVoteable")
	defer span.End()

	uuidVoteable, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	err = s.db.Table(tableName).
		Put(Voteable{
			UUID:     uuidVoteable.String(),
			Question: req.Question,
			Answers:  req.Answers,
			Votes:    make([]int64, len(req.Answers)),
		}).
		RunWithContext(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	votable := &api.CreateVoteableResponse{
		Uuid: uuidVoteable.String(),
	}
	s.logger.Debug().Msgf("created votable %s", votable.Uuid)
	span.SetAttributes(attribute.String("votable", votable.Uuid))
	return votable, nil
}

func (s *VotingServiceServer) ListVoteables(ctx context.Context, req *api.ListVoteablesRequest) (*api.ListVoteablesResponse, error) {
	ctx, span := tracer.Start(ctx, "server.ListVoteables")
	defer span.End()

	scan := s.db.Table(tableName).Scan()
	if req.PageSize > 0 {
		scan = scan.SearchLimit(req.PageSize)
	}

	if len(req.PagingKey) > 0 {
		var pagingKey dynamo.PagingKey
		if err := json.Unmarshal(req.PagingKey, &pagingKey); err != nil {
			span.RecordError(err)
			return nil, err
		}
		scan = scan.StartFrom(pagingKey)
	}

	var results []Voteable
	lastEvaluatedKey, err := scan.AllWithLastEvaluatedKeyContext(ctx, &results)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	pagingKey, err := json.Marshal(lastEvaluatedKey)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var voteables []*api.Voteable
	for _, r := range results {
		var answers = make([]string, len(r.Answers))
		copy(answers, r.Answers)
		voteables = append(voteables, &api.Voteable{
			Uuid:     r.UUID,
			Question: r.Question,
			Answers:  answers,
		})
	}

	return &api.ListVoteablesResponse{
		Votables:  voteables,
		PagingKey: pagingKey,
	}, nil
}

func (s VotingServiceServer) CastVote(ctx context.Context, req *api.CastVoteRequest) (*api.CastVoteResponse, error) {
	ctx, span := tracer.Start(ctx, "server.CastVote")
	defer span.End()
	span.SetAttributes(attribute.String("Uuid", req.Uuid), attribute.String("AnswerIndex", strconv.Itoa(int(req.AnswerIndex))))

	err := s.db.Table(tableName).
		Update("ID", req.Uuid).
		SetExpr("Votes[$] = Votes[$] + ?", req.AnswerIndex, req.AnswerIndex, 1).
		RunWithContext(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &api.CastVoteResponse{}, nil
}
