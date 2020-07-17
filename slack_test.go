package slacknimate

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

func TestUpdater(t *testing.T) {
	// slacktest module is pretty bare bones, and doesn't support chat.update
	// API post which is the core of our functionality.  So patch in a very
	// rudimentary handler to just register that we got the updates.
	testServer := slacktest.NewTestServer()
	var serverChatUpdate int
	testServer.Handle("/chat.update", func(w http.ResponseWriter, r *http.Request) {
		serverChatUpdate++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	})

	// start up the test server and configure an API client
	testServer.Start()
	defer testServer.Stop()
	client := slack.New("ABCD123",
		slack.OptionAPIURL(testServer.GetAPIURL()),
		slack.OptionDebug(false),
	)

	// goroutine to generate a few test frames and then close
	const numTestFrames = 10
	frames := testFrameGenerator(context.Background(), numTestFrames)

	// ctx, cf := context.WithCancel(context.Background())
	ctx := context.Background()
	var callbacksSeen int
	err := Updater(ctx, client, "#testing", frames, UpdaterOptions{
		UpdateFunc: func(u Update) {
			callbacksSeen++
			if err := u.Err; err != nil {
				t.Errorf("%#v", err)
			}
		},
	})
	if err != nil {
		t.Fatal("Updater fatal err:", err)
	}

	if callbacksSeen != numTestFrames {
		t.Errorf(
			"client callbacks seen want %v got %v",
			numTestFrames, callbacksSeen)
	}
	if serverChatUpdate != numTestFrames-1 {
		t.Errorf(
			"server chat.update posts received want %v got %v",
			numTestFrames-1, serverChatUpdate)
	}
}

// testFrameGenerator creates a background goroutine which will send n mock
// frame updates over the returned channel
func testFrameGenerator(ctx context.Context, n uint) <-chan string {
	frames := make(chan string)
	go func() {
		defer close(frames)
		for i := uint(0); i < n; i++ {
			select {
			case frames <- fmt.Sprintf("frame%v", i):
			case <-ctx.Done():
				return
			}
		}
	}()
	return frames
}
