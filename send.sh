# /bin/bash
for ((i=1; i<=10; i++)); do
    grpcurl -plaintext -import-path ./api -proto service.proto  -d '{"question":"Do you like survey '${i}'?","answers":["yes","no"]}' localhost:3000 VotingService/CreateVoteable
done
grpcurl -plaintext -import-path ./api -proto service.proto  -d '{"page_size":10,"paging_key":""}' localhost:3000 VotingService/ListVoteables