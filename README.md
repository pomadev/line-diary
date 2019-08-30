## ディレクトリ構成

    .
    ├── line-diary        # LINE Webhook URLに登録しているURLに紐づくハンドラー
    ├── line-notify       # Cloud Watch Eventsから叩かれるハンドラー
    ├── Makefile          
    └── template.yaml     # AWS SAM template file

## 構成図

![構成図](./line-diary.png)
