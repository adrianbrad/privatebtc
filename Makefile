lint:
	golangci-lint run --fix

test:
	go test -mod=mod -shuffle=on -race -v .

test-ci:
	go test -mod=mod -shuffle=on -race -timeout 3000s -coverprofile=coverage.txt -covermode=atomic .

benchmark:
	go test -bench=.  -benchmem

.PHONY: mocks
mocks:
	moq -rm -out ./mock/mock_rpc_client.go -pkg mock . RPCClient:RPCClient
	moq -rm -out ./mock/mock_rpc_client_factory.go -pkg mock . RPCClientFactory:RPCClientFactory
	moq -rm -out ./mock/mock_node_service.go -pkg mock . NodeService:NodeService
	moq -rm -out ./mock/mock_node_handler.go -pkg mock . NodeHandler:NodeHandler
