# Go のベースイメージを指定
FROM golang:1.20.5-alpine

# コマンドを実行するコンテナ内のディレクトリをworkに指定
WORKDIR /work

# ローカルのカレントディレクトリをコンテナのカレントディレクトリ(work)にコピー
COPY . .

# Go のプログラムをビルド
RUN go build -o app

# ビルドしたものを実行
ENTRYPOINT ./app
