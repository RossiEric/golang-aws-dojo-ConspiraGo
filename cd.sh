#!/bin/bash

# Nome da função Lambda
FUNCTION_NAME="ConspiraGo-Lambda"

# Nome do bucket do Amazon S3 para armazenar o pacote da função
S3_BUCKET="conspirago-documents"

# Nome do arquivo .zip que conterá o código da função
ZIP_FILE="function.zip"

# Caminho para o código-fonte da função em Go
CODE_PATH="./cmd/analyzer-lambda"

# Compila o código em um arquivo .zip
cd $CODE_PATH
GOOS=linux go build -o main
zip $ZIP_FILE main

# Envia o arquivo .zip para o bucket do Amazon S3
aws s3 cp $ZIP_FILE s3://$S3_BUCKET/$ZIP_FILE

# Cria a função Lambda
aws lambda create-function \
  --function-name $FUNCTION_NAME \
  --runtime go1.x \
  --handler main \
  --code S3Bucket=$S3_BUCKET,S3Key=$ZIP_FILE \
  --role arn:aws:iam::369407194787:role/tmp-dojo-golang

# Concede permissões completas de acesso ao Amazon S3
aws lambda add-permission \
  --function-name $FUNCTION_NAME \
  --action lambda:InvokeFunction \
  --principal s3.amazonaws.com \
  --source-arn arn:aws:s3:::$S3_BUCKET \
  --statement-id lambda-s3 \
  --source-account 369407194787

# Limpa os arquivos temporários
rm $ZIP_FILE main
