package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
    telegramAPIURL = "https://api.telegram.org/bot"
)

var (
    configuration Config
    telegramBotToken string
    configFilePath string
    whitelistMap map[int]User
    pollingRate time.Duration
)

//********//
// CONFIG //
//********//
type Config struct {
    adminId int

    // TODO: Define the config file structure
}

//***********//
// WHITELIST //
//***********//
type User struct {
	Username string
    Locale string
}

//**********//
// TELEGRAM // // TODO: Redefine all the structs from https://core.telegram.org/bots/api
//**********//
type Update struct {
    UpdateID int                `json:"update_id"`
    Message  TelegramMessage    `json:"message"`
    CallbackQuery CallbackQuery `json:"callback_query,omitempty"`
}

type CallbackQuery struct {
    ID      string          `json:"id"`
    Data    string          `json:"data"`
    Message TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageId int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type Chat struct {
	Id int            `json:"id"`
    Username *string  `json:"username,omitempty"`
    Firstname *string `json:"first_name,omitempty"`
    Lastname *string  `json:"last_name,omitempty"`
}

//********//
//  main  //
//********//
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

    //whitelistMap = loadConfigFile()

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
        chatId =  update.Message.Chat.Id
    }

    _, isAuthorizedUser := whitelistMap[chatId]
    if !isAuthorizedUser {
        fmt.Printf("Unauthorized telegram id %d tried to acccess the bot\n", chatId)
        return
    }

    if update.CallbackQuery.Data != "" {
        handleCallbackQuery(update.CallbackQuery.Data, chatId)
        return
    }

    //input := update.Message.Text

    //TODO: Handle input
}

func loadConfigFile() {
    //file, _ := os.ReadFile(configFilePath)
    //TODO: Handle config file parsing knowing it will be a pretty json
    loadWhitelist()
}

func loadWhitelist() {
    //TODO: Handle whitelist loading
    //whitelistMap =
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
    //TODO: update config file
    
    writeToConfigFile()
}

func writeToConfigFile() error {
    updatedJSON, err := json.MarshalIndent(configuration, "", "    ") // TODO: Check that this makes pretty JSONs
    if err != nil {
        return err
    }

    return os.WriteFile(configFilePath, updatedJSON, 0644)
}
