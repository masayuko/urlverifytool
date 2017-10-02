# 手順

## node.jsのインストール

https://nodejs.org/ja/ から最新版をダウンロードしてインストールします。
インストーラーが起動しますので、促されるがままにインストールしてください。

## node.jsのコマンドプロンプトを起動

node.jsをインストールすると、メニューに「Node.js command prompt」が存在しています。
それを起動します。そうするとDOS窓が開きます。
`node -v` と入力して、nodeが起動するのを確認してください。

## site-checkerのインストール

作業用フォルダを一つ作成して、そこに移動してください。
例えば、urlverifyというフォルダを作成します。

```
> mkdir urlverify
> cd urlverify
```

npmコマンドを使って、site-checkerをインストールします。

```
> npm install site-checker
```

インストールが完了すると、`node_modules\.bin\site-checker` を起動することができるようになります。

```
> node_modules\.bin\site-checker
```

## CSVファイルからsite-checker用のURL一覧ファイルを作成

添付した urlverifytool.exe を作業用フォルダにコピーしてください。
このツールを使って、URL一覧ファイルを作成します。

```
> urlverifytool urllist -j 1000 disease_refs_prediction_omim.csv
```

この場合、1000件ずつに分割されたURLリストが作成されます。
disease_refs_prediction_omim.csvには2928件ありますので、3つのファイルになります。
作成されるファイルは、以下のようになります。

```
disease_refs_prediction_omim_url0.txt
disease_refs_prediction_omim_url1.txt
disease_refs_prediction_omim_url2.txt
```

-jオプションに渡す数を変更すると出力されるファイル数も変わります。
指定しないと一つのファイルに全てのURLが出力されます。

## site-checkerを起動し、スクリーンショットを取得します。

スクリーンショットはpngファイルとして保存されます。

```
> node_modules\.bin\site-checker -l disease_refs_prediction_omim_url0.txt -f
```

-fオプションを指定するとページ全体の画像が作成されます。

バッチ的に処理が行われ、したらく待つと終了します。
pngファイルは disease_refs_prediction_omim_url0 というフォルダの下に作成されます。
-o omim_url0 などと指定すると、そのフォルダの下に作成されます。

処理結果が記されたファイルが disease_refs_prediction_omim_url0/result.json に書き出されます。
このファイルには正常に取得、404 Not Foundやなんらかのエラーが発生したなどが記されています。

## 元のCSVファイルに結果をマージします。

urlverifytool.exeを使って、マージします。

```
> urlverifytool merge disease_refs_prediction_omim.csv disease_refs_prediction_omim_url0/result.json
```

３つのファイルが出力されます。

* disease_refs_prediction_omim_merged.csv - マージ後のCSV
* disease_refs_prediction_omim_error.txt - 応答がなくタイムアウトなどでエラーが発生したURL一覧
* disease_refs_prediction_omim_not200.txt - 404 Not Foundなどの応答があったURL一覧

disease_refs_prediction_omim_merged.csv には、
スクリーンショットの画像へのファイルパスが付与されています。
エラーや404などの情報も記載されています。

disease_refs_prediction_omim_error.txt を使って、再度site-checkerを起動することで、
スクリーンショットが取得できる可能性があります。

disease_refs_prediction_omim_not200.txt のURLは、サイトがなくなったか移動してしまった可能性があります。

## ドクター確認作業

disease_refs_prediction_omim_merged.csv をExcelで読み出し、
ドクターに確認作業を依頼することもできると思いますが、
奥村先生のおっしゃるように、専用のチェックツールを作成できればとも思っています。
これには来週頭ぐらいまでかかりそうです。
