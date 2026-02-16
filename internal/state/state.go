package state

import (
    "bootdevblog/internal/config"
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
    "bootdevblog/internal/database"
    "os"
    "bootdevblog/rss"
)

type State struct {
    Config *config.Config
    DB     *database.Queries
}

type Command struct {
    Name string
    Args []string
}

type Commands struct {
    handlers map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {
    if len(cmd.Args) == 0 {
        return fmt.Errorf("usage: login <username>")
    }
    username := cmd.Args[0]

    _, err := s.DB.GetUserByUsername(context.Background(), username)
    if err != nil {
        return fmt.Errorf("user %s does not exist", username)
    }

    err = s.Config.SetUser(username)
    if err != nil {
        return err
    }

    fmt.Printf("User set to: %s\n", username)
    return nil
}


func (c *Commands) Register(name string, f func(*State, Command) error) {
    if c.handlers == nil {
        c.handlers = make(map[string]func(*State, Command) error)
    }
    c.handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
    handler, exists := c.handlers[cmd.Name]
    if !exists {
        return fmt.Errorf("unknown command: %s", cmd.Name)
    }
    return handler(s, cmd)
}

func NewCommands() *Commands {
    return &Commands{
        handlers: make(map[string]func(*State, Command) error),
    }
}

func HandlerRegister(s *State, cmd Command) error {
    if len(cmd.Args) == 0 {
        return fmt.Errorf("usage: register <username>")
    }
    username := cmd.Args[0]

    _, err := s.DB.GetUserByUsername(context.Background(), username)
    if err == nil {
        fmt.Printf("User with name %s already exists\n", username)
        os.Exit(1)
    }

    userID := uuid.New()
    now := time.Now().UTC()

    _, err = s.DB.CreateUser(context.Background(), database.CreateUserParams{
        ID:        userID,
        Username:  username,
        CreatedAt: now,
        UpdatedAt: now,
    })
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    err = s.Config.SetUser(username)
    if err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    fmt.Printf("User '%s' created successfully\n", username)
    fmt.Printf("Debug: ID=%s, CreatedAt=%s\n", userID, now.Format(time.RFC3339))

    return nil
}

func HandlerReset(s *State, cmd Command) error {
    err := s.DB.DeleteAllUsers(context.Background())
    if err != nil {
        return fmt.Errorf("failed to reset database: %w", err)
    }
    fmt.Println("Database reset successfully")
    return nil
}

func HandlerUsers(s *State, cmd Command) error {
    users, err := s.DB.GetUsers(context.Background())
    if err != nil {
        return fmt.Errorf("failed to fetch users: %w", err)
    }

    currentUser := s.Config.CurrentUserName
    for _, u := range users {
        if u.Username == currentUser {
            fmt.Printf("* %s (current)\n", u.Username)
        } else {
            fmt.Printf("  %s\n", u.Username)
        }
    }

    return nil
}

func HandlerAgg(s *State, cmd Command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: agg <time_between_reqs>")
    }
    duration, err := time.ParseDuration(cmd.Args[0])
    if err != nil {
        return fmt.Errorf("invalid duration: %w", err)
    }

    fmt.Printf("Collecting feeds every %s\n", duration)

    ticker := time.NewTicker(duration)
    defer ticker.Stop()

    if err := scrapeFeeds(s); err != nil {
        fmt.Printf("Error on first scrape: %v\n", err)
    }

    for range ticker.C {
        if err := scrapeFeeds(s); err != nil {
            fmt.Printf("Error scraping feeds: %v\n", err)
        }
    }
}

func HandlerAddFeed(s *State, cmd Command) error {
    if len(cmd.Args) != 2 {
        return fmt.Errorf("usage: addfeed <name> <url>")
    }
    name := cmd.Args[0]
    url := cmd.Args[1]

    currentUser, err := s.GetCurrentUser(context.Background())
    if err != nil {
        return err
}

    feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        Name:      name,
        Url:       url,
        UserID:    currentUser.ID,
    })
    if err != nil {
        return fmt.Errorf("failed to create feed: %w", err)
    }

    _, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        UserID:    currentUser.ID,
        FeedID:    feed.ID,
    })
    if err != nil {
        return fmt.Errorf("failed to follow feed: %w", err)
    }

    fmt.Printf("Feed created and followed: %s (%s)\n", feed.Name, feed.Url)
    return nil
}

func HandlerListFeeds(s *State, cmd Command) error {
    feeds, err := s.DB.GetFeedsWithUser(context.Background())
    if err != nil {
        return fmt.Errorf("failed to fetch feeds: %w", err)
    }

    fmt.Printf("Feeds:\n")
    for _, feed := range feeds {
        fmt.Printf(" * %s (%s) - %s\n", feed.Name, feed.Url, feed.UserName)
    }
    return nil
}

