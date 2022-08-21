package local

import (
	"fyne.io/fyne/v2/data/binding"
	"github.com/kiamev/moogle-mod-manager/config"
	"github.com/kiamev/moogle-mod-manager/mods"
	"github.com/kiamev/moogle-mod-manager/mods/managed"
	ci "github.com/kiamev/moogle-mod-manager/ui/config-installer"
	"github.com/kiamev/moogle-mod-manager/ui/state"
	"github.com/kiamev/moogle-mod-manager/ui/util"
	"path/filepath"
)

type enableBind struct {
	binding.Bool
	tm   *mods.TrackedMod
	done mods.DoneCallback
}

func newEnableBind(tm *mods.TrackedMod, done mods.DoneCallback) *enableBind {
	b := &enableBind{
		Bool: binding.NewBool(),
		tm:   tm,
		done: done,
	}
	_ = b.Set(tm.Enabled)
	b.AddListener(b)
	return b
}

func (b *enableBind) DataChanged() {
	isChecked, _ := b.Get()
	if isChecked != b.tm.Enabled {
		if isChecked {
			if err := b.EnableMod(); err != nil {
				if err != nil {
					util.ShowErrorLong(err)
					_ = b.Set(false)
				}
			}
		} else {
			b.done(b.DisableMod())
			_ = b.Set(b.tm.Enabled)
		}
	}
}

func (b *enableBind) EnableMod() (err error) {
	var (
		tm  = b.tm
		tis []*mods.ToInstall
	)
	if len(b.tm.Mod.Configurations) > 0 {
		err = b.enableModWithConfig()
	} else {
		tis, err = mods.NewToInstallForMod(tm.Mod.ModKind.Kind, tm.Mod, tm.Mod.AlwaysDownload)
		if err == nil {
			// Success
			err = b.enableMod(tis)
		}
	}
	return err
}

func (b *enableBind) enableModWithConfig() (err error) {
	modPath := filepath.Join(config.Get().GetModsFullPath(*state.CurrentGame), b.tm.GetDirSuffix())
	if err = state.GetScreen(state.ConfigInstaller).(ci.ConfigInstaller).Setup(b.tm.Mod, modPath, b.enableMod); err != nil {
		// Failed to set up config installer screen
		return
	}
	state.ShowScreen(state.ConfigInstaller)
	return
}

func (b *enableBind) enableMod(toInstall []*mods.ToInstall) (err error) {
	return managed.EnableMod(mods.NewModEnabler(*state.CurrentGame, b.tm, toInstall, func(err error) {
		_ = b.Set(b.tm.Enabled)
		b.done(err)
	}))
}

func (b *enableBind) DisableMod() error {
	return managed.DisableMod(*state.CurrentGame, b.tm)
}

func (b *enableBind) OnConflict() {

}
