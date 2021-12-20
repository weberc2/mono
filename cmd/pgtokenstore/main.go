package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/weberc2/auth/pkg/pgtokenstore"
)

type Store = pgtokenstore.PGTokenStore

func main() {
	app := cli.App{
		Name:        "pgtokenstore",
		Description: "a command line `PGTokenStore` interface",
		Commands: []*cli.Command{{
			Name:        "table",
			Description: "commands for interacting with the backing pg table",
			Subcommands: []*cli.Command{{
				Name:        "ensure",
				Aliases:     []string{"make", "create"},
				Description: "create the table if it doesn't already exist",
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.EnsureTable()
				}),
			}, {
				Name:        "drop",
				Aliases:     []string{"delete", "destroy"},
				Description: "drop the postgres table",
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.DropTable()
				}),
			}, {
				Name:        "reset",
				Description: "delete and recreate the postgres table",
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.ResetTable()
				}),
			}, {
				Name: "clear",
				Description: "clear the rows from the table without " +
					"dropping it",
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.ClearTable()
				}),
			}},
		}, {
			Name:        "tokens",
			Description: "commands for managing tokens",
			Subcommands: []*cli.Command{{
				Name:        "put",
				Aliases:     []string{"add", "create", "make", "insert"},
				Description: "put a token into the token store",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "token",
						Usage:    "the token string itself",
						Required: true,
					},
					&cli.TimestampFlag{
						Name: "expires",
						Usage: "the timestamp for token expiry. Defaults to " +
							"7 days from now.",
						Required: false,
						Layout:   time.RFC3339,
					},
				},
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					token := ctx.String("token")
					exp := ctx.Timestamp("expires")
					if exp == nil {
						t := time.Now().UTC().Add(7 * 24 * time.Hour)
						exp = &t
					}
					return store.Put(token, *exp)
				}),
			}, {
				Name:        "list",
				Description: "list tokens in the store",
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					tokens, err := store.List()
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(tokens, "", "  ")
					if err != nil {
						return fmt.Errorf("marshaling tokens to JSON: %w", err)
					}
					if _, err := fmt.Printf("%s\n", data); err != nil {
						return fmt.Errorf("writing JSON to stdout: %w", err)
					}
					return nil
				}),
			}, {
				Name: "exists",
				Description: "returns no output if the token exists or else " +
					"an error message",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "token",
						Usage:    "the token to search for",
						Required: true,
					},
				},
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.Exists(ctx.String("token"))
				}),
			}, {
				Name:        "delete",
				Aliases:     []string{"rm", "remove"},
				Description: "delete a token from the store",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "token",
						Usage:    "the token to delete",
						Required: true,
					},
				},
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					return store.Delete(ctx.String("token"))
				}),
			}, {
				Name:        "delete-expired",
				Description: "delete all expired tokens from the store",
				Flags: []cli.Flag{
					&cli.TimestampFlag{
						Name: "now",
						Usage: "the time for which to evaluate expiration. " +
							"defaults to the current time.",
						Layout:   time.RFC3339,
						Required: false,
					},
				},
				Action: withStore(func(store *Store, ctx *cli.Context) error {
					now := ctx.Timestamp("now")
					if now == nil {
						t := time.Now().UTC()
						now = &t
					}
					return store.DeleteExpired(*now)
				}),
			}},
		}},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func withStore(f func(*Store, *cli.Context) error) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		store, err := pgtokenstore.OpenEnv()
		if err != nil {
			return fmt.Errorf("opening PGTokenStore: %w", err)
		}
		return f(store, ctx)
	}
}
