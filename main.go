package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
	"github.com/virtcanhead/mohead/jy901"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	optDevice string
	optBaud   int
	optSetup  bool
	optBind   string

	angles     Angles
	anglesCond = sync.NewCond(&sync.Mutex{})
)

type Angles struct {
	ID    uint64
	Roll  float64
	Pitch float64
	Yaw   float64
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	flag.StringVar(&optDevice, "device", "/dev/tty.HC-06-DevB", "serial device file to open")
	flag.IntVar(&optBaud, "baud", 115200, "bound rate of serial device")
	flag.BoolVar(&optSetup, "setup", false, "setup the device")
	flag.StringVar(&optBind, "bind", "127.0.0.1:6770", "bind address of NDJSON API service")
	flag.Parse()

	log.Info().Str("device", optDevice).Int("baud", optBaud).Bool("calibrate", optSetup).Str("bind", optBind).Msg("options loaded")

	var err error
	var p *serial.Port
	if p, err = serial.OpenPort(&serial.Config{Name: optDevice, Baud: optBaud}); err != nil {
		log.Error().Err(err).Msg("failed to open serial device")
		return
	}
	defer p.Close()

	if optSetup {
		calibrate(p)
		return
	}

	var l net.Listener
	if l, err = net.Listen("tcp", optBind); err != nil {
		log.Error().Err(err).Msg("failed to bind")
		return
	}
	defer l.Close()

	go listenerRoutine(l)

	frameReaderRoutine(p)
}

func listenerRoutine(l net.Listener) {
	var err error
	for {
		var c net.Conn
		if c, err = l.Accept(); err != nil {
			log.Error().Err(err).Msg("failed to accept")
			return
		}
		go connectionRoutine(c)
	}
}

func connectionRoutine(c net.Conn) {
	var err error
	defer c.Close()
	var localID uint64
	for {
		if err = connectionLoop(c, &localID); err != nil {
			break
		}
	}
}

func connectionLoop(c net.Conn, localID *uint64) (err error) {
	anglesCond.L.Lock()
	defer anglesCond.L.Unlock()
	// check duplicated fire
	for *localID == angles.ID {
		anglesCond.Wait()
	}
	// marshal angles
	var buf []byte
	if buf, err = json.Marshal(&angles); err != nil {
		log.Error().Err(err).Msg("failed to marshal angles")
		return
	}
	// write NDJSON
	if _, err = fmt.Fprintln(c, string(buf)); err != nil {
		log.Error().Err(err).Msgf("failed to write connection")
		return
	}
	// update id
	*localID = angles.ID
	return
}

func frameReaderRoutine(p io.Reader) {
	var err error
	fr := jy901.NewFrameReader(p)
	for {
		var f jy901.Frame
		if f, err = fr.NextFrame(); err != nil {
			log.Error().Err(err).Msg("failed to read next frame")
			return
		}
		if f.Type == jy901.FrameTypeAngles {
			f.GetAngles(&angles.Roll, &angles.Pitch, &angles.Yaw)
			atomic.AddUint64(&angles.ID, 1)
			anglesCond.Broadcast()
		}
	}
}

func calibrate(p io.Writer) {
	var err error
	if _, err = p.Write(jy901.ResetCommand); err != nil {
		log.Error().Err(err).Msg("failed to send reset command")
		return
	}
	log.Info().Msg("module reset")

	time.Sleep(time.Second)

	if _, err = p.Write(jy901.SelectFramesCommand); err != nil {
		log.Error().Err(err).Msg("failed to send reduce frames command")
		return
	}
	log.Info().Msg("frames selected")

	time.Sleep(time.Second)

	countdown("starting acceleration calibration", 5)

	if _, err = p.Write(jy901.AccelerationCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send acceleration calibration command")
	}

	countdown("acceleration calibration in progress", 5)

	if _, err = p.Write(jy901.ExitCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send exit calibration command")
	}

	log.Info().Msg("acceleration calibration complete")

	countdown("starting magnetic field calibration", 5)

	if _, err = p.Write(jy901.MagneticFieldCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send magnetic calibration command")
	}

	countdown("magnetic calibration in progress", 20)

	if _, err = p.Write(jy901.ExitCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send exit calibration command")
	}

	log.Info().Msg("magnetic calibration complete")

	time.Sleep(time.Second)

	if _, err = p.Write(jy901.SelectSpeedCommand); err != nil {
		log.Error().Err(err).Msg("failed to send speed mode command")
		return
	}
	log.Info().Msg("speed selected")

	time.Sleep(time.Second)

	if _, err = p.Write(jy901.SaveCommand); err != nil {
		log.Error().Err(err).Msg("failed to send save command")
		return
	}
	log.Info().Msg("saved")
}

func countdown(title string, i int) {
	log.Info().Msgf("%s: %d...", title, i)
	for {
		time.Sleep(time.Second)
		i--
		if i <= 0 {
			break
		}
		log.Info().Msgf("%s: %d...", title, i)
	}
}
