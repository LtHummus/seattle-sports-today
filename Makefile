.PHONY: build deploy clean

build:
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags "-s -w" -trimpath -o bin/bootstrap ./main.go

deploy:
	cd cdk && cdk deploy --build

clean:
	rm -rf ./bin
