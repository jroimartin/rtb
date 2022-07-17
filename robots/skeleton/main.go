package main

import (
	"math"

	"github.com/jroimartin/rtb"
)

func main() {
	settings := rtb.ListenSettings{
		SendRotationReached: 2,
		ChanBufferCapacity:  100,
	}
	msgs := rtb.Listen(settings)
loop:
	for msg := range msgs {
		switch m := msg.(type) {
		case rtb.MessageInitialize:
			if !m.First {
				continue
			}
			rtb.Name("Skeleton")
			rtb.Colour("00ff00", "ff0000")
		case rtb.MessageGameStarts:
			rtb.Sweep(rtb.PartRadar, math.Pi/4, -math.Pi/2, math.Pi/2)
		case rtb.MessageRadar:
			rtb.Debugf("radar: distance=%v object=%v angle=%v", m.Distance, m.Object, m.RadarAngle)
		case rtb.MessageExitRobot:
			break loop
		default:
			rtb.Debugf("ignored message: %#v", msg)
		}
	}
	rtb.Debugf("done")
}
