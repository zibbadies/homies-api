package checks;

import (
	"fmt"
	"regexp"

	"github.com/zibbadies/homies/internal/homies/models"
)

var CheckersData map[string]models.Checker;

var colorRegex = regexp.MustCompile(`^([A-Fa-f0-9]{6})$`)

var rfc3339Regex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)

func init() {
	fmt.Println("Initializing checks...")
	CheckersData = make(map[string]models.Checker);

	// Password
	CheckersData["password"] = models.Checker{
		Check: models.Check{
			FriendlyName: "password",
			MinLength: 8,
			MaxLength: 100,
		},
		Checker: BasicStringCheck("password"),
	}

	// Username
	CheckersData["username"] = models.Checker{
		Check: models.Check{
			FriendlyName: "username",
			MinLength: 3,
			MaxLength: 12,
		},
		Checker: BasicStringCheck("username"),
	}

	// Home name
	CheckersData["house_name"] = models.Checker{
		Check: models.Check{
			FriendlyName: "home name",
			MinLength: 4,
			MaxLength: 25,
		},
		Checker: BasicStringCheck("house_name"),
	}

	// List Item text
	CheckersData["list_item_text"] = models.Checker{
		Check: models.Check{
			FriendlyName: "item",
			MinLength: 3,
			MaxLength: 512,
		},
		Checker: BasicStringCheck("list_item_text"),
	}

	// List Item due date (RFC3339)
	CheckersData["list_item_due"] = models.Checker{
		Check: models.Check{
			FriendlyName: "due date",
			MinLength: 20,
			MaxLength: 35,
		},
		Checker: BasicRegexCheck("list_item_due", rfc3339Regex),
	}

	// Avatar Color
	CheckersData["color"] = models.Checker{
		Check: models.Check{
			FriendlyName: "color",
			MinLength: 6,
			MaxLength: 6,
		},
		Checker: BasicRegexCheck("color", colorRegex),
	}

	// Avatar bezier
	CheckersData["bezier"] = models.Checker{
		Check: models.Check{
			FriendlyName: "bezier",
			MinLength: 3,
			MaxLength: 20,
		},
		Checker: BezierCheck,
	}
}

func Check(key string, value string) error {
	return CheckersData[key].Checker(value)
}

func BasicStringCheck(key string) func(string) error {
	return func(value string) error {
		checker, ok := CheckersData[key]
		if (!ok) {
			return &models.CheckError{
				Message: fmt.Sprintf("Internal error: Checker %s not found!", key),
				ErrorCode: models.InternalError,
			}
		}

		if (len(value) < checker.MinLength) {
			return &models.CheckError{
				Message: fmt.Sprintf("Your %s must be at least %d characters long!", checker.FriendlyName, checker.MinLength),
				ErrorCode: models.BasicCheckError,
			}
		} else if (len(value) >= checker.MaxLength) {
			return &models.CheckError{
				Message: fmt.Sprintf("Your %s is too long! max %d characters.", checker.FriendlyName, checker.MaxLength),
				ErrorCode: models.BasicCheckError,
			}
		}

		return nil;
	}
}

func BasicRegexCheck(key string, regex *regexp.Regexp) func(string) error {
	return func(value string) error {
		checker, ok := CheckersData[key]
		if (!ok) {
			return &models.CheckError{
				Message: fmt.Sprintf("Internal error: Checker %s not found!", key),
				ErrorCode: models.InternalError,
			}
		}

		if (!regex.MatchString(value)) {
			return &models.CheckError{
				Message: fmt.Sprintf("Your %s is not valid!", checker.FriendlyName),
				ErrorCode: models.BasicCheckError,
			}
		}

		return nil;
	}
}
