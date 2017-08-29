package dicom_test

import (
	"bytes"
	"encoding/binary"
	"github.com/yasushi-saito/go-dicom"
	"io"
	"testing"
)

func TestBasic(t *testing.T) {
	e := dicom.NewEncoder(binary.BigEndian, dicom.UnknownVR)
	e.WriteByte(10)
	e.WriteByte(11)
	e.WriteUInt16(0x123)
	e.WriteUInt32(0x234)
	e.WriteZeros(12)
	e.WriteString("abcde")

	encoded, err := e.Finish()
	if err != nil {
		t.Fatal(encoded)
	}
	d := dicom.NewDecoder(
		bytes.NewBuffer(encoded), int64(len(encoded)),
		binary.BigEndian, dicom.ImplicitVR)
	if v := d.ReadByte(); v != 10 {
		t.Errorf("ReadByte %v", v)
	}
	if v := d.ReadByte(); v != 11 {
		t.Errorf("ReadByte %v", v)
	}
	if v := d.ReadUInt16(); v != 0x123 {
		t.Errorf("ReadUint16 %v", v)
	}
	if v := d.ReadUInt32(); v != 0x234 {
		t.Errorf("ReadUint32 %v", v)
	}
	d.Skip(12)
	if v := d.ReadString(5); v != "abcde" {
		t.Errorf("ReadString %v", v)
	}
	if d.Len() != 0 {
		t.Errorf("Len %d", d.Len())
	}
	if d.Error() != nil {
		t.Errorf("!Error %v", d.Error())
	}
	// Read past the buffer. It should flag an error
	if _ = d.ReadByte(); d.Error() == nil {
		t.Errorf("Error %v %v", d.Error())
	}
}

func TestSkip(t *testing.T) {
	e := dicom.NewEncoder(binary.BigEndian, dicom.UnknownVR)
	e.WriteString("abcdefghijk")
	encoded, err := e.Finish()
	if err != nil {
		t.Fatal(encoded)
	}

	d := dicom.NewBytesDecoder(encoded, binary.BigEndian, dicom.UnknownVR)
	d.Skip(3)
	if d.Len() != 8 {
		t.Error("Skip 3; len")
	}
	if d.ReadString(8) != "defghijk" {
		t.Error("Skip 3; read")
	}
}

func TestPartialData(t *testing.T) {
	e := dicom.NewEncoder(binary.BigEndian, dicom.UnknownVR)
	e.WriteByte(10)
	encoded, err := e.Finish()
	if err != nil {
		t.Fatal(encoded)
	}
	// Read uint16, when there's only one byte in buffer.
	d := dicom.NewDecoder(bytes.NewBuffer(encoded), int64(len(encoded)),
		binary.BigEndian, dicom.ImplicitVR)
	if _ = d.ReadUInt16(); d.Error() == nil {
		t.Errorf("ReadUint16")
	}
}

func TestLimit(t *testing.T) {
	e := dicom.NewEncoder(binary.BigEndian, dicom.UnknownVR)
	e.WriteByte(10)
	e.WriteByte(11)
	e.WriteByte(12)
	encoded, err := e.Finish()
	if err != nil {
		t.Error(encoded)
	}
	// Allow reading only the first two bytes
	d := dicom.NewDecoder(bytes.NewBuffer(encoded), int64(len(encoded)),
		binary.BigEndian, dicom.ImplicitVR)
	if d.Len() != 3 {
		t.Errorf("Len %d", d.Len())
	}
	d.PushLimit(2)
	if d.Len() != 2 {
		t.Errorf("Len %d", d.Len())
	}
	v0, v1 := d.ReadByte(), d.ReadByte()
	if d.Len() != 0 {
		t.Errorf("Len %d", d.Len())
	}
	_ = d.ReadByte()
	if v0 != 10 || v1 != 11 || d.Error() != io.EOF {
		t.Error("Limit: %v %v %v", v0, v1, d.Error())
	}

}
