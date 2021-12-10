package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli"
	"github.com/weberc2/comments/pkg/pgcommentsstore"
	"github.com/weberc2/comments/pkg/types"
)

func main() {
	app := cli.App{
		Name: "pgcommentsstore",
		Commands: []cli.Command{
			{
				Name: "table",
				Subcommands: cli.Commands{
					{
						Name:        "ensure",
						Description: "ensure that the `comments` table exists",
						ShortName:   "e",
						Action: withStore(func(
							store *pgcommentsstore.PGCommentsStore,
							ctx *cli.Context,
						) error {
							return store.EnsureTable()
						}),
					},
					{
						Name:        "drop",
						Description: "drop the `comments` table",
						ShortName:   "d",
						Action: withStore(func(
							store *pgcommentsstore.PGCommentsStore,
							ctx *cli.Context,
						) error {
							return store.DropTable()
						}),
					},
					{
						Name:        "reset",
						Description: "reset the `comments` table (drop + ensure)",
						ShortName:   "r",
						Action: withStore(func(
							store *pgcommentsstore.PGCommentsStore,
							ctx *cli.Context,
						) error {
							return store.DropTable()
						}),
					},
					{
						Name:        "clear",
						Description: "clear the `comments` table",
						ShortName:   "c",
						Action: withStore(func(
							store *pgcommentsstore.PGCommentsStore,
							ctx *cli.Context,
						) error {
							return store.ClearTable()
						}),
					},
				},
			},
			{
				Name:        "put",
				Aliases:     []string{"add", "create"},
				Description: "add a comment",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "The comment's ID. Defaults to a UUID.",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "post",
						Usage:    "The comment's post. Required.",
						Required: true,
					},
					&cli.StringFlag{
						Name: "parent",
						Usage: "The comment's parent. Defaults to " +
							"toplevel comment.",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "author",
						Usage:    "The comment's author. Required.",
						Required: true,
					},
					&cli.StringFlag{
						Name: "created",
						Usage: "The time the comment was created. Defaults " +
							"to the current time.",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "body",
						Usage:    "The comment's body. Defaults to empty.",
						Required: false,
					},
				},
				Action: withStore(func(
					store *pgcommentsstore.PGCommentsStore,
					ctx *cli.Context,
				) error {
					input := &types.Comment{
						ID:     types.CommentID(ctx.String("id")),
						Post:   types.PostID(ctx.String("post")),
						Parent: types.CommentID(ctx.String("parent")),
						Author: types.UserID(ctx.String("author")),
						Body:   ctx.String("body"),
					}
					createdValue := ctx.String("created")
					if createdValue != "" {
						t, err := time.Parse(time.RFC3339, createdValue)
						if err != nil {
							log.Fatalf("parsing `created` flag: %v", err)
						}
						input.Created = t
					} else {
						input.Created = time.Now().UTC()
					}
					modifiedValue := ctx.String("modified")
					if modifiedValue != "" {
						t, err := time.Parse(time.RFC3339, modifiedValue)
						if err != nil {
							log.Fatalf("parsing `modified` flag: %v", err)
						}
						input.Modified = t
					} else {
						input.Modified = time.Now().UTC()
					}

					if input.ID == "" {
						input.ID = types.CommentID(uuid.NewString())
					}
					comment, err := store.Put(input)
					data, err := json.MarshalIndent(comment, "", "  ")
					if err != nil {
						return err
					}
					_, err = fmt.Printf("%s\n", data)
					return err
				}),
			},
			{
				Name:        "comment",
				Aliases:     []string{"get", "fetch"},
				Description: "Retrieve a comment.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "post",
						Usage:    "The comment's post.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "id",
						Usage:    "The comment's ID.",
						Required: true,
					},
				},
				Action: withStore(func(
					store *pgcommentsstore.PGCommentsStore,
					ctx *cli.Context,
				) error {
					c, err := store.Comment(
						types.PostID(ctx.String("post")),
						types.CommentID(ctx.String("id")),
					)
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(c, "", "  ")
					if err != nil {
						return err
					}
					_, err = fmt.Printf("%s\n", data)
					return err
				}),
			},
			{
				Name:        "replies",
				Description: "Fetch replies to a comment or post.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "post",
						Usage:    "The comment's post.",
						Required: true,
					},
					&cli.StringFlag{
						Name: "parent",
						Usage: "The parent of the replies to be fetched. " +
							"If this is omitted, all comments associated " +
							"with the provided post will be retrieved.",
						Required: false,
					},
				},
				Action: withStore(func(
					store *pgcommentsstore.PGCommentsStore,
					ctx *cli.Context,
				) error {
					replies, err := store.Replies(
						types.PostID(ctx.String("post")),
						types.CommentID(ctx.String("parent")),
					)
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(replies, "", "  ")
					if err != nil {
						return err
					}
					_, err = fmt.Printf("%s\n", data)
					return err
				}),
			},
			{
				Name:        "delete",
				Description: "Deletes a comment.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "post",
						Usage:    "The comment's post.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "id",
						Usage:    "The comment's ID.",
						Required: true,
					},
				},
				Action: withStore(func(
					store *pgcommentsstore.PGCommentsStore,
					ctx *cli.Context,
				) error {
					return store.Delete(
						types.PostID(ctx.String("post")),
						types.CommentID(ctx.String("id")),
					)
				}),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func withStore(
	f func(
		store *pgcommentsstore.PGCommentsStore,
		ctx *cli.Context,
	) error,
) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		store, err := pgcommentsstore.OpenEnv()
		if err != nil {
			log.Fatalf("opening comments store: %v", err)
		}
		return f(store, ctx)
	}
}
