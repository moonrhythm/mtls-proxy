COMMIT_SHA=$(shell git rev-parse HEAD)

build:
	buildctl build \
		--frontend dockerfile.v0 \
		--local context=. \
		--local dockerfile=. \
		--output type=image,name=gcr.io/moonrhythm-containers/mtls-proxy:$(COMMIT_SHA),push=true

	buildctl build \
		--frontend dockerfile.v0 \
		--local context=. \
		--local dockerfile=. \
		--output type=image,name=gcr.io/moonrhythm-containers/mtls-proxy:latest,push=true
