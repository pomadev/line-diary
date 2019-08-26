.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./hello-world/hello-world

build:
	GOOS=linux GOARCH=amd64 go build -o line-diary/line-diary ./line-diary
	GOOS=linux GOARCH=amd64 go build -o line-notify/line-notify ./line-notify

# 追加
# TODO: s3作成処理をtemplate.yamlに引っ越し？
s3:
	aws s3 mb s3://pomadev-line-diary --profile pomadev

# 追加
package:
	sam package \
	--template-file ./template.yaml \
	--output-template-file ./output-template.yaml \
	--s3-bucket pomadev-line-diary \
	--profile pomadev

# 追加
deploy:
	sam deploy \
	--template-file ./output-template.yaml \
	--stack-name line-diary \
	--capabilities CAPABILITY_IAM \
	--profile pomadev

# 追加
# TODO: dynamodb作成処理をtemplate.yamlに引っ越し
dynamodb:
	aws dynamodb create-table \
	--table-name Diary \
	--key-schema AttributeName=LineID,KeyType=HASH AttributeName=Date,KeyType=RANGE \
	--attribute-definitions AttributeName=LineID,AttributeType=S AttributeName=Date,AttributeType=S \
	--provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
	--profile pomadev
