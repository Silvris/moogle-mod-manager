package discover

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kiamev/moogle-mod-manager/config"
	"github.com/kiamev/moogle-mod-manager/mods"
	"github.com/kiamev/moogle-mod-manager/mods/managed"
	"github.com/kiamev/moogle-mod-manager/remote"
	"github.com/kiamev/moogle-mod-manager/repo"
	cw "github.com/kiamev/moogle-mod-manager/ui/custom-widgets"
	mp "github.com/kiamev/moogle-mod-manager/ui/mod-preview"
	"github.com/kiamev/moogle-mod-manager/ui/state"
	"github.com/kiamev/moogle-mod-manager/ui/util"
	"golang.org/x/sync/errgroup"
	"sort"
	"strings"
)

func New() state.Screen {
	return &discoverUI{}
}

type discoverUI struct {
	selectedMod *mods.Mod
	data        binding.UntypedList
	split       *container.Split
	mods        []*mods.Mod
	localMods   map[string]bool
	prevSearch  string
}

func (ui *discoverUI) OnClose() {}

func (ui *discoverUI) PreDraw(w fyne.Window, args ...interface{}) (err error) {
	var (
		d          = dialog.NewInformation("", "Finding Mods...", w)
		remoteMods []*mods.Mod
		repoMods   []*mods.Mod
		found      *mods.Mod
		repoGetter = repo.NewGetter()
		eg         errgroup.Group
		ok         bool
	)
	defer d.Hide()
	d.Show()

	ui.localMods = make(map[string]bool)
	for _, tm := range args[0].([]interface{})[0].([]*mods.TrackedMod) {
		ui.localMods[tm.GetModID()] = true
	}

	eg.Go(func() (e error) {
		remoteMods, e = remote.GetMods(*state.CurrentGame)
		return
	})
	eg.Go(func() (e error) {
		repoMods, e = repoGetter.GetMods(*state.CurrentGame)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}

	lookup := make(map[string]*mods.Mod)
	for _, m := range repoMods {
		if _, ok = lookup[m.ModUniqueID(*state.CurrentGame)]; !ok {
			lookup[m.ModUniqueID(*state.CurrentGame)] = m
		}
	}
	for _, m := range remoteMods {
		if found, ok = lookup[m.ModUniqueID(*state.CurrentGame)]; !ok {
			lookup[m.ModUniqueID(*state.CurrentGame)] = m
		} else {
			found.Merge(m)
		}
	}

	ui.mods = make([]*mods.Mod, 0, len(lookup))
	for _, m := range lookup {
		if _, ok = ui.localMods[m.ID]; !ok {
			ui.mods = append(ui.mods, m)
		}
	}
	return
}

func (ui *discoverUI) DrawAsDialog(w fyne.Window) {
	ui.draw(w, true)
}

func (ui *discoverUI) Draw(w fyne.Window) {
	ui.draw(w, false)
}

func (ui *discoverUI) draw(w fyne.Window, isPopup bool) {
	if len(ui.mods) == 0 {
		// TODO
		return
	}
	ui.data = binding.NewUntypedList()
	modList := widget.NewListWithData(
		ui.data,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(item binding.DataItem, co fyne.CanvasObject) {
			var m *mods.Mod
			if i, ok := cw.GetValueFromDataItem(item); ok {
				if m, ok = i.(*mods.Mod); ok {
					co.(*widget.Label).SetText(m.Name)
				}
			}
		})
	if err := ui.showSorted(ui.mods); err != nil {
		util.ShowErrorLong(err, w)
		return
	}

	ui.split = container.NewHSplit(modList, container.NewMax())
	ui.split.SetOffset(0.25)

	modList.OnSelected = func(id widget.ListItemID) {
		data, err := ui.data.GetItem(id)
		if err != nil {
			util.ShowErrorLong(err, w)
			return
		}
		if i, ok := cw.GetValueFromDataItem(data); ok {
			ui.selectedMod = i.(*mods.Mod)
		}
		ui.split.Trailing = container.NewCenter(widget.NewLabel("Loading..."))
		ui.split.Refresh()
		ui.split.Trailing = container.NewBorder(
			container.NewHBox(widget.NewButton("Include Mod", func() {
				mod := ui.selectedMod
				if err := managed.AddMod(*state.CurrentGame, mods.NewTrackerMod(mod, *state.CurrentGame)); err != nil {
					util.ShowErrorLong(err, w)
					return
				}
				for i, m := range ui.mods {
					if m == mod {
						ui.mods = append(ui.mods[:i], ui.mods[i+1:]...)
						break
					}
				}
				sl := make([]interface{}, len(ui.mods))
				for i, m := range ui.mods {
					sl[i] = m
				}
				if err := ui.data.Set(sl); err != nil {
					util.ShowErrorLong(err, w)
					return
				}
				ui.selectedMod = nil
				ui.split.Trailing = container.NewMax()
				ui.split.Refresh()
				state.UpdateCurrentScreen()
			})), nil, nil, nil,
			mp.CreatePreview(ui.selectedMod))
		ui.split.Refresh()
	}

	searchTb := widget.NewEntry()
	searchTb.OnChanged = func(s string) {
		if err := ui.search(s); err != nil {
			util.ShowErrorLong(err, w)
		}
	}

	w.SetContent(container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle(config.GameNameString(*state.CurrentGame), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
		), nil, nil, nil, container.NewBorder(
			container.NewAdaptiveGrid(8, container.NewHBox(widget.NewButton("Back", func() {
				if isPopup {
					state.ClosePopupWindow()
				} else {
					state.ShowPreviousScreen()
				}

			})), widget.NewLabelWithStyle("Search", fyne.TextAlignTrailing, fyne.TextStyle{}), searchTb), nil, nil, nil,
			ui.split)))
}

func (ui *discoverUI) search(s string) error {
	if s == ui.prevSearch || (len(s) < 3 && ui.prevSearch == "") {
		s = ""
		if ui.data.Length() == len(ui.mods) {
			return nil
		}
	}
	s = strings.ToLower(s)
	ui.prevSearch = s

	var ms []*mods.Mod
	for _, m := range ui.mods {
		if strings.Contains(strings.ToLower(m.Name), s) ||
			strings.Contains(strings.ToLower(string(m.Category)), s) ||
			strings.Contains(strings.ToLower(m.Description), s) ||
			strings.Contains(strings.ToLower(m.Author), s) {
			ms = append(ms, m)
		}
	}
	return ui.showSorted(ms)
}

func (ui *discoverUI) showSorted(ms []*mods.Mod) error {
	lookup := make(map[string]*mods.Mod)
	sorted := make([]string, len(ms))
	for i, m := range ms {
		key := fmt.Sprintf("%s%s", m.Name, m.ID)
		lookup[key] = m
		sorted[i] = key
	}
	sort.Strings(sorted)

	_ = ui.data.Set(nil)
	for _, s := range sorted {
		if err := ui.data.Append(lookup[s]); err != nil {
			return err
		}
	}
	return nil
}
