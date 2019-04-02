package database

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/strongjz/contino-bucks/cbuck"
	"log"
)

var (
	buf               bytes.Buffer
	debug    bool
	logger            = log.New(&buf, "logger: ", log.LstdFlags)
)

type DB struct{

	svc *dynamodb.DynamoDB
}

func New(endpoint string) *DB {

	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(endpoint)},
	)

	if err != nil {
		logger.Fatalf("Could not creation to Database: %s",err)
	}

	return &DB{
		svc:dynamodb.New(sess),
	}


}

func (d *DB) CreateTables(){

	// Create table Contino_Bucks_Total
	tableName := "Contino_Bucks_Total"

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("receiver"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("receiver"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("total"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_, err := d.svc.CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
	}

}

func (d *DB) createItem(g *cbuck.Gift){


}