package dicom

//go:generate ./generate_dimse_messages.py

// Standard DICOM tag definitions.
//
// ftp://medical.nema.org/medical/dicom/2011/11_06pu.pdf

import (
	"fmt"
	"strconv"
	"strings"
	"v.io/x/lib/vlog"
)

// Tag is a <group, element> tuple that identifies an element type in a DICOM
// file. List of standard tags are defined in tag.go. See also:
//
// ftp://medical.nema.org/medical/dicom/2011/11_06pu.pdf
type Tag struct {
	// Group and element are results of parsing the hex-pair tag, such as (1000,10008)
	Group   uint16
	Element uint16
}

// Return a string of form "(0008,1234)", where 0x0008 is t.Group,
// 0x1234 is t.Element.
func (t *Tag) String() string {
	return fmt.Sprintf("(%04x,%04x)", t.Group, t.Element)
}

type TagInfo struct {
	Tag Tag
	// Data encoding "UL", "CS", etc.
	VR string
	// Human-readable name of the tag, e.g., "CommandDataSetType"
	Name string
	// Cardinality (# of values expected in the element)
	VM string
}

const TagMetadataGroup = 2

// VRKind defines the golang encoding of a VR.
type VRKind int

const (
	VRString VRKind = iota
	VRBytes
	VRUInt16
	VRUInt32
	VRInt16
	VRInt32
	VRFloat32
	VRFloat64
	VRSequence
	VRItem
	VRTag
	VRPixelData
)

// GetVRKind returns the encoding type of VR (e.g., VRUint32) given a
// two-character VR string (e.g., "UL").
func GetVRKind(vr string) VRKind {
	switch vr {
	case "AT":
		return VRTag
	case "OW", "OB":
		return VRBytes
	case "UL":
		return VRUInt32
	case "SL":
		return VRInt32
	case "US":
		return VRUInt16
	case "SS":
		return VRInt16
	case "FL":
		return VRFloat32
	case "FD":
		return VRFloat64
	default:
		return VRString
	}
}

// LookupTag finds information about the given tag. If the tag is not part of
// the DICOM standard, or is retired from the standard, it returns an error.
func LookupTag(tag Tag) (TagInfo, error) {
	maybeInitTagDict()
	entry, ok := tagDict[tag]
	if !ok {
		// (0000-u-ffff,0000)	UL	GenericGroupLength	1	GENERIC
		if tag.Group%2 == 0 && tag.Element == 0x0000 {
			entry = TagInfo{tag, "UL", "GenericGroupLength", "1"}
		} else {
			return TagInfo{}, fmt.Errorf("Could not find tag (0x%x, 0x%x) in dictionary", tag.Group, tag.Element)
		}
	}
	return entry, nil
}

// Like LookupTag, but panics on error.
func MustLookupTag(tag Tag) TagInfo {
	e, err := LookupTag(tag)
	if err != nil {
		vlog.Fatalf("tag %s not found: %s", tag, err)
	}
	return e
}

// LookupTag finds information about the tag with the given name. If the tag is not part of
// the DICOM standard, or is retired from the standard, it returns an error.
//
//   Example: LookupTagByName("TransferSyntaxUID")
func LookupTagByName(name string) (TagInfo, error) {
	maybeInitTagDict()
	for _, ent := range tagDict {
		if ent.Name == name {
			return ent, nil
		}
	}
	return TagInfo{}, fmt.Errorf("Could not find tag with name %s", name)
}

// TagString returns a human-readable diagnostic string for the tag
func TagString(tag Tag) string {
	e, err := LookupTag(tag)
	if err != nil {
		return fmt.Sprintf("(%04x,%04x)[??]", tag.Group, tag.Element)
	}
	return fmt.Sprintf("(%04x,%04x)[%s]", tag.Group, tag.Element, e.Name)
}

// Split a tag into a group and element, represented as a hex value
// TODO: support group ranges (6000-60FF,0803)
func parseTag(tag string) (Tag, error) {
	parts := strings.Split(strings.Trim(tag, "()"), ",")
	group, err := strconv.ParseInt(parts[0], 16, 0)
	if err != nil {
		return Tag{}, err
	}
	elem, err := strconv.ParseInt(parts[1], 16, 0)
	if err != nil {
		return Tag{}, err
	}
	return Tag{Group: uint16(group), Element: uint16(elem)}, nil
}
