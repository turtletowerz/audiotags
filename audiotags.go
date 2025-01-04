/***************************************************************************
   copyright            : (C) 2014 by Nick Sellen
   email                : code@nicksellen.co.uk
***************************************************************************/

/***************************************************************************
 *   This library is free software; you can redistribute it and/or modify  *
 *   it  under the terms of the GNU Lesser General Public License version  *
 *   2.1 as published by the Free Software Foundation.                     *
 *                                                                         *
 *   This library is distributed in the hope that it will be useful, but   *
 *   WITHOUT ANY WARRANTY; without even the implied warranty of            *
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU     *
 *   Lesser General Public License for more details.                       *
 *                                                                         *
 *   You should have received a copy of the GNU Lesser General Public      *
 *   License along with this library; if not, write to the Free Software   *
 *   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307  *
 *   USA                                                                   *
 ***************************************************************************/

package audiotags

/*
#cgo pkg-config: taglib_c
#include "audiotags.h"
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	JPEG = iota
	PNG
)

type File C.TagLib_File

func (f *File) HasMedia() bool {
	return !f.ReadAudioProperties().isEmpty()
}

type AudioProperties struct {
	Length, Bitrate, Samplerate, Channels int
}

func (props *AudioProperties) isEmpty() bool {
	return props.Bitrate == 0 && props.Length == 0 && props.Samplerate == 0 && props.Channels == 0
}

func Open(filename string) (*File, error) {
	fp := C.CString(filename)
	defer C.free(unsafe.Pointer(fp))
	f := C.audiotags_file_new(fp)
	if f == nil {
		return nil, fmt.Errorf("cannot process %s", filename)
	}
	return (*File)(f), nil
}

func (f *File) Close() {
	C.taglib_file_free((*C.TagLib_File)(f))
}

func (f *File) ReadTags() keyMap {
	id := mapsNextID.Add(1)
	defer maps.Delete(id)

	m := keyMap{}
	maps.Store(id, m)
	C.audiotags_file_properties((*C.TagLib_File)(f), C.int(id))
	return m
}

func (f *File) WriteTags(tagMap keyMap) bool {
	if len(tagMap) == 0 {
		return bool(C.audiotags_clear_properties((*C.TagLib_File)(f)))
	}

	tagFields := make([]*C.char, len(tagMap))
	tagValues := make([]*C.char, len(tagMap))
	var i int
	for field, values := range tagMap {
		fieldC := C.CString(field)
		tagFields[i] = fieldC

		valueC := C.CString(strings.Join(values, "\v"))
		tagValues[i] = valueC
		i++
	}
	defer func() {
		for i := 0; i < len(tagMap); i++ {
			C.free(unsafe.Pointer(tagFields[i]))
			C.free(unsafe.Pointer(tagValues[i]))
		}
	}()

	return bool(C.audiotags_write_properties((*C.TagLib_File)(f), C.uint(len(tagMap)), &tagFields[0], &tagValues[0]))
}

func (f *File) ReadAudioProperties() *AudioProperties {
	props := C.audiotags_file_audioproperties((*C.TagLib_File)(f))
	if props == nil {
		return nil
	}
	p := AudioProperties{}
	p.Length = int(C.audiotags_audioproperties_length(props))
	p.Bitrate = int(C.audiotags_audioproperties_bitrate(props))
	p.Samplerate = int(C.audiotags_audioproperties_samplerate(props))
	p.Channels = int(C.audiotags_audioproperties_channels(props))
	return &p
}

func (f *File) ReadImage() (image.Image, error) {
	id := mapsNextID.Add(1)
	defer maps.Delete(id)

	C.audiotags_read_picture((*C.TagLib_File)(f), C.int(id))
	v, ok := maps.Load(id)
	if !ok {
		return nil, nil
	}
	img, _, err := image.Decode(v.(*bytes.Reader))
	return img, err
}

func (f *File) ReadImageRaw() *bytes.Reader {
	id := mapsNextID.Add(1)
	defer maps.Delete(id)

	C.audiotags_read_picture((*C.TagLib_File)(f), C.int(id))
	v, ok := maps.Load(id)
	if !ok {
		return nil
	}

	return v.(*bytes.Reader)
}

func (f *File) WriteImage(img image.Image, format int) error {
	var formatStr string
	buff := bytes.NewBuffer([]byte{})
	switch format {
	case JPEG:
		if err := jpeg.Encode(buff, img, &jpeg.Options{Quality: 65}); err != nil {
			return err
		}
		formatStr = "image/jpeg"
	case PNG:
		if err := png.Encode(buff, img); err != nil {
			return err
		}
		formatStr = "image/png"
	default:
		return fmt.Errorf("unsuppported image format")
	}

	data := buff.Bytes()
	if len(data) == 0 {
		return fmt.Errorf("can't write empty image")
	}

	if !f.WriteImageData(data, formatStr, img.Bounds().Size().X, img.Bounds().Size().Y) {
		return fmt.Errorf("can't write image")
	}

	return nil
}

func (f *File) WriteImageData(data []byte, format string, w, h int) bool {
	if len(data) == 0 {
		return false
	}

	formatCStr := C.CString(format)
	defer C.free(unsafe.Pointer(formatCStr))

	return bool(C.audiotags_write_picture((*C.TagLib_File)(f), (*C.char)(unsafe.Pointer(&data[0])), C.uint(len(data)), C.int(w), C.int(h), formatCStr))
}

func (f *File) RemovePictures() bool {
	return bool(C.audiotags_remove_pictures((*C.TagLib_File)(f)))
}

var maps sync.Map
var mapsNextID atomic.Uint64

type keyMap = map[string][]string

//export goTagPut
func goTagPut(id C.int, key *C.char, val *C.char) {
	r, _ := maps.Load(uint64(id))
	m := r.(keyMap)
	k := strings.ToLower(C.GoString(key))
	v := C.GoString(val)
	m[k] = append(m[k], v)
}

//export goPutImage
func goPutImage(id C.int, data *C.char, size C.int) {
	maps.Store(uint64(id), bytes.NewReader(C.GoBytes(unsafe.Pointer(data), size)))
}
