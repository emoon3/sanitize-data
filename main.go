package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

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
}

func handleRequest(ctx context.Context, event json.RawMessage) error {

	type Newfile struct {
		Detail struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"detail"`
	}

	type Record struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		DOB       string `json:"dob"`
		Country   string `json:"country"`
	}

	var record Record
	var newfile Newfile
	outputbucket := newfile.Detail.Bucket.Name + "-output"

	err := json.Unmarshal([]byte(event), &newfile)
	if err != nil {
		fmt.Printf("Problem loading filename. %s\n", err)
	}

	fileInput := &s3.GetObjectInput{
		Bucket: aws.String(newfile.Detail.Bucket.Name),
		Key:    aws.String(newfile.Detail.Object.Key),
	}

	file, err := client.GetObject(ctx, fileInput)
	if err != nil {
		fmt.Printf("JSON failed to be parsed. %s\n", err)
	}

	// capture all bytes from upload
	body, err := io.ReadAll(file.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(body), &record)
	if err != nil {
		fmt.Printf("JSON failed to be parsed. %s\n", err)
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
		fmt.Printf("JSON failed to be parsed. %s\n", err)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(outputbucket),
		Key:    aws.String(newfile.Detail.Object.Key),
		Body:   bytes.NewReader(jsonOutput),
	})
	if err != nil {
		fmt.Printf("Error writing file: %s\n", err)
	}
	return nil
}
