SHELL := cmd

.PHONY: all start db stop resume clean

all: start

start:
	docker-compose up -d --build

db:
	docker-compose up -d --build pgsql pgadmin

clean:
	docker-compose down --rmi local --remove-orphans

purge:
	docker-compose down --rmi local --volumes --remove-orphans

stop:
	docker-compose stop

resume:
	docker-compose start

status:
	docker-compose ps
