package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type Issue struct {
	Title    string
	Body     string
	Comments []string
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Issue number is required")
	}

	issueNumber, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid issue number: %v", err)
	}

	owner := os.Getenv("REPO_OWNER")
	repo := os.Getenv("REPO_NAME")

	if owner == "" || repo == "" {
		log.Fatal("REPO_OWNER and REPO_NAME environment variables are required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	processIssue(client, owner, repo, issueNumber)
}

func getIssueDetails(client *github.Client, owner, repo string, issueNumber int) (*Issue, error) {
	issue, _, err := client.Issues.Get(context.Background(), owner, repo, issueNumber)
	if err != nil {
		return nil, err
	}

	// TODO: add comments to the issue as well as the ability to follow up to comments?
	return &Issue{
		Title:    *issue.Title,
		Body:     *issue.Body,
		Comments: []string{},
	}, nil
}

func postComment(client *github.Client, owner, repo string, issueNumber int, commentBody string) error {
	comment := &github.IssueComment{
		Body: github.String(commentBody),
	}
	_, _, err := client.Issues.CreateComment(context.Background(), owner, repo, issueNumber, comment)
	return err
}

func processIssue(client *github.Client, owner, repo string, issueNumber int) {
	issue, err := getIssueDetails(client, owner, repo, issueNumber)
	if err != nil {
		log.Fatalf("Failed to get issue details: %v", err)
	}

	fmt.Printf("Title: %s\n", issue.Title)
	fmt.Printf("Body: %s\n", issue.Body)
	for i, comment := range issue.Comments {
		fmt.Printf("Comment %d: %s\n", i+1, comment)
	}

	followUpQuestions := callGemmaAPI(issue.Body)

	err = postComment(client, owner, repo, issueNumber, followUpQuestions)
	if err != nil {
		log.Fatalf("Failed to post comment: %v", err)
	}

	fmt.Printf("Successfully posted follow-up questions as a comment!\n")
}

func callGemmaAPI(text string) string {
	ctx := context.Background()

	// Get Google AI API key from environment
	// Must be set GOOGLE_AI_API_KEY in screts in github repo which is using this code
	// You can read how to create an API key for gemma here => https://ai.google.dev/gemini-api/docs/api-key
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: GOOGLE_AI_API_KEY not set, using placeholder response")
		return "Please provide more details about this issue. What steps have you tried so far?"
	}

	// Initialize the Google AI client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating AI client: %v", err)
		return "Please provide more details about this issue."
	}
	defer client.Close()

	// TODO: Make model configurable
	model := client.GenerativeModel("gemini-1.5-flash")
	// TODO: Add repo context? Hopefully without hallucination
	prompt := fmt.Sprintf(`Analyze the following GitHub issue and generate helpful follow-up questions that would help clarify the problem or gather more information needed to resolve it.

Issue content: %s

Please generate 2-3 specific, actionable follow-up questions that would help the maintainers better understand and resolve this issue. Format your response as a clear, helpful comment.`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return "Please provide more details about this issue. What steps have you tried so far?"
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("## AI-Generated Follow-up Questions\n\n%v", resp.Candidates[0].Content.Parts[0])
	}

	return "Please provide more details about this issue."
}
