// Package slacknimate provides facilities for posting continuous realtime
// status updates to a single Slack message.
package slacknimate

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// UpdaterOptions contains optional configuration for the Update function.
type UpdaterOptions struct {
	MinDelay   time.Duration // minimum delay between frames
	UpdateFunc func(Update)  // callback to perform upon each update result

	Username  string // override bot username
	IconEmoji string // override bot icon with Emoji
	IconURL   string // override bot icon with URL
}

func (opts UpdaterOptions) slackMsgOptions() slack.MsgOption {
	var msgOpts []slack.MsgOption
	if opts.Username != "" {
		msgOpts = append(msgOpts, slack.MsgOptionUsername(opts.Username))
	}
	if opts.IconEmoji != "" {
		msgOpts = append(msgOpts, slack.MsgOptionIconEmoji(opts.IconEmoji))
	}
	if opts.IconURL != "" {
		msgOpts = append(msgOpts, slack.MsgOptionIconURL(opts.IconURL))
	}
	return slack.MsgOptionCompose(msgOpts...)
}

// Updater posts and updates the "animated" message via the Slack API. It
// consumes the required frames chan, posting the initial frame as a Slack
// message to the provided destination Slack channel, and using each subsequent
// frame to update the text of the posted message.
//
// The Slack channel can be an encoded ID, or a name.
//
// The Slack api client should be configured using an authentication token that
// is bearing appropriate OAuth scopes for its destination and options.
//
// Results
//
// This function blocks until the provided frame chan is closed, or it
// encounters a fatal condition. This fatal condition will be returned as a
// non-nil error, an example would be not being able to make the initial post to
// Slack. Subsequent message update errors may be transient and thus are not
// considered fatal errors, and can be monitored or handled via the
// UpdaterOptions.UpdateFunc callback.
//
// Monitoring Realtime Updates
//
// If you wish to monitor or act upon individual updates to the Updater
// completing, you can set an UpdateFunc callback in the opts. For example, to
// simply log intermediate errors:
//
//     opts.UpdateFunc = func(u Update) {
//         if u.Err != nil {
//             log.Println(err)
//         }
//     }
//
// Or to get the updates sent back to you on a buffered channel:
//
//     updateChan := make(chan Update, 50)
//     opts.UpdateFunc = func(u Update) {
//         updateChan <- res
//     }
//
// This allows the consumer the most flexibility in how to consume these
// updates.
func Updater(ctx context.Context,
	api *slack.Client,
	channelID string,
	frames <-chan string,
	opts UpdaterOptions) error {

	var delayTicker *time.Ticker
	if opts.MinDelay > 0 {
		delayTicker = time.NewTicker(opts.MinDelay)
		defer delayTicker.Stop()
	}

	msgOpts := opts.slackMsgOptions()

	var dst, ts string
	for frame := range frames {
		// if context is already cancelled, exit immediately
		if err := ctx.Err(); err != nil {
			return err
		}

		// If we have a minDelay ticker, ensure at least that much time has
		// passed before proceeding. Also continue to check for context
		// completion just in case, so we can handle that situation immediately
		// if it occurs while we're waiting for the minDelay.
		if delayTicker != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-delayTicker.C:
			}
		}

		// If no messages have been posted, post the initial message; otherwise,
		// update using the previous channel/timestamp pairing as identifier.
		msgText := slack.MsgOptionText(frame, true)
		var err error
		if dst == "" || ts == "" {
			dst, ts, err = api.PostMessageContext(ctx, channelID, msgText, msgOpts)
			if err != nil {
				return fmt.Errorf("FATAL: Could not post initial frame: %w", err)
			}
		} else {
			_, _, _, err = api.UpdateMessageContext(ctx, dst, ts, msgText, msgOpts)
		}
		if opts.UpdateFunc != nil {
			opts.UpdateFunc(Update{dst, ts, frame, err})
		}
	}

	return nil
}

// Update represents the status returned from the Slack API for a specific
// frame message post or update.
type Update struct {
	Dst   string // target message destination channel ID
	TS    string // target message timestamp in Slack API format
	Frame string // text sent as message payload
	Err   error
}
