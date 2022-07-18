package rtb

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		msg    any
		nilErr bool
	}{
		// Initialize
		{
			"Initialize first",
			"Initialize 1",
			MessageInitialize{
				First: true,
			},
			true,
		},
		{
			"Initialize not first",
			"Initialize 2",
			MessageInitialize{
				First: false,
			},
			true,
		},

		// YourName
		{
			"YourName",
			"YourName foo",
			MessageYourName{
				Name: "foo",
			},
			true,
		},
		{
			"YourName spaces",
			"YourName foo bar",
			MessageYourName{
				Name: "foo bar",
			},
			true,
		},

		// YourColour
		{
			"YourColour",
			"YourColour 11aa22",
			MessageYourColour{
				Colour: "11aa22",
			},
			true,
		},

		// GameOption
		{
			"GameOption",
			"GameOption 8 1.234",
			MessageGameOption{
				Option: GOptionShotSpeed,
				Value:  1.234,
			},
			true,
		},

		// GameStarts
		{
			"GameStarts",
			"GameStarts",
			MessageGameStarts{},
			true,
		},

		// Radar
		{
			"Radar",
			"Radar 1.2 3 4.5",
			MessageRadar{
				Distance:   1.2,
				Object:     ObjectCookie,
				RadarAngle: 4.5,
			},
			true,
		},

		// Info
		{
			"Info",
			"Info 1.2 3.4 5.6",
			MessageInfo{
				Time:        1.2,
				Speed:       3.4,
				CannonAngle: 5.6,
			},
			true,
		},

		// Coordinates
		{
			"Coordinates",
			"Coordinates 1.2 3.4 5.6",
			MessageCoordinates{
				X:     1.2,
				Y:     3.4,
				Angle: 5.6,
			},
			true,
		},

		// RobotInfo
		{
			"RobotInfo enemy",
			"RobotInfo 1.2 0",
			MessageRobotInfo{
				EnergyLevel: 1.2,
				TeamMate:    false,
			},
			true,
		},
		{
			"RobotInfo teammate",
			"RobotInfo 1.2 1",
			MessageRobotInfo{
				EnergyLevel: 1.2,
				TeamMate:    true,
			},
			true,
		},
		{
			"RobotInfo unknown",
			"RobotInfo 1.2 -1",
			nil,
			false,
		},

		// RotationReached
		{
			"RotationReached",
			"RotationReached 3",
			MessageRotationReached{
				PartRobot | PartCannon,
			},
			true,
		},

		// Energy
		{
			"Energy",
			"Energy	1.2",
			MessageEnergy{
				EnergyLevel: 1.2,
			},
			true,
		},

		// RobotsLeft
		{
			"RobotsLeft",
			"RobotsLeft 123",
			MessageRobotsLeft{
				NumRobots: 123,
			},
			true,
		},

		// Collision
		{
			"Collision",
			"Collision 2 3.4",
			MessageCollision{
				Object: ObjectWall,
				Angle:  3.4,
			},
			true,
		},

		// Warning
		{
			"Warning",
			"Warning 2 foo",
			MessageWarning{
				Warning: WarningMessageSentInIllegalState,
				Message: "foo",
			},
			true,
		},
		{
			"Warning spaces",
			"Warning 2 foo bar",
			MessageWarning{
				Warning: WarningMessageSentInIllegalState,
				Message: "foo bar",
			},
			true,
		},

		// Dead
		{
			"Dead",
			"Dead",
			MessageDead{},
			true,
		},

		// GameFinishes
		{
			"GameFinishes",
			"GameFinishes",
			MessageGameFinishes{},
			true,
		},

		// ExitRobot
		{
			"ExitRobot",
			"ExitRobot",
			MessageExitRobot{},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseMessage(tt.line)
			if (err == nil) != tt.nilErr {
				t.Errorf("unexpected error: got=%v", err)
			}
			if msg != tt.msg {
				t.Errorf("wrong message: got=%#v want=%#v", msg, tt.msg)
			}
		})
	}
}

func TestListen(t *testing.T) {
	osStdin = bytes.NewBufferString(`
		GameStarts
		YourName foo bar
		RobotInfo 1.23 0
		Warning 4 foo bar
	`)
	osStdout = io.Discard
	defer func() {
		osStdin = os.Stdin
		osStdout = os.Stdout
	}()

	want := []any{
		MessageGameStarts{},
		MessageYourName{
			Name: "foo bar",
		},
		MessageRobotInfo{
			EnergyLevel: 1.23,
			TeamMate:    false,
		},
		MessageWarning{
			Warning: WarningObsoleteKeyword,
			Message: "foo bar",
		},
	}

	var got []any
	for msg := range Listen(ListenSettings{}) {
		got = append(got, msg)
	}

	if len(got) != len(want) {
		t.Fatalf("invalid number of messages: got=%v want=%v", len(got), len(want))
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("unexpected message: got=%#v want=%#v", got, want)
		}
	}
}

