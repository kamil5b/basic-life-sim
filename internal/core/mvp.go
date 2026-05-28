package core

import (
	"basic-life-sim/internal/constant"
	"basic-life-sim/internal/model"
	"fmt"
)

const (
	MVPTitleMenu = `
BASIC LIFE SIMULATOR

1. New Game
2. Load Game
3. Exit
`

	MVPMainMenu = `
=============================================
================= MAIN MENU =================
=============================================
Money: %.2f $
Needs:
	- Food       [%d/%d] %d
	- Energy     [%d/%d] %d
	- Hygiene    [%d/%d] %d
	- Confidence [%d/%d] %d
	- Strength   [%d/%d] %d

Actions:
1. Check Room
2. Do Activity
3. Buy Item
4. Save Game
5. Exit Game
`
)

func mainMenu(char model.Character) {
	fmt.Printf(MVPMainMenu,
		char.CurrentStats.Money,
		char.CurrentStats.Food, char.MaxFood, char.CurrentStats.Food*100/char.MaxFood,
		char.CurrentStats.Energy, char.MaxEnergy, char.CurrentStats.Energy*100/char.MaxEnergy,
		char.CurrentStats.Hygiene, char.MaxHygiene, char.CurrentStats.Hygiene*100/char.MaxHygiene,
		char.CurrentStats.Confidence, char.MaxConfidence, char.CurrentStats.Confidence*100/char.MaxConfidence,
		char.CurrentStats.Strength, char.MaxStrength, char.CurrentStats.Strength*100/char.MaxStrength,
	)
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		fmt.Println("Checking room...")
		// Implement room checking logic here
	case 2:
		fmt.Println("Doing activity...")
		// Implement activity logic here
	case 3:
		fmt.Println("Buying item...")
		// Implement item purchasing logic here
	case 4:
		fmt.Println("Saving game...")
		// Implement game saving logic here
	case 5:
		fmt.Println("Exiting game. Goodbye!")
		return
	default:
		fmt.Println("Invalid choice. Please try again.")
		mainMenu(char) // Restart the menu on invalid input
	}
}

func newGame() (model.Character, error) {
	// Initialize character with default values
	char := model.Character{
		Age:           18,
		MaxFood:       50,
		MaxEnergy:     50,
		MaxHygiene:    50,
		MaxConfidence: 50,
		MaxStrength:   50,
		CurrentStats: model.Stats{
			Money:      3000,
			Food:       50,
			Energy:     50,
			Hygiene:    50,
			Confidence: 50,
			Strength:   50,
		},
	}

	fmt.Println("What is your name?")
	fmt.Scanln(&char.Name)

	fmt.Println("Are you a male or female (M/F)?")
	var gender string
	fmt.Scanln(&gender)
	if gender == "M" || gender == "m" {
		char.IsMale = true
	}

	fmt.Println("Choose your home:")
	for i, home := range constant.Level1HomeTypes {
		fmt.Printf("%d. %s\n", i+1, home.Name)
		home.PrintLayout()
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
	var char model.Character
	var err error
	fmt.Println(MVPTitleMenu)
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		fmt.Println("Starting a new game...")
		char, err = newGame()
		if err != nil {
			fmt.Println("Error creating character:", err)
			return
		}
		mainMenu(char)
		// Initialize game state here
	case 2:
		fmt.Println("Loading game...")
		// Load game state here
	case 3:
		fmt.Println("Exiting game. Goodbye!")
		return
	default:
		fmt.Println("Invalid choice. Please try again.")
		RunMVP() // Restart the menu on invalid input
	}
}
