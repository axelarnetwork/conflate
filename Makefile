VERSION ?= 1.2.5
REPO_NAME ?= conflate
REGISTRY ?= ""

TAG := $(REPO_NAME):$(VERSION)
ifneq ($(REGISTRY), "")
	TAG := $(REGISTRY)/$(TAG)
endif

.PHONY: build
build:
	docker build --platform linux/amd64 --build-arg "VERSION=$(VERSION)" -t $(TAG) .

.PHONY: push
push: build
	docker push $(TAG)