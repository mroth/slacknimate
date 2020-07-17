// Command slacknimate is a basic CLI client for the slacknimate library that
// posts each line from os.Stdin.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/mroth/slacknimate"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v2"
)

var version = "development"

type options struct {
	apiToken string
	channel  string
	delay    float64
	loop     bool
	preview  bool

	slackUsername  string
	slackIconURL   string
	slackIconEmoji string
}

func main() {
	app := cli.App{
		Name:            "slacknimate",
		Usage:           "text animation for Slack messages",
		Version:         version,
		UsageText:       "slacknimate [options]",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"a"},
				Usage:   "Slack API token*",
				EnvVars: []string{"SLACK_TOKEN"},
			},
			&cli.StringFlag{
				Name:    "channel",
				Aliases: []string{"c"},
				Usage:   "Slack channel*",
				EnvVars: []string{"SLACK_CHANNEL"},
			},
			&cli.StringFlag{
				Name:    "username",
				Usage:   "Slack username",
				EnvVars: []string{"SLACK_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "icon-url",
				Usage:   "Slack icon from url",
				EnvVars: []string{"SLACK_ICON_URL"},
			},
			&cli.StringFlag{
				Name:    "icon-emoji",
				Usage:   "Slack icon from emoji",
				EnvVars: []string{"SLACK_ICON_EMOJI"},
			},
			&cli.Float64Flag{
				Name:    "delay",
				Aliases: []string{"d"},
				Usage:   "minimum delay between frames",
				Value:   1,
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "loop content upon reaching EOF",
			},
			&cli.BoolFlag{
				Name:  "preview",
				Usage: "preview on terminal only",
			},
		},
		Action: func(c *cli.Context) error {
			opts, err := parseOpts(c)
			if err != nil {
				return err
			}
			return post(c.Context, opts)
		},
	}

	// handle graceful shutdowns
	ctx, cancel := context.WithCancel(context.Background())
	interruptC := make(chan os.Signal, 1)
	signal.Notify(interruptC, os.Interrupt)
	go func() {
		c := <-interruptC
		log.Printf("Got %v signal. Aborting...", c)
		cancel()
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

// functionality to extract CLI options with some custom error handling
// that would be annoying to model via the cli module.
func parseOpts(c *cli.Context) (options, error) {
	opts := options{
		apiToken: c.String("token"),
		channel:  c.String("channel"),
		delay:    c.Float64("delay"),
		loop:     c.Bool("loop"),
		preview:  c.Bool("preview"),

		slackUsername:  c.String("username"),
		slackIconURL:   c.String("icon-url"),
		slackIconEmoji: c.String("icon-emoji"),
	}
	if !opts.preview {
		if opts.apiToken == "" {
			return opts, errors.New("api-token is required")
		}
		if opts.channel == "" {
			return opts, errors.New("channel is required")
		}
		if opts.delay < 0.001 {
			return opts, errors.New("delay must be >= 0.001 to avoid creating a time paradox")
		}
	}
	return opts, nil
}

func post(ctx context.Context, opts options) error {
	var frames <-chan string
	if opts.loop {
		frames = slacknimate.NewLoopingLineScanner(ctx, os.Stdin, 4096).Frames()
	} else {
		frames = slacknimate.NewLineScanner(ctx, os.Stdin).Frames()
	}

	delay := time.Millisecond * time.Duration(opts.delay*1000)
	if opts.preview {
		return previewer(ctx, frames, delay)
	}

	api := slack.New(opts.apiToken)
	err := slacknimate.Updater(context.Background(), api, opts.channel, frames, slacknimate.UpdaterOptions{
		Username:  opts.slackUsername,
		IconURL:   opts.slackIconURL,
		IconEmoji: opts.slackIconEmoji,
		MinDelay:  delay,
		UpdateFunc: func(u slacknimate.Update) {
			if u.Err == nil {
				log.Printf("posted frame %v/%v: %v",
					u.Dst, u.TS, u.Frame,
				)
			} else {
				log.Printf("ERROR updating %v/%v with frame %v: %v",
					u.Dst, u.TS, u.Frame, u.Err)
			}
		},
	})
	return err
}

func previewer(ctx context.Context, frames <-chan string, delay time.Duration) error {
	delayTicker := time.NewTicker(delay)
	defer delayTicker.Stop()
	for frame := range frames {
		select {
		case <-delayTicker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
		fmt.Printf("\033[2K\r%s", frame)
	}
	return nil
}
