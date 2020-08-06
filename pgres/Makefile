
.PHONY: run stop clean attach container

run:
	docker container start dev-postgres

container:
	docker run -d --name dev-postgres -e POSTGRES_PASSWORD=hallo123 -v ${pwd}/postgres-data:/var/lib/postgresql/data -p 5432:5432 pgres_custom

build: Dockerfile
	docker build -t pgres_custom .
attach:
	docker exec -it dev-postgres bash
stop:
	docker container stop dev-postgres
clean:
	docker container rm dev-postgres
