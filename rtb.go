// Package rtb provides support for writing RealTimeBattle robots.
package rtb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// rawf sends a raw message. It returns error if the message is longer than 128
// characters.
func rawf(format string, a ...any) error {
	s := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	if len(s) > 128 {
		return fmt.Errorf("message is too long (%v)", len(s))
	}

	fmt.Print(s)

	return nil
}

// rOption represents a robot option.
type rOption int

const (
	// rOptionSendSignal tells the server to send SIGUSR1 when there is a
	// message waiting. Send this message (with argument 1 (= true)) as
	// soon as you are prepared to receive the signal. Default is false.
	rOptionSendSignal rOption = 0

	// rOptionSendRotationReached tells the server to send a
	// RotationReached message when a rotation is finished. With a value of
	// 1, the message is sent when a RotateTo or a RotateAmount is
	// finished, with a value of 2, changes in sweep direction are also
	// notified. Default is 0, i.e. no messages are sent.
	rOptionSendRotationReached = 1

	// rOptionSignal tells the server to send a signal when there is a
	// message waiting. The argument will determine which signal. Send this
	// message as soon as you are prepared to receive the signal. Default
	// is 0, which means don't send any signals.
	rOptionSignal = 2

	// rOptionUseNonBlocking selects how to reading messages works. This
	// option should be sent exactly once as soon as the program starts.
	// Since it should always be given, there is no default value.
	rOptionUseNonBlocking = 3
)

// robotOption sets option to value.
func robotOption(option rOption, value int) error {
	return rawf("RobotOption %d %d", option, value)
}

// Name sets the name of the robot. When receiving the Initialize message with
// argument 1, indicating that this is the first sequence, you should send your
// name. If your name ends with the string "Team: teamname", you will be in the
// team "teamname". For example "foo Team: bar" will assign you to the team
// "bar" and your name will be "foo".
func Name(name string) error {
	return rawf("Name %s", name)
}

// Colour sets your colour. The colours are like normal football shirts, the
// home colour is used unless it is already used. Otherwise the away colour or,
// as a last resort, a non-occupied colour is selected randomly. Colours are
// specified using a hex string of the form "11aa22".
func Colour(homeColour, awayColour string) error {
	return rawf("Colour %s %s", homeColour, awayColour)
}

// Part represents a part of the robot. Part values can be or'ed to specify
// multiple parts at the same time.
type Part int

const (
	// PartRobot is the robot.
	PartRobot Part = 1

	// PartCannon is the cannon of the robot.
	PartCannon = 2

	// PartRadar is the radar of the robot.
	PartRadar = 4
)

func (p Part) String() string {
	var parts []string

	if p&PartRobot != 0 {
		parts = append(parts, "Robot")
	}
	if p&PartCannon != 0 {
		parts = append(parts, "Cannon")
	}
	if p&PartRadar != 0 {
		parts = append(parts, "Radar")
	}

	if len(parts) == 0 {
		return "unknown"
	}

	return strings.Join(parts, "|")
}

// Rotate sets the angular velocity for the robot, its cannon and/or its radar.
// The angular velocity is given in radians per second and is limited by Robot
// (cannon/radar) max rotate speed.
func Rotate(what Part, v float64) error {
	return rawf("Rotate %d %f", what, v)
}

// RotateTo is like Rotate, but will rotate to a given angle. Note that radar
// and cannon angles are relative to the robot angle. You cannot use this
// command to rotate the robot itself, use RotateAmount instead.
func RotateTo(what Part, v, end float64) error {
	return rawf("RotateTo %d %f %f", what, v, end)
}

// RotateAmount is like Rotate, but will rotate relative to the current angle.
func RotateAmount(what Part, v, angle float64) error {
	return rawf("RotateAmount %d %f %f", what, v, angle)
}

// Sweep is like Rotate, but sets the radar and/or the cannon (not available
// for the robot itself) in a sweep mode.
func Sweep(what Part, v, rightAngle, leftAngle float64) error {
	return rawf("Sweep %d %f %f %f", what, v, rightAngle, leftAngle)
}

// Accelerate sets the robot acceleration. Value is bounded by Robot max/min
// acceleration.
func Accelerate(value float64) error {
	return rawf("Accelerate %f", value)
}

