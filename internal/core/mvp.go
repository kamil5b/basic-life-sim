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
		Money: %d $
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
		fmt.Println(char)
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
