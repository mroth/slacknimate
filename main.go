package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "slacknimate"
	app.Usage = "text animation for Slack messages"
	app.Version = "1.0.1"
	app.UsageText = "slacknimate [options]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "api-token, a",
			Usage:  "API token*",
			EnvVar: "SLACK_TOKEN",
		},
		cli.Float64Flag{
			Name:  "delay, d",
			Usage: "minimum delay between frames",
			Value: 1,
		},
		cli.StringFlag{
			Name:   "channel, c",
			Usage:  "channel/destination*",
			EnvVar: "SLACK_CHANNEL",
		},
		cli.BoolFlag{
			Name:  "loop, l",
			Usage: "loop content upon reaching end",
		},
		cli.BoolFlag{
			Name:  "preview",
			Usage: "preview on terminal instead of posting",
		},
	}
	app.Action = func(c *cli.Context) {
		apiToken := c.String("api-token")
		channel := c.String("channel")
		delay := c.Float64("delay")
		noop := c.Bool("preview")

		if !noop {
			stderr := log.New(os.Stderr, "", 0) // log to stderr with no timestamps
			if apiToken == "" {
				stderr.Fatal("API token is required.",
					" Use --api-token or set SLACK_TOKEN env variable.")
			}
			if channel == "" {
				stderr.Fatal("Destination is required.",
					" Use --channel or set SLACK_CHANNEL env variable.")
			}
			if delay < 0.001 {
				stderr.Fatal("You must have a delay >=0.001 to avoid creating a time paradox.")
			}
		}

		ctx := context.Background()
		var frames <-chan string
		if c.Bool("loop") {
			frames = NewLoopingLineScanner(ctx, os.Stdin, 4096).Frames()
		} else {
			frames = NewLineScanner(ctx, os.Stdin).Frames()
		}

		// TODO: restore noop case
		/*
			for frame := range frames {
				<-tickerChan
				if noop {
					fmt.Printf("\033[2K\r%s", frame)
				}
		*/

		err := Updater(context.Background(), apiToken, channel, frames, UpdaterOptions{
			MinDelay: time.Millisecond * time.Duration(delay*1000),
			UpdateFunc: func(u Update) {
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
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Done!")
	}

	app.Run(os.Args)
}
