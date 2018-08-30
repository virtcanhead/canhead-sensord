# canhead-sensord

Bluetooth Motion Sensor Daemon for canhead

## Device

* WIT-MOTION BWT901BCL

## Usage

**build the command**

```bash
go build

./canhead-sensord --help
```

```text
Usage of ./canhead-sensord:
  -baud int
        bound rate of serial device (default 115200)
  -bind string
        bind address of NDJSON API service (default "127.0.0.1:6770")
  -calibrate
        start calibration
  -device string
        serial device file to open (default "/dev/tty.HC-06-DevB")
```

**pair the device**

Pair the device with computer, this will automatically create a BT SPP device.

By default, the device is `/dev/tty.HC-06-DevB`, if is not, you have to use the `-device` option

**calibrate the device (if needed)**

```bash
./canhead-sensord --calibrate
```

**run the canhead-sensord**

```bash
./canhead-sensord
```

**read the values**

```bash
nc 127.0.0.1 6770
```

```text

```

## Lisense

MIT License

canhead <hi@canhead.xyz>
