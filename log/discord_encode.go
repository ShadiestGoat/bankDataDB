package log

import (
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

type discEnc struct {
	strRepl *strings.Replacer
	strings.Builder
}

func (d *discEnc) write(k, v string) {
	d.WriteString(k + ": " + v + "\n")
}

func (d *discEnc) AddArray(key string, _ zapcore.ArrayMarshaler) error {
	d.write(key, "[array]")
	return nil
}

func (d *discEnc) AddObject(k string, _ zapcore.ObjectMarshaler) error {
	d.write(k, "{object}")
	return nil
}

func (d *discEnc) AddBinary(k string, arr []byte) {
	if len(arr) == 0 {
		d.write(k, "<no bin data>")

		return
	}

	str := ""
	for _, v := range arr {
		str += strconv.FormatInt(int64(v), 16) + " "
	}

	d.write(k, str[:len(str)-1])
}

func (d *discEnc) addString(k string, prefix string, v string) {
	d.write(k, prefix + `"` + d.strRepl.Replace(v) + `"`)
}

func (d *discEnc) AddByteString(k string, v []byte) {
	d.addString(k, "b", string(v))
}

func (d *discEnc) AddString(k, v string) {
	d.addString(k, "", v)
}

func (d *discEnc) AddBool(k string, v bool) {
	if v {
		d.write(k, "true")
	} else {
		d.write(k, "false")
	}
}

func (d *discEnc) AddComplex128(k string, v complex128) {
	d.write(k, strconv.FormatComplex(v, 'f', 3, 64))
}

func (d *discEnc) AddDuration(k string, v time.Duration) {
	d.write(k, v.String())
}

func (d *discEnc) AddFloat64(k string, v float64) {
	d.write(k, strconv.FormatFloat(v, 'f', 3, 64))
}

func (d *discEnc) AddInt(k string, v int) {
	d.write(k, strconv.Itoa(v))
}

func (d *discEnc) AddTime(k string, v time.Time) {
	d.write(k, v.Format(time.DateTime))
}

func (d *discEnc) AddUint64(k string, v uint64) {
	d.write(k, strconv.FormatUint(v, 10))
}

func (d *discEnc) AddUintptr(k string, v uintptr) {
	d.write(k, "*ptr")
}

func (d *discEnc) AddReflected(k string, _ any) error {
	d.write(k, "???")
	return nil
}

// No namespacing in this country
func (d *discEnc) OpenNamespace(k string) {}

// Uint format
// FormatUint takes a uint64 value so thats why that is the base instead of AddUint
func (d *discEnc) AddUint(k string, v uint) { d.AddUint64(k, uint64(v)) }
func (d *discEnc) AddUint32(k string, v uint32) { d.AddUint64(k, uint64(v)) }
func (d *discEnc) AddUint16(k string, v uint16) { d.AddUint64(k, uint64(v)) }
func (d *discEnc) AddUint8(k string, v uint8) { d.AddUint64(k, uint64(v)) }

// Int format
func (d *discEnc) AddInt64(k string, v int64) { d.AddInt(k, int(v)) }
func (d *discEnc) AddInt32(k string, v int32) { d.AddInt(k, int(v)) }
func (d *discEnc) AddInt16(k string, v int16) { d.AddInt(k, int(v)) }
func (d *discEnc) AddInt8(k string, v int8) { d.AddInt(k, int(v)) }

// Float format
func (d *discEnc) AddFloat32(k string, v float32) { d.AddFloat64(k, float64(v)) }

// Complex format
func (d *discEnc) AddComplex64(k string, v complex64) { d.AddComplex128(k, complex128(v)) }
