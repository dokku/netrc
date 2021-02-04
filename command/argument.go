package command

import (
	"fmt"
	"strconv"
	"strings"
)

type Argument struct {
	Name     string
	Optional bool
	Type     ArgumentType
	Value    interface{}
	HasValue bool
}

// ArgumentType is an enum to define what arguments are present
type ArgumentType uint

const (
	ArgumentString ArgumentType = 0
	ArgumentInt    ArgumentType = 1 << iota
	ArgumentBool   ArgumentType = 2 << iota
	ArgumentList   ArgumentType = 3 << iota
)

func (a Argument) BoolValue() bool {
	if a.Type != ArgumentBool {
		panic(fmt.Errorf("Unexpected argument type for %s when calling BoolValue()", a.Name))
	}

	return a.Value.(bool)
}

func (a Argument) IntValue() int {
	if a.Type != ArgumentInt {
		panic(fmt.Errorf("Unexpected argument type for %s when calling IntValue()", a.Name))
	}

	return a.Value.(int)
}

func (a Argument) StringValue() string {
	if a.Type != ArgumentString {
		panic(fmt.Errorf("Unexpected argument type for %s when calling StringValue()", a.Name))
	}

	return a.Value.(string)
}

func (a Argument) ListValue() []string {
	if a.Type != ArgumentList {
		panic(fmt.Errorf("Unexpected argument type for %s when calling ListValue()", a.Name))
	}

	return a.Value.([]string)
}

func ArgumentAsString(arguments []Argument) string {
	argumentString := []string{}

	for _, argument := range arguments {
		if argument.Optional {
			argumentString = append(argumentString, fmt.Sprintf("[%s]", argument.Name))
		} else {
			argumentString = append(argumentString, fmt.Sprintf("<%s>", argument.Name))
		}
	}

	return strings.Join(argumentString, " ")
}

func ParseArguments(args []string, arguments []Argument) (map[string]Argument, error) {
	returnArguments := map[string]Argument{}
	if err := validateArguments(arguments); err != nil {
		return returnArguments, err
	}

	maxArgs := len(arguments)
	minArgs := 0
	for _, argument := range arguments {
		if !argument.Optional {
			minArgs++
		}
	}

	argumentWord := "argument"
	if maxArgs != 1 {
		argumentWord = "arguments"
	}
	errorMessage := fmt.Sprintf("This command requires %d", minArgs)
	if minArgs != maxArgs {
		errorMessage = fmt.Sprintf("%s and at most %d %s", errorMessage, maxArgs, argumentWord)
	} else {
		errorMessage = fmt.Sprintf("%s %s", errorMessage, argumentWord)
	}

	errorMessage = fmt.Sprintf("%s: %s", errorMessage, ArgumentAsString(arguments))

	if len(args) == 0 {
		if len(arguments) == 0 {
			return returnArguments, nil
		}

		if !arguments[0].Optional {
			return returnArguments, fmt.Errorf(errorMessage)
		}
	}

	if len(args) < minArgs {
		return returnArguments, fmt.Errorf(errorMessage)
	}

	hasListArgument := false
	listIndex := 0
	for i, value := range args {
		if hasListArgument {
			arguments[listIndex].HasValue = true
			arguments[listIndex].Value = append(arguments[listIndex].Value.([]string), value)
		} else {
			arguments[i].HasValue = true
			if arguments[i].Type == ArgumentList {
				hasListArgument = true
				listIndex = i
				arguments[i].Value = []string{value}
			} else {
				if arguments[i].Type == ArgumentInt {
					intValue, err := strconv.Atoi(value)
					if err != nil {
						return returnArguments, fmt.Errorf("Invalid value for argument %s", arguments[i].Name)
					}
					arguments[i].Value = intValue
				} else {
					arguments[i].Value = value
				}
			}
		}
	}

	for _, argument := range arguments {
		if argument.Value == nil {
			if argument.Type == ArgumentBool {
				argument.Value = false
			} else if argument.Type == ArgumentInt {
				argument.Value = 0
			} else if argument.Type == ArgumentList {
				argument.Value = []string{}
			} else if argument.Type == ArgumentString {
				argument.Value = ""
			}
			argument.HasValue = false
		}
		returnArguments[argument.Name] = argument
	}

	return returnArguments, nil
}

func validateArguments(arguments []Argument) error {
	reachedOptional := false
	reachedList := false
	listArgument := ""
	for _, arg := range arguments {
		if reachedOptional {
			if !arg.Optional {
				return fmt.Errorf("Argument %s must be placed before all optional arguments", arg.Name)
			}
		} else if arg.Optional {
			reachedOptional = true
		}

		if reachedList {
			return fmt.Errorf("List Argument %s must be placed after all other arguments", listArgument)
		} else if arg.Type == ArgumentList {
			listArgument = arg.Name
			reachedList = true
		}
	}
	return nil
}
