package main

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
	"github.com/virtcanhead/sensorhead/bt901"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	optDevice    string
	optBaud      int
	optCalibrate bool

	waitGroup = &sync.WaitGroup{}

	isShuttingDown bool
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	flag.StringVar(&optDevice, "d", "/dev/tty.HC-06-DevB", "serial device file to open")
	flag.IntVar(&optBaud, "b", 115200, "bound rate of serial device")
	flag.BoolVar(&optCalibrate, "c", false, "start calibration")
	flag.Parse()

	var err error
	var p *serial.Port
	if p, err = serial.OpenPort(&serial.Config{Name: optDevice, Baud: optBaud}); err != nil {
		log.Error().Err(err).Msg("failed to open serial device")
		return
	}
	defer p.Close()

	if optCalibrate {
		calibrate(p)
		return
	}

	go signalRoutine()
	waitGroup.Add(1)
	go frameRoutine(p)
	waitGroup.Wait()
}

func frameRoutine(p io.Reader) {
	var err error
	defer waitGroup.Done()
	fr := bt901.NewFrameReader(p)
	for {
		var f bt901.Frame
		if f, err = fr.NextFrame(); err != nil {
			log.Error().Err(err).Msg("failed to read next frame")
			return
		}
		if f.Type == bt901.FrameTypeAngles {
			var rol, pit, yaw float64
			f.GetAngles(&rol, &pit, &yaw)
			log.Info().Msgf("Roll, Pitch, Yaw = %04.4f, %04.4f, %04.4f", rol, pit, yaw)
		}
		if isShuttingDown {
			log.Debug().Msg("read routine exiting")
			break
		}
	}
}

func signalRoutine() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)
	sig := <-done
	log.Info().Str("signal", sig.String()).Msg("signal received")
	isShuttingDown = true
}

func calibrate(p io.Writer) {
	var err error
	if _, err = p.Write(bt901.ResetCommand); err != nil {
		log.Error().Err(err).Msg("failed to send reset command")
		return
	}
	log.Info().Msg("module reset")

	time.Sleep(time.Second)

	if _, err = p.Write(bt901.ReduceFramesCommand); err != nil {
		log.Error().Err(err).Msg("failed to send reduce frames command")
		return
	}
	log.Info().Msg("frame reduced")

	time.Sleep(time.Second)

	countdown("starting acceleration calibration", 5)

	if _, err = p.Write(bt901.AccelerationCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send acceleration calibration command")
	}

	countdown("acceleration calibration in progress", 5)

	if _, err = p.Write(bt901.ExitCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send exit calibration command")
	}

	log.Info().Msg("acceleration calibration complete")

	countdown("starting magnetic field calibration", 5)

	if _, err = p.Write(bt901.MagneticFieldCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send magnetic calibration command")
	}

	countdown("magnetic calibration in progress", 20)

	if _, err = p.Write(bt901.ExitCalibrationCommand); err != nil {
		log.Error().Err(err).Msg("failed to send exit calibration command")
	}

	log.Info().Msg("magnetic calibration complete")

	time.Sleep(time.Second)

	if _, err = p.Write(bt901.SpeedModeCommand); err != nil {
		log.Error().Err(err).Msg("failed to send speed mode command")
		return
	}
	log.Info().Msg("more speed")

	time.Sleep(time.Second)

	if _, err = p.Write(bt901.SaveCommand); err != nil {
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
