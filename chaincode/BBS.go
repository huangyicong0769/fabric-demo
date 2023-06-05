/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Comment
type SmartContract struct {
	contractapi.Contract
}

// Comment describes basic details of what makes up a Comment
type Comment struct {
	User string `json:"user"`
	Text string `json:"text"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Comment
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	comments := []Comment{
		Comment{User: "ADSR", Text: "4chan"},
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

// CreatePost adds a new Comment to the world state with given details
func (s *SmartContract) CreateCommet(ctx contractapi.TransactionContextInterface, CommentID int, user string, text string) error {
	comment := Comment{
		User: user,
		Text: text,
	}

	commentAsBytes, _ := json.Marshal(comment)

	return ctx.GetStub().PutState("COMMENT"+strconv.Itoa(CommentID), commentAsBytes)
}

// QueryCar returns the car stored in the world state with given id
func (s *SmartContract) QueryCar(ctx contractapi.TransactionContextInterface, CommentID int) (*Comment, error) {
	commentAsBytes, err := ctx.GetStub().GetState("COMMENT"+strconv.Itoa(CommentID))

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
// Should only uesd in test
// Would not work if there are lists
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

		comment := new(Comment)
		_ = json.Unmarshal(queryResponse.Value, comment)

		queryResult := QueryResult{Key: queryResponse.Key, Record: comment}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeCarOwner updates the owner field of car with given id in world state
// func (s *SmartContract) ChangeCarOwner(ctx contractapi.TransactionContextInterface, carNumber string, newOwner string) error {
// 	car, err := s.QueryCar(ctx, carNumber)

// 	if err != nil {
// 		return err
// 	}

// 	car.Owner = newOwner

// 	carAsBytes, _ := json.Marshal(car)

// 	return ctx.GetStub().PutState(carNumber, carAsBytes)
// }

// Sync Lists of Topics, Posts, Comments
// Unfinished
// func (s *SmartContract) SnycLists(ctx contractapi.TransactionContextInterface, ListID string, newList []string) ([]string, error) {
// 	oldListAsBytes, err := ctx.GetStub().GetState(ListID)

// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
// 	}

// 	if oldListAsBytes == nil {
// 		return nil, fmt.Errorf("%s does not exist", ListID)
// 	}

// 	var oldList []string
// 	_ = json.Unmarshal(oldListAsBytes, oldList)

// 	if len(oldList) >= len(newList) {
// 		return oldList, err
// 	}

// 	newListAsBytes, _ := json.Marshal(newList)
// 	return newList, ctx.GetStub().PutState(ListID, newListAsBytes)
// }

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
