/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Comment
type SmartContract struct {
	contractapi.Contract
}

// Comment describes basic details of what makes up a Comment
type Comment struct {
	User    string `json:"user"`
	Text    string `json:"text"`
	TopicID string `json:topicID`
	PostID  string `json:postID`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Comment
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	comments := []Comment{
		Comment{User: "ADSR", Text: "4chan", TopicID: "TOPIC0", PostID: "POST0"},
		Comment{User: "Luv Letter", Text: "WDC", TopicID: "TOPIC0", PostID: "POST0"},
	}

	for i, comment := range comments {
		commentAsBytes, _ := json.Marshal(comment)
		err := ctx.GetStub().PutState("COMMENT"+strconv.Itoa(i), commentAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// CreateComment adds a new Comment to the world state with given details
func (s *SmartContract) CreateComment(ctx contractapi.TransactionContextInterface, CommentID string, user string, text string, topicID string, postID string) error {
	comment := Comment{
		User: user,
		Text: text,
		TopicID: topicID,
		PostID: postID,
	}

	commentAsBytes, _ := json.Marshal(comment)

	return ctx.GetStub().PutState(CommentID, commentAsBytes)
}

// QueryComment returns the comment stored in the world state with given id
func (s *SmartContract) QueryComment(ctx contractapi.TransactionContextInterface, CommentID string) (*Comment, error) {
	commentAsBytes, err := ctx.GetStub().GetState(CommentID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if commentAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", CommentID)
	}

	comment := new(Comment)
	_ = json.Unmarshal(commentAsBytes, comment)

	return comment, nil
}

// QueryAllComments returns all comments found in world state
func (s *SmartContract) QueryAllComments(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		if strings.Contains(queryResponse.Key, "COMMENT") {
			comment := new(Comment)
			_ = json.Unmarshal(queryResponse.Value, comment)

			queryResult := QueryResult{Key: queryResponse.Key, Record: comment}
			results = append(results, queryResult)
		}
	}

	return results, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create BBS chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting BBS chaincode: %s", err.Error())
	}
}
