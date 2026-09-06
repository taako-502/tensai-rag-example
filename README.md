# tensai-rag-example

[mattn/tensai](https://github.com/mattn/tensai)を使った、Go製の小さなRAG（Retrieval-Augmented Generation）CLIです。

## 構成

1. `docs/`以下のMarkdown・テキストファイルを読み込む
2. 文書をチャンクに分割する
3. TF-IDFベクトルを作る
4. `tensai.DotVec`でコサイン類似度を計算し、関連チャンクを検索する
5. 検索結果をtensaiのOpenAI互換APIへ渡して回答を生成する

検索部分は意味埋め込みではなく、英数字の単語と日本語の文字2-gramを使った疎ベクトル検索です。外部の埋め込みAPIやベクトルDBを使わず、RAGの処理の流れを小さく確認することを優先しています。

## 必要なもの

- Go 1.27以上
- 回答生成に使うGGUFモデル（検索だけ試す場合は不要）

## 検索だけ試す

```sh
go run ./cmd/rag -docs ./docs -retrieve-only "RAGはどうやって文書を検索しますか？"
```

実行例:

```text
indexed 2 chunks from 2 documents
[1] rag.md (score: 0.6803)
# RAG
...
```

## 回答生成まで試す

最初のターミナルでtensaiサーバーを起動します。`-model`にはローカルのGGUFファイル、またはtensaiが取得できるモデル名を指定してください。

```sh
go run github.com/mattn/tensai/cmd/tensai@v0.0.26 serve \
  -model ./path/to/model.gguf \
  -q8 \
  -addr 127.0.0.1:8080


# モデルを指定せずに起動するのであれば以下でも問題ありません
go run github.com/mattn/tensai/cmd/tensai@v0.0.26 serve -addr 127.0.0.1:8080
```

別のターミナルでRAG CLIを実行します。

```sh
go run ./cmd/rag -docs ./docs "tensaiはどのようなライブラリですか？"
```

接続先を変える場合は`-endpoint`または`TENSAI_ENDPOINT`を指定します。tensaiサーバーを`-api-key`付きで起動した場合は、同じ値を`TENSAI_API_KEY`へ設定してください。

```sh
TENSAI_ENDPOINT=http://127.0.0.1:9000 \
TENSAI_API_KEY=secret \
go run ./cmd/rag "RAGとは何ですか？"
```

## 環境変数

起動時にカレントディレクトリの`.env`を自動で読み込みます。`.env.example`をコピーして利用できます。

```sh
cp .env.example .env
```

| 環境変数 | 説明 | 既定値 |
| --- | --- | --- |
| `TENSAI_ENDPOINT` | 回答生成に使用するtensaiサーバーのURLです。`-endpoint`を指定した場合は、コマンドラインの値が優先されます。 | `http://127.0.0.1:8080` |
| `TENSAI_API_KEY` | tensaiサーバーを`-api-key`付きで起動した場合に、そのAPIキーを指定します。APIキーを使わない場合は設定不要です。 | なし |

すでにシェルで同名の環境変数が設定されている場合は、`.env`の値で上書きしません。

これらは回答生成時だけ使用します。`-retrieve-only`で検索だけを試す場合は設定不要です。

## 文書を追加する

`docs/`以下へ`.md`または`.txt`ファイルを置き、CLIを再実行します。起動時に索引を作り直すため、別途データベースを用意する必要はありません。

## テスト

```sh
go test ./...
```

## 制約

- TF-IDFなので、同義語や言い換えの検索は苦手です。
- 索引はメモリ上にだけ保持し、起動ごとに作り直します。
- チャンク分割は段落境界と文字数だけを見る単純な実装です。
- tensaiサーバーはリクエストを直列処理するため、このサンプルはローカルでの動作確認向けです。

本格運用する場合は、埋め込みモデル、永続ベクトル索引、メタデータフィルタ、ストリーミング応答などを追加する必要があります。
