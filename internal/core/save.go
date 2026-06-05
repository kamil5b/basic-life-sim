package core

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/kamil5b/basic-life-sim/internal/experience"
	"github.com/kamil5b/basic-life-sim/internal/model"
)

const (
	saveVersion  = 1
	saveDir      = "saves"
	maxSaveSlots = 5
)

// SaveFile is the on-disk format for a save game.
type SaveFile struct {
	Version   int
	SavedAt   time.Time
	Character model.Character
}

// SaveInfo describes a single save slot for the UI.
type SaveInfo struct {
	Name     string
	CharName string
	Age      uint8
	Date     string
	SavedAt  string
	Exists   bool
}

func ensureSaveDir() error {
	return os.MkdirAll(saveDir, 0755)
}

func savePath(name string) string {
	return filepath.Join(saveDir, name+".gob")
}

func slotName(i int) string {
	return fmt.Sprintf("slot%d", i+1)
}

func saveExists(name string) bool {
	_, err := os.Stat(savePath(name))
	return err == nil
}

// listSaveFiles returns all existing save names sorted by modification time (newest first).
func listSaveFiles() ([]string, error) {
	if err := ensureSaveDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		return nil, err
	}
	type info struct {
		name string
		mod  time.Time
	}
	var files []info
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".gob" {
			fi, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, info{name: e.Name()[:len(e.Name())-4], mod: fi.ModTime()})
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.After(files[j].mod)
	})
	var out []string
	for _, f := range files {
		out = append(out, f.name)
	}
	return out, nil
}

func getSaveInfos() []SaveInfo {
	var infos []SaveInfo
	for i := 0; i < maxSaveSlots; i++ {
		name := slotName(i)
		info := SaveInfo{Name: name}
		path := savePath(name)
		if _, err := os.Stat(path); err == nil {
			file, err := os.Open(path)
			if err == nil {
				dec := gob.NewDecoder(file)
				var f SaveFile
				if dec.Decode(&f) == nil {
					info.Exists = true
					info.CharName = f.Character.Name
					info.Age = f.Character.Age
					y, m, d := f.Character.CurrentDate.Unpack()
					info.Date = fmt.Sprintf("%04d-%02d-%02d", y, m, d)
					info.SavedAt = f.SavedAt.Format("2006-01-02 15:04")
				}
				file.Close()
			}
		}
		infos = append(infos, info)
	}
	return infos
}

func peekSaveInfo(name string) SaveInfo {
	info := SaveInfo{Name: name}
	path := savePath(name)
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err == nil {
			dec := gob.NewDecoder(file)
			var f SaveFile
			if dec.Decode(&f) == nil {
				info.Exists = true
				info.CharName = f.Character.Name
				info.Age = f.Character.Age
				y, m, d := f.Character.CurrentDate.Unpack()
				info.Date = fmt.Sprintf("%04d-%02d-%02d", y, m, d)
				info.SavedAt = f.SavedAt.Format("2006-01-02 15:04")
			}
			file.Close()
		}
	}
	return info
}

func saveGame(name string, char model.Character) error {
	if err := ensureSaveDir(); err != nil {
		return err
	}
	preSave(&char)
	f := SaveFile{
		Version:   saveVersion,
		SavedAt:   time.Now(),
		Character: char,
	}
	path := savePath(name)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	return enc.Encode(f)
}

func loadGame(name string) (model.Character, error) {
	path := savePath(name)
	file, err := os.Open(path)
	if err != nil {
		return model.Character{}, err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	var f SaveFile
	if err := dec.Decode(&f); err != nil {
		return model.Character{}, err
	}
	char := f.Character
	postLoad(&char)
	return char, nil
}

func deleteSave(name string) error {
	return os.Remove(savePath(name))
}

// preSave nils out all function fields so gob can encode the struct tree.
func preSave(char *model.Character) {
	for i := range char.CurrentHome.RoomItems {
		item := &char.CurrentHome.RoomItems[i]
		item.Item.DoAction = nil
		for j := range item.Utilities {
			item.Utilities[j].Item.DoAction = nil
			for k := range item.Utilities[j].OnSurface {
				item.Utilities[j].OnSurface[k].Food.OnEat = nil
			}
		}
		for j := range item.Stored {
			if item.Stored[j].Item.Kind == model.FloorKindFood {
				item.Stored[j].Item.Food.OnEat = nil
			} else {
				item.Stored[j].Item.Utility.DoAction = nil
			}
		}
		for j := range item.OnSurface {
			item.OnSurface[j].Food.OnEat = nil
		}
	}
	for i := range char.CurrentHome.FloorItems {
		fi := &char.CurrentHome.FloorItems[i]
		if fi.Item.Kind == model.FloorKindFood {
			fi.Item.Food.OnEat = nil
		} else {
			fi.Item.Utility.DoAction = nil
		}
	}
	for i := range char.Experiences {
		char.Experiences[i].Experience.ApplyExperience = nil
	}
}

// postLoad restores function fields from the in-memory registries after decoding.
func postLoad(char *model.Character) {
	for i := range char.CurrentHome.RoomItems {
		item := &char.CurrentHome.RoomItems[i]
		if reg, ok := model.RoomItemRegistry[item.Item.Name]; ok {
			item.Item = reg
		}
		for j := range item.Utilities {
			if reg, ok := model.UtilityRegistry[item.Utilities[j].Item.Name]; ok {
				item.Utilities[j].Item = reg
			}
			for k := range item.Utilities[j].OnSurface {
				sf := &item.Utilities[j].OnSurface[k]
				if reg, ok := model.FoodRegistry[sf.Food.Name]; ok {
					sf.Food.OnEat = reg.OnEat
				}
			}
		}
		for j := range item.Stored {
			s := &item.Stored[j]
			if s.Item.Kind == model.FloorKindFood {
				if reg, ok := model.FoodRegistry[s.Item.Food.Name]; ok {
					s.Item.Food.OnEat = reg.OnEat
				}
			} else {
				if reg, ok := model.UtilityRegistry[s.Item.Utility.Name]; ok {
					s.Item.Utility = reg
				}
			}
		}
		for j := range item.OnSurface {
			sf := &item.OnSurface[j]
			if reg, ok := model.FoodRegistry[sf.Food.Name]; ok {
				sf.Food.OnEat = reg.OnEat
			}
		}
	}
	for i := range char.CurrentHome.FloorItems {
		fi := &char.CurrentHome.FloorItems[i]
		if fi.Item.Kind == model.FloorKindFood {
			if reg, ok := model.FoodRegistry[fi.Item.Food.Name]; ok {
				fi.Item.Food.OnEat = reg.OnEat
			}
		} else {
			if reg, ok := model.UtilityRegistry[fi.Item.Utility.Name]; ok {
				fi.Item.Utility = reg
			}
		}
	}
	for i := range char.Experiences {
		if reg, ok := model.ExperienceRegistry[char.Experiences[i].Experience.Name]; ok {
			char.Experiences[i].Experience = reg
		}
	}
}
