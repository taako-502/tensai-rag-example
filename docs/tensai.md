# tensai

tensaiは、学習と実験を目的としたGo製の小さな機械学習フレームワークです。行列演算、ニューラルネットワークのレイヤー、最適化、逆伝播などをpure Goで実装しています。

tensaiコマンドはGGUFまたはsafetensors形式の言語モデルを読み込めます。`serve`サブコマンドを使うと、`/v1/chat/completions`にOpenAI互換のHTTP APIを公開します。

通常のビルドは外部依存やCGOを必要としません。WebGPUはオプションです。
