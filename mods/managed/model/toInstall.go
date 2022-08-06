package model

import (
	"fmt"
	"github.com/kiamev/moogle-mod-manager/config"
	"github.com/kiamev/moogle-mod-manager/mods"
	"github.com/kiamev/moogle-mod-manager/util"
	"os"
	"path/filepath"
)

type ToInstall struct {
	kind          mods.Kind
	Download      *mods.Download
	DownloadFiles []*mods.DownloadFiles
	downloadDir   string
}

func NewToInstall(kind mods.Kind, download *mods.Download, downloadFiles *mods.DownloadFiles) *ToInstall {
	return &ToInstall{
		kind:          kind,
		Download:      download,
		DownloadFiles: []*mods.DownloadFiles{downloadFiles},
	}
}

func NewToInstallForMod(kind mods.Kind, mod *mods.Mod, downloadFiles []*mods.DownloadFiles) (result []*ToInstall, err error) {
	lookup := make(map[string]*mods.Download)
	for _, dl := range mod.Downloadables {
		lookup[dl.Name] = dl
	}

	for _, f := range downloadFiles {
		dl, ok := lookup[f.DownloadName]
		if !ok {
			return nil, fmt.Errorf("could not find download %s for mod %s", f.DownloadName, mod.Name)
		}
		result = append(result, NewToInstall(kind, dl, f))
	}
	return
}

func (ti *ToInstall) GetDownloadLocation(game config.Game, tm *TrackedMod) (string, error) {
	if ti.kind == mods.Hosted {
		return ti.getHostedDownloadLocation(game, tm)
	}
	return ti.getNexusDownloadLocation(game, tm)
}

func (ti *ToInstall) getHostedDownloadLocation(game config.Game, tm *TrackedMod) (string, error) {
	if ti.downloadDir == "" {
		v := ti.Download.Version
		if v == "" {
			v = "nv"
		}
		ti.downloadDir = filepath.Join(config.Get().GetDownloadFullPath(game), tm.GetDirSuffix(), util.CreateFileName(v))
		if err := createPath(ti.downloadDir); err != nil {
			return "", err
		}
	}
	return ti.downloadDir, nil
}

func (ti *ToInstall) getNexusDownloadLocation(game config.Game, tm *TrackedMod) (string, error) {
	if ti.downloadDir == "" {
		ti.downloadDir = filepath.Join(config.Get().GetDownloadFullPath(game), tm.GetDirSuffix(), util.CreateFileName(ti.Download.Version))
		if err := createPath(ti.downloadDir); err != nil {
			return "", err
		}
	}
	return ti.downloadDir, nil
}

func createPath(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		err = fmt.Errorf("failed to create mod directory: %v", err)
		return err
	}
	return nil
}