// Brake sets the brake. Full brake (portion = 1.0) means that the friction in
// the robot direction is equal to Slide friction.
func Brake(portion float64) error {
	return rawf("Break %f", portion)
}

// Shoot with the given energy.
func Shoot(energy float64) error {
	return rawf("Shoot %f", energy)
}

// Printf prints a message on the message window.
func Printf(format string, a ...any) error {
	return rawf("Print "+format, a...)
}

// Debugf prints a message on the message window if in debug-mode.
func Debugf(format string, a ...any) error {
	return rawf("Debug "+format, a...)
}

// DebugLine draws a line direct to the arena. This is only allowed in the
// highest debug level(5), otherwise a warning message is sent. The arguments
// are the start and end point of the line given in polar coordinates relative
// to the robot.
func DebugLine(angle1, radius1, angle2, radius2 float64) error {
	return rawf("DebugLine %f %f %f %f", angle1, radius1, angle2, radius2)
}

// DebugCircle is similar to DebugLine, but draws a circle. The first two
// arguments are the angle and radius of the central point of the circle
// relative to the robot. The third argument gives the radius of the circle.
func DebugCircle(centerAngle, centerRadius, circleRadius float64) error {
	return rawf("DebugCircle %f %f %f", centerAngle, centerRadius, circleRadius)
}

// GOption represents a game option.
type GOption int

const (
	// GOptionRobotMaxRotate is how fast the robot itself may rotate in
	// radians/s .
	GOptionRobotMaxRotate GOption = 0

	// GOptionRobotCannonMaxRotate is the maximum cannon rotate speed. Note
	// that the cannon moves relative to the robot, so the actual rotation
	// speed may be higher.
	GOptionRobotCannonMaxRotate = 1

	// GOptionRobotRadarMaxRotate is the maximum radar rotate speed. Note
	// that the radar moves relative to the robot, so the actual rotation
	// speed may be higher.
	GOptionRobotRadarMaxRotate = 2

	// GOptionRobotMaxAcceleration indicates that robots are not allowed to
	// accelerate faster than this.
	GOptionRobotMaxAcceleration = 3

	// GOptionRobotMinAcceleration indicates that robots are not allowed to
	// accelerate slower than this.
	GOptionRobotMinAcceleration = 4

	// GOptionRobotStartEnergy is the amount of energy the robots will have
	// at the beginning of each game.
	GOptionRobotStartEnergy = 5

	// GOptionRobotMaxEnergy is the maximum amount of energy a robot can
	// get. By eating a cookie, the robot can increase its energy; not
	// more than this, though.
	GOptionRobotMaxEnergy = 6

	// GOptionRobotEnergyLevels decides how many discretation levels will
	// be used.
	GOptionRobotEnergyLevels = 7

	// GOptionShotSpeed is speed of the shot in the direction of the
	// cannon. Shots move at this speed plus the velocity of the robot.
	GOptionShotSpeed = 8

	// GOptionShotMinEnergy is the lowest shot energy allowed. A robot
	// trying to shoot with less energy will fail to shoot.
	GOptionShotMinEnergy = 9

	// GOptionShotMaxEnergy is the maximum shot energy.
	GOptionShotMaxEnergy = 10

	// GOptionShotEnergyIncreaseSpeed determines how fast the robots shot
	// energy will increase in energy/s .
	GOptionShotEnergyIncreaseSpeed = 11

	// GOptionTimeout is the longest time a game will take. When the time
	// is up all remaining robots are killed, without getting any more
	// points.
	GOptionTimeout = 12

	// GOptionDebugLevel is the debug level. From 0 (no debug) to 5
	// (highest debug level).
	GOptionDebugLevel = 13

	// GOptionSendRobotCoordinates determines how coordinates are send to
	// the robots. The following options are available:
	//
	// - 0: No coordinates.
	// - 1: Coordinates are given relative the starting position.
	// - 2: Absolute coordinates.
	GOptionSendRobotCoordinates = 14
)

