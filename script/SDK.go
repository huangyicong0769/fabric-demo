package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

type Post struct {
	PostID      string `json:"postID"`
	Caption     string `json:"caption"`
	CommentList []Comment
}

type Topic struct {
	TopicID   string `json:"topicID"`
	TopicName string `json:"topicName"`
	PostList  []Post
}

var (
	TopicList    []Topic
	CommentTotal int
)

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

func (s *Post) Post2str() string {
	result := ""

	return result
}

func main() {
	CommentTotal = 2

	r := gin.Default()

	//Only for test
	r.GET("/initList", func(c *gin.Context) {
		TopicList = append(TopicList, Topic{TopicID: "Topic0", TopicName: "Anime"})
		TopicList[0].PostList = append(TopicList[0].PostList, Post{PostID: "Post0", Caption: "New Macross project started"})
		TopicList[0].PostList[0].CommentList = append(TopicList[0].PostList[0].CommentList, Comment{CommentID: "COMMENT" + strconv.Itoa(2), User: "尼古拉斯赵四", Text: "rt"}, Comment{CommentID: "COMMENT" + strconv.Itoa(3), User: "LRSzwei", Text: "cy"})
		CommentTotal = 4

		cnt := 0
		for _, comment := range TopicList[0].PostList[0].CommentList {
			result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text)})
			fmt.Println(result)
			if err != nil {
				log.Fatalf("Failed to evaluate transaction: %s\n", err)
			}
			cnt++
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Create Success",
			"result":  "add " + strconv.Itoa(cnt) + " comments",
		})
	})

	//Shuold not be used
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

	//Should not be used
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

	r.GET("/queryTopicList", func(c *gin.Context) {
		result := "["
		for i, topic := range TopicList {
			if i != 0 {
				result += ","
			}
			result += "{" + "\\\"topicID:\\\"" + topic.TopicID + "\\\",\\\"topicName\\\":\\\"" + topic.TopicName + "\\\"" + "}"
		}
		result += "]"

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query Success",
			"result":  result,
		})
	})

	r.POST("/queryPostList", func(c *gin.Context) {})

	r.POST("/queryCommentList", func(c *gin.Context) {})

	r.Run(":9099")
}
