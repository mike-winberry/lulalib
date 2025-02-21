package message

import (
	"fmt"
)

func PromptForConfirmation(spinner *Spinner) bool {
	// Prompt the user to confirm the action
	if spinner != nil {
		spinnerText := spinner.Pause()
		defer spinner.Updatef("%s\n", spinnerText)
	}

	fmt.Println("Do you want to run executable validations? (y/n)")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	return response == "y" || response == "Y"
}
