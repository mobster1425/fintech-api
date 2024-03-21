# Define variables
CONTAINER_NAME = digital-fintech-api
DB_NAME = digital-fintech-api-db
DB_USER = postgres
OWNER=postgres
name=init_schema
# Makefile rule to create the database
createdb:
	docker exec -it $(CONTAINER_NAME) createdb --username=$(DB_USER) --owner=$(OWNER) $(DB_NAME)

dropdb:
	docker exec -it $(CONTAINER_NAME) dropdb --username=$(DB_USER) $(DB_NAME)

postgres:
	docker run --name digital-fintech-api -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres:14-alpine

migrateup:
	migrate -path db/migrations -database "postgresql://postgres:mysecretpassword@localhost:5432/digital-fintech-api-db?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://postgres:mysecretpassword@localhost:5432/digital-fintech-api-db?sslmode=disable" -verbose down

runmigration:
	migrate -path db/migrations -database "postgresql://postgres:mysecretpassword@localhost:5432/digital-fintech-api-db?sslmode=disable" -verbose up $(VERSION)


rundownmigration:
	migrate -path db/migrations -database "postgresql://postgres:mysecretpassword@localhost:5432/digital-fintech-api-db?sslmode=disable" -verbose down $(VERSION)

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

mock:
	mockgen -package mockdb -destination db/mock/store.go feyin/digital-fintech-api/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go feyin/digital-fintech-api/worker TaskDistributor

.PHONY: createdb dropdb postgres migrateup migratedown runmigration rundownmigration sqlc test new_migration redis mock
