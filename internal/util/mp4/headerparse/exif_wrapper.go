package headerparse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"seneca/api/senecaerror"
	"strings"
)

const exifCommand = "exiftool -ee -g3 -j %s"

func runExifCommand(filePath string) (map[string]interface{}, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, senecaerror.NewBadStateError(fmt.Errorf("%q does not exist", filePath))
	}

	commandString := fmt.Sprintf(exifCommand, filePath)
	commandStringParts := strings.Split(commandString, " ")
	cmd := exec.Command(commandStringParts[0], commandStringParts[1:]...)

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error fetching cmd.StdoutPipe()")
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdOut)
	scanner.Split(bufio.ScanWords)

	output := ""
	for scanner.Scan() {
		// Add a space between words.
		out := scanner.Text() + " "
		output = output + out
	}

	cmd.Wait()

	var topLevelArray []interface{}
	if err := json.NewDecoder(strings.NewReader(output)).Decode(&topLevelArray); err != nil {
		return nil, fmt.Errorf("error parsing json: %w, exiftool may not be installed", err)
	}

	if len(topLevelArray) != 1 {
		return nil, fmt.Errorf("expect length of 1 for top level array, got %d", len(topLevelArray))
	}

	topMap, ok := topLevelArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expect map[string]interface{} for topLevelArray[0], got %T", topLevelArray[0])
	}

	return topMap, nil
}
