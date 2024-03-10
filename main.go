package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	telegramAPIURL = "https://api.telegram.org/bot"
)

var (
	telegramBotToken string
	configFilePath   string
	configuration    Config
	pollingRate      time.Duration // TODO: Define conditions for this to be changed

)

// ********//
// CONFIG //
// ********//
type Config struct {
	AdminId      int                       `json:"adminTelegramId"`
	Whitelist    map[int]User              `json:"whitelist"` //TODO: Understand if "0" will be actually deserialized instead of 0
	Localization map[int]map[string]Status `json:"localization"`
	Hub          map[string]Service        `json:"hub"`
}

type User struct {
	Username string `json:"username"`
	Locale   string `json:"locale"`
}

type Status struct {
	Text string `json:"text"`
}

type Service struct {
	Path string `json:"path"`
}

// **********//
// TELEGRAM //
// **********//
type Update struct {
	UpdateID      int           `json:"update_id"`
	Message       Message       `json:"message"`
	CallbackQuery CallbackQuery `json:"callback_query,omitempty"`
}

type CallbackQuery struct {
	Data    string  `json:"data"`
	Message Message `json:"message"`
}

type Message struct {
	MessageId int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type Chat struct {
	Id        int     `json:"id"`
	Username  *string `json:"username,omitempty"`
	Firstname *string `json:"first_name,omitempty"`
	Lastname  *string `json:"last_name,omitempty"`
}

// ********//
//
//	main  //
//
// ********//
func main() {
	telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramBotToken == "" {
		fmt.Println("Error: TELEGRAM_BOT_TOKEN must be set")
		os.Exit(1)
	}
	configFilePath = os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		fmt.Println("Error: CONFIG_FILE_PATH must be set")
		os.Exit(1)
	}

	loadConfigFile()

	offset := 0
	pollingRate = 5 * time.Second
	fmt.Println("Starting polling...")
	for {
		updates, err := getUpdates(offset)
		if err != nil {
			fmt.Println("Error getting updates:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			go processUpdate(update)
			offset = update.UpdateID + 1
		}

		time.Sleep(pollingRate)
	}
}

func getUpdates(offset int) ([]Update, error) {
	resp, err := http.Get(telegramAPIURL + telegramBotToken + "/getUpdates?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func processUpdate(update Update) {
	fmt.Println("Found an update from telegram...")

	var chatId int
	if update.CallbackQuery.Data != "" {
		chatId = update.CallbackQuery.Message.Chat.Id
	} else {
		chatId = update.Message.Chat.Id
	}

	_, isAuthorizedUser := configuration.Whitelist[chatId]
	if !isAuthorizedUser {
		fmt.Printf("Unauthorized telegram id %d tried to acccess the bot\n", chatId)
		return
	}

	if update.CallbackQuery.Data != "" {
		handleCallbackQuery(update.CallbackQuery.Data, chatId)
		return
	}

	handleInput(update.Message.Text, chatId)
}

func handleInput(input string, chatId int) {
	if strings.HasPrefix(input, "/") {
		handleCommand(strings.ToLower(input[1:]), chatId)
	} else {
		// TODO: Handle text not starting with "/" maybe by saying that the bot doesn't want to chat
	}
}

func handleCommand(input string, chatId int) {
	switch input {
	case "start":
		err := sendLanguageSelectionButtons(chatId, "ㅤㅤ( ﾉ ﾟｰﾟ)ﾉ")
		if err != nil {
			fmt.Println("Error sending message: ", err)
		}
		return
	case "list":
		// TODO: Handle /list command and remember to write that if the bot got rebooted in the day, it could be inconsistent
		return
	case "help":
		// TODO: Handle /help command
		return
	case "stop":
		if chatId == configuration.AdminId {
            handleService(input, chatId, "stop.sh")
		} else {
		// TODO: Handle unauthorized user
		}
		return
	default:
		handleService(input, chatId, "start.sh")
		return
	}
}

func handleService(input string, chatId int, command string) {
	service, supported := configuration.Hub[input]
	if !supported {
		// TODO: Handle not supported command and write about the /help command
	} else {
		err := commandService(service, command)
		if err != nil {
			// TODO: Handle error response
			return
		}
		// TODO: Handle success response
	}
}

func commandService(service Service, scriptName string) error {
	cmd := exec.Command("sudo", "-u", "root", service.Path+scriptName)
	return cmd.Run()
}

func loadConfigFile() {
	file, _ := os.ReadFile(configFilePath)

	json.Unmarshal(file, &configuration)
}

func sendMessage(chatID int, text string) error {
	url := fmt.Sprintf("%s%s/sendMessage?chat_id=%d&text=%s", telegramAPIURL, telegramBotToken, chatID, text)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func handleCallbackQuery(locale string, chatId int) {
	switch locale {
	case "it":
		updateLocale("it", chatId)
	default:
		updateLocale("en", chatId)
	}
}

func updateLocale(locale string, chatId int) {
	if userEntry, ok := configuration.Whitelist[chatId]; ok {
		userEntry.Locale = locale
		configuration.Whitelist[chatId] = userEntry
	}

	writeToConfigFile()
}

func writeToConfigFile() error {
	updatedJSON, err := json.MarshalIndent(configuration, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, updatedJSON, 0644)
}

func sendLanguageSelectionButtons(chatID int, text string) error {
	keyboard := struct {
		InlineKeyboard [][]struct {
			Text         string `json:"text"`
			CallbackData string `json:"callback_data"`
		} `json:"inline_keyboard"`
	}{
		InlineKeyboard: [][]struct {
			Text         string `json:"text"`
			CallbackData string `json:"callback_data"`
		}{
			{
				{Text: "🇮🇹", CallbackData: "it"},
				{Text: "🇬🇧", CallbackData: "en"},
			},
		},
	}

	keyboardJSON, err := json.Marshal(keyboard)
	if err != nil {
		return err
	}

	formData := fmt.Sprintf("chat_id=%d&text=%s&reply_markup=%s", chatID, text, keyboardJSON)
	contentType := "application/x-www-form-urlencoded"

	url := fmt.Sprintf("%s%s/sendMessage", telegramAPIURL, telegramBotToken)
	resp, err := http.Post(url, contentType, strings.NewReader(formData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
