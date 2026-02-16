package main

import (
    "bootdevblog/internal/config"
    "bootdevblog/internal/state"
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
    "bootdevblog/internal/database"
)

func main() {
    cfg, err := config.Read()
    if err != nil {
        log.Fatal(err)
    }

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil { log.Fatal(err) }
    defer db.Close()

    dbQueries := database.New(db)

    s := &state.State{
        Config: &cfg,
        DB:     dbQueries,
    }

    cmds := state.NewCommands()
    cmds.Register("login", state.HandlerLogin)
    cmds.Register("register", state.HandlerRegister)
    cmds.Register("reset", state.HandlerReset)
    cmds.Register("users", state.HandlerUsers)
    cmds.Register("agg", state.HandlerAgg)
    cmds.Register("addfeed", state.HandlerAddFeed)
    cmds.Register("feeds", state.HandlerListFeeds)
    cmds.Register("follow", state.HandlerFollow)
    cmds.Register("following", state.HandlerFollowing)
    cmds.Register("unfollow", state.HandlerUnfollow)
    cmds.Register("browse", state.HandlerBrowse)

    if len(os.Args) < 2 {
        fmt.Println("error: not enough arguments")
        fmt.Println("usage: gator <command> [<args>...]")
        os.Exit(1)
    }

    cmd := state.Command{
        Name: os.Args[1],
        Args: os.Args[2:],
    }

    err = cmds.Run(s, cmd)
    if err != nil {
        log.Fatal(err)
    }

}
