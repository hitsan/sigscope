# sigscope

VCD (Value Change Dump) ファイルをターミナルで操作するツールセット。

## 主な機能

- **波形ビューア**: エンジニアがターミナル上で波形確認できます
- **JSON出力**: AIエージェントが信号リストや波形データをプログラマティックに取得できます

## インストール

### バイナリのビルド

```bash
git clone <repository-url>
cd sigscope
go build -o sigscope .
```

または直接実行:

```bash
go run . <command> [options]
```

### Claude Code Skillのインストール

このツールをClaude Codeから使用するためのスキルが含まれています。

```bash
cp -r .claude/skills/sigscope <path-to-project>/.claude/skills/
```

インストール後、Claude Codeで `/sigscope` コマンドが使用可能になります。

## 使い方

### 1. 波形ビューア

VCDファイルをインタラクティブに閲覧します。

```bash
sigscope path/to/file.vcd
```

#### 波形表示スタイル

1ビット信号は以下の文字で表示されます：

```
clk:  __/‾‾\__/‾‾\__/‾‾\__
```

- `_` = LOW (0)
- `‾` = HIGH (1)
- `/` = 立ち上がりエッジ (0→1)
- `\` = 立ち下がりエッジ (1→0)
- `?` = 不明値 (x)
- `Z` = High-Z (z)

マルチビット信号（バス）は16進数で表示されます：

```
data[7:0]: X--2A---X--FF---X--00---
```

- `X` = 値の変化点
- `-` = 値の継続
- `2A`, `FF` = 16進数の値

#### TUI操作

- `q` / `Ctrl+C`: 終了
- `j` / `k` / `↑` / `↓`: シグナル移動
- `h` / `l` / `←` / `→`: 時間ウィンドウをスクロール
- `H` / `L` / `Shift+←` / `Shift+→`: ページ単位でスクロール
- `+` / `-` / `0`: ズームイン / ズームアウト / リセット
- `g` / `G`: 先頭 / 末尾へジャンプ
- `c`: カーソル表示の切替
- `[` / `]`: 前後の変化点へジャンプ
- `/`: 検索モード
- `s`: シグナル選択モード切替
- `space`: 表示/非表示の切替（選択モードのみ）
- `a` / `A`: 全表示 / 全非表示（選択モードのみ）

### 2. 信号リスト取得

VCDファイル内の全信号とメタデータをJSON形式で出力します。

```bash
sigscope list path/to/file.vcd
```

**出力例:**
```json
{
  "signals": [
    {"name": "TOP.module.clk", "width": 1},
    {"name": "TOP.module.data", "width": 8}
  ],
  "timescale": "1ps",
  "time_range": [0, 1000000]
}
```

### 3. 波形データ抽出（query）

時系列イベントをJSON形式で出力します。変化した信号のみを記録する差分形式です。

```bash
sigscope query [OPTIONS] path/to/file.vcd
```

**オプション:**
- `-s, --signals <pattern>`: 信号名パターン（部分一致、繰り返し可能）
- `-t, --time-start <time>`: 開始時刻（デフォルト: 0）
- `-e, --time-end <time>`: 終了時刻（デフォルト: VCD終了時刻）

**使用例:**
```bash
# 全信号の波形データ
sigscope query waveform.vcd

# 特定信号のみ（部分一致）
sigscope query -s clk -s data waveform.vcd

# 時間範囲指定
sigscope query -t 1000 -e 5000 waveform.vcd

# 組み合わせ
sigscope query -s "udp_rx" -t 1000 -e 10000 waveform.vcd
```

**出力例:**
```json
{
  "timescale": "1ps",
  "defs": {
    "clk": {"w": 1},
    "data": {"w": 8, "radix": "hex"}
  },
  "clock": {
    "name": "clk",
    "period": 10000,
    "edge": "posedge"
  },
  "init": {
    "clk": "0",
    "data": "00"
  },
  "events": [
    {"t": 15000, "set": {"data": "2A"}},
    {"t": 25000, "set": {"data": "FF"}}
  ]
}
```

**出力形式の詳細:**
- `timescale`: VCDファイルのタイムスケール（例: `"1ps"`, `"1ns"`）
- `defs`: 各信号のビット幅と基数（hex/bin）
- `clock`: 自動検出されたクロック情報（検出失敗時は`null`）
- `init`: 開始時刻における各信号の初期値
- `events`: 時刻順の変化イベント（変化した信号のみ記録）

AIエージェント向けの詳細は[AGENT.md](./AGENT.md)を参照してください。

## 技術仕様

- **言語**: Go 1.24.3
- **依存ライブラリ**: Bubble Tea, Lip Gloss, fsnotify
