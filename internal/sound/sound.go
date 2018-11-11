package sound

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/gobuffalo/packr"
)

var box = packr.NewBox("./data")

// Beep makes a beep sound.
func Beep() error {
	b, err := box.Find("beep.wav")
	if err != nil {
		return err
	}

	s, format, err := wav.Decode(ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan struct{})
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))
	<-done

	return nil
}
