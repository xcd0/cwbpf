# モールス信号用可変BPFオーディオフィルター 仕様書 v0.0.2

## v0.0.2 内容

ADCとDACを強化する。

目的に対しオーバースペックだが、秋月電子で手に入るPCM1808とPCM5102Aを使用する。
- PCM1808
    - 24bit, ~192kHz, 98dB, I2S
    - https://akizukidenshi.com/catalog/g/g112988/
    - http://www.easyaudiokit.com/bekkan3/PCM1808/PCM1808.html
- PCM5102A
    - 32bit, ~384kHz, 112dB, I2S
    - https://akizukidenshi.com/catalog/g/g111836/
    - http://www.easyaudiokit.com/bekkan/PCM5102/PCM5102.html
を使用する。

将来的には、raspberry pi picoの処理能力の余力次第で、下記を目標にしていく。
- アナログLPFの回路を減らして、サンプリングレートを上げ、デジタル処理側でのフィルタ処理を増やしたい。(アナログ回路の実装面積削減)
- 更に処理能力に余力があれば、人の声を対象に、MFCCとかを使用して特徴量抽出、再変換して、声を聴きやすくする機能を実装したい。(CW以外の音声処理)


### PCM1808とPCM5102A解説 (Raspberry Pi Pico用)

下記は Claude Sonnet 4 の解説。

