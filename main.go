package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("SDK config failed: %v", err)
	}
	client = s3.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	type Bucket struct {
		Name string `json:"name"`
	}

	type Object struct {
		Key       string `json:"key"`
		Size      int    `json:"size"`
		Etag      string `json:"etag"`
		VersionID string `json:"version-id"`
		Sequencer string `json:"sequencer"`
	}

	type Detail struct {
		Version         string `json:"version"`
		Bucket          Bucket `json:"bucket"`
		Object          Object `json:"object"`
		RequestID       string `json:"request-id"`
		Requester       string `json:"requester"`
		SourceIPAddress string `json:"source-ip-address"`
		Reason          string `json:"reason"`
	}

	type Newfile struct {
		Version    string   `json:"version"`
		ID         string   `json:"id"`
		DetailType string   `json:"detail-type"`
		Source     string   `json:"source"`
		Account    string   `json:"account"`
		Time       string   `json:"time"`
		Region     string   `json:"region"`
		Resources  []string `json:"resources"`
		Detail     Detail   `json:"detail"`
	}

	type Record struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		DOB       string `json:"dob"`
		Country   string `json:"country"`
	}

	var record Record
	var newfile Newfile

	err := json.Unmarshal(event, &newfile)
	if err != nil {
		fmt.Printf("Problem loading filename. %s\n", err)
	}

	file, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(newfile.Detail.Bucket.Name),
		Key:    aws.String(newfile.Detail.Object.Key),
	})
	if err != nil {
		fmt.Printf("JSON parse failed. %s %s %s", err, newfile.Detail.Object.Key, newfile.Detail.Bucket.Name)
	}

	fmt.Println(file)

	// capture all bytes from upload
	body, err := io.ReadAll(file.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(body), &record)
	if err != nil {
		fmt.Printf("Error loading file. %s\n", err)
	}

	firstname := []rune(record.Firstname)
	replacementValue := []rune("*******")
	anonFirstname := make([]rune, 0, 8)
	anonFirstname = append(anonFirstname, firstname[0])
	anonFirstname = append(anonFirstname, replacementValue...)
	record.Firstname = string(anonFirstname)

	lastname := []rune(record.Lastname)
	anonLastname := make([]rune, 0, 8)
	anonLastname = append(anonLastname, lastname[0])
	anonLastname = append(anonLastname, replacementValue...)
	record.Lastname = string(anonLastname)

	dob := record.DOB[0:4]
	record.DOB = dob

	jsonOutput, err := json.MarshalIndent(record, "", "	")
	if err != nil {
		fmt.Printf("JSON failed to be encoded. %s\n", err)
	}

	outputbucket := fmt.Sprintf("%s-output", newfile.Detail.Bucket.Name)
	log.Println(outputbucket)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(outputbucket),
		Key:    aws.String(newfile.Detail.Object.Key),
		Body:   bytes.NewReader(jsonOutput),
	})
	if err != nil {
		fmt.Printf("Error writing file: %s\n", err)
	}

	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(newfile.Detail.Bucket.Name),
		Key:    aws.String(newfile.Detail.Object.Key),
	})
	if err != nil {
		fmt.Printf("Unredacted file could not be deleted. %s\n", err)
	}

	return nil
}
