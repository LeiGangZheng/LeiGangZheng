help:
	@sleep 1

test-create:
	docker run --name test-db \
		-e MYSQL_ROOT_PASSWORD=db123456 \
		-e MYSQL_DATABASE=db_test \
		-p 3306:3306 \
		-d hub.deepin.com/library/mariadb:latest

test-clean:
	-docker stop test-db
	-docker rm   test-db

test:
	go test
