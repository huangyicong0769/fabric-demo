/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
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

// QueryComment structure used for handling result of query
type QueryComment struct {
	Key    string `json:"Key"`
	Record *Comment
}

// Comment describes basic details of what makes up a Post
type Post struct {
	Caption string `json:caption`
	TopicID string `json:topicID`
}

// QueryComment structure used for handling result of query
type QueryPost struct {
	Key    string `json:"Key"`
	Record *Post
}

// Comment describes basic details of what makes up a Topic
type Topic struct {
	TopicName string `json:topicName`
}

// QueryComment structure used for handling result of query
type QueryTopic struct {
	Key    string `json:"Key"`
	Record *Topic
}

// InitLedger do nothing
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// CreateComment adds a new Comment to the world state with given details
func (s *SmartContract) CreateComment(ctx contractapi.TransactionContextInterface, CommentID string, user string, text string, topicID string, postID string) error {
	comment := Comment{
		User:    user,
		Text:    text,
		TopicID: topicID,
		PostID:  postID,
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
func (s *SmartContract) QueryAllComments(ctx contractapi.TransactionContextInterface) ([]QueryComment, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryComment{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		if strings.Contains(queryResponse.Key, "COMMENT") {
			comment := new(Comment)
			_ = json.Unmarshal(queryResponse.Value, comment)

			queryComment := QueryComment{Key: queryResponse.Key, Record: comment}
			results = append(results, queryComment)
		}
	}

	return results, nil
}

// CreateTopic adds a new Topic to the world state with given details
func (s *SmartContract) CreateTopic(ctx contractapi.TransactionContextInterface, TopicID string, topicName string) error {
	topic := Topic{
		TopicName: topicName,
	}

	topicAsBytes, _ := json.Marshal(topic)

	return ctx.GetStub().PutState(TopicID, topicAsBytes)
}

// QueryTopic returns the topic name stored in the world state with given id
func (s *SmartContract) QueryTopic(ctx contractapi.TransactionContextInterface, TopicID string) (*Topic, error) {
	topicAsBytes, err := ctx.GetStub().GetState(TopicID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if topicAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", TopicID)
	}

	topic := new(Topic)
	_ = json.Unmarshal(topicAsBytes, topic)

	return topic, nil
}

// CreatePost adds a new Post to the world state with given details
func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, PostID string, caption string, topicID string) error {
	post := Post{
		Caption: caption,
		TopicID: topicID,
	}

	postAsBytes, _ := json.Marshal(post)

	return ctx.GetStub().PutState(PostID, postAsBytes)
}

// QueryPost returns the Post stored in the world state with given id
func (s *SmartContract) QueryPost(ctx contractapi.TransactionContextInterface, PostID string) (*Post, error) {
	postAsBytes, err := ctx.GetStub().GetState(PostID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if postAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", PostID)
	}

	post := new(Post)
	_ = json.Unmarshal(postAsBytes, post)

	return post, nil
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
