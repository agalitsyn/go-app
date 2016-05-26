APPLICATION ?= $$(basename $(CURDIR))
BUILD_DIR ?= bin

IMAGE ?= goapi-web:latest
REGISTRY ?=

DOCKER_RUN_OPTS := --publish 5000:5000 \
					--env LOG_LEVEL=debug \
					--env DATABASE_URL=postgres://docker:docker@172.17.0.1:5432/docker?sslmode=disable&connect_timeout=1
DOCKER_RUN_ARGS :=
