package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

	type Newfile struct {
		Detail struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"detail"`
	}

	var newfile Newfile

	filejson, err := os.ReadFile("s3.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal([]byte(filejson), &newfile)

	fmt.Println(newfile.Detail.Object)

	//_, err := client.ListObjectsV2()
}

// func uploadToS3() {
// 	_, err := client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket: "test",
// 		Key:    "test",
// 		Body:   "test",
// 	})
// 	if err != nil {
// 		log.Fatalf("File upload failed: %v", err)
// 	}
// }

func handleRequest(ctx context.Context, event json.RawMessage) error {

	filejson, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Println(err)
	}

	type Record struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		DOB       string `json:"dob"`
		Country   string `json:"country"`
	}

	var record Record

	err = json.Unmarshal([]byte(filejson), &record)
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

	err = os.WriteFile("anondata.json", jsonOutput, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %s\n", err)
	}
	return nil
}