func TestRawf(t *testing.T) {
	var buf bytes.Buffer
	osStdout = &buf
	defer func() { osStdout = os.Stdout }()

	tests := []struct {
		name   string
		s      string
		want   string
		nilErr bool
	}{
		{
			"Short string",
			"foo",
			"foo\n",
			true,
		},
		{
			"Valid edge case",
			strings.Repeat("x", 127),
			strings.Repeat("x", 127) + "\n",
			true,
		},
		{
			"Invalid edge case",
			strings.Repeat("x", 128),
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rawf(tt.s)
			if (err == nil) != tt.nilErr {
				t.Errorf("unexpected error: got=%v", err)
			}
			got, err := io.ReadAll(&buf)
			if err != nil {
				t.Fatalf("error reading bytes.Buffer")
			}
			if string(got) != tt.want {
				t.Errorf("unexpected output: got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestRobotMessages(t *testing.T) {
	var buf bytes.Buffer
	osStdout = &buf
	defer func() { osStdout = os.Stdout }()

	tests := []struct {
		name string
		f    func()
		want string
	}{
		{
			"RobotOption",
			func() { robotOption(rOptionUseNonBlocking, 0) },
			"RobotOption 3 0\n",
		},
		{
			"Name",
			func() { Name("foo") },
			"Name foo\n",
		},
		{
			"Colour",
			func() { Colour("11aa22", "bb33cc") },
			"Colour 11aa22 bb33cc\n",
		},
		{
			"Rotate",
			func() { Rotate(PartCannon|PartRadar, 1.23) },
			"Rotate 6 1.230000\n",
		},
		{
			"RotateTo",
			func() { RotateTo(PartCannon|PartRadar, 1.23, 4.56) },
			"RotateTo 6 1.230000 4.560000\n",
		},
		{
			"RotateAmount",
			func() { RotateAmount(PartCannon|PartRadar, 1.23, 4.56) },
			"RotateAmount 6 1.230000 4.560000\n",
		},
		{
			"Sweep",
			func() { Sweep(PartCannon|PartRadar, 1.23, 4.56, 7.89) },
			"Sweep 6 1.230000 4.560000 7.890000\n",
		},
		{
			"Accelerate",
			func() { Accelerate(1.23) },
			"Accelerate 1.230000\n",
		},
		{
			"Brake",
			func() { Brake(1.23) },
			"Brake 1.230000\n",
		},
		{
			"Shoot",
			func() { Shoot(1.23) },
			"Shoot 1.230000\n",
		},
		{
			"Print",
			func() { Printf("foo bar %v", PartRobot|PartRadar) },
			"Print foo bar Robot|Radar\n",
		},
		{
			"Debug",
			func() { Debugf("foo bar %v", PartCannon|PartRadar) },
			"Debug foo bar Cannon|Radar\n",
		},
		{
			"DebugLine",
			func() { DebugLine(1.23, 4.56, 7.89, 10.11) },
			"DebugLine 1.230000 4.560000 7.890000 10.110000\n",
		},
		{
			"DebugCircle",
			func() { DebugCircle(1.23, 4.56, 7.89) },
			"DebugCircle 1.230000 4.560000 7.890000\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			got, err := io.ReadAll(&buf)
			if err != nil {
				t.Fatalf("error reading bytes.Buffer")
			}
			if string(got) != tt.want {
				t.Errorf("unexpected output: got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestPartString(t *testing.T) {
	tests := []struct {
		p    Part
		want string
	}{
		{PartRobot, "Robot"},
		{PartCannon, "Cannon"},
		{PartRadar, "Radar"},
		{PartRobot | PartCannon, "Robot|Cannon"},
		{PartRobot | PartRadar, "Robot|Radar"},
		{PartCannon | PartRadar, "Cannon|Radar"},
		{PartRobot | PartCannon | PartRadar, "Robot|Cannon|Radar"},
		{Part(15), "Robot|Cannon|Radar"},
		{Part(16), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("unexpected string: got=%q want=%q", got, tt.want)
		}
	}
}