func (opt GOption) String() string {
	switch opt {
	case GOptionRobotMaxRotate:
		return "RobotMaxRotate"
	case GOptionRobotCannonMaxRotate:
		return "RobotCannonMaxRotate"
	case GOptionRobotRadarMaxRotate:
		return "RobotRadarMaxRotate"
	case GOptionRobotMaxAcceleration:
		return "RobotMaxAcceleration"
	case GOptionRobotMinAcceleration:
		return "RobotMinAcceleration"
	case GOptionRobotStartEnergy:
		return "RobotStartEnergy"
	case GOptionRobotMaxEnergy:
		return "RobotMaxEnergy"
	case GOptionRobotEnergyLevels:
		return "RobotEnergyLevels"
	case GOptionShotSpeed:
		return "ShotSpeed"
	case GOptionShotMinEnergy:
		return "ShotMinEnergy"
	case GOptionShotMaxEnergy:
		return "ShotMaxEnergy"
	case GOptionShotEnergyIncreaseSpeed:
		return "ShotEnergyIncreaseSpeed"
	case GOptionTimeout:
		return "Timeout"
	case GOptionDebugLevel:
		return "DebugLevel"
	case GOptionSendRobotCoordinates:
		return "SendRobotCoordinates"
	default:
		return "unknown"
	}
}

// Object represents an object type.
type Object int

const (
	// ObjectNoObject means that there isn't any object. This should never
	// happen.
	ObjectNoObject Object = -1

	// ObjectRobot means that the observed object is a robot.
	ObjectRobot = 0

	// ObjectShot means that the observed object is a shot.
	ObjectShot = 1

	// ObjectWall means that the observed object is a wall.
	ObjectWall = 2

	// ObjectCookie means that the observed object is a cookie.
	ObjectCookie = 3

	// ObjectMine means that the observed object is a mine.
	ObjectMine = 4
)

func (obj Object) String() string {
	switch obj {
	case ObjectNoObject:
		return "NoObject"
	case ObjectRobot:
		return "Robot"
	case ObjectShot:
		return "Shot"
	case ObjectWall:
		return "Wall"
	case ObjectCookie:
		return "Cookie"
	case ObjectMine:
		return "Mine"
	default:
		return "unknown"
	}
}

// Warning represents a warning sent by the server.
type Warning int

const (
	// WarningUnknownMessage means that the server received a message it
	// couldn't recognize.
	WarningUnknownMessage Warning = 0

	// WarningProcessTimeLow means that the CPU usage has reached the CPU
	// warning percentage. Only in competition-mode.
	WarningProcessTimeLow = 1

	// WarningMessageSentInIllegalState means that the message received
	// couldn't be handled in this state of the program. For example Rotate
	// is sent before the game has started.
	WarningMessageSentInIllegalState = 2

	// WarningUnknownOption means that the robot sent a robot option with
	// either illegal option name or illegal argument to that option.
	WarningUnknownOption = 3

	// WarningObsoleteKeyword means that the keyword sent is obsolete and
	// should not be used any more.
	WarningObsoleteKeyword = 4

	// WarningNameNotGiven means that the robot has not sent its name
	// before the game begins. This happens if the robot startup time is
	// too short or the robot does not send its name early enough.
	WarningNameNotGiven = 5

	// WarningColourNotGiven means that the robot has not sent its colour
	// before the game begins.
	WarningColourNotGiven = 6
)

func (warn Warning) String() string {
	switch warn {
	case WarningUnknownMessage:
		return "UnknownMessage"
	case WarningProcessTimeLow:
		return "ProcessTimeLow"
	case WarningMessageSentInIllegalState:
		return "MessageSentInIllegalState"
	case WarningUnknownOption:
		return "UnknownOption"
	case WarningObsoleteKeyword:
		return "ObsoleteKeyword"
	case WarningNameNotGiven:
		return "NameNotGiven"
	case WarningColourNotGiven:
		return "ColourNotGiven"
	default:
		return "unknown"
	}
}

