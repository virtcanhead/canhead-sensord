package bt901

var (
	SaveCommand                     = NewCommand(0x00, 0x00, 0x00)
	ResetCommand                    = NewCommand(0x00, 0x01, 0x00)
	ExitCalibrationCommand          = NewCommand(0x01, 0x00, 0x00)
	AccelerationCalibrationCommand  = NewCommand(0x01, 0x01, 0x00)
	MagneticFieldCalibrationCommand = NewCommand(0x01, 0x02, 0x00)
	ReduceFramesCommand             = NewCommand(0x02, 0x0f, 0x00)
	SpeedModeCommand                = NewCommand(0x03, 0x08, 0x00)
)

func NewCommand(addr byte, dl byte, dh byte) []byte {
	return []byte{0xFF, 0xAA, addr, dl, dh}
}
