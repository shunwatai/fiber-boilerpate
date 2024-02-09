.PHONY: dev build clean \
	docker-dev docker-dev-build docker-dev-up docker-dev-down docker-dev-log \
	docker-prod docker-prod-build docker-prod-up docker-prod-down docker-prod-log

dev:
	air -c .air.toml

build: 
	go build -race -o ./build/fiber-api .

# ref: https://unix.stackexchange.com/a/669683
clean: 
	find ./build/ -type f -executable -delete

docker-dev: docker-dev-build docker-dev-up    # docker-compose -f compose-dev.yaml up --build -d
docker-prod: docker-prod-build docker-prod-up # docker-compose -f compose-prod.yaml up --build -d

docker-dev-build: 
	docker-compose -f compose-dev.yaml build fiber-api-dev --build-arg UID=$$(id -u)

docker-prod-build: 
	docker-compose -f compose-prod.yaml build fiber-api-prod

docker-dev-up: 
	docker-compose -f compose-dev.yaml up -d

docker-prod-up: 
	docker-compose -f compose-prod.yaml up -d

docker-dev-down: 
	docker-compose -f compose-dev.yaml down

docker-prod-down: 
	docker-compose -f compose-prod.yaml down

docker-dev-log: 
	docker-compose -f compose-dev.yaml logs -f

docker-prod-log: 
	docker-compose -f compose-prod.yaml logs -f
