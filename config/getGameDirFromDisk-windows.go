//go:build windows
// +build windows

package config

import (
	"golang.org/x/sys/windows/registry"
)

func (c *Configs) getGameDirFromDisk(gameId string) (dir string) {
	//only poke into registry for Windows, there's probably a similar method for Mac/Linux
	//remove the if here, since this will only build on windows
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, windowsRegLookup+gameId, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	if dir, _, err = key.GetStringValue("InstallLocation"); err != nil {
		dir = ""
	}
	return

}
