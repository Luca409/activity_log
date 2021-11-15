package chatter

import (
	"activity_log/api/apperror"
	"activity_log/api/constants"
	"activity_log/api/constructs"
	"activity_log/internal/dao"
	"activity_log/internal/user_input"
	"activity_log/internal/user_output"
	"activity_log/internal/util"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MaxLastRecordMinutesDefault = time.Hour

type ChatterConfig struct {
	ResponseWait        time.Duration
	MaxConfusionRetries int
}

type Chatter struct {
	userListener  *user_input.UserListener
	userMessenger *user_output.UserMessenger
	userSchemaDAO dao.UserSchemaDAO
	userDataDAO   dao.UserDataDAO

	chatterConfig  *ChatterConfig
	lastRecordTime time.Time
}

func NewChatter(
	userListener *user_input.UserListener,
	userMessenger *user_output.UserMessenger,
	userSchemaDAO dao.UserSchemaDAO,
	userDataDAO dao.UserDataDAO,
	chatterConfig *ChatterConfig,
) *Chatter {
	return &Chatter{
		userListener:  userListener,
		userMessenger: userMessenger,
		userSchemaDAO: userSchemaDAO,
		userDataDAO:   userDataDAO,
		chatterConfig: chatterConfig,

		lastRecordTime: time.Now(),
	}
}

func (ctr *Chatter) Run() {
	go ctr.setUpReminders()

	time.Sleep(time.Second * time.Duration(3))

	for {
		if err := ctr.round(); err != nil {
			log.Fatalf(err.Error())
		}
	}
}

func (ctr *Chatter) round() error {

	userSchema, err := ctr.getUserSchema()
	if err != nil {
		return fmt.Errorf("failed to get user schema: %w", err)
	}

	existingSchema := userSchema.Schema.ToRegularMap()

	if err := ctr.writeRound([]string{}, userSchema.Schema); err != nil {
		if err := ctr.userMessenger.Send(fmt.Sprintf("ERROR: %v", err)); err != nil {
			return fmt.Errorf("couldn't log error to user. err: %w", err)
		}
	}

	if err := util.NestedMapsEqual(existingSchema, userSchema.Schema.ToRegularMap()); err != nil {
		if err := ctr.userSchemaDAO.Dump(userSchema, true); err != nil {
			return fmt.Errorf("couldn't record schema change, err: %w", err)
		}
	}

	return nil
}

func (ctr *Chatter) writeRound(path []string, expandingMap *util.ExpandingMap) error {
	subExpandingMap, err := expandingMap.GetSubMap(path)
	if err != nil {
		return fmt.Errorf("GetSubMap(%v) returns err: %w", path, err)
	}

	userInput, options, err := ctr.getOptionOrText(path, subExpandingMap)
	if err != nil {
		return fmt.Errorf("getOptionOrText() returns err: %w", err)
	}

	if choiceDigit, err := strconv.Atoi(userInput.Text); err == nil {
		// Get input.
		path = append(path, options[choiceDigit])

		if subMap, err := expandingMap.GetSubMap(path); err == nil {
			if !subMap.IsEmpty() {
				return ctr.writeRound(path, expandingMap)
			}
		} else {
			return fmt.Errorf("no submap at this choice -- implementation error")
		}

		if err := ctr.userMessenger.Send(fmt.Sprintf("%s -- how many minutes did you do this for?", options[choiceDigit])); err != nil {
			return fmt.Errorf("userMessenger.Send() returns err: %w", err)
		}

		userInput, err = ctr.userListener.GetUserInput(
			ctr.chatterConfig.ResponseWait,
			ctr.chatterConfig.MaxConfusionRetries,
			func(ui *constructs.UserInput) error {
				return nil
			},
		)
		if err != nil {
			return fmt.Errorf("GetUserInput() returns err: %w", err)
		}

		// Record or add option.
		if userInput.Text == "" {
			if time.Since(ctr.lastRecordTime) > MaxLastRecordMinutesDefault {
				return fmt.Errorf("time since last record is greater than %v, please specify minutes", MaxLastRecordMinutesDefault)
			}
			userInput.Text = fmt.Sprintf("%d", int(time.Since(ctr.lastRecordTime).Minutes()))
		}

		if digit, err := strconv.Atoi(userInput.Text); err == nil {
			if err := ctr.recordValue(path, digit); err != nil {
				return fmt.Errorf("recordValue() returns err: %v", err)
			}
			return nil
		} else {
			if choiceDigit == 0 {
				return fmt.Errorf("cannot expand first option")
			}

			if err := expandingMap.AddSubMapIncludingParent(path, userInput.Text); err != nil {
				return fmt.Errorf("AddSubMapIncludingParent(%v, %s) returns err: %w", path, userInput.Text, err)
			}

			return ctr.writeRound(path, expandingMap)
		}
	} else {
		// Add option.
		if _, ok := subExpandingMap.ToRegularMap()[userInput.Text]; ok {
			return fmt.Errorf("option already exists")
		}

		if err := expandingMap.AddSubMap(path, userInput.Text); err != nil {
			return fmt.Errorf("AddSubMap(%v, %s) returns err: %w", path, userInput.Text, err)
		}
		return ctr.writeRound(path, expandingMap)
	}
}

func (ctr *Chatter) getOptionOrText(path []string, expandingMap *util.ExpandingMap) (*constructs.UserInput, map[int]string, error) {
	firstVal := constants.DEFAULT_FIRST_OPTION

	if len(path) != 0 {
		firstVal = path[len(path)-1]
	}

	options := mapToOptions(expandingMap.ToRegularMap())
	optionsKeys := []int{}
	for k := range options {
		optionsKeys = append(optionsKeys, k)
	}
	sort.Slice(optionsKeys, func(i, j int) bool { return optionsKeys[i] < optionsKeys[j] })

	// Swap default value to beginning.
	currentFirst := options[optionsKeys[0]]
	for idx, opt := range optionsKeys {
		if options[opt] == firstVal {
			options[optionsKeys[0]] = firstVal
			options[optionsKeys[idx]] = currentFirst
		}
	}

	userQuery := ""
	for _, key := range optionsKeys {
		userQuery += fmt.Sprintf("%d .) %s\n", key, options[key])
	}

	if err := ctr.userMessenger.Send(userQuery); err != nil {
		return nil, nil, fmt.Errorf("userMessenger.Send() returns err: %w", err)
	}

	if err := ctr.userMessenger.Send("Choose an option from the list above, or type something new to add it."); err != nil {
		return nil, nil, fmt.Errorf("userMessenger.Send() returns err: %w", err)
	}

	rangeMin := 0
	rangeMax := len(options) - 1

	userInput, err := ctr.userListener.GetUserInput(
		ctr.chatterConfig.ResponseWait,
		ctr.chatterConfig.MaxConfusionRetries,
		func(ui *constructs.UserInput) error {
			if digit, err := strconv.Atoi(ui.Text); err == nil {
				if digit < int(rangeMin) || digit > int(rangeMax) {
					return fmt.Errorf("input not in range [%d, %d]", rangeMin, rangeMax)
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	if userInput.Text == "" {
		userInput.Text = "0"
	}

	return userInput, options, nil
}

func (ctr *Chatter) recordValue(path []string, value int) error {
	userData := &constructs.UserData{
		Data: map[string]interface{}{
			string(constructs.Activity):     strings.Join(path, "."),
			string(constructs.MinutesSpent): value,
		},
		TimestampMS: time.Now().UnixNano() / int64(time.Millisecond),
	}

	if err := ctr.userDataDAO.Append(userData); err != nil {
		return fmt.Errorf("userDataDAO.Append() returns err: %w", err)
	}

	ctr.lastRecordTime = time.Now()

	return nil
}

func (ctr *Chatter) getUserSchema() (*constructs.UserSchema, error) {
	us, err := ctr.userSchemaDAO.Load()
	if err != nil {
		if apperror.IsNotFoundError(err) {
			if err := ctr.userMessenger.Send(err.Error()); err != nil {
				return nil, fmt.Errorf("userMessenger.Send returns err: %w", err)
			}

			if err := ctr.userMessenger.Send("Would you like to create a new schema?"); err != nil {
				return nil, fmt.Errorf("userMessenger.Send returns err: %w", err)
			}

			if _, err := ctr.userListener.GetUserInput(
				ctr.chatterConfig.ResponseWait,
				0,
				func(ui *constructs.UserInput) error {
					if strings.ToUpper(ui.Text) != "YES" {
						return fmt.Errorf("well then you'll need to create it yourself then")
					}
					return nil
				},
			); err != nil {
				return nil, err
			}

			us, err = ctr.userSchemaDAO.Init()
			if err != nil {
				return nil, fmt.Errorf("userSChemaDAO.Init() returns err: %w", err)
			}
		} else {
			return nil, fmt.Errorf("userSchemaDAO.Load() returns err: %w", err)
		}
	}

	return us, nil
}

func mapToOptions(subSchema map[string]interface{}) map[int]string {
	options := map[int]string{}
	counter := 0
	for _, key := range util.SortedMapKeysAsc(subSchema) {
		options[counter] = key
		counter++
	}
	return options
}

func (ctr *Chatter) setUpReminders() error {
	if err := ctr.userMessenger.Send("How often, in minutes, would you like to be reminded to record activity? 0 for never."); err != nil {
		log.Fatalf("userMessenger.Send() returns err: %v", err)
	}

	userInput, err := ctr.userListener.GetUserInput(
		ctr.chatterConfig.ResponseWait,
		ctr.chatterConfig.MaxConfusionRetries,
		func(ui *constructs.UserInput) error {
			if _, err := strconv.Atoi(ui.Text); err != nil {
				return fmt.Errorf("input must be a digit")
			}
			return nil
		},
	)

	if err != nil {
		log.Fatalf("user input error: %v", err)
	}

	userDigit, err := strconv.Atoi(userInput.Text)
	if err != nil {
		log.Fatalf("user input error: %v", err)
	}

	if userDigit == 0 {
		return nil
	}

	triggerReminderLoop(int64(userDigit))

	return nil
}

func triggerReminderLoop(minutesToWait int64) {

	for {
		time.Sleep(time.Minute * time.Duration(minutesToWait))

		cmd := exec.Command("/bin/sh", "internal/chatter/reminder.bash")
		if err := cmd.Run(); err != nil {
			// HACK
			if strings.Contains(err.Error(), "already started") {
				continue
			}

			log.Fatalf("Failed to execute reminder command: %v", err)
		}
	}
}
