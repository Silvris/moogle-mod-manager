//go:build !windows
// +build !windows

package config

//this will show an error, it's just getting mad because of not being built normally
import (
	"fmt"
	"os"
)

type vdfObject struct {
	key    string
	strVal string
	objVal []vdfObject
}

func (v *vdfObject) parseVdf(data []byte) (remaining []byte) {
	//vdf features three control characters, we assume we start out with the key
	var (
		key       string
		seekpoint int
		strVal    string
		//		objVal    vdfObject
	)
	if data[0] == '"' {
		for i := 1; i < (len(data) - 1); i++ {
			//iterate until we reach the end character
			if data[i] == '"' {
				seekpoint = i
				key = string(data[1:seekpoint])
				break
			}
		}
	}
	//now that we have our key, retrieve the next control character
	for i := seekpoint; i < (len(data) - 1); i++ {
		if data[i] == '"' || data[i] == '{' {
			seekpoint = i
			break
		}
	}
	//change what we do based on the control character
	if data[seekpoint] == '"' {
		for i := seekpoint; i < (len(data) - 1); i++ {

		}
	}
	//set our values onto the vdf itself
	v.key = key
	if strVal != "" {
		v.strVal = strVal
	}
	/*
		if objVal != nil {
			v.objVal = objVal
		}*/
	var returnVal = []byte{0}
	return returnVal
}

func (c *Configs) getGameDirFromDisk(gameId string) (dir string) {
	//linux does not have the concept of a common registry, so instead we poll steam for some information
	var (
		b    []byte
		file string = "~/.steam/root/steamapps/libraryfolders.vdf"
		err  error
	)
	if _, err = os.Stat(file); err != nil {
		return ""
	}
	if b, err = os.ReadFile(file); err != nil {
		return ""
	}
	fmt.Println(b[0])
	//vdf = new(vdfObject)
	//vdf.parseVdf(b)
	return ""
}
