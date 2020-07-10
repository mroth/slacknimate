# slacknimate
> text animation for Slack messages :dancers:

Useful for ChatOps:
![deployinator](https://cloud.githubusercontent.com/assets/40650/26273321/0cd49fda-3cfc-11e7-90ce-78f369e783ac.gif)



...Or for comedy:

![slacknimate_fine2](https://cloud.githubusercontent.com/assets/40650/26273332/613cc17e-3cfc-11e7-9365-88b0043c17ef.gif)

...Or maybe art:

![slacknimate_nyan](https://cloud.githubusercontent.com/assets/40650/26273350/ad3b0d56-3cfc-11e7-9359-83c92f440a03.gif)


## Installation
Download a binary from the [Releases Page](https://github.com/mroth/slacknimate/releases) and put it somewhere on your `$PATH`.

_macOS Homebrew users, you can also just `brew install slacknimate`._

## Authentication
Generate your Slack app and generate an API token. The app will need appropriate
OAuth scopes to post messages to your desired destination.

You'll need to either pass it to the program via the `--api-token` flag or store
it as `SLACK_TOKEN` environment variable.

## Usage

```
NAME:
   slacknimate - text animation for Slack messages

USAGE:
   slacknimate [options]

VERSION:
   1.0.0

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --api-token, -a      API token* [$SLACK_TOKEN]
   --delay, -d "1"      minimum delay between frames
   --channel, -c        channel/destination* [$SLACK_CHANNEL]
   --loop, -l           loop content upon reaching end
   --preview            preview on terminal instead of posting
   --help, -h           show help
   --version, -v        print the version
```

### Simple animation loops

    $ slacknimate -c "#general" --loop < examples/emoji.txt

![slacknimate1](https://cloud.githubusercontent.com/assets/40650/13275355/32f5997c-da82-11e5-8a9d-61c53f94c718.gif)

### Realtime process monitoring
Why spam a chatroom with periodic monitoring messages when you can have realtime
status updates so that a message is never out of date?

See for example this example:

```
$ ./examples/process.sh 5 | slacknimate -c "#devops"
2016/02/23 19:03:14 initial frame G07AJU0SH/1456272194.000086: Processing items: 0/5
2016/02/23 19:03:15 updated frame G07AJU0SH/1456272194.000086: Processing items: 1/5
2016/02/23 19:03:16 updated frame G07AJU0SH/1456272194.000086: Processing items: 2/5
2016/02/23 19:03:17 updated frame G07AJU0SH/1456272194.000086: Processing items: 3/5
2016/02/23 19:03:18 updated frame G07AJU0SH/1456272194.000086: Processing items: 4/5
2016/02/23 19:03:19 updated frame G07AJU0SH/1456272194.000086: Processing items: 5/5

Done!
```

![slacknimate2](https://cloud.githubusercontent.com/assets/40650/13275356/36c8f15c-da82-11e5-93c1-ef8e6d3e556e.gif)


### Preview in terminal
If you aren't certain about your source, you can preview what the animation
would look like in the terminal via the `--preview` flag.

    $ slacknimate --preview --loop -d 0.25 < examples/sample.txt

![slacknimate3](https://cloud.githubusercontent.com/assets/40650/13275357/3b04b6ac-da82-11e5-9fab-1a7704c98b12.gif)
