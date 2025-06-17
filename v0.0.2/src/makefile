# TinyGo用Makefile - Raspberry Pi Pico向けオーディオパススルーシステム

# ターゲット設定
TARGET = pico
MAIN_FILE = main.go
OUTPUT_FILE = main.uf2

# TinyGoビルド設定
TINYGO_FLAGS = -target=$(TARGET) -opt=2 -gc=conservative

# デフォルトターゲット: ビルド実行
all: build

# メインビルドターゲット: TinyGoでRaspberry Pi Pico用にコンパイル
build:
	tinygo build $(TINYGO_FLAGS) -o $(OUTPUT_FILE) $(MAIN_FILE)

# フラッシュターゲット: Picoにプログラムを書き込み
flash: build
	tinygo flash $(TINYGO_FLAGS) $(MAIN_FILE)

# モニターターゲット: シリアル出力を監視
monitor:
	tinygo monitor -target=$(TARGET)

# サイズ確認: バイナリサイズを表示
size: build
	ls -lh $(OUTPUT_FILE)

# クリーンアップ: 生成ファイルを削除
clean:
	rm -f $(OUTPUT_FILE)
	rm -f main.hex

# 依存関係確認: 必要なツールの存在確認
check-deps:
	@which tinygo > /dev/null || (echo "TinyGoがインストールされていません" && exit 1)
	@echo "TinyGoバージョン:"
	@tinygo version

# デバッグビルド: デバッグ情報付きでビルド
debug:
	tinygo build $(TINYGO_FLAGS) -debug -o $(OUTPUT_FILE) $(MAIN_FILE)

# リリースビルド: 最適化を最大にしてビルド
release:
	tinygo build -target=$(TARGET) -opt=z -gc=conservative -o $(OUTPUT_FILE) $(MAIN_FILE)

# ヘルプ: 使用可能なターゲットを表示
help:
	@echo "使用可能なターゲット:"
	@echo "  build    - TinyGoでビルド (デフォルト)"
	@echo "  flash    - Picoにプログラムを書き込み"
	@echo "  monitor  - シリアル出力を監視"
	@echo "  size     - バイナリサイズを確認"
	@echo "  clean    - 生成ファイルを削除"
	@echo "  debug    - デバッグビルド"
	@echo "  release  - リリースビルド"
	@echo "  help     - このヘルプを表示"

.PHONY: all build flash monitor size clean check-deps debug release help
