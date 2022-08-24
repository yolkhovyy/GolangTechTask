# Golang gRPC Demo

Forked from [here](https://github.com/buffup/golang-grpc-demo)

## Content

- `api/` - gRPC service definition for a voting service, contains RPCs for creating, listing and voting on `voteable` items
- `cmd/main.go` & `service/service.go` - service implementation
- `config/config.go` - service configuration
- `telemetry` - metrics, tracing and profiling
- `docker-compose.yml` - for running `Amazon DynamoDB` locally
## Running

Install

- [Docker](https://docs.docker.com/get-docker/)
- [Go](https://go.dev/doc/install)
- [Protocol Buffer Compiler](https://grpc.io/docs/protoc-installation/)
- [gRPCurl](https://github.com/fullstorydev/grpcurl)

Clone

```shell
git clone git@github.com:yolkhovyy/golang-grpc-demo.git
```

### Start

```shell
$ ./start.sh
{"level":"info","time":"2022-08-02T11:29:03+02:00","message":"configuration loaded"}                                                                          
{"level":"info","service":"ggd","time":"2022-08-02T11:29:03+02:00","time":"2022-08-02T11:29:03+02:00","message":"AWS session created"}
...
```
### Send messages

In the second shell

```shell
$ ./send.sh
{
  "uuid": "aa271b8e-1245-11ed-8692-02426fbaa4b9"
}
{
  "uuid": "aa2df569-1245-11ed-8692-02426fbaa4b9"
}
...
```
### Test

In the second shell

```shell
$ go test -v -count=1 ./...
=== RUN   TestVotingServiceBasic
--- PASS: TestVotingServiceBasic (0.10s)
=== RUN   TestVotingServicePaging
--- PASS: TestVotingServicePaging (0.28s)
PASS
ok      github.com/yolkhovyy/golang-grpc-demo/service   0.407s
```

### Stop

```shell
./stop.sh
```