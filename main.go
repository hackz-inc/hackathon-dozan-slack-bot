package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/api/option"
)

func main() {
	// ======== ここからが諸岡コード！！！！！！！！！ ==========
	// ctxを再利用する為下記のように書きます。
	ctx := context.Background()
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Println("接続エラー。")
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		fmt.Println("error getting Auth client: \n", err)
	}

	// ======== ここまでが諸岡コード！！！！！！！！！ ==========
	
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

					snapshot, firestoreErr := client.Collection("users").Doc(userInfo.ID).Get(ctx)

					if firestoreErr != nil {
						fmt.Println("Failed adding alovelace:", firestoreErr)
					}

					if snapshot.Data() != nil {
						getUserFromFirestore := snapshot.Data()
						fmt.Println(getUserFromFirestore)
						fmt.Println(getUserFromFirestore["technology"])
						// if res, state := getUserFromFirestore["technology"].([]string); state {
						// 	fmt.Println("test", state)
						// 	fmt.Println("test", res)
						// }

						// fmt.Println("test", getUserFromFirestore)

						res, state := getUserFromFirestore["technology"].([]string)
							fmt.Println("test", state)
							fmt.Println("test", res)
						// fmt.Println("test", state)
						// fmt.Println("test", res)

						// var initialTechnology []string
						// technology := append(res, reaction)
						
						// fmt.Println("test", technology)
					}

					var initialTechnology []string
					technology := append(initialTechnology, reaction)

					// type User struct {
					// 	id          string
					// 	userName    string
					// 	email       string
					// 	technology []string
					// }

					// user := User {
					// 	userInfo.ID,
					// 	userInfo.Profile.RealName,
					// 	userInfo.Profile.Email,
					// 	technology,
					// }

					// データ追加
					_, err = client.Collection("users").Doc(userInfo.ID).Set(ctx, map[string]interface{} {
						"id":  userInfo.ID,
						"userName": userInfo.Profile.RealName,
						"email": userInfo.Profile.Email,
						"technology": technology,
					})

					if(err != nil) {
						fmt.Println("Failed adding alovelace:", err)
					}

					fmt.Println("success")
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
