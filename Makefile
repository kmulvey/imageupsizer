REPOPATH = github.com/kmulvey/imagedup
BUILDS := auto imageupsizer manual verify

build: 
	for target in $(BUILDS); do \
		go build -v -x -ldflags="-s -w" -o ./cmd/$$target ./cmd/$$target; \
	done
