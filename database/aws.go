package database

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

var (
	buf        bytes.Buffer
	debug      bool
	logger     = log.New(&buf, "logger: ", log.LstdFlags)
	totalTable = "Bucks_Total"
)

type Gift struct {
	Receiver string
	Giver    string
	Amount   float64
}

type DB struct {
	svc *dynamodb.DynamoDB
}

func New(endpoint string) *DB {

	logger.Printf("[INFO] Creating new DB connection to endpoint %s", endpoint)

	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(endpoint)},
	)
	if err != nil {
		logger.Printf("[ERROR] Could not connect to Database: %s", err)
		return nil
	}

	db := DB{
		svc: dynamodb.New(sess),
	}

	return &db

}

func (d *DB) createTables() error {

	logger.Print("Creating Tables")
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Receiver"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("Total"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Receiver"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Total"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(totalTable),
	}

	_, err := d.svc.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceInUseException:
				logger.Print(dynamodb.ErrCodeResourceInUseException, aerr.Error())
				return nil //Table already exists
			case dynamodb.ErrCodeLimitExceededException:
				logger.Print(dynamodb.ErrCodeLimitExceededException, aerr.Error())
				return err
			case dynamodb.ErrCodeInternalServerError:
				logger.Print(dynamodb.ErrCodeInternalServerError, aerr.Error())
				return err
			default:
				logger.Print(aerr.Error())
				return err
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Print(err.Error())
			return err
		}
	}

	return nil

}

func (d *DB) WriteGift(g *Gift) error {

	//get receiver current total
	oldGift, err := d.getGift(g.Receiver)
	if err != nil {
		//first time getting bucks
		err = d.updateGift(g)
		if err != nil {
			logger.Printf(fmt.Sprintf("[ERROR] Updating Databases, %v", err))
			return err
		}
	}

	//update it
	newAmount := oldGift.Amount + g.Amount
	g.Amount = newAmount

	//write it back to db
	err = d.updateGift(g)
	if err != nil {
		logger.Printf(fmt.Sprintf("[ERROR] Updating Databases, %v", err))
		return err
	}

	return nil
}

func (d *DB) getGift(id string) (*Gift, error) {

	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(totalTable),
		Key: map[string]*dynamodb.AttributeValue{
			"Receiver": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	gift := Gift{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &gift)
	if err != nil {
		logger.Panicf(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		return nil, err
	}

	return &gift, nil
}

func (d *DB) updateGift(g *Gift) error {

	av, err := dynamodbattribute.MarshalMap(g)
	if err != nil {
		logger.Printf("Got error marshalling new movie item:")
		logger.Printf(err.Error())
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Contino_Bucks_Total"),
	}

	_, err = d.svc.PutItem(input)
	if err != nil {
		logger.Printf("Got error calling PutItem:")
		logger.Printf(err.Error())
		return err
	}

	return nil
}
