package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pereslava/grpc_zerolog"
	"github.com/rs/zerolog/log"

	"github.com/yolkhovyy/golang-grpc-demo/api"
	"github.com/yolkhovyy/golang-grpc-demo/config"
	"github.com/yolkhovyy/golang-grpc-demo/service"
	"github.com/yolkhovyy/golang-grpc-demo/telemetry"
)

const (
	serviceTag      = "service"
	serviceTagValue = "ggd"
)

var (
	configFile = kingpin.Flag("config", "Config file").Default("config.yml").ExistingFile()
)

func main() {
	// options
	switch kingpin.Parse() {
	case "config":
		log.Info().Msgf("using config file %s", *configFile)
	}

	// configuration
	err := config.Service.Load(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("configuration load failed")
	}
	log.Info().Msg("configuration loaded")

	// logger
	logger := log.Level(config.Service.LogLevel).With().
		Timestamp().Str(serviceTag, serviceTagValue).Logger()

	// AWS session/configuration
	awsSession, err := session.NewSession()
	if err != nil {
		logger.Fatal().Err(err).Msg("AWS session create failed")
	}
	logger.Info().Msg("AWS session created")
	awsConfig := aws.Config{
		Region:   aws.String(config.Service.AWS.Region),
		Endpoint: aws.String("http://" + config.Service.DB.Host + ":" + strconv.Itoa(config.Service.DB.Port)),
		Credentials: credentials.NewStaticCredentials(config.Service.AWS.Id,
			config.Service.AWS.Secret, config.Service.AWS.Token),
	}

	// trace provider
	shutdown, err := telemetry.StartTrace()
	if err != nil {
		logger.Fatal().Err(err).Msg("trace provider start failed")
	}
	defer shutdown()

	// metrics
	metrics, err := telemetry.NewMetricsServer()
	if err != nil {
		logger.Fatal().Err(err).Msg("metrics server create failed")
	}
	defer metrics.Shutdown(config.Service.Metrics.ShutdownTimeout)
	go metrics.Serve()

	// profiler
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	// profiler
	_, err = telemetry.PyroscopeStart(config.Service.Profiler)
	if err != nil {
		logger.Error().Err(err).Msg("profiler start failed")
	}

	// gRPC server options
	grpcServerOptions := []grpc.ServerOption{
		// zerolog
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(logger),
			grpc_zerolog.NewUnaryServerInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewPayloadStreamServerInterceptor(logger),
			grpc_zerolog.NewStreamServerInterceptor(logger),
		),
		// opentelemetry
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	// replace internal gRPC library logger with zerolog
	// https://pkg.go.dev/github.com/pereslava/grpc_zerolog
	grpc_zerolog.ReplaceGrpcLogger(logger)

	// gRPC server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	grpcServer := grpc.NewServer(grpcServerOptions...)
	defer func() {
		logger.Info().Msg("stopping grpc server")
		grpcServer.GracefulStop()
	}()
	if config.Service.GRPC.Reflection {
		reflection.Register(grpcServer)
	}

	// voting service
	votingService := service.NewVotingServiceServer(ctx, &logger, dynamo.New(awsSession, &awsConfig))
	api.RegisterVotingServiceServer(grpcServer, votingService)
	logger.Info().Msg("registered voting service")

	// listen
	address := net.JoinHostPort(config.Service.GRPC.Host, strconv.Itoa(config.Service.GRPC.Port))
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		logger.Fatal().Err(err).Msg("listen failed")
	}
	defer listener.Close()
	logger.Info().Msgf("listening on %v", listener.Addr().String())

	// serve gRPC
	chErr := make(chan error)
	go func() {
		chErr <- grpcServer.Serve(listener)
	}()

	// shutdown
	chStop := make(chan os.Signal, 1)
	signal.Notify(chStop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-chErr:
		logger.Fatal().Err(err).Msg("gRPC server error")
	case signal := <-chStop:
		log.Info().Msgf("received %v signal", signal)
	}
}