func HandlerFollow(s *State, cmd Command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: follow <feed_url>")
    }
    feedURL := cmd.Args[0]

    currentUser, err := s.GetCurrentUser(context.Background())
    if err != nil {
        return err
    }

    feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
    if err != nil {
        return fmt.Errorf("feed not found: %w", err)
    }

    follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
        ID:        uuid.New(),
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        UserID:    currentUser.ID,
        FeedID:    feed.ID,
    })
    if err != nil {
        return fmt.Errorf("failed to follow feed: %w", err)
    }

    fmt.Printf("Following feed: %s (user: %s)\n", follow.FeedName, follow.UserName)
    return nil
}

func HandlerFollowing(s *State, cmd Command) error {
    currentUser, err := s.GetCurrentUser(context.Background())
    if err != nil {
        return err
    }

    follows, err := s.DB.GetFeedFollowsForUser(context.Background(), currentUser.ID)
    if err != nil {
        return fmt.Errorf("failed to get follows: %w", err)
    }

    fmt.Printf("Feeds %s is following:\n", currentUser.Username)
    for _, f := range follows {
        fmt.Printf(" * %s\n", f.FeedName)
    }
    return nil
}

func (s *State) GetCurrentUser(ctx context.Context) (database.User, error) {
    if s.Config.CurrentUserName == "" {
        return database.User{}, fmt.Errorf("no user logged in")
    }
    user, err := s.DB.GetUserByUsername(ctx, s.Config.CurrentUserName)
    if err != nil {
        return database.User{}, fmt.Errorf("user not found: %w", err)
    }
    return user, nil
}

func HandlerUnfollow(s *State, cmd Command) error {
    if len(cmd.Args) != 1 {
        return fmt.Errorf("usage: unfollow <feed_url>")
    }
    feedURL := cmd.Args[0]

    currentUser, err := s.GetCurrentUser(context.Background())
    if err != nil {
        return err
    }

    feed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
    if err != nil {
        return fmt.Errorf("feed not found: %w", err)
    }

    err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
        UserID: currentUser.ID,
        FeedID: feed.ID,
    })
    if err != nil {
        return fmt.Errorf("failed to unfollow feed: %w", err)
    }

    fmt.Printf("Unfollowed feed: %s\n", feed.Name)
    return nil
}

func scrapeFeeds(s *State) error {
    feed, err := s.DB.GetNextFeedToFetch(context.Background())
    if err != nil {
        return fmt.Errorf("failed to get feed: %w", err)
    }

    rssFeed, err := rss.FetchFeed(context.Background(), feed.Url)
    if err != nil {
        return fmt.Errorf("failed to fetch RSS feed: %w", err)
    }

    for _, item := range rssFeed.Channel.Items {
        publishedAt := sql.NullTime{Valid: false}
        if t, err := time.Parse(time.RFC1123, item.PubDate); err == nil {
            publishedAt = sql.NullTime{Time: t, Valid: true}
        }

        _, err := s.DB.CreatePost(context.Background(), database.CreatePostParams{
            ID:         uuid.New(),
            CreatedAt:  time.Now().UTC(),
            UpdatedAt:  time.Now().UTC(),
            Title:      item.Title,
            Url:        item.Link,
            Description: sql.NullString{String: item.Description, Valid: true},
            PublishedAt: publishedAt,
            FeedID:     feed.ID,
        })
        if err != nil {
            if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
                continue // Ignore duplicate URL errors
            }
            return fmt.Errorf("failed to create post: %w", err)
        }
    }

    _, err = s.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
        ID:           feed.ID,
        LastFetchedAt: time.Now().UTC(),
    })
    if err != nil {
        return fmt.Errorf("failed to mark feed fetched: %w", err)
    }

    return nil
}

func HandlerBrowse(s *State, cmd Command) error {
    var limit int = 2
    if len(cmd.Args) > 0 {
        parsed, err := strconv.Atoi(cmd.Args[0])
        if err != nil {
            return fmt.Errorf("invalid limit: must be a number")
        }
        limit = parsed
    }

    currentUser, err := s.GetCurrentUser(context.Background())
    if err != nil {
        return err
    }

    posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
        UserID: currentUser.ID,
        Limit:  int32(limit),
    })
    if err != nil {
        return fmt.Errorf("failed to get posts: %w", err)
    }

    fmt.Printf("Showing %d recent posts for %s:\n", len(posts), currentUser.Username)
    for _, post := range posts {
        feedName := "Unknown"
        if post.FeedName.Valid {
            feedName = post.FeedName.String
        }
        fmt.Printf("\n- %s\n  (%s)\n", post.Title, feedName)
        if post.Description.Valid {
            fmt.Printf("  %s\n", post.Description.String)
        }
    }
    return nil
}
