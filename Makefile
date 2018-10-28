REPOSITORY?=mixmaxhq/jqd
TAG?=latest

build:
	echo "building $(REPOSITORY):$(TAG)..."
	# Build it
	GOOS=linux GOARCH=amd64 go build -o jqd .

	# Make it executable
	chmod +x jqd

	# Build the docker image
	docker build --rm -t $(REPOSITORY):$(TAG) .

push: build
	echo "pushing $(REPOSITORY):$(TAG)..."
	docker push $(REPOSITORY):$(TAG)


test: build
	echo "starting $(REPOSITORY):$(TAG)..."
	docker run -it -p 9090:9090/tcp \
		-e JIRA_API_TOKEN=$(JIRA_API_TOKEN) \
		-e JIRA_USER=$(JIRA_USER) \
		-e JIRA_BASE_URL=$(JIRA_BASE_URL) $(REPOSITORY):$(TAG)

.PHONY: build