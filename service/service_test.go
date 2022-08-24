package service

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/rs/zerolog"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/yolkhovyy/golang-grpc-demo/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var listener *bufconn.Listener

func init() {
	// AWS session/config
	awsSession, err := session.NewSession()
	if err != nil {
		log.WithError(err).Fatal("AWS session create failed")
	}
	log.Info("AWS session created")
	awsConfig := aws.NewConfig().
		WithEndpoint("http://localhost:8000").
		WithRegion("us-west-2").
		WithCredentials(credentials.NewStaticCredentials("id", "secret", "token"))

	// Voting service server
	votingService := NewVotingServiceServer(context.Background(), &zerolog.Logger{}, dynamo.New(awsSession, awsConfig))
	grpcServer := grpc.NewServer()
	api.RegisterVotingServiceServer(grpcServer, votingService)

	go func() {
		listener = bufconn.Listen(bufSize)
		if err := grpcServer.Serve(listener); err != nil {
			log.WithError(err).Fatalf("gRPC serve failed")
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return listener.Dial()
}

func TestVotingServiceBasic(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial bufnet %v", err)
	}
	defer conn.Close()

	client := api.NewVotingServiceClient(conn)

	resVoteable, err := client.CreateVoteable(ctx, &api.CreateVoteableRequest{
		Question: "foo-0",
		Answers:  []string{"bar-0", "baz-0"},
	})

	require.NoError(t, err)
	_, err = uuid.Parse(resVoteable.Uuid)
	require.NoError(t, err)

	resListVoteables, err := client.ListVoteables(ctx, &api.ListVoteablesRequest{})
	require.NoError(t, err)
	l := len(resListVoteables.Votables)
	require.Greater(t, l, 0)
	for i := 0; i < l; i++ {
		if resListVoteables.Votables[i].Uuid == resVoteable.Uuid {
			require.Equal(t, "foo-0", resListVoteables.Votables[i].Question)
			require.Equal(t, "bar-0", resListVoteables.Votables[i].Answers[0])
			require.Equal(t, "baz-0", resListVoteables.Votables[i].Answers[1])
		}
	}

	_, err = client.CastVote(ctx, &api.CastVoteRequest{
		Uuid:        resVoteable.Uuid,
		AnswerIndex: 0,
	})
	require.NoError(t, err)

	// TODO The API provides no way to retieve number of votes
}

func TestVotingServicePaging(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial bufnet %v", err)
	}
	defer conn.Close()

	client := api.NewVotingServiceClient(conn)

	numOfVotes := 100
	for i := 0; i < numOfVotes; i++ {
		it := strconv.Itoa(i)
		resVoteable, err := client.CreateVoteable(ctx, &api.CreateVoteableRequest{Question: "foo-" + it, Answers: []string{"bar-" + it, "baz-" + it}})
		require.NoError(t, err)
		_, err = uuid.Parse(resVoteable.Uuid)
		require.NoError(t, err)
	}

	resListVoteables, err := client.ListVoteables(ctx, &api.ListVoteablesRequest{})
	require.NoError(t, err)
	// There are vote(s) created by the other test
	require.Greater(t, len(resListVoteables.Votables)%(numOfVotes+1), 0)

	const pageSize = 10
	var pagingKey []byte
	numOfPages := len(resListVoteables.Votables) / pageSize
	for i := 0; i <= numOfPages; i++ {
		resListVoteables, err := client.ListVoteables(ctx,
			&api.ListVoteablesRequest{PageSize: int64(pageSize), PagingKey: pagingKey})
		require.NoError(t, err)
		l := len(resListVoteables.Votables)
		if i < numOfPages {
			require.Equal(t, pageSize, l)
		} else {
			require.GreaterOrEqual(t, l, 0)
		}
		pagingKey = resListVoteables.PagingKey
	}

}