type (
	// MessageInitialize is the very first message the robot will get.
	MessageInitialize struct {
		// First means it is the first sequence in the tournament and
		// the robot should send its name and colour to the server,
		// otherwise it should wait for MessageYourName and
		// MessageYourColour messages.
		First bool
	}

	// MessageYourName is the current name of the robot. Don't change it if
	// you don't have very good reasons.
	MessageYourName struct {
		// Current name of the robot.
		Name string
	}

	// MessageYourColour is the current colour of the robot, change it if
	// you find it ugly. All robots in a team will have the same colour.
	MessageYourColour struct {
		// Current colour of the robot.
		Colour string
	}

	// MessageGameOption [optionnr (int)] [value (float64)]. At the
	// beginning of each game the robots will be sent a number of settings,
	// which can be useful for the robot.
	MessageGameOption struct {
		// Game option.
		Option GOption

		// Value of the game option.
		Value float64
	}

	// MessageGameStarts is sent when the game starts.
	MessageGameStarts struct{}

	// MessageRadar gives information from the radar each turn.
	MessageRadar struct {
		// Distance to the observed object.
		Distance float64

		// Object is the type of the observed object.
		Object Object

		// Radar Angle relative to the robot front given in radians.
		RadarAngle float64
	}

	// MessageInfo does always follow the Radar message. It gives more
	// general information on the state of the robot.
	MessageInfo struct {
		// Time is the game-time elapsed since the start of the game.
		// This is not necessarily the same as the real time elapsed,
		// due to time scale and max timestep.
		Time float64

		// Current speed of the robot.
		Speed float64

		// Current angle of the cannon.
		CannonAngle float64
	}

	// MessageCoordinates tells you the current robot position. It is only
	// sent if the option GOptionSendRobotCoordinates is 1 or 2. If it is 1
	// the coordinates are sent relative the starting position, which has
	// the effect that the robot doesn't know where it is starting, but
	// only where it has moved since.
	MessageCoordinates struct {
		// Current position of the robot.
		X, Y float64

		// Current angle of the robot.
		Angle float64
	}

	// MessageRobotInfo. If you detect a robot with your radar, this
	// message will follow, giving some information on the robot.
	MessageRobotInfo struct {
		// EnergyLevel is the energy level of the observed robot
		// discretized into a number of energy levels.
		EnergyLevel float64

		// TeamMate is true when the observed robot is an team mate.
		TeamMate bool
	}

	// MessageRotationReached is sent when a rotation (with RotateTo or
	// RotateAmount) has finished or the direction has changed (when
	// sweeping). The option SendRotationReached has to be set
	// appropriately.
	MessageRotationReached struct {
		// Part identifies the rotated part.
		Part Part
	}

	// MessageEnergy is sent at the end of each round so the robot knows
	// its energy level.
	MessageEnergy struct {
		// EnergyLevel is the current energy level of the robot
		// discretized into a number of energy levels.
		EnergyLevel float64
	}

	// MessageRobotsLeft is sent at the beginning of the game and when a
	// robot is killed.
	MessageRobotsLeft struct {
		// NumRobots is the number of remaining robots.
		NumRobots int
	}

	// MessageCollision is sent when a robot hits (or is hit by) something.
	// It does not include how severe the collision was. This can, however,
	// be determined indirectly (approximately) by the loss of energy.
	MessageCollision struct {
		// Object is the type of the object hitting you.
		Object Object

		// Angle is the angle from where the collision occurred
		// relative the robot.
		Angle float64
	}

	// MessageWarning can be sent when the robot has to be notified on
	// different problems which have occured.
	MessageWarning struct {
		// Warning is the type of warning.
		Warning Warning

		// Message is the message related to the warning.
		Message string
	}

	// MessageDead is sent when the bobot died. Do not try to send more
	// messages to the server until the end of the game, the server doesn't
	// read them.
	MessageDead struct{}

	// MessageGameFinishes is sent when the current game is finished.
	MessageGameFinishes struct{}

	// MessageExitRobot means that you have to exit immediately. Otherwise
	// the robot program will be killed forcefully.
	MessageExitRobot struct{}
)

// ListenSettings defines the settings passed to Listen.
type ListenSettings struct {
	// SendRotationReached tells the server to send a RotationReached
	// message when a rotation is finished. With a value of 1, the message
	// is sent when a RotateTo or a RotateAmount is finished, with a value
	// of 2, changes in sweep direction are also notified. Default is 0,
	// i.e. no messages are sent.
	SendRotationReached int

	// ChanBufferCapacity is the buffer capacity of the channel returned by
	// Listen. If zero, an unbuffered channel is used.
	ChanBufferCapacity int
}

