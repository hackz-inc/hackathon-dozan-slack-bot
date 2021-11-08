# hackathon-dozan-slack-bot
ハッカソン参加者の使用技術を集計するslack bot  
  
その名も **どーざんbot** !!!

## Product
https://hackathon-dozan-slack-bot.herokuapp.com

## Description

- [ハックちゅう for Slack Bot](https://topaz.dev/projects/e7e5a03b79924031da7b)

## Depends

#### ▼ Packages

- Golang 1.17

#### ▼ インフラ, 外部サービス

- Heroku (デプロイ先)

## SetUp

環境変数を設定する必要があります。  
`serviceAccountKey.json`が必要となるので、誰かに教えてもらいましょう。  

また、下記の環境変数も誰かに教えてもらい、`.env`ファイルに設定しましょう。

```bash
SLACK_TOKEN=
BOT_ID=
```
