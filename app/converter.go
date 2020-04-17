package app

import (
	"io"
	"os"
	"os/exec"
)

func convert(src, temp, dest string, logging bool) {
	c := exec.Command(
		"ffmpeg",
		"-i", src,
		"-c:v", "libx264",
		"-b:v", "4M", "-maxrate", "4M", "-bufsize", "2M",
		"-vf", "scale=1280:720:flags=lanczos",
		"-preset", "fast",
		temp,
	)

	if logging {
		err, e := c.StderrPipe()
		if e == nil {
			go func() {
				_, _ = io.Copy(os.Stdout, err)
			}()
		}
	}

	if logErr("Convert", c.Run()) {
		logErr("Remove temp", os.Remove(temp))
	} else {
		logErr("Rename temp -> dest", os.Rename(temp, dest))
	}

	logErr("Remove src", os.Remove(src))
}