// Listen initializes the RTB communication channel and listens to RTB
// messages. It returns a channel on which the received messages are delivered.
func Listen(settings ListenSettings) <-chan any {
	// We dedicate a goroutine to read from stdin, so we use blocking mode.
	// Blocking mode is also simpler and more predictable.
	robotOption(rOptionUseNonBlocking, 0)

	robotOption(rOptionSendRotationReached, settings.SendRotationReached)

	stdin := stdinReader()
	msgs := make(chan any, settings.ChanBufferCapacity)
	go func() {
		defer close(msgs)

		for {
			line, ok := <-stdin
			if !ok {
				dbgf("stdin channel is closed")
				return
			}
			msg, err := parseMessage(line)
			if err != nil {
				dbgf("error parsing message %q: %v", line, err)
				continue
			}
			msgs <- msg
		}
	}()

	return msgs
}

// stdinReader reads lines from standard input. It returns a channel on which
// the lines are delivered.
func stdinReader() <-chan string {
	c := make(chan string)

	go func() {
		defer close(c)

		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			c <- s.Text()
		}
		if err := s.Err(); err != nil {
			dbgf("error reading from stdin: %v", err)
			return
		}
	}()

	return c
}

// parsers maps a message type to the corresponding parser.
var parsers = map[string]func([]string) (any, error){
	"Initialize":      parseInitialize,
	"YourName":        parseYourName,
	"YourColour":      parseYourColour,
	"GameOption":      parseGameOption,
	"GameStarts":      parseGameStarts,
	"Radar":           parseRadar,
	"Info":            parseInfo,
	"Coordinates":     parseCoordinates,
	"RobotInfo":       parseRobotInfo,
	"RotationReached": parseRotationReached,
	"Energy":          parseEnergy,
	"RobotsLeft":      parseRobotsLeft,
	"Collision":       parseCollision,
	"Warning":         parseWarning,
	"Dead":            parseDead,
	"GameFinishes":    parseGameFinishes,
	"ExitRobot":       parseExitRobot,
}

// parseMessage parses a message string.
func parseMessage(s string) (msg any, err error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return nil, errors.New("empty string")
	}

	fields := strings.Fields(s)

	f, ok := parsers[fields[0]]
	if !ok {
		return nil, errors.New("unknown message")
	}

	return f(fields)
}

func parseInitialize(fields []string) (msg any, err error) {
	if len(fields) != 2 {
		return nil, errors.New("wrong number of arguments")
	}
	msg = MessageInitialize{
		First: fields[1] == "1",
	}
	return msg, nil
}

func parseYourName(fields []string) (msg any, err error) {
	if len(fields) < 2 {
		return nil, errors.New("wrong number of arguments")
	}
	msg = MessageYourName{
		Name: strings.Join(fields[1:], " "),
	}
	return msg, nil
}

func parseYourColour(fields []string) (msg any, err error) {
	if len(fields) < 2 {
		return nil, errors.New("wrong number of arguments")
	}
	msg = MessageYourColour{
		Colour: strings.Join(fields[1:], " "),
	}
	return msg, nil
}

func parseGameOption(fields []string) (msg any, err error) {
	if len(fields) != 3 {
		return nil, errors.New("wrong number of arguments")
	}
	option, err := strconv.ParseInt(fields[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse option %q: %v", fields[1], err)
	}
	value, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse value %q: %v", fields[2], err)
	}
	msg = MessageGameOption{
		Option: GOption(option),
		Value:  value,
	}
	return msg, nil
}

func parseGameStarts(fields []string) (msg any, err error) {
	if len(fields) != 1 {
		return nil, errors.New("wrong number of arguments")
	}
	return MessageGameStarts{}, nil
}

func parseRadar(fields []string) (msg any, err error) {
	if len(fields) != 4 {
		return nil, errors.New("wrong number of arguments")
	}
	distance, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse distance %q: %v", fields[1], err)
	}
	object, err := strconv.ParseInt(fields[2], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse object type %q: %v", fields[2], err)
	}
	radarAngle, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse angle %q: %v", fields[3], err)
	}
	msg = MessageRadar{
		Distance:   distance,
		Object:     Object(object),
		RadarAngle: radarAngle,
	}
	return msg, nil
}

