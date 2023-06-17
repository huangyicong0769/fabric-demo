package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"encoding/json"

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
	CommentList []string
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

func SaveList() error{
	for _, topic := range TopicList{
		_, err := ChannelExecute("CreateTopic", [][]byte{[]byte(topic.TopicID), []byte(topic.TopicName)})
		if err != nil {
			log.Fatalf("Failed to save lists: %s\n", err)
		}

		for _, post := range topic.PostList {
			_, err := ChannelExecute("CreatePost", [][]byte{[]byte(topic.TopicID + post.PostID), []byte(post.Caption), []byte(topic.TopicID),})
			if err != nil {
				log.Fatalf("Failed to save lists: %s\n", err)
			}
		}
	}
	return nil
}

func main() {
	CommentTotal = 2

	r := gin.Default()

	//Only for test
	r.GET("/initList", func(c *gin.Context) {
		TopicList = append(TopicList, Topic{TopicID: "TOPIC0", TopicName: "Anime"})
		TopicList[0].PostList = append(TopicList[0].PostList, Post{PostID: "POST0", Caption: "New Macross project started"})
		TopicList[0].PostList[0].CommentList = append(TopicList[0].PostList[0].CommentList, "COMMENT" + strconv.Itoa(0), "COMMENT" + strconv.Itoa(1))
		CommentTotal = 2

		commentList := []Comment{{CommentID: "COMMENT" + strconv.Itoa(0), User: "尼古拉斯赵四", Text: "rt"}, {CommentID: "COMMENT" + strconv.Itoa(1), User: "LRSzwei", Text: "cy"}}

		cnt := 0
		for _, comment := range commentList {
			result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text), []byte("TOPIC0"), []byte("POST0")})
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

	//Only for testing chaincode
	// r.POST("/createComment", func(c *gin.Context) {
	// 	var comment Comment
	// 	c.BindJSON(&comment)
	// 	var result channel.Response
	// 	result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text)})
	// 	fmt.Println(result)
	// 	if err != nil {
	// 		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	// 	}
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code":    "200",
	// 		"message": "Create Success",
	// 		"result":  string(result.Payload),
	// 	})
	// })

	r.GET("/queryTopicList", func(c *gin.Context) {
		result := "["
		for i, topic := range TopicList {
			if i != 0 {
				result += ","
			}
			result += "{" + "\"topicID:\"" + topic.TopicID + "\",\"topicName\":\"" + topic.TopicName + "\"" + "}"
		}
		result += "]"

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query Success",
			"result":  result,
		})
	})

	r.POST("/queryPostList", func(c *gin.Context) {
		topicID := c.PostForm("topicID")

		topicIndex, err := strconv.Atoi(string(topicID[5:]))

		if err != nil {
			log.Fatalf("Failed to query list: %s\n", err)
		}

		result := "["
		for i, post := range TopicList[topicIndex].PostList {
			if i != 0 {
				result += ","
			}
			result += "{" + "\"postID:\"" + post.PostID + "\",\"caption\":\"" + post.Caption + "\"" + "}"
		}
		result += "]"

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query Success",
			"result":  result,
		})
	})

	r.POST("/queryCommentList", func(c *gin.Context) {
		topicID, postID := c.PostForm("topicID"), c.PostForm("postID")

		topicIndex, err := strconv.Atoi(topicID[5:])
		if err != nil {
			log.Fatalf("Failed to query list: %s\n", err)
		}

		postIndex, err := strconv.Atoi(postID[4:])
		if err != nil {
			log.Fatalf("Failed to query list: %s\n", err)
		}

		result := "["
		for i, commentID := range TopicList[topicIndex].PostList[postIndex].CommentList {
			qcomment, err := ChannelExecute("QueryComment", [][]byte{[]byte(commentID)})
			fmt.Println(result)
			if err != nil {
				log.Fatalf("Failed to evaluate transaction :%s\n", err)
			}

			comment := new(Comment)
			_ = json.Unmarshal(qcomment.Payload, comment)

			if i != 0 {
				result += ","
			}
			result += "{" + "\"commentID:\"" + commentID + "\",\"user\":\"" + comment.User + "\",\"text\":\"" + comment.Text + "}"
		}
		result += "]"

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Query Success",
			"result":  result,
		})
	})

	r.POST("/createComment", func(c *gin.Context) {
		topicID, postID := c.PostForm("topicID"), c.PostForm("postID")
		var comment Comment
		comment.CommentID, comment.User, comment.Text = "COMMENT"+strconv.Itoa(CommentTotal), c.PostForm("user"), c.PostForm("text")
		CommentTotal++

		topicIndex, err := strconv.Atoi(topicID[5:])
		if err != nil {
			log.Fatalf("Failed to create comment: %s\n", err)
		}

		postIndex, err := strconv.Atoi(postID[4:])
		if err != nil {
			log.Fatalf("Failed to create comment: %s\n", err)
		}

		TopicList[topicIndex].PostList[postIndex].CommentList = append(TopicList[topicIndex].PostList[postIndex].CommentList, comment.CommentID)
		result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text), []byte(topicID), []byte(postID)})
		fmt.Println(result)
		if err != nil {
			log.Fatalf("Failed to create comment: %s\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Create Success",
			"result":  string(result.Payload),
		})
	})

	r.POST("/createPost", func(c *gin.Context) {
		topicID := c.PostForm("topicID")
		var post Post
		post.Caption = c.PostForm("caption")
		var comment Comment
		comment.CommentID, comment.User, comment.Text = "COMMENT"+strconv.Itoa(CommentTotal), c.PostForm("user"), c.PostForm("text")
		post.CommentList = append(post.CommentList, comment.CommentID)
		CommentTotal++

		topicIndex, err := strconv.Atoi(topicID[5:])
		if err != nil {
			log.Fatalf("Failed to create comment: %s\n", err)
		}

		post.PostID = "POST" + strconv.Itoa(len(TopicList[topicIndex].PostList))
		TopicList[topicIndex].PostList = append(TopicList[topicIndex].PostList, post)
		result, err := ChannelExecute("CreateComment", [][]byte{[]byte(comment.CommentID), []byte(comment.User), []byte(comment.Text), []byte(topicID), []byte(post.PostID)})
		fmt.Println(result)
		if err != nil {
			log.Fatalf("Failed to create comment: %s\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Create Success",
			"result":  string(result.Payload),
		})
	})

	r.GET("/saveList", func(c *gin.Context) {
		err := SaveList()
		if err != nil {
			log.Fatalf("Failed to save lists: %s\n", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Save Success",
			"result":  "",
		})
	})

	r.GET("/rebuildList", func(c *gin.Context) {
		TopicList = nil

		qtopicNum, err := ChannelExecute("QueryTopicNumber", [][]byte{})
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}

		topicNum, err := strconv.Atoi(string(qtopicNum.Payload))
		if err != nil {
			log.Fatalf("Failed to rebuild list: %s\n", err)
		}
		
		i := 0
		for i < topicNum {
			TopicID := "TOPIC" + strconv.Itoa(i)
			result, err := ChannelExecute("QueryTopic", [][]byte{[]byte(TopicID)})
			if err != nil {
				log.Fatalf("Failed to evaluate transaction: %s\n", err)
			}

			if result.Payload == nil {
				break
			}

			topic := new(Topic)
			_ = json.Unmarshal(result.Payload, topic)
			topic.TopicID = TopicID

			qpostNum, err := ChannelExecute("QueryPostNumber", [][]byte{[]byte(TopicID)})
			if err != nil {
				log.Fatalf("Failed to evaluate transaction: %s\n", err)
			}
	
			postNum, err := strconv.Atoi(string(qpostNum.Payload))
			if err != nil {
				log.Fatalf("Failed to rebuild list: %s\n", err)
			}

			j := 0
			for j < postNum {
				PostID := TopicID + "POST" + strconv.Itoa(j)
				result, err := ChannelExecute("QueryPost", [][]byte{[]byte(PostID)})
				if err != nil {
					log.Fatalf("Failed to evaluate transaction: %s\n", err)
				}
	
				if result.Payload == nil {
					break
				}
	
				post := new(Post)
				_ = json.Unmarshal(result.Payload, post)
				post.PostID = PostID[(len(TopicID)):]

				topic.PostList = append(topic.PostList, *post)

				j++
			}

			TopicList = append(TopicList, *topic)

			i++
		}

		qcommentNum, err := ChannelExecute("QueryCommentNumber", [][]byte{})
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}

		commentNum, err := strconv.Atoi(string(qcommentNum.Payload))
		if err != nil {
			log.Fatalf("Failed to rebuild list: %s\n", err)
		}

		CommentTotal = 0
		for CommentTotal < commentNum{
			CommentID := "COMMENT" + strconv.Itoa(CommentTotal)
			result, err := ChannelExecute("QueryComment", [][]byte{[]byte(CommentID)})
			if err != nil {
				log.Fatalf("Failed to evaluate transaction: %s\n", err)
			}

			if result.Payload == nil {
				break
			}

			type RichComment struct {
				User    string `json:"user"`
				Text    string `json:"text"`
				TopicID string `json:topicID`
				PostID  string `json:postID`
			}

			comment := new(RichComment)
			_ = json.Unmarshal(result.Payload, comment)

			topicIndex, err := strconv.Atoi(comment.TopicID[5:])
			if err != nil {
				log.Fatalf("Failed to create comment: %s\n", err)
			}
	
			postIndex, err := strconv.Atoi(comment.PostID[4:])
			if err != nil {
				log.Fatalf("Failed to create comment: %s\n", err)
			}

			TopicList[topicIndex].PostList[postIndex].CommentList = append(TopicList[topicIndex].PostList[postIndex].CommentList, CommentID)

			CommentTotal++
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    "200",
			"message": "Rebuild Success",
			"result":  "",
		})
	})

	r.Run(":9099")
}
