// File: simulator.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const baseURL = "http://localhost:8080"

type Action string

const (
	ActionRegister     Action = "register"
	ActionLogin        Action = "login"
	ActionCreateSub    Action = "create_subreddit"
	ActionJoinSub      Action = "join_subreddit"
	ActionLeaveSub     Action = "leave_subreddit"
	ActionPost         Action = "post"
	ActionComment      Action = "comment"
	ActionVote         Action = "vote"
	ActionGetFeed      Action = "get_feed"
	ActionSendMessage  Action = "send_message"
	ActionGetMessages  Action = "get_messages"
	ActionReplyMessage Action = "reply_message"
)

// StartSimulation triggers the actor to start user's scenario
type StartSimulation struct {
	UserIndex int
}

// StopSimulation stops the actor
type StopSimulation struct{}

type UserActor struct {
	userIndex   int
	username    string
	password    string
	userID      int
	subredditID int
	token       string
	email       string
	logFile     *os.File
}

func NewUserActor(userIndex int) actor.Actor {
	return &UserActor{userIndex: userIndex}
}

func (u *UserActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case StartSimulation:
		u.userIndex = msg.UserIndex
		u.username = fmt.Sprintf("user%d", u.userIndex)
		u.password = "pass123"
		u.email = fmt.Sprintf("user%d@ufl.edu", u.userIndex)
		// Open log file for this user
		os.MkdirAll("logs", 0755)
		logFileName := fmt.Sprintf("logs/user_%d.log", u.userIndex)
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Failed to open log file for user %d: %v", u.userIndex, err)
			ctx.Send(ctx.Self(), &StopSimulation{})
			return
		}
		u.logFile = f
		u.logf("Starting simulation for %s", u.username)

		u.runScenario()

		ctx.Send(ctx.Self(), &StopSimulation{})

	case *StopSimulation:
		u.logf("Stopping simulation for %s", u.username)
		if u.logFile != nil {
			u.logFile.Close()
		}
		ctx.Stop(ctx.Self())
	}
}