func parseInfo(fields []string) (msg any, err error) {
	if len(fields) != 4 {
		return nil, errors.New("wrong number of arguments")
	}
	time, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse time %q: %v", fields[1], err)
	}
	speed, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse speed %q: %v", fields[2], err)
	}
	cannonAngle, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse cannon angle %q: %v", fields[3], err)
	}
	msg = MessageInfo{
		Time:        time,
		Speed:       speed,
		CannonAngle: cannonAngle,
	}
	return msg, nil
}

func parseCoordinates(fields []string) (msg any, err error) {
	if len(fields) != 4 {
		return nil, errors.New("wrong number of arguments")
	}
	x, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse x %q: %v", fields[1], err)
	}
	y, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse y %q: %v", fields[2], err)
	}
	angle, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse angle %q: %v", fields[3], err)
	}
	msg = MessageCoordinates{
		X:     x,
		Y:     y,
		Angle: angle,
	}
	return msg, nil
}

func parseRobotInfo(fields []string) (msg any, err error) {
	if len(fields) != 3 {
		return nil, errors.New("wrong number of arguments")
	}
	energyLevel, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse energy level %q: %v", fields[1], err)
	}
	msg = MessageRobotInfo{
		EnergyLevel: energyLevel,
		TeamMate:    fields[2] == "1",
	}
	return msg, nil
}

func parseRotationReached(fields []string) (msg any, err error) {
	if len(fields) != 2 {
		return nil, errors.New("wrong number of arguments")
	}
	part, err := strconv.ParseInt(fields[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse robot part %q: %v", fields[1], err)
	}
	msg = MessageRotationReached{
		Part: Part(part),
	}
	return msg, nil
}

func parseEnergy(fields []string) (msg any, err error) {
	if len(fields) != 2 {
		return nil, errors.New("wrong number of arguments")
	}
	energyLevel, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse energy level %q: %v", fields[1], err)
	}
	msg = MessageEnergy{
		EnergyLevel: energyLevel,
	}
	return msg, nil
}

func parseRobotsLeft(fields []string) (msg any, err error) {
	if len(fields) != 2 {
		return nil, errors.New("wrong number of arguments")
	}
	numRobots, err := strconv.ParseInt(fields[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse number of robots %q: %v", fields[1], err)
	}
	msg = MessageRobotsLeft{
		NumRobots: int(numRobots),
	}
	return msg, nil
}

func parseCollision(fields []string) (msg any, err error) {
	if len(fields) != 3 {
		return nil, errors.New("wrong number of arguments")
	}
	object, err := strconv.ParseInt(fields[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse object type %q: %v", fields[1], err)
	}
	angle, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse angle %q: %v", fields[2], err)
	}
	msg = MessageCollision{
		Object: Object(object),
		Angle:  angle,
	}
	return msg, nil
}

func parseWarning(fields []string) (msg any, err error) {
	if len(fields) < 2 {
		return nil, errors.New("wrong number of arguments")
	}

	warning, err := strconv.ParseInt(fields[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("could not parse warning type %q: %v", fields[1], err)
	}

	warnMsg := ""
	if len(fields) > 2 {
		warnMsg = strings.Join(fields[2:], " ")
	}

	msg = MessageWarning{
		Warning: Warning(warning),
		Message: warnMsg,
	}
	return msg, nil
}

func parseDead(fields []string) (msg any, err error) {
	if len(fields) != 1 {
		return nil, errors.New("wrong number of arguments")
	}
	return MessageDead{}, nil
}

func parseGameFinishes(fields []string) (msg any, err error) {
	if len(fields) != 1 {
		return nil, errors.New("wrong number of arguments")
	}
	return MessageGameFinishes{}, nil
}

func parseExitRobot(fields []string) (msg any, err error) {
	if len(fields) != 1 {
		return nil, errors.New("wrong number of arguments")
	}
	return MessageExitRobot{}, nil
}

// Debug allows to enable debug messages.
var Debug = false

// dbg sends a debug message if Debug is true.
func dbgf(format string, a ...any) error {
	if !Debug {
		return nil
	}
	return Debugf(format, a...)
}
