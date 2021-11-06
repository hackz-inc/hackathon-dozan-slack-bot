package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	godotenv.Load(".env")

	// SlackClientの構築

	// ======== .envファイルにSLACK_TOKEN追加してね！！！！！！！！！ ==========
	api := slack.New(os.Getenv("SLACK_TOKEN"))

	// ルートにアクセスがあった時の処理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// リクエスト内容を取得
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// イベント内容を取得
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// イベント内容によって処理を分岐
		switch eventsAPIEvent.Type {
		case slackevents.URLVerification: // URL検証の場合の処理
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case slackevents.CallbackEvent: // コールバックイベントの場合の処理
			innerEvent := eventsAPIEvent.InnerEvent

			// イベントタイプで分岐
			switch event := innerEvent.Data.(type) {
			case *slackevents.ReactionAddedEvent:
				reaction := event.Reaction // スタンプ名
				user := event.User         // ユーザーID

				userInfo, err := api.GetUserInfo(user) // ユーザー情報の取得

				if err != nil {
					log.Println(err)
				}

				fmt.Printf("ID: %s, Fullname: %s, Email: %s, Reaction: %s\n", userInfo.ID, userInfo.Profile.RealName, userInfo.Profile.Email, reaction)
			case *slackevents.AppMentionEvent: // メンションイベント

				// スペースを区切り文字として配列に格納
				message := strings.Split(event.Text, " ")
				fmt.Printf("message %v\n", message)
				// テキストが送信されていない場合は終了
				if len(message) < 2 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// 送信されたテキストを取得
				command := message[1]

				if err != nil {
					log.Println(err)
				}

				// 送信元のユーザIDを取得
				user := event.User

				switch command {
				case "hello": // helloが送られた場合
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("<@"+user+"> world", false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
