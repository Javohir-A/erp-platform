.PHONY: proto tidy run-gateway docker-up

proto:
	cd proto && go run github.com/bufbuild/buf/cmd/buf@v1.47.2 generate
	cd genproto && go mod tidy

tidy:
	cd genproto && go mod tidy
	for d in services/auth-service services/gateway services/hr-service services/procurement-service services/warehouse-service services/finance-service; do \
		(cd $$d && go mod tidy); \
	done

docker-up:
	cd deploy && docker compose up --build

run-gateway:
	cd services/gateway && go run ./cmd
