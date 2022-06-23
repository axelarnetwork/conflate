VERSION ?= 1.2.4
REPO_NAME ?= conflate
REGISTRY ?= ""

TAG := $(REPO_NAME):$(VERSION)
ifneq ($(REGISTRY), "")
	TAG := $(REGISTRY)/$(TAG)
endif

.PHONY: build
build:
	docker build --build-arg "VERSION=$(VERSION)" -t $(TAG) .

.PHONY: push
push: build
	docker push $(TAG)