package rtb

import "testing"

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
				Distance: 1.2,
				Object: ObjectCookie,
				RadarAngle: 4.5,
			},
			true,
		},

		// Info
		{
			"Info",
			"Info 1.2 3.4 5.6",
			MessageInfo{
				Time: 1.2,
				Speed: 3.4,
				CannonAngle: 5.6,
			},
			true,
		},

		// Coordinates
		{
			"Coordinates",
			"Coordinates 1.2 3.4 5.6",
			MessageCoordinates{
				X: 1.2,
				Y: 3.4,
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
				TeamMate: false,
			},
			true,
		},
		{
			"RobotInfo teammate",
			"RobotInfo 1.2 1",
			MessageRobotInfo{
				EnergyLevel: 1.2,
				TeamMate: true,
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
				PartRobot|PartCannon,
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
				Angle: 3.4,
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
