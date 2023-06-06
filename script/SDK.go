package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	SDK           *fabsdk.FabricSDK
	channelClient *channel.Client
	channelName   = "mychannel"
	chaincodeName = "fabcar" //to make it work in fabcar testnetwork. DO NOT CHANGE!
	orgName       = "Org1"
	orgAdmin      = "Admin"
	org1Peer0     = "peer0.org1.example.com"
	org2Peer0     = "peer0.org2.example.com"
)

type Comment struct {
	CommentID string `json:"commentID"`
	User      string `json:"user"`
	Text      string `json:"text"`
}

func ChannelExecute(funcName string, args [][]byte) (channel.Response, error) {
	var err error
	configPath := "./config.yaml"
	configProvider := config.FromFile(configPath)
	SDK, err = fabsdk.New(configProvider)
	if err != nil {
		log.Fatalf("Failed to create new SDK: %s", err)
	}
	ctx := SDK.ChannelContext(channelName, fabsdk.WithOrg(orgName), fabsdk.WithUser(orgAdmin))
	channelClient, err = channel.New(ctx)
	response, err := channelClient.Execute(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         funcName,
		Args:        args,
	})
	if err != nil {
		return response, err
	}
	SDK.Close()
	return response, nil
}

func main() {
	r := gin.Default()

	r.GET("/queryAllComments", func(c *gin.Context) {
		var result channel.Response
		result, err := ChannelExecute("QueryAllComments", [][]byte{})
		fmt.Println(result)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query All Success",
			"result":  string(result.Payload),
		})
	})

	r.POST("/queryComment", func(c *gin.Context) {
		var comment Comment
		c.BindJSON(&comment)
		var result channel.Response
		result, err := ChannelExecute("QueryComment", [][]byte{[]byte(comment.CommentID)})
		fmt.Println(result)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction :%s\n", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query All Success",
			"result":  string(result.Payload),
		})
	})

	r.POST("/createComment", func(c *gin.Context) {
		var comment Comment
		c.BindJSON(&comment)
		var result channel.Response
		result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text)})
		fmt.Println(result)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Create Success",
			"result":  string(result.Payload),
		})
	})

	r.Run(":9099")
}
