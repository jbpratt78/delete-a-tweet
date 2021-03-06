package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/browser"

	"github.com/c-bata/go-prompt"
)

var (
	client  *twitter.Client
	user    *twitter.User
	current twitter.Tweet
	tweets  []twitter.Tweet
	curr    int
)

func main() {
	consumerKey := flag.String("consumerKey", "", "Twitter Consumer Key")
	consumerSecret := flag.String("consumerSecret", "", "Twitter Consumer Secret")
	accessToken := flag.String("accessToken", "", "Twitter Access Token")
	accessSecret := flag.String("accessSecret", "", "Twitter Access Secret")
	flag.Parse()

	newClient, newUser, err := login(*consumerKey, *consumerSecret, *accessToken, *accessSecret)
	if err != nil {
		panic(err)
	}

	client = newClient
	user = newUser

	fmt.Println(`
	██████╗ ███████╗██╗     ███████╗████████╗███████╗     █████╗    ████████╗██╗    ██╗███████╗███████╗████████╗
	██╔══██╗██╔════╝██║     ██╔════╝╚══██╔══╝██╔════╝    ██╔══██╗   ╚══██╔══╝██║    ██║██╔════╝██╔════╝╚══██╔══╝
	██║  ██║█████╗  ██║     █████╗     ██║   █████╗█████╗███████║█████╗██║   ██║ █╗ ██║█████╗  █████╗     ██║
	██║  ██║██╔══╝  ██║     ██╔══╝     ██║   ██╔══╝╚════╝██╔══██║╚════╝██║   ██║███╗██║██╔══╝  ██╔══╝     ██║
	██████╔╝███████╗███████╗███████╗   ██║   ███████╗    ██║  ██║      ██║   ╚███╔███╔╝███████╗███████╗   ██║
	╚═════╝ ╚══════╝╚══════╝╚══════╝   ╚═╝   ╚══════╝    ╚═╝  ╚═╝      ╚═╝    ╚══╝╚══╝ ╚══════╝╚══════╝   ╚═╝
	`)
	fmt.Println("Please type 'exit' to terminate this program")
	fmt.Println("You are signed in as", user.ScreenName)
	fmt.Println("Begin by typing 'load' to load all of your tweets")
	p := prompt.New(executor, completer)
	p.Run()
}

func reviewTweet() {
	if curr > len(tweets) {
		fmt.Println("All out of tweets")
		return
	}
	current = tweets[curr]
	time, err := current.CreatedAtTime()
	if err != nil {
		return
	}

	fmt.Printf(`
%s                                    
									 
%d (R) %d (r) %d (f)
Created at: %s
(%d)
`,
		current.Text, current.ReplyCount, current.RetweetCount,
		current.FavoriteCount, time.String(), current.ID)
}

func executor(in string) {
	switch in {
	case "e", "exit":
		fmt.Println("exiting...")
		os.Exit(0)
	case "load":
		err := loadTweets(user.StatusesCount)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Loaded %d tweets\n", len(tweets))
		fmt.Println("Type 'review' to begin..")
	case "d", "delete":
		fmt.Println("deleting tweet ", current.ID)
		if err := delete(current.ID); err != nil {
			log.Fatal(err)
		}
	case "review":
		fmt.Println("Beginning review...")
		fmt.Println("'n' or 'next' to continue")
		fmt.Println("'o' or 'open' to open current tweet in browser")
		fmt.Println("'d' or 'delete' to delete the current tweet")
		fmt.Println("'b' or 'back' to navigate to the previous tweet")
		fmt.Println("'e' or 'exit' to exit")
		reviewTweet()
		curr++
	case "n", "next":
		reviewTweet()
		curr++
	case "b", "back":
		if curr == 0 {
			fmt.Println("Already as far back as you can go")
			return
		}
		curr--
		reviewTweet()
	case "o", "open":
		browser.OpenURL("https://twitter.com/" + user.ScreenName + "/status/" + current.IDStr)
	default:
		return
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{
		{Text: "exit"},
		{Text: "load", Description: "Load all of your tweets"},
	}
	if len(tweets) > 0 {
		suggestions = append(suggestions,
			prompt.Suggest{Text: "delete", Description: "Delete the current tweet"},
			prompt.Suggest{Text: "review", Description: "Review loaded tweets"},
			prompt.Suggest{Text: "next", Description: "Continue on reviewing the next tweet"},
			prompt.Suggest{Text: "open", Description: "Open current tweet in default browser"},
		)
	}
	if curr > 0 {
		suggestions = append(suggestions, prompt.Suggest{Text: "back", Description: "Navigate to the previous tweet"})
	}
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

func login(consumerKey, consumerSecret, accessToken, accessSecret string) (*twitter.Client, *twitter.User, error) {
	if consumerKey == "" || consumerSecret == "" {
		return nil, nil, fmt.Errorf("Please provide a consumer key and secret")
	}
	if accessToken == "" || accessSecret == "" {
		return nil, nil, fmt.Errorf("Must supply access token and secret for user context")
	}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	client := twitter.NewClient(config.Client(oauth1.NoContext, oauth1.NewToken(accessToken, accessSecret)))

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(false),
	}

	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to validate credentials: %v", err)

	}

	return client, user, nil
}

func loadTweets(count int) error {
	params := &twitter.UserTimelineParams{UserID: user.ID, Count: count}
	x, _, err := client.Timelines.UserTimeline(params)
	if err != nil {
		return fmt.Errorf("Failed to get user timeline: %v", err)
	}
	tweets = x
	return nil
}

func delete(id int64) error {
	params := &twitter.StatusDestroyParams{ID: id}
	_, _, err := client.Statuses.Destroy(id, params)
	if err != nil {
		return fmt.Errorf("Failed to destroy tweet: %v", err)
	}
	return nil
}
