package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/oto/v2"
)

var (
	sampleRate      = flag.Int("samplerate", 44100, "sample rate")
	channelNum      = flag.Int("channelnum", 2, "number of channel")
	bitDepthInBytes = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
)

type SineWave struct {
	freq   float64
	length int64
	pos    int64

	remaining []byte
}

func NewSineWave(freq float64, duration time.Duration) *SineWave {
	l := int64(*channelNum) * int64(*bitDepthInBytes) * int64(*sampleRate) * int64(duration) / int64(time.Second)
	l = l / 4 * 4
	return &SineWave{
		freq:   freq,
		length: l,
	}
}

func (s *SineWave) Read(buf []byte) (int, error) {
	if len(s.remaining) > 0 {
		n := copy(buf, s.remaining)
		copy(s.remaining, s.remaining[n:])
		s.remaining = s.remaining[:len(s.remaining)-n]
		return n, nil
	}

	if s.pos == s.length {
		return 0, io.EOF
	}

	eof := false
	if s.pos+int64(len(buf)) > s.length {
		buf = buf[:s.length-s.pos]
		eof = true
	}

	var origBuf []byte
	if len(buf)%4 > 0 {
		origBuf = buf
		buf = make([]byte, len(origBuf)+4-len(origBuf)%4)
	}

	length := float64(*sampleRate) / float64(s.freq)

	num := (*bitDepthInBytes) * (*channelNum)
	p := s.pos / int64(num)
	switch *bitDepthInBytes {
	case 1:
		for i := 0; i < len(buf)/num; i++ {
			const max = 127
			b := int(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < *channelNum; ch++ {
				buf[num*i+ch] = byte(b + 128)
			}
			p++
		}
	case 2:
		for i := 0; i < len(buf)/num; i++ {
			const max = 32767
			b := int16(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < *channelNum; ch++ {
				buf[num*i+2*ch] = byte(b)
				buf[num*i+1+2*ch] = byte(b >> 8)
			}
			p++
		}
	}

	s.pos += int64(len(buf))

	n := len(buf)
	if origBuf != nil {
		n = copy(origBuf, buf)
		s.remaining = buf[n:]
	}

	if eof {
		return n, io.EOF
	}
	return n, nil
}

func play(context *oto.Context, freq float64, duration time.Duration) oto.Player {
	p := context.NewPlayer(NewSineWave(freq, duration))
	p.Play()
	return p
}

func uwu() error {

	c, ready, err := oto.NewContext(*sampleRate, *channelNum, *bitDepthInBytes)
	if err != nil {
		return err
	}
	<-ready

	var wg sync.WaitGroup
	var players []oto.Player

	wg.Add(1)
	go func() {
		freq := map[string]float64{
			"D": 587,
			"A": 880,
			"P": 830,
			"G": 784,
			"F": 698,
			"Ñ": 932,
			"C": 523,
			"B": 987,
		}

		duration := map[string]time.Duration{
			"D": time.Millisecond * 150,
			"A": time.Millisecond * 150,
			"P": time.Millisecond * 150,
			"G": time.Millisecond * 150,
			"F": time.Millisecond * 150,
			"Ñ": time.Millisecond * 150,
			"C": time.Millisecond * 150,
			"B": time.Millisecond * 150,
		}

		/*
					NOTES:
					D = 587
					A = 880
					P = 830
					G = 784
					F = 698
					Ñ = 932

			        D D D A P G F D F G C C
					D A P G F D F G B B
					D A P G F D F G Ñ Ñ D A P G F D
					F G F F F F D D D F F F G P G F D F G F F F G P
					A C A D D D A D C A A A A G G G A A A A G A C A G D
					A G F C G F E D D D D F C F D F G P G F

					// https://www.reddit.com/r/UndertaleMusic/comments/6eeah8/what_are_the_letter_notes_for_megalovania/

		*/

		notes := map[string]string{
			"D": "DDDAPGFDFGCC",
			"A": "DAPGFDFGBB",
			"P": "DAPGFDFGÑÑDAPGFD",
			"G": "FGFFFFDDDFFFGPGFDFGFFGP",
			"F": "ACADDDADCAAAAGGGAAAAGACAGD",
			"Ñ": "AGFCGFEDDDDFCFDFGPGF",
		}
		for {

			// iterate over notes and play them
			for _, note := range notes {
				for _, n := range note {
					p := play(c, freq[string(n)], duration[string(n)])
					players = append(players, p)
					time.Sleep(200 * time.Millisecond)
					fmt.Println([...]string{string(n), "played"})
				}
			}

		}
	}()

	wg.Wait()
	runtime.KeepAlive(players)

	return nil

}

func main() {
	if err := uwu(); err != nil {
		panic(err)
	}
}
