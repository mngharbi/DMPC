/*
	Interaction with user through CLI
*/

package startup

import (
	"bufio"
	"fmt"
	"os"
)

const (
	confirmSuffix string = " (y/n)"
)

var (
	confirmMapping map[string]bool = map[string]bool{
		"y":   true,
		"Y":   true,
		"yes": true,
		"YES": true,
		"n":   false,
		"N":   false,
		"no":  false,
		"NO":  false,
	}
)

/*
	General
*/
type InputHandler func(string) (interface{}, bool)

func cliRead(handler InputHandler) interface{} {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if ret, ok := handler(text); ok {
			return ret
		}
	}
	return nil
}

func cliWrite(str string) {
	fmt.Print(str)
}

type InputVerifier func(string) bool
type InputTransformer func(string) interface{}

func cliGet(text string, errorText string, verifier InputVerifier, transformer InputTransformer) interface{} {
	cliWrite(text + " ")
	return cliRead(func(s string) (interface{}, bool) {
		ok := verifier(s)
		if !ok {
			cliWrite(errorText)
			return nil, false
		} else {
			return transformer(s), true
		}
	})
}

func cliConfirm(question string) bool {
	fullText := question + confirmSuffix
	verifier := func(s string) bool {
		_, ok := confirmMapping[s]
		return ok
	}
	transformer := func(s string) interface{} {
		val, _ := confirmMapping[s]
		return val
	}
	return cliGet(fullText, fullText, verifier, transformer).(bool)
}

func cliGetFilePath(text string) string {
	verifier := func(s string) bool {
		file, err := os.Stat(s)
		if err != nil {
			return false
		}
		return file.Mode().IsRegular()
	}
	transformer := func(s string) interface{} { return s }
	return cliGet(text, "Unable to read file. Choose a different file: ", verifier, transformer).(string)
}

func cliGetString(text string) string {
	verifier := func(s string) bool { return true }
	transformer := func(s string) interface{} { return s }
	return cliGet(text, "", verifier, transformer).(string)
}
