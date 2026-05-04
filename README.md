# IIJmio プラン組み合わせ計算機

IIJmio の料金プランを複数回線・複数 SIM タイプで組み合わせ、指定した合計容量の範囲内で最安値を計算するツールです。

## 機能

- 音声 / SMS / データ eSIM / データ の各 SIM タイプについて最低・最大回線数を指定可能
- 指定条件の中で合計容量を満たす最安の組み合わせを表示
- 容量を 1GB 単位で一覧表示するモード
- CLI（Go バイナリ）と Web ブラウザ（WebAssembly）の両方に対応

## 制約

- 音声 SIM は最大 5 回線
- 全 SIM タイプ合計で最大 10 回線

## 料金プラン

| 容量 | 音声 | SMS | データ eSIM | データ |
|------|------|-----|------------|--------|
| 2GB  | ¥850 | ¥820 | ¥440 | ¥740 |
| 5GB  | ¥950 | ¥930 | ¥650 | ¥860 |
| 10GB | ¥1,400 | ¥1,370 | ¥1,050 | ¥1,300 |
| 15GB | ¥1,600 | ¥1,580 | ¥1,320 | ¥1,530 |
| 25GB | ¥2,000 | ¥1,980 | ¥1,650 | ¥1,950 |
| 35GB | ¥2,400 | ¥2,380 | ¥2,240 | ¥2,340 |
| 45GB | ¥3,300 | ¥3,280 | ¥2,940 | ¥3,240 |
| 55GB | ¥3,900 | ¥3,880 | ¥3,540 | ¥3,840 |

> 複数回線割引: ¥100/回線/月（全回線合計から差し引き）

## 使い方

### CLI

```sh
# ビルド
go build -o iij-kumiawase .

# 例: 音声2〜4回線・合計 25〜45GB の最安値を計算
./iij-kumiawase -voice-min 2 -voice-max 4 -min 25 -max 45

# 例: データeSIMのみ・1GB 単位で一覧表示
./iij-kumiawase -esim-min 1 -min 10 -max 30 -all

# 例: 制約なし（全タイプ自由）で合計 50GB 以上
./iij-kumiawase -min 50 -max 100
```

#### オプション

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `-voice-min` | 0 | 音声 SIM の最低回線数 |
| `-voice-max` | -1 | 音声 SIM の最大回線数（-1 = 制限なし、システム上限 5） |
| `-sms-min` | 0 | SMS SIM の最低回線数 |
| `-sms-max` | -1 | SMS SIM の最大回線数（-1 = 制限なし） |
| `-esim-min` | 0 | データ eSIM の最低回線数 |
| `-esim-max` | -1 | データ eSIM の最大回線数（-1 = 制限なし） |
| `-data-min` | 0 | データ SIM の最低回線数 |
| `-data-max` | -1 | データ SIM の最大回線数（-1 = 制限なし） |
| `-min` | 25 | 最低合計容量 (GB) |
| `-max` | 45 | 最高合計容量 (GB) |
| `-all` | false | min〜max を 1GB 単位で一覧表示 |

### Web (WebAssembly)

`index.html`・`main.wasm`・`wasm_exec.js` を同じディレクトリに置き、HTTP サーバーで配信します。

```sh
python3 -m http.server 8080
```

ブラウザで `http://localhost:8080` を開いてください。

#### WASM のビルド

```sh
make wasm
# または
GOOS=js GOARCH=wasm go build -tags 'js wasm' -o main.wasm .
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```