> # PCM1808とPCM5102A解説 (Raspberry Pi Pico用)
> 
> ## PCM1808 (24bit ADC)
> 
> ### 概要
> - **用途**: アナログ音声信号 → デジタル信号変換 (録音・収音)
> - **分解能**: 24bit
> - **サンプリングレート**: 8kHz～192kHz
> - **SNR**: 98dB
> - **インターフェース**: I2S出力
> 
> ### 主要ピン機能
> - **VINL+/VINL-**: 左チャンネル差動入力
> - **VINR+/VINR-**: 右チャンネル差動入力  
> - **DOUT**: デジタル音声データ出力
> - **BCKO**: ビットクロック出力
> - **LROUT**: L/Rチャンネル選択クロック出力
> - **MD0/MD1**: サンプリングレート設定
> - **FMT0/FMT1**: 出力フォーマット設定
> 
> ## PCM5102A (32bit DAC)
> 
> ### 概要
> - **用途**: デジタル信号 → アナログ音声信号変換 (音声出力)
> - **分解能**: 32bit
> - **サンプリングレート**: 最大384kHz
> - **SNR**: 112dB
> - **インターフェース**: I2S入力
> 
> ### 主要ピン機能
> - **DIN**: デジタル音声データ入力
> - **BCK**: ビットクロック入力
> - **LRCK**: L/Rチャンネル選択クロック入力
> - **OUTL/OUTR**: アナログ音声出力
> - **XSMT**: ミュート制御
> - **FMT**: オーディオフォーマット選択
> 
> ## Raspberry Pi Pico接続回路図
> 
> ### PCM1808 (ADC) 接続
> ```
> Raspberry Pi Pico              PCM1808 ADC
>                               
> GPIO 18 (I2S BCK) <----------- BCKO (Bit Clock Output)
> GPIO 19 (I2S WS)  <----------- LROUT (Word Select Output)  
> GPIO 20 (I2S DIN) <----------- DOUT (Data Output)
> 3V3 OUT ---------------------> VDD (3.3V Digital)
> 5V OUT (VBUS) --------------> VCC (5V Analog)
> GND ---------------------> DGND/AGND
> 
> 設定ピン:
> MD1 ---------> GND (単一速度モード)
> MD0 ---------> VDD (8-50kHz)
> FMT1 --------> GND (I2Sフォーマット)
> FMT0 --------> GND  
> OSR1 --------> GND
> OSR0 --------> VDD
> HPF ---------> VDD (ハイパスフィルタ有効)
> SCKI --------> GND (内蔵クロック使用)
> 
> アナログ入力:
> Audio Input L --[100nF]--- VINL+
>                     |
>                  [47kΩ]
>                     |
>                   AGND ---- VINL-
> 
> Audio Input R --[100nF]--- VINR+
>                     |
>                  [47kΩ]
>                     |
>                   AGND ---- VINR-
> ```
> 
> ### PCM5102A (DAC) 接続
> ```
> Raspberry Pi Pico              PCM5102A DAC
>                               
> GPIO 21 (I2S BCK) -----------> BCK (Bit Clock)
> GPIO 22 (I2S WS)  -----------> LRCK (Word Select)  
> GPIO 23 (I2S DIN) -----------> DIN (Data Input)
> 3V3 OUT ---------------------> VDD (3.3V)
> GND ---------------------> GND
> 
> 設定ピン:
> SCK ---------> GND (外部クロック使用しない)
> XSMT --------> VDD (ミュート解除)
> FMT ---------> GND (I2Sフォーマット)
> DEMP --------> GND (デエンファシス無効)
> FLT ---------> GND (ノーマルレイテンシ)
> AVDD --------> VDD または 5V (より良い音質)
> 
> アナログ出力:
> OUTL --------> Left Audio Output
> OUTR --------> Right Audio Output
> ```
> 
> ### 完全システム接続図
> ```
>                     Raspberry Pi Pico
>     Audio Input     ┌─────────────────┐     Audio Output
>          │          │                 │          │
>     ┌────▼────┐     │  GPIO 18 ←──────┼─┐    ┌───▼────┐
>     │ PCM1808 │     │  GPIO 19 ←──────┼─┼─┐  │PCM5102A│
>     │   ADC   │     │  GPIO 20 ←──────┼─┼─┼─┐│  DAC   │
>     │         │     │                 │ │ │ ││        │
>     │ BCKO ───┼─────┼─→ GPIO 18       │ │ │ ││        │
>     │ LROUT───┼─────┼─→ GPIO 19       │ │ │ ││        │
>     │ DOUT ───┼─────┼─→ GPIO 20       │ │ │ ││        │
>     │         │     │                 │ │ │ ││        │
>     │         │     │  GPIO 21 ───────┼─┼─┼─┼┼─→ BCK  │
>     │         │     │  GPIO 22 ───────┼─┼─┼─┼┼─→ LRCK │
>     │         │     │  GPIO 23 ───────┼─┼─┼─┼┼─→ DIN  │
>     │         │     │                 │ │ │ ││        │
>     │ VDD ────┼─────┼── 3V3           │ │ │ ││── VDD  │
>     │ VCC ────┼─────┼── 5V            │ │ │ ││        │
>     │ GND ────┼─────┼── GND ──────────┼─┼─┼─┼┼── GND  │
>     └─────────┘     └─────────────────┘ │ │ │└────────┘
>                                         │ │ │
>                                         │ │ └─ GPIO 23
>                                         │ └─── GPIO 22  
>                                         └───── GPIO 21
> ```
> 
> ## ピン割り当て表
> 
> | 用途 | PCM1808 | Pico GPIO | PCM5102A |
> |------|---------|-----------|----------|
> | BCK | BCKO (出力) | 18 | BCK (入力) |
> | WS/LRCK | LROUT (出力) | 19 | LRCK (入力) |
> | データ | DOUT (出力) | 20 | - |
> | データ | - | 21 | BCK (入力) |
> | データ | - | 22 | LRCK (入力) |
> | データ | - | 23 | DIN (入力) |
> 
> ## 実装上の注意点
> 
> ### 電源設計
> - **PCM1808**: VDD=3.3V (デジタル), VCC=5V (アナログ、音質向上)
> - **PCM5102A**: VDD=3.3V, AVDD=3.3Vまたは5V (音質向上)
> - 各電源に100nF + 10µFバイパスコンデンサ配置
> 
> ### クロック設計
> - PCM1808がマスター (クロック出力)
> - PCM5102AがスレーブPIOで同期
> - または独立したI2Sポートで個別制御
> 
> ### グランド設計
> - デジタルグランド (DGND) とアナロググランド (AGND) を1点で接続
> - 適切なグランドプレーン配置でノイズ対策
> 
> この構成により、Raspberry Pi Picoで高品質な音声録音・再生システムを構築できます。
