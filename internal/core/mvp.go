package core

import (
	"basic-life-sim/internal/constant"
	"basic-life-sim/internal/model"
	"fmt"
)

const (
	mvpTitleMenu = `
BASIC LIFE SIMULATOR

1. New Game
2. Load Game
3. Exit`

	mvpMainMenu = `
=============================================
================= MAIN MENU =================
=============================================
Money: %.2f $
Needs:
	- Food       [%d/%d] %d%%
	- Energy     [%d/%d] %d%%
	- Hygiene    [%d/%d] %d%%
	- Confidence [%d/%d] %d%%
	- Strength   [%d/%d] %d%%

Actions:
1. Check Room
2. Do Activity
3. Buy Item
4. Save Game
5. Exit Game`
)

func needPct(need model.NeedStat) uint16 {
	if need.Max == 0 {
		return 0
	}
	return need.Current * 100 / need.Max
}

func printHomeLayout(h model.HomeType) {
	fmt.Println("===========================================")
	fmt.Println("Home Type:", h.Name)
	fmt.Println("Max Height:", h.MaxHeight)
	fmt.Println("Shared Bathroom:", h.SharedBathroom)
	fmt.Println("Shared Kitchen:", h.SharedKitchen)
	fmt.Printf("Upfront Cost: $%.2f\n", h.UpfrontCost)
	fmt.Printf("Monthly Cost: $%.2f\n", h.MonthlyCost)
	fmt.Println("Layout:")
	for _, row := range h.Layout {
		for _, cell := range row {
			switch cell {
			case model.HomeCellWall:
				fmt.Print("█")
			case model.HomeCellFloor:
				fmt.Print(" ")
			case model.HomeCellDoor:
				fmt.Print("D")
			default:
				fmt.Print("?")
			}
		}
		fmt.Println()
	}
	fmt.Println("===========================================")
}

func mainMenu(char *model.Character) {
	for {
		fmt.Printf(mvpMainMenu,
			char.CurrentStats.Money,
			char.Food.Current, char.Food.Max, needPct(char.Food),
			char.Energy.Current, char.Energy.Max, needPct(char.Energy),
			char.Hygiene.Current, char.Hygiene.Max, needPct(char.Hygiene),
			char.Confidence.Current, char.Confidence.Max, needPct(char.Confidence),
			char.Strength.Current, char.Strength.Max, needPct(char.Strength),
		)

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			checkRoomMenu(char)
		case 2:
			fmt.Println("Doing activity...")
			// Implement activity logic here
		case 3:
			buyItemMenu(char)
		case 4:
			fmt.Println("Saving game...")
			// Implement game saving logic here
		case 5:
			fmt.Println("Exiting game. Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func newGame() (model.Character, error) {
	char := model.Character{
		Age:         18,
		CurrentDate: model.NewCompactDate(2025, 1, 1),
		Food:        model.NeedStat{Current: 50, Max: 50},
		Energy:      model.NeedStat{Current: 50, Max: 50},
		Hygiene:     model.NeedStat{Current: 50, Max: 50},
		Confidence:  model.NeedStat{Current: 50, Max: 50},
		Strength:    model.NeedStat{Current: 50, Max: 50},
		CurrentStats: model.Stats{
			Money: 3000,
		},
	}

	fmt.Println("What is your name?")
	fmt.Scanln(&char.Name)

	fmt.Println("Are you a male or female (M/F)?")
	var gender string
	fmt.Scanln(&gender)
	switch gender {
	case "M", "m":
		char.IsMale = true
	case "F", "f":
		char.IsMale = false
	default:
		return char, fmt.Errorf("invalid gender choice: %q — expected M or F", gender)
	}

	fmt.Println("Choose your home:")
	for i, home := range constant.Level1HomeTypes {
		fmt.Printf("%d. %s\n", i+1, home.Name)
		printHomeLayout(home)
	}
	var chooseHome int
	fmt.Scanln(&chooseHome)
	if chooseHome < 1 || chooseHome > len(constant.Level1HomeTypes) {
		return char, fmt.Errorf("invalid home choice")
	}
	char.CurrentStats.Money -= constant.Level1HomeTypes[chooseHome-1].UpfrontCost
	char.CurrentHome = model.Home{
		Type: constant.Level1HomeTypes[chooseHome-1],
	}

	return char, nil
}

func RunMVP() {
	for {
		fmt.Println(mvpTitleMenu)
		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Println("Starting a new game...")
			char, err := newGame()
			if err != nil {
				fmt.Println("Error creating character:", err)
				return
			}
			mainMenu(&char)
			return
		case 2:
			fmt.Println("Loading game...")
			// Load game state here
			return
		case 3:
			fmt.Println("Exiting game. Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