func (u *UserActor) runScenario() {
	actions := []Action{
		ActionRegister,
		ActionLogin,
		ActionCreateSub,
		ActionJoinSub,
		ActionPost,
		ActionComment,
		ActionVote,
		ActionGetFeed,
		ActionSendMessage,
		ActionGetMessages,
		ActionReplyMessage,
		ActionLeaveSub,
	}

	for _, action := range actions {
		if err := u.performAction(action); err != nil {
			u.logf("Action %s failed: %v", action, err)
			break
		} else {
			u.logf("Action %s succeeded", action)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (u *UserActor) performAction(a Action) error {
	switch a {
	case ActionRegister:
		return u.register()
	case ActionLogin:
		return u.login()
	case ActionCreateSub:
		return u.createSubreddit()
	case ActionJoinSub:
		return u.joinSubreddit()
	case ActionLeaveSub:
		return u.leaveSubreddit()
	case ActionPost:
		return u.createPost()
	case ActionComment:
		return u.createComment(false)
	case ActionVote:
		return u.voteOnPostOrComment()
	case ActionGetFeed:
		return u.getFeed()
	case ActionSendMessage:
		return u.sendMessage()
	case ActionGetMessages:
		return u.getMessages()
	case ActionReplyMessage:
		return u.replyToMessage()
	default:
		return fmt.Errorf("unknown action")
	}
}

func (u *UserActor) register() error {
	payload := map[string]string{
		"username": u.username,
		"password": u.password,
		"email":    u.email,
	}
	resp, err := u.makeRequest("POST", "/register", payload, 0)
	if err != nil {
		return err
	}
	var userRes struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(resp, &userRes); err != nil {
		return err
	}
	u.userID = userRes.UserID
	return nil
}

func (u *UserActor) login() error {
	payload := map[string]string{
		"username": u.username,
		"password": u.password,
	}
	resp, err := u.makeRequest("POST", "/login", payload, u.userID)
	if err != nil {
		return err
	}
	var loginRes struct {
		UserID int    `json:"user_id"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(resp, &loginRes); err != nil {
		return err
	}
	u.token = loginRes.Token
	// user_id might already be known from registration, but if login returns it, set it again:
	//u.userID = loginRes.UserID
	return nil
}

func (u *UserActor) createSubreddit() error {
	name := fmt.Sprintf("r/testsub_%d", u.userIndex)
	payload := map[string]interface{}{
		"name":        name,
		"description": "A test subreddit",
		"created_by":  u.userID,
	}
	resp, err := u.makeRequest("POST", "/subreddits", payload, u.userID)
	if err != nil {
		return err
	}
	var subRes struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp, &subRes); err != nil {
		return err
	}
	u.subredditID = subRes.ID
	return nil
}

func (u *UserActor) joinSubreddit() error {
	if u.subredditID == 0 {
		return fmt.Errorf("no subreddit to join")
	}

	payload := map[string]interface{}{
		"user_id":  u.userID,
	}

	// If the server identifies user from X-User-ID, no payload needed:
	_, err := u.makeRequest("POST", fmt.Sprintf("/subreddits/%d/join", u.subredditID), payload, u.userID)
	return err
}

func (u *UserActor) leaveSubreddit() error {
	if u.subredditID == 0 {
		return fmt.Errorf("no subreddit to leave")
	}

	payload := map[string]interface{}{
		"user_id":  u.userID,
	}

	_, err := u.makeRequest("POST", fmt.Sprintf("/subreddits/%d/leave", u.subredditID), payload, u.userID)
	return err
}

func (u *UserActor) createPost() error {
	if u.subredditID == 0 {
		return fmt.Errorf("no subreddit for post")
	}
	payload := map[string]interface{}{
		"title":        "Hello World",
		"content":      "This is a test post",
		"subreddit_id": u.subredditID,
		"author_id":    u.userID,
	}
	_, err := u.makeRequest("POST", fmt.Sprintf("/subreddits/%d/posts", u.subredditID), payload, u.userID)
	return err
}

func (u *UserActor) createComment(replyToComment bool) error {
	postID := 1 // Example: in real scenario, get from feed
	var parentID *int
	if replyToComment {
		cid := 100
		parentID = &cid
	}

	payload := map[string]interface{}{
		"content":    "This is a comment",
		"post_id":    postID,
		"author_id":  u.userID,      // must include author_id
		"parent_id":  parentID,      // for top-level comment, this can be nil
	}
	_, err := u.makeRequest("POST", fmt.Sprintf("/posts/%d/comments", postID), payload, u.userID)
	return err
}

func (u *UserActor) voteOnPostOrComment() error {
	postID := 1
	// The Vote model requires `vote_type` ("upvote" or "downvote"), `user_id`, and either `post_id` or `comment_id`.
	// Assume we upvote this post:
	payload := map[string]interface{}{
		"post_id":   postID,
		"user_id":   u.userID,
		"vote_type": "upvote",
	}
	_, err := u.makeRequest("POST", fmt.Sprintf("/posts/%d/vote", postID), payload, u.userID)
	return err
}

func (u *UserActor) getFeed() error {
	_, err := u.makeRequest("GET", "/feed", nil, u.userID)
	return err
}

func (u *UserActor) sendMessage() error {
	toUserID := 2
	// The Message model requires sender_id, receiver_id, content
	payload := map[string]interface{}{
		"sender_id":   u.userID,
		"receiver_id": toUserID,
		"content":     "Hello from " + u.username,
	}
	_, err := u.makeRequest("POST", "/messages", payload, u.userID)
	return err
}

func (u *UserActor) getMessages() error {
	_, err := u.makeRequest("GET", fmt.Sprintf("/users/%d/messages", u.userID), nil, u.userID)
	return err
}

func (u *UserActor) replyToMessage() error {
	messageID := 1
	// For a reply, we need sender_id, receiver_id, content, and parent_id
	receiverID := 2 // Example: same user as above or determined from getMessages
	parentID := messageID
	payload := map[string]interface{}{
		"sender_id":   u.userID,
		"receiver_id": receiverID,
		"content":     "Replying to your message",
		"parent_id":   parentID,
	}
	_, err := u.makeRequest("POST", fmt.Sprintf("/messages/%d/reply", messageID), payload, u.userID)
	return err
}

func (u *UserActor) logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if u.logFile != nil {
		io.WriteString(u.logFile, time.Now().Format(time.RFC3339)+" "+msg+"\n")
	} else {
		log.Println(msg)
	}
}

func (u *UserActor) makeRequest(method, path string, payload interface{}, userID int) ([]byte, error) {
	var reqBody []byte
	var err error

	if payload != nil {
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	url := baseURL + path
	u.logf("Request: %s %s payload=%v userID=%d", method, url, payload, userID)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if userID > 0 {
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		u.logf("Response error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	u.logf("Response status=%d body=%s", resp.StatusCode, string(body))
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return body, nil
}

type Supervisor struct {
	totalUsers int
}

func (s *Supervisor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case StartSimulation:
		_ = msg
		s.totalUsers = 10 // for demonstration
		numUsers := s.totalUsers

		numActors := 5
		usersPerActor := (numUsers + numActors - 1) / numActors

		for i := 0; i < numActors; i++ {
			props := actor.PropsFromProducer(func(i int) func() actor.Actor {
				return func() actor.Actor {
					return NewUserActor(i + 1)
				}
			}(i))
			workerPID := ctx.Spawn(props)

			start := i * usersPerActor
			end := (i + 1) * usersPerActor
			if end > numUsers {
				end = numUsers
			}

			for u := start; u < end; u++ {
				ctx.Send(workerPID, StartSimulation{UserIndex: u + 1})
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &Supervisor{}
	})
	supervisorPID := system.Root.Spawn(props)

	// start simulation
	system.Root.Send(supervisorPID, StartSimulation{})

	// Let simulation run for a while
	time.Sleep(30 * time.Second)
	log.Println("Simulation finished.")
}
