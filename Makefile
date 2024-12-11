run:
	go run main.go

up:
	migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path migrations up

down:
	migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -path migrations down

test:
	go test ./...