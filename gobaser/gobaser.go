package gobaser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func WriteToFile(mname string, mtime time.Time, value int64) {
	var file *os.File

	log.Println("[DEBUG]: WriteToFile (started)")
	pathString := strings.Replace(mname, ".", "/", -1)
	pathString = pathString + ".graph"
	_, err := OpenOrCreateFile(pathString)
	if err != nil {
		return
	}
	defer file.Close()

	toWrite := new(bytes.Buffer)
	// UnixTime
	err = binary.Write(toWrite, binary.BigEndian, mtime.Unix())
	if err != nil {
		return
	}

	//

	//mfile.Write(toWrite.Bytes())
}

func OpenOrCreateFile(fpath string) (mfile *os.File, err error) {
	needWriteBody := false
	dir := path.Dir(fpath)
	fname := path.Base(fpath)
	log.Printf("[DEBUG]: fpath = '%v'", fpath)
	log.Printf("[DEBUG]: dir = '%v'", dir)
	log.Printf("[DEBUG]: fname = '%v'", fname)
	if _, err := os.Stat(fpath); err == nil {
		log.Printf("[DEBUG]: file '%v' exists", fpath)
	} else {
		log.Printf("[DEBUG]: file '%v' NOT exists", fpath)
		needWriteBody = true
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("[ERROR]: cant create dir '%v'", dir)
			return mfile, err
		}
	}

	mfile, err = os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Printf("[ERROR]: cant create file '%v'", fpath)
	}
	if needWriteBody {
		WriteClearBody(mfile)
	}
	return mfile, err
}

func WriteClearBody(mfile *os.File) (err error) {
	log.Printf("[DEBUG]: write new body")
	toWrite := new(bytes.Buffer)
	magic := uint16(0xbaaf)
	version := uint32(0x00000001)
	aggrType := uint8(0x01)            // 0x1 -- SUM
	maxRetention := uint32(0x00000000) // reserved
	archCount := uint8(0x02)

	err = binary.Write(toWrite, binary.BigEndian, magic)
	if err != nil {
		log.Printf("[ERROR]: cant write buffer '%v'", err)
		return err
	}

	err = binary.Write(toWrite, binary.BigEndian, version)
	if err != nil {
		log.Printf("[ERROR]: cant write buffer '%v'", err)
		return err
	}

	err = binary.Write(toWrite, binary.BigEndian, aggrType)
	if err != nil {
		log.Printf("[ERROR]: cant write buffer '%v'", err)
		return err
	}

	err = binary.Write(toWrite, binary.BigEndian, maxRetention)
	if err != nil {
		log.Printf("[ERROR]: cant write buffer '%v'", err)
		return err
	}

	err = binary.Write(toWrite, binary.BigEndian, archCount)
	if err != nil {
		log.Printf("[ERROR]: cant write buffer '%v'", err)
		return err
	}

	mfile.Write(toWrite.Bytes())
	return nil
}

func CheckFile(fn string) {
	fmt.Printf("CheckFile '%v'\n", fn)
}
