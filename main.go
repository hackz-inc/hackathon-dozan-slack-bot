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

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/api/option"
)

func main() {
	var SlackReactionTechs = []string{
	"html5",
	"css3",
	"sass",
	"javascript",
	"typescript",
	"vue",
	"nuxt",
	"react",
	"nextjs",
	"node",
	"ruby_on_rails",
	"java",
	"php",
	"laravel",
	"python",
	"django",
	"c_sharp",
	"unity",
	"swift",
	"flutter",
	"aws",
	"azure",
	"gcp",
	"docker",
	"firebase",
}

	// ctxを再利用する為下記のように書きます。
	ctx := context.Background()
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)

	if err != nil {
		fmt.Println("接続エラー。")
	}

	client, firestoreErr := app.Firestore(ctx)
	if err != nil {
		fmt.Println("error getting Auth client: \n", firestoreErr)
	}

	// SlackClientの構築
	// ======== .envファイルにSLACK_TOKENと、BOT_IDを追加してね！！！！！！！！！ ==========
	godotenv.Load(".env")
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	botId := os.Getenv("BOT_ID")

	// イベントがあった時の処理
	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
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

				log.Println("ahiahi1")
				// イベントタイプで分岐
				switch event := innerEvent.Data.(type) {
					case *slackevents.ReactionAddedEvent:
						reaction := event.Reaction // スタンプ名
						user := event.User         // ユーザーID
						itemUser := event.ItemUser // 投稿者

						if botId != itemUser {
							log.Printf("Botに対してのスタンプじゃあねぇなあ！")
							return
						}

						userInfo, err := api.GetUserInfo(user) // ユーザー情報の取得
						if err != nil {
							log.Println(err)
						}

						snapshot, firestoreErr := client.Collection("users").Doc(userInfo.ID).Get(ctx)
						if firestoreErr != nil {
							fmt.Println("Failed adding alovelace:", firestoreErr)
						}

						// 既にユーザーデータがある場合は、スタンプだけをアップデート
						if snapshot.Data() != nil {
							getUserFromFirestore := snapshot.Data()

							if res, state := getUserFromFirestore["technology"].([]interface{}); state {
								for _, v := range res {
									if v == reaction {
										fmt.Println("スタンプが重複しているよ")
										return
									}
								}

								technology := append(res, reaction)

								client.Collection("users").Doc(userInfo.ID).Set(ctx, map[string]interface{} {
									"technology": technology,
								}, firestore.MergeAll)
							}

							fmt.Println("データあるルート")
							return
						}

						var initialTechnology []string
						technology := append(initialTechnology, reaction)
						fmt.Println("データなしルート")

						// データ追加
						_, err = client.Collection("users").Doc(userInfo.ID).Set(ctx, map[string]interface{} {
							"id":  userInfo.ID,
							"userName": userInfo.Profile.RealName,
							"email": userInfo.Profile.Email,
							"technology": technology,
						})

						if err != nil {
							fmt.Println("Failed adding alovelace:", err)
						}

						fmt.Printf("ID: %s, Fullname: %s, Email: %s, Reaction: %s\n", userInfo.ID, userInfo.Profile.RealName, userInfo.Profile.Email, reaction)

					case *slackevents.ReactionRemovedEvent:
						reaction := event.Reaction // スタンプ名
						user := event.User         // ユーザーID
						itemUser := event.ItemUser // 投稿者

						if botId != itemUser {
							log.Printf("Botに対してのスタンプじゃあねぇなあ！")
							return
						}

						userInfo, err := api.GetUserInfo(user) // ユーザー情報の取得

						if err != nil {
							log.Println(err)
						}

						snapshot, firestoreErr := client.Collection("users").Doc(userInfo.ID).Get(ctx)

						if firestoreErr != nil {
							fmt.Println("Failed adding alovelace:", firestoreErr)
						}

						// スタンプをデリート
						if snapshot.Data() != nil {
							getUserFromFirestore := snapshot.Data()

							if res, state := getUserFromFirestore["technology"].([]interface{}); state {
								technology := []string{}
								for _, v := range res {
									if v != reaction {
										technology = append(technology, v.(string))
									}
								}

								client.Collection("users").Doc(userInfo.ID).Set(ctx, map[string]interface{} {
									"technology": technology,
								}, firestore.MergeAll)
							}
							fmt.Println("スタンプ消したルート")
						}

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
							case "tech":
								if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("今回使用している技術をスタンプで欲しいっチュ！！！\n（押されてないものは自分で追加してね！）", false)); err != nil {
									log.Println(err)
									w.WriteHeader(http.StatusInternalServerError)

									return
								}
						}

					case *slackevents.MessageEvent:
						if event.Text == "今回使用している技術をスタンプで欲しいっチュ！！！\n（押されてないものは自分で追加してね！）" {
							ref := slack.NewRefToMessage(event.Channel, event.TimeStamp)

							for _, value := range SlackReactionTechs {
								log.Println(value)
								api.AddReaction(value, ref)
							}
						}
				}
		}
	})

	log.Println("[INFO] Server listening")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatal(err)
	}
}
