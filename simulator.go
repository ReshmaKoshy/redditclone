package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const (
	baseURL = "http://localhost:8080"

	// Simulation parameters
	numUsers                  = 90
	numSubreddits             = 15
	minActionsPerUser         = 4 // Minimum actions each user will perform
	maxActionsPerUser         = 9 // Maximum actions each user will perform
	maxMembersForTopSubreddit = 79

	// Action weights (probability distribution)
	weightPost     = 0.1
	weightComment  = 0.3
	weightVote     = 0.3
	weightJoinSub  = 0.03
	weightLeaveSub = 0.07
	weightMessage  = 0.2
)

// Messages for actor communication
type SimulateUsers struct {
	Users []User
	State *SimState
}

type SimulationComplete struct {
	ProcessedUsers int
}

// UserSimulatorActor handles simulation for a batch of users
type UserSimulatorActor struct {
	processedUsers int
	actorIndex     int
	wg             *sync.WaitGroup
}

type FeedPost struct {
	ID            int    `json:"post_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	SubredditID   int    `json:"subreddit_id"`
	SubredditName string `json:"subreddit_name"`
	AuthorID      int    `json:"author_id"`
	AuthorName    string `json:"author_name"`
	CreatedAt     string `json:"created_at"`
	VoteCount     struct {
		Upvotes   int `json:"upvotes"`
		Downvotes int `json:"downvotes"`
	} `json:"vote_count"`
}

func NewUserSimulatorActor(index int, wg *sync.WaitGroup) *UserSimulatorActor {
	return &UserSimulatorActor{
		actorIndex: index,
		wg:         wg,
	}
}

func (state *UserSimulatorActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *SimulateUsers:
		for _, user := range msg.Users {
			// Random number of actions for this user
			numActions := rand.Intn(maxActionsPerUser-minActionsPerUser+1) + minActionsPerUser

			fmt.Printf("Actor %v simulating user %d with %d actions...\n",
				state.actorIndex, user.ID, numActions)

			// Perform random actions for this user
			for actionNum := 0; actionNum < numActions; actionNum++ {
				action := randomAction()

				switch action {
				case "post":
					simulatePost(msg.State, user)
				case "comment":
					simulateComment(msg.State, user)
				case "vote":
					simulateVote(msg.State, user)
				case "join":
					simulateJoinSubreddit(msg.State, user)
				case "leave":
					simulateLeaveSubreddit(msg.State, user)
				case "message":
					simulateDirectMessage(msg.State, user)
				}

				// Add small delay between actions
				time.Sleep(50 * time.Millisecond)
			}
			state.processedUsers++
		}

		// Notify parent about completion
		context.Send(context.Parent(), &SimulationComplete{
			ProcessedUsers: state.processedUsers,
		})
		state.wg.Done()
	}
}

// SupervisorActor manages the pool of UserSimulatorActors
type SupervisorActor struct {
	completedUsers int
	totalUsers     int
}

func (state *SupervisorActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *SimulateUsers:
		state.totalUsers = len(msg.Users)

		// Create a wait group to track completion
		wg := &sync.WaitGroup{}

		// Define number of actors and users per actor
		numActors := 5 // Can be adjusted based on needs
		usersPerActor := (len(msg.Users) + numActors - 1) / numActors

		// Create and start actor pool
		for i := 0; i < numActors; i++ {
			wg.Add(1)

			props := actor.PropsFromProducer(func(actorIndex int) func() actor.Actor {
				return func() actor.Actor {
					return NewUserSimulatorActor(actorIndex+1, wg)
				}
			}(i))

			workerPID := context.Spawn(props)

			// Calculate user batch for this actor
			start := i * usersPerActor
			end := (i + 1) * usersPerActor
			if end > len(msg.Users) {
				end = len(msg.Users)
			}

			// Send users to this actor
			context.Send(workerPID, &SimulateUsers{
				Users: msg.Users[start:end],
				State: msg.State,
			})
		}

		// Wait for all actors to complete in a separate goroutine
		go func() {
			wg.Wait()
			fmt.Println("All user simulations completed....")
		}()

	case *SimulationComplete:
		state.completedUsers += msg.ProcessedUsers
		if state.completedUsers >= state.totalUsers {
			fmt.Printf("Simulation completed for all %d users.....\n", state.totalUsers)
		}
	}
}

// Simulation state
type SimState struct {
	users        []User
	subreddits   []Subreddit
	posts        []Post
	comments     []Comment
	messages     []Message
	memberCounts map[int]int // subredditID -> member count
}

type User struct {
	ID       int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Karma    int    `json:"karma"`
}

type Subreddit struct {
	ID          int    `json:"subreddit_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Post struct {
	ID          int    `json:"post_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	SubredditID int    `json:"subreddit_id"`
}

type Comment struct {
	ID       int    `json:"comment_id"`
	Content  string `json:"content"`
	PostID   int    `json:"post_id"`
	ParentID *int   `json:"parent_comment_id"`
}

type Message struct {
	ID        int    `json:"message_id"`
	Content   string `json:"content"`
	FromUser  int    `json:"from_user_id"`
	ToUser    int    `json:"to_user_id"`
	Timestamp string `json:"timestamp"`
}

type SendMessageRequest struct {
	ToUserID int    `json:"to_user_id"`
	Content  string `json:"content"`
}

type topUsers struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Karma        int    `json:"karma"`
	PostCount    int    `json:"post_count"`
	CommentCount int    `json:"comment_count"`
}

func main() {

	totalStart := time.Now()

	rand.Seed(time.Now().UnixNano())
	state := &SimState{
		memberCounts: make(map[int]int),
	}

	// Initialize simulation
	err := createInitialUsers(state)
	if err != nil {
		log.Fatalf("Failed to create users: %v", err)
	}

	err = createInitialSubreddits(state)
	if err != nil {
		log.Fatalf("Failed to create subreddits: %v", err)
	}

	// Distribute members according to Zipf's law
	distributeMembersZipf(state)

	simulateUsers(state)

	// Print final statistics
	printStats(state)

	// Print total execution time
	totalDuration := time.Since(totalStart)
	log.Printf("\nExecution time: %v", totalDuration)

}

func createInitialUsers(state *SimState) error {
	for i := 0; i < numUsers; i++ {
		username := fmt.Sprintf("user%d", i+1)
		password := fmt.Sprintf("pass%d", i+1)

		payload := map[string]string{
			"username": username,
			"password": password,
		}

		resp, err := makeRequest("POST", "/register", payload, "")
		if err != nil {
			return fmt.Errorf("failed to create user %s: %v", username, err)
		}

		var user User
		err = json.Unmarshal(resp, &user)
		if err != nil {
			return fmt.Errorf("failed to parse user response: %v", err)
		}

		state.users = append(state.users, user)
	}
	return nil
}

func createInitialSubreddits(state *SimState) error {
	topics := []string{"Stupidity", "Science", "IndieGaming", "Movies", "Books", "Music", "Art",
		"RareInsults", "TravelVlogs", "FitnessHealth", "Photography", "FashionInsta", "Sports", "NSFW", "Humor",
		"DIY", "Pets", "NatureIsLit", "HistoryBuff", "PhilosophyHelp"}

	for i := 0; i < numSubreddits; i++ {
		name := fmt.Sprintf("r/%s", topics[i])
		description := fmt.Sprintf("A community for discussing %s", strings.ToLower(topics[i]))

		payload := map[string]string{
			"name":        name,
			"description": description,
		}

		// Use a random user as creator
		creator := state.users[rand.Intn(len(state.users))]

		resp, err := makeRequest("POST", "/subreddits", payload, creator.ID)
		if err != nil {
			return fmt.Errorf("failed to create subreddit %s: %v", name, err)
		}

		var subreddit Subreddit
		err = json.Unmarshal(resp, &subreddit)
		if err != nil {
			return fmt.Errorf("failed to parse subreddit response: %v", err)
		}

		state.subreddits = append(state.subreddits, subreddit)
	}
	return nil
}

func distributeMembersZipf(state *SimState) {
	// Calculate Zipf distribution for member counts
	for i, subreddit := range state.subreddits {
		// Using Zipf's law: N(k) ∝ 1/k where k is rank
		rank := float64(i + 1)
		memberCount := int(float64(maxMembersForTopSubreddit) / rank)
		state.memberCounts[subreddit.ID] = memberCount

		// Randomly select members up to the calculated count
		members := rand.Perm(len(state.users))[:memberCount]
		for _, memberIdx := range members {
			user := state.users[memberIdx]
			_, err := makeRequest("POST", fmt.Sprintf("/subreddits/%d/join", subreddit.ID), nil, user.ID)
			if err != nil {
				log.Printf("Failed to add user %d to subreddit %d: %v", user.ID, subreddit.ID, err)
			}
		}
	}
}

// Modified main simulation function to use actors
func simulateUsers(state *SimState) {
	fmt.Printf("Starting simulation for %d users using Proto.Actor...\n", len(state.users))

	// Create actor system
	system := actor.NewActorSystem()

	// Create supervisor actor
	supervisorProps := actor.PropsFromProducer(func() actor.Actor {
		return &SupervisorActor{}
	})

	supervisorPID := system.Root.Spawn(supervisorProps)

	// Start simulation
	system.Root.Send(supervisorPID, &SimulateUsers{
		Users: state.users,
		State: state,
	})

	// Wait for a moment to allow simulation to complete

	//time.Sleep(time.Duration(len(state.users)) * time.Second)
	time.Sleep(time.Duration(20) * time.Second)

}

func randomAction() string {
	r := rand.Float64()
	switch {
	case r < weightPost:
		return "post"
	case r < weightPost+weightComment:
		return "comment"
	case r < weightPost+weightComment+weightVote:
		return "vote"
	case r < weightPost+weightComment+weightVote+weightJoinSub:
		return "join"
	case r < weightPost+weightComment+weightVote+weightJoinSub+weightLeaveSub:
		return "leave"
	}
	return "message"
}

func simulatePost(state *SimState, user User) {
	if len(state.subreddits) == 0 {
		return
	}

	subreddit := state.subreddits[rand.Intn(len(state.subreddits))]

	payload := map[string]interface{}{
		"title":        generateRandomTitle(),
		"content":      generateRandomContent(),
		"subreddit_id": subreddit.ID,
	}

	resp, err := makeRequest("POST", "/posts", payload, user.ID)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		return
	}

	var post Post
	if err := json.Unmarshal(resp, &post); err == nil {
		state.posts = append(state.posts, post)
	}
}

func simulateComment(state *SimState, user User) {
	if len(state.posts) == 0 {
		return
	}

	post := state.posts[rand.Intn(len(state.posts))]
	var parentID *int

	// 30% chance of replying to existing comment
	if len(state.comments) > 0 && rand.Float64() < 0.3 {
		comment := state.comments[rand.Intn(len(state.comments))]
		parentID = &comment.ID
	}

	payload := map[string]interface{}{
		"content":           generateRandomComment(),
		"post_id":           post.ID,
		"parent_comment_id": parentID,
	}

	resp, err := makeRequest("POST", "/comments", payload, user.ID)
	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		return
	}

	var comment Comment
	if err := json.Unmarshal(resp, &comment); err == nil {
		state.comments = append(state.comments, comment)
	}
}

func simulateVote(state *SimState, user User) {
	// Decide whether to vote on post or comment
	voteOnPost := rand.Float64() < 0.7

	if voteOnPost && len(state.posts) > 0 {
		post := state.posts[rand.Intn(len(state.posts))]
		payload := map[string]interface{}{
			"target_id":   post.ID,
			"target_type": "post",
			"value":       []int{-1, 1}[rand.Intn(2)],
		}

		_, err := makeRequest("POST", "/vote", payload, user.ID)
		if err != nil {
			log.Printf("Failed to vote on post: %v", err)
		}
	} else if len(state.comments) > 0 {
		comment := state.comments[rand.Intn(len(state.comments))]
		payload := map[string]interface{}{
			"target_id":   comment.ID,
			"target_type": "comment",
			"value":       []int{-1, 1}[rand.Intn(2)],
		}

		_, err := makeRequest("POST", "/vote", payload, user.ID)
		if err != nil {
			log.Printf("Failed to vote on comment: %v", err)
		}
	}
}

func simulateJoinSubreddit(state *SimState, user User) {
	if len(state.subreddits) == 0 {
		return
	}

	subreddit := state.subreddits[rand.Intn(len(state.subreddits))]
	_, err := makeRequest("POST", fmt.Sprintf("/subreddits/%d/join", subreddit.ID), nil, user.ID)
	if err != nil {
		log.Printf("Failed to join subreddit: %v", err)
		return
	}

	state.memberCounts[subreddit.ID]++
}

func simulateLeaveSubreddit(state *SimState, user User) {
	if len(state.subreddits) == 0 {
		return
	}

	subreddit := state.subreddits[rand.Intn(len(state.subreddits))]
	_, err := makeRequest("POST", fmt.Sprintf("/subreddits/%d/leave", subreddit.ID), nil, user.ID)
	if err != nil {
		log.Printf("Failed to leave subreddit: %v", err)
		return
	}

	if state.memberCounts[subreddit.ID] > 0 {
		state.memberCounts[subreddit.ID]--
	}
}

func simulateDirectMessage(state *SimState, user User) {
	if len(state.users) <= 1 {
		return
	}

	// Select random recipient (different from sender)
	var recipient User
	for {
		recipient = state.users[rand.Intn(len(state.users))]
		if recipient.ID != user.ID {
			break
		}
	}

	payload := SendMessageRequest{
		ToUserID: recipient.ID,
		Content:  generateRandomMessage(),
	}

	resp, err := makeRequest("POST", "/message", payload, user.ID)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return
	}

	var message Message
	if err := json.Unmarshal(resp, &message); err == nil {
		state.messages = append(state.messages, message)
	}
}

// Add generateRandomMessage function
func generateRandomMessage() string {
	messages := []string{
		"Hi, I came across your recent post and thought of reaching out!",
		"How do you feel about the latest updates in this space?",
		"I’m curious about the point you raised in the discussion—could we talk more about it?",
		"It would be great to team up on a post sometime!",
		"Appreciate your thoughtful input—it really stands out!",
		"I wanted to share something I thought you might find fascinating.",
		"Could we dive deeper into your knowledge on this topic?",
		"Your perspective really caught my attention—thanks for sharing!",
	}
	return messages[rand.Intn(len(messages))] + " " + randomSentence()
}
func makeRequest(method, path string, payload interface{}, userID interface{}) ([]byte, error) {
	var reqBody []byte
	var err error

	if payload != nil {
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	if userID != nil {
		req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getUserDetails(userID int) (*User, error) {
	username := fmt.Sprintf("user%d", userID)
	resp, err := makeRequest("GET", fmt.Sprintf("/users/%s", username), nil, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %v", err)
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %v", err)
	}
	return &user, nil
}

// Add function to get user's feed
func getUserFeed(userID int) ([]FeedPost, error) {
	resp, err := makeRequest("GET", "/feed", nil, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %v", err)
	}

	var feed []FeedPost
	if err := json.Unmarshal(resp, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse feed response: %v", err)
	}
	return feed, nil
}

// Add function to get user's direct messages
func getUserMessages(userID int) ([]Message, error) {
	resp, err := makeRequest("GET", "/message", nil, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}

	var messages []Message
	if err := json.Unmarshal(resp, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse messages response: %v", err)
	}
	return messages, nil
}

func getKarmaLeaders(userID int) {
	resp, err := makeRequest("GET", "/users/top", nil, userID)
	if err != nil {
		fmt.Printf("failed to get Karma leaderboard....: %v", err)
	}

	var topUser []topUsers
	if err := json.Unmarshal(resp, &topUser); err != nil {
		fmt.Printf("failed to parse karma leaderboard response.....: %v", err)
	}

	fmt.Println("------------------------")
	fmt.Printf("\nKarma Leaderboard\n")
	fmt.Println("------------------------")

	for i := 0; i < len(topUser); i++ {
		fmt.Printf("Username :%s\n", topUser[i].Username)
		fmt.Printf("Karma :%d\n", topUser[i].Karma)
	}

}

// Add function to print user details
func printUserDetails(user User, feed []FeedPost, messages []Message) {
	fmt.Printf("\n=== User: %s ===\n", user.Username)
	fmt.Printf("Karma: %d\n", user.Karma)

	fmt.Printf("\nFeed (latest 5 posts):\n")
	fmt.Println("------------------------")
	maxPosts := 5
	if len(feed) < maxPosts {
		maxPosts = len(feed)
	}
	for i := 0; i < maxPosts; i++ {
		post := feed[i]
		fmt.Printf("[%s] %s\n", post.SubredditName, post.Title)
		fmt.Printf("Posted by u/%s |\n", post.AuthorName)
		fmt.Printf("Content: %s\n", post.Content)
		fmt.Printf("Upvotes :%d  Downvotes :%d", post.VoteCount.Upvotes, post.VoteCount.Downvotes)
		fmt.Println("------------------------")
	}

	fmt.Printf("\nDirect Messages (latest 5):\n")
	fmt.Println("------------------------")
	maxMessages := 5
	if len(messages) < maxMessages {
		maxMessages = len(messages)
	}
	for i := 0; i < maxMessages; i++ {
		msg := messages[i]
		fmt.Printf("From User %d: %s\n", msg.FromUser, msg.Content)
		fmt.Println("------------------------")
	}
}

func printStats(state *SimState) {
	// Original statistics code...
	fmt.Println("\nSimulation Statistics:")
	fmt.Printf("Total Users: %d\n", len(state.users))
	fmt.Printf("Total Subreddits: %d\n", len(state.subreddits))
	fmt.Printf("Total Posts: %d\n", len(state.posts))
	fmt.Printf("Total Comments: %d\n", len(state.comments))
	fmt.Printf("Total Direct Messages: %d\n", len(state.messages))

	fmt.Println("\nSubreddit Member Distribution:")
	type subCount struct {
		name  string
		count int
	}

	counts := make([]subCount, 0, len(state.subreddits))
	for _, sub := range state.subreddits {
		counts = append(counts, subCount{sub.Name, state.memberCounts[sub.ID]})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	for _, sc := range counts {
		fmt.Printf("%s: %d members\n", sc.name, sc.count)
	}

	getKarmaLeaders(3)

	// Add detailed information for 3 random users
	fmt.Println("\n=== Random User Details ===")

	// Get 3 random unique users
	userIndices := rand.Perm(len(state.users))[:1]

	for _, idx := range userIndices {
		user := state.users[idx]

		// Get user details including karma
		userDetails, err := getUserDetails(user.ID)
		if err != nil {
			log.Printf("Failed to get details for user %d: %v", user.ID, err)
			continue
		}

		// Get user's feed
		feed, err := getUserFeed(user.ID)
		if err != nil {
			log.Printf("Failed to get feed for user %d: %v", user.ID, err)
			continue
		}

		// Get user's messages
		messages, err := getUserMessages(user.ID)
		if err != nil {
			log.Printf("Failed to get messages for user %d: %v", user.ID, err)
			continue
		}

		// Print detailed information for this user
		printUserDetails(*userDetails, feed, messages)
	}
}

// Helper functions for generating random content
func generateRandomTitle() string {
	titles := []string{
		"I came across something fascinating today!",
		"How do you feel about this?",
		"A unique take on this caught my attention.",
		"Could really use your input on something.",
		"Take a look at this—it’s intriguing!",
		"I stumbled upon something you might find interesting.",
		"What’s your opinion on this idea?",
		"I’d love to hear your thoughts on this discovery.",
	}
	return titles[rand.Intn(len(titles))]
}

func generateRandomContent() string {
	contents := []string{
		"This has been on my mind lately...",
		"Let me share my perspective on this...",
		"How do you see this situation?",
		"I’m curious to know your viewpoint...",
		"We should definitely have a conversation about this...",
		"What’s your take on this topic?",
		"I’d value your insights on this...",
		"This feels like an important topic to dive into...",
	}
	return contents[rand.Intn(len(contents))] + " " + randomSentence()
}

func generateRandomComment() string {
	comments := []string{
		"That's a solid perspective!",
		"Totally on board with this.",
		"What if we looked at it this way?",
		"Appreciate you bringing this up!",
		"This makes me think of something similar I read.",
		"Great insight, thanks for posting!",
		"Have you thought about it from this angle?",
		"This connects so well with another discussion I saw!",
	}
	return comments[rand.Intn(len(comments))] + " " + randomSentence()
}

func randomSentence() string {
	sentences := []string{
		"The depth of discussion here is truly commendable.",
		"This angle is really intriguing and new to me.",
		"It’d be great to dive deeper into this together.",
		"Can’t wait to see more viewpoints on this topic.",
		"This idea could lead to some exciting possibilities.",
		"I appreciate the effort everyone’s putting into this conversation.",
		"This perspective opens up some unique questions.",
		"I’d love to unpack this concept further with others.",
	}
	return sentences[rand.Intn(len(sentences))]
}
