.PHONY: build deploy clean

build:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bin/bootstrap ./main.go

deploy:
	cd cdk && cdk deploy --build

clean:
	rm -rf ./bin
