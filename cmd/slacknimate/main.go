package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
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
				Name:    "api-token",
				Aliases: []string{"a"},
				Usage:   "API token*",
				EnvVars: []string{"SLACK_TOKEN"},
			},
			&cli.Float64Flag{
				Name:    "delay",
				Aliases: []string{"d"},
				Usage:   "minimum delay between frames",
				Value:   1,
			},
			&cli.StringFlag{
				Name:    "channel",
				Aliases: []string{"c"},
				Usage:   "channel/destination*",
				EnvVars: []string{"SLACK_CHANNEL"},
			},
			&cli.BoolFlag{
				Name:    "loop",
				Aliases: []string{"l"},
				Usage:   "loop content upon reaching end",
			},
			&cli.BoolFlag{
				Name:  "preview",
				Usage: "preview on terminal instead of posting",
			},
		},
		Action: func(c *cli.Context) error {
			opts, err := parseOpts(c)
			if err != nil {
				return err
			}
			return post(opts)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// functionality to extract CLI options with some custom error handling
// that would be annoying to model via the cli module.
func parseOpts(c *cli.Context) (options, error) {
	opts := options{
		apiToken: c.String("api-token"),
		channel:  c.String("channel"),
		delay:    c.Float64("delay"),
		loop:     c.Bool("loop"),
		preview:  c.Bool("preview"),
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

func post(opts options) error {
	// for now, just use default context, but will want to adjust in future
	ctx := context.TODO()

	// setup frame source
	var frames <-chan string
	if opts.loop {
		frames = slacknimate.NewLoopingLineScanner(ctx, os.Stdin, 4096).Frames()
	} else {
		frames = slacknimate.NewLineScanner(ctx, os.Stdin).Frames()
	}

	delay := time.Millisecond * time.Duration(opts.delay*1000)
	if opts.preview {
		previewer(ctx, frames, delay)
		os.Exit(0)
	}

	api := slack.New(opts.apiToken)
	err := slacknimate.Updater(context.Background(), api, opts.channel, frames, slacknimate.UpdaterOptions{
		// Username:  "Animation Funtime",
		// IconEmoji: "cat",
		MinDelay: delay,
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

func previewer(ctx context.Context, frames <-chan string, delay time.Duration) {
	delayTicker := time.NewTicker(delay)
	defer delayTicker.Stop()
	for frame := range frames {
		select {
		case <-delayTicker.C:
		case <-ctx.Done():
			return
		}
		fmt.Printf("\033[2K\r%s", frame)
	}
}
