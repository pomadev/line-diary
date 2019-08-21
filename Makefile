.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./hello-world/hello-world

build:
	GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world

# 追加
package:
	sam package \
	--template-file sam-app/template.yaml \
	--output-template-file sam-app/output-template.yaml \
	--s3-bucket template-store \
	--profile pomadev

# 追加
deploy:
	sam deploy \
	--template-file sam-app/output-template.yaml \
	--stack-name go-lambda \
	--capabilities CAPABILITY_IAM \
	--profile pomadev

# 追加
dynamodb:
	aws dynamodb create-table \
	--table-name Diary \
	--key-schema AttributeName=LineID,KeyType=HASH AttributeName=Date,KeyType=RANGE \
	--attribute-definitions AttributeName=LineID,AttributeType=S AttributeName=Date,AttributeType=S \
	--provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
	--profile pomadev
