package omxplayer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

// binary name
const bin string = "omxplayer"

// commands
const (
	stop      string = "q"
	pause     string = "p"
	subs      string = "s"
	fwd       string = "\x1B[C"
	back      string = "\x1B[D"
	fastFwd   string = "\x1B[A"
	fastBack  string = "\x1B[B"
	volUp     string = "+"
	volDown   string = "-"
	next      string = "o"
	prev      string = "i"
	info      string = "z"
	nextAudio string = "k"
	nextSub   string = "m"
)

// audio outputs
const (
	AnalogAudio  = iota
	DigitalAudio = iota
)

// file descriptors connected to the running application
var (
	stdin  io.Writer
	stdout io.Reader
)

var cmd *exec.Cmd                   // command running player application
var running bool = false            // keep track of running status
var audio_output int = DigitalAudio // audio output sink, default to HDMI
var paused bool = false             // paused state of the video stream

// Start the external player application
func start(uri string, audioOut int) error {
	var err error
	options := make([]string, 0)
	switch audioOut {
	case DigitalAudio:
		options = append(options, "-o", "hdmi")
	case AnalogAudio:
		options = append(options, "-o", "local")
	}
	cmd = exec.Command(bin, append(options, uri)...)

	// Get std pipes (in/out)
	if stdin, err = cmd.StdinPipe(); err != nil {
		return err
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return err
	}

	// start
	if err = cmd.Start(); err != nil {
		return err
	}
	running = true
	return nil
}

// send command to shell and omxplayer
func sendCmd(cmd string) error {
	if !running {
		return nil
	}
	_, e := fmt.Fprint(stdin, cmd)
	// if there was an error 'stop' the player
	if e != nil {
		Stop()
	}
	return e
}

// Start playback
func Play(uri string) error {
	// if the player is already running, stop it and start the new file
	if running {
		Stop()
	}
	if e := start(uri, audio_output); e != nil {
		return e
	}
	return nil
}

// Stop the player (exits)
func Stop() error {
	// in case the player is not running, do nothing
	if !running {
		return nil
	}
	e := sendCmd(stop)
	wait()
	return e
}

// Pause playback
func Pause() error {
	paused = true
	return sendCmd(pause)
}

// Resume playback
func Resume() error {
	paused = false
	return sendCmd(pause)
}

// Seek forward
func Fwd() error {
	return sendCmd(fwd)
}

// Seek backward
func Bwd() error {
	return sendCmd(back)
}

// Go to next chapter
func Next() error {
	return sendCmd(next)
}

// Go back to previous chapter
func Prev() error {
	return sendCmd(prev)
}

// Next audio stream
func NextAudio() error {
	return sendCmd(nextAudio)
}

// Next subtitle stream
func NextSub() error {
	return sendCmd(nextSub)
}

// Show info
func Info() error {
	return sendCmd(info)
}

// Toggle subtitles
func Subs() error {
	return sendCmd(subs)
}

// Set audio output sink
func SetAudioOutput(mode int) {
	audio_output = mode
}

// Return paused status.
func Paused() bool {
	return paused
}

// Function to be used to read output from the player application and interpret
func parseOutput() {
	r := bufio.NewReader(stdout)
	for {
		if line, error := r.ReadString('\n'); error != nil {
			break
		} else {
			fmt.Printf(" [OMXPLAYER] (Application output) \n%s\n", line)
		}
	}
}

// Cleanup after stopping the player
func wait() {
	cmd.Wait()
	running = false
}
