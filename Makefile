SHELL := /bin/bash

ifeq ($(OS),Windows_NT)
    OPEN_CMD = cmd /c start
else
    OPEN_CMD = xdg-open
endif

all: up open

up:
	docker-compose up -d

open:
	$(OPEN_CMD) http://localhost:8080

stop:
	docker-compose down
