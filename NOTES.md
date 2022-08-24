# Notes

## Compile api/service.proto

```shell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/service.proto
```

## AWS DynamoDB

Configurte AWS

```shell
sudo apt install awscli
$ aws configure
AWS Access Key ID [None]: id
AWS Secret Access Key [None]: secret
Default region name [None]: token
Default output format [None]: 
```

DynamoDB

```
aws dynamodb scan --table-name voteables \
--select ALL_ATTRIBUTES \
--endpoint-url="http://localhost:8000" --region us-west-2
```

## Install grpcurl

```shell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

## grpcurl

```shell
# Reflection enabled
$ grpcurl -plaintext localhost:3000 list
VotingService
grpc.reflection.v1alpha.ServerReflection
# ...

# Using proto services
$ grpcurl -import-path ./api -proto service.proto list
VotingService
$ grpcurl -import-path ./api -proto service.proto list VotingService
VotingService.CastVote
VotingService.CreateVoteable
VotingService.ListVoteables

$ grpcurl -plaintext -import-path ./api -proto service.proto  -d '{"question":"Do you like this survey?","answers":["yes","no"]}' localhost:3000 VotingService/CreateVoteable
{
  "uuid": "bdd942fb-0f03-11ed-990e-0242633d156d"
}

$ grpcurl -plaintext -import-path ./api -proto service.proto  -d '{"uuid":"bdd942fb-0f03-11ed-990e-0242633d156d","answer_index":1}' localhost:3000 VotingService/CastVote
{
}

$ grpcurl -plaintext -import-path ./api -proto service.proto  -d '{"page_size":10,"paging_key":""}' localhost:3000 VotingService/ListVoteables
{
  "votables": [
    {
      "uuid": "bdd942fb-0f03-11ed-990e-0242633d156d",
      "question": "Do you like this survey?",
      "answers": [
        "yes",
        "no"
      ]
    }
  ],
"pagingKey": "eyJJRCI6eyJCIjpudWxsLCJCT09MIjpudWxsLCJCUyI6bnVsbCwiTCI6bnVsbCwiTSI6bnVsbCwiTiI6bnVsbCwiTlMiOm51bGwsIk5VTEwiOm51bGwsIlMiOiJmMDIyOTI4NS0wZjA3LTExZWQtOTkwZS0wMjQyNjMzZDE1NmQiLCJTUyI6bnVsbH19" <- MongoDB pagination key (base64-encoded), there are more data
  "pagingKey": "bnVsbA==" <- null (base64 encoded), no more data
}


```

## Go profiling

### Google Go Profiler

Profiling outside of [Google Cloud](https://cloud.google.com/profiler/docs/profiling-external)

Configuration

```yml
# https://cloud.google.com/profiler/docs/profiling-go
# https://github.com/googleapis/google-cloud-go/blob/4a5b2449980c5b6b31ba95326237c61f70ae1702/profiler/profiler.go#L108
GoProfiler:
  Service: ""
  ProjetcID: ""
  DebugLogging: false
  MutexProfiling: false
  NoCPUProfiling: false
  NoAllocProfiling: false
  AllocForceGC: false
  NoHeapProfiling: false
  NoGoroutineProfiling: false
```

Starting profiler

```go
import "cloud.google.com/go/profiler"

	// Google Go profiler
	// https://cloud.google.com/profiler/docs/profiling-go
	if config.Service.GoProfiler.Service != "" {
		err = profiler.Start(config.Service.GoProfiler)
		if err != nil {
			logger.Error().Err(err).Msg("Google Go profiler start failed")
		}
	}
```

### Pyroscope

Pyroscope [docs](https://pyroscope.io/docs/)

Pyroscope [docker guide](https://pyroscope.io/docs/docker-guide/)

Starting pyroscope with docker compose

```yml
service:
  pyroscope:
    image: "pyroscope/pyroscope:latest"
    ports:
      - "4040:4040"
    command:
      - "server"
```

Starting Pyroscope [Golang](https://pyroscope.io/docs/golang/)

```go
import "github.com/pyroscope-io/client"

  // These 2 lines are only required if you're using mutex or block profiling
	// Read the explanation below for how to set these rates:
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	_, err = pyroscope.Start(pyroscope.Config{
		ApplicationName: "gtt",

		// replace this with the address of pyroscope server
		ServerAddress: "http://localhost:4040",

		// you can disable logging by setting this to nil
		// Logger: pyroscope.StandardLogger,
		Logger: nil,

		// optionally, if authentication is enabled, specify the API key:
		// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Pyroscope profiler start failed")
	}
```

