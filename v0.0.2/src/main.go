package main

// Claude Sonnet 4製
// 大幅最適化版：シリアル出力を1秒間隔に短縮

import (
	"machine"
	"time"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

const (
	// PM1808PWR (入力系)
	ADC_DATA_PIN  = machine.GPIO2
	ADC_CLOCK_PIN = machine.GPIO3
	ADC_LR_PIN    = machine.GPIO4

	// PCM5102A (出力系)
	DAC_DATA_PIN  = machine.GPIO16
	DAC_CLOCK_PIN = machine.GPIO17
	DAC_LR_PIN    = machine.GPIO18

	// 可変抵抗（ADC）接続
	LOWCUT_POT_PIN  = machine.ADC1 // GPIO27/ADC1
	HIGHCUT_POT_PIN = machine.ADC2 // GPIO28/ADC2

	// 制御系
	STATUS_LED = machine.GPIO25
	DEBUG_PIN  = machine.GPIO15

	SAMPLE_RATE = 44100
	BUFFER_SIZE = 32    // バッファサイズを大幅削減。
	
	// 大幅最適化設定
	ADC_RESOLUTION = 4096
	VREF = 3.3
	
	// 周波数範囲
	LOWCUT_CENTER  = 400.0
	LOWCUT_RANGE   = 400.0
	HIGHCUT_CENTER = 1400.0
	HIGHCUT_RANGE  = 700.0
	
	// 更新間隔を大幅最適化
	FILTER_UPDATE_INTERVAL = 2000 // フィルタ更新を大幅削減。
	SERIAL_OUTPUT_INTERVAL = 50   // シリアル出力を大幅短縮。
	
	PI = 3.14159265359
)

// 軽量フィルタ構造体
type SimpleBiquad struct {
	b0, b1, b2 float32
	a1, a2     float32
	x1, x2     float32
	y1, y2     float32
}

type SimpleBPF struct {
	hp SimpleBiquad // ハイパス1段のみ
	lp SimpleBiquad // ローパス1段のみ
}

var (
	inputSM     pio.StateMachine
	outputSM    pio.StateMachine
	i2sInput    *piolib.I2S
	i2sOutput   *piolib.I2S
	audioBuffer [BUFFER_SIZE]uint32
	statusLED   machine.Pin
	debugPin    machine.Pin
	
	// ADC（必要最小限）
	adc1 machine.ADC
	adc2 machine.ADC
	
	// 簡略化フィルタ
	leftBPF  SimpleBPF
	rightBPF SimpleBPF
	
	// 周波数制御（簡略化）
	currentLowcut  float32
	currentHighcut float32
	
	// カウンタ
	sampleCount uint32
	serialCount uint32
	filterCount uint32
	
	// 簡易CPU負荷測定
	lastTime uint64
	cpuLoad  float32
)

func init() {
	// 最小限の初期化
	statusLED = STATUS_LED
	statusLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	
	debugPin = DEBUG_PIN
	debugPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	
	// シリアル通信初期化
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	time.Sleep(time.Millisecond * 500) // 短縮
	
	// ADC初期化（必要最小限）
	machine.InitADC()
	adc1 = machine.ADC{Pin: LOWCUT_POT_PIN}
	adc1.Configure(machine.ADCConfig{})
	adc2 = machine.ADC{Pin: HIGHCUT_POT_PIN}
	adc2.Configure(machine.ADCConfig{})
	
	// 初期値設定
	currentLowcut = LOWCUT_CENTER
	currentHighcut = HIGHCUT_CENTER
	
	// 簡略フィルタ初期化
	initSimpleFilters()
	
	println("# Optimized Audio BPF System")
	println("# Format: ADC1,ADC2,LowCut,HighCut,CPULoad")
}

func main() {
	initializeSystemFast()
	startOptimizedPassthrough()
}

func initializeSystemFast() {
	time.Sleep(time.Millisecond * 50) // 短縮
	statusLED.High()
	
	var err error
	
	inputSM, err = pio.PIO0.ClaimStateMachine()
	if err != nil {
		panic("Input SM failed")
	}
	
	outputSM, err = pio.PIO1.ClaimStateMachine()
	if err != nil {
		panic("Output SM failed")
	}
	
	// I2S初期化（簡略化）
	i2sInput, err = piolib.NewI2S(inputSM, ADC_DATA_PIN, ADC_CLOCK_PIN)
	if err != nil {
		panic("I2S input failed")
	}
	i2sInput.SetSampleFrequency(SAMPLE_RATE)
	
	i2sOutput, err = piolib.NewI2S(outputSM, DAC_DATA_PIN, DAC_CLOCK_PIN)
	if err != nil {
		panic("I2S output failed")
	}
	i2sOutput.SetSampleFrequency(SAMPLE_RATE)
	
	println("# System ready")
}

func initSimpleFilters() {
	// 1段フィルタで軽量化
	leftBPF.hp = createSimpleHighpass(currentLowcut)
	leftBPF.lp = createSimpleLowpass(currentHighcut)
	rightBPF.hp = createSimpleHighpass(currentLowcut)
	rightBPF.lp = createSimpleLowpass(currentHighcut)
}

func createSimpleHighpass(freq float32) SimpleBiquad {
	var filter SimpleBiquad
	omega := 2.0 * PI * freq / SAMPLE_RATE
	cosOmega := fastCos(omega)
	sinOmega := fastSin(omega)
	alpha := sinOmega / 1.414
	
	a0 := 1.0 + alpha
	filter.b0 = (1.0 + cosOmega) / 2.0 / a0
	filter.b1 = -(1.0 + cosOmega) / a0
	filter.b2 = (1.0 + cosOmega) / 2.0 / a0
	filter.a1 = -2.0 * cosOmega / a0
	filter.a2 = (1.0 - alpha) / a0
	
	return filter
}

func createSimpleLowpass(freq float32) SimpleBiquad {
	var filter SimpleBiquad
	omega := 2.0 * PI * freq / SAMPLE_RATE
	cosOmega := fastCos(omega)
	sinOmega := fastSin(omega)
	alpha := sinOmega / 1.414
	
	a0 := 1.0 + alpha
	filter.b0 = (1.0 - cosOmega) / 2.0 / a0
	filter.b1 = (1.0 - cosOmega) / a0
	filter.b2 = (1.0 - cosOmega) / 2.0 / a0
	filter.a1 = -2.0 * cosOmega / a0
	filter.a2 = (1.0 - alpha) / a0
	
	return filter
}

func processSimpleBiquad(filter *SimpleBiquad, input float32) float32 {
	output := filter.b0*input + filter.b1*filter.x1 + filter.b2*filter.x2 - filter.a1*filter.y1 - filter.a2*filter.y2
	
	filter.x2 = filter.x1
	filter.x1 = input
	filter.y2 = filter.y1
	filter.y1 = output
	
	return output
}

func startOptimizedPassthrough() {
	statusLED.High()
	lastTime = getMicroseconds()
	
	for {
		startTime := getMicroseconds()
		
		// 音声処理（簡略化）
		if readAudioDataFast() {
			processAudioFast()
			writeAudioDataFast()
			sampleCount += BUFFER_SIZE
		}
		
		// フィルタ更新（低頻度）
		if filterCount >= FILTER_UPDATE_INTERVAL {
			updateFiltersFromPots()
			filterCount = 0
		}
		filterCount++
		
		// CPU負荷計算（簡略）
		endTime := getMicroseconds()
		processingTime := float32(endTime - startTime)
		totalTime := float32(endTime - lastTime)
		if totalTime > 0 {
			cpuLoad = processingTime / totalTime * 100.0
			if cpuLoad > 100.0 {
				cpuLoad = 100.0
			}
		}
		lastTime = endTime
		
		// 高頻度シリアル出力
		if serialCount >= SERIAL_OUTPUT_INTERVAL {
			outputMinimalData()
			serialCount = 0
		}
		serialCount++
		
		// 動的スリープ
		if cpuLoad > 90.0 {
			time.Sleep(time.Microsecond * 200)
		} else if cpuLoad > 70.0 {
			time.Sleep(time.Microsecond * 100)
		} else {
			time.Sleep(time.Microsecond * 20)
		}
	}
}

func readAudioDataFast() bool {
	// 簡易テスト信号（軽量化）
	for i := 0; i < BUFFER_SIZE; i++ {
		t := float32(sampleCount+uint32(i)) / SAMPLE_RATE
		signal := 0.3*fastSin(2*PI*500*t) + 0.3*fastSin(2*PI*1000*t)
		sample := int16(signal * 16384)
		audioBuffer[i] = uint32(uint16(sample)) | (uint32(uint16(sample)) << 16)
	}
	return true
}

func processAudioFast() {
	// 1段フィルタで高速処理
	for i := 0; i < BUFFER_SIZE; i++ {
		leftSample := int16(audioBuffer[i] & 0xFFFF)
		rightSample := int16((audioBuffer[i] >> 16) & 0xFFFF)
		
		leftFloat := float32(leftSample) / 32768.0
		rightFloat := float32(rightSample) / 32768.0
		
		// 1段ずつ処理
		leftHP := processSimpleBiquad(&leftBPF.hp, leftFloat)
		leftFiltered := processSimpleBiquad(&leftBPF.lp, leftHP)
		
		rightHP := processSimpleBiquad(&rightBPF.hp, rightFloat)
		rightFiltered := processSimpleBiquad(&rightBPF.lp, rightHP)
		
		// クリッピング（簡略）
		if leftFiltered > 1.0 {
			leftFiltered = 1.0
		} else if leftFiltered < -1.0 {
			leftFiltered = -1.0
		}
		if rightFiltered > 1.0 {
			rightFiltered = 1.0
		} else if rightFiltered < -1.0 {
			rightFiltered = -1.0
		}
		
		leftInt := int16(leftFiltered * 32767.0)
		rightInt := int16(rightFiltered * 32767.0)
		
		audioBuffer[i] = uint32(uint16(leftInt)) | (uint32(uint16(rightInt)) << 16)
	}
}

func writeAudioDataFast() {
	i2sOutput.WriteStereo(audioBuffer[:])
}

func updateFiltersFromPots() {
	// ADC読み取り（平均化なし）
	adc1Val := adc1.Get()
	adc2Val := adc2.Get()
	
	// 周波数計算
	lowcutVoltage := float32(adc1Val) * VREF / ADC_RESOLUTION
	highcutVoltage := float32(adc2Val) * VREF / ADC_RESOLUTION
	
	newLowcut := (lowcutVoltage / VREF) * (2.0 * LOWCUT_RANGE)
	if newLowcut < 20.0 {
		newLowcut = 20.0
	}
	
	newHighcut := HIGHCUT_CENTER - HIGHCUT_RANGE + (highcutVoltage / VREF) * (2.0 * HIGHCUT_RANGE)
	if newHighcut > 20000.0 {
		newHighcut = 20000.0
	}
	
	if newHighcut <= newLowcut {
		newHighcut = newLowcut + 100.0
	}
	
	// 大きな変化があった場合のみ更新
	if fastAbs(newLowcut - currentLowcut) > 20.0 || fastAbs(newHighcut - currentHighcut) > 20.0 {
		currentLowcut = newLowcut
		currentHighcut = newHighcut
		
		// フィルタ再計算
		leftBPF.hp = createSimpleHighpass(currentLowcut)
		leftBPF.lp = createSimpleLowpass(currentHighcut)
		rightBPF.hp = createSimpleHighpass(currentLowcut)
		rightBPF.lp = createSimpleLowpass(currentHighcut)
		
		debugPin.High()
		time.Sleep(time.Microsecond * 50)
		debugPin.Low()
	}
}

func outputMinimalData() {
	// 最小限のデータ出力
	adc1Val := adc1.Get()
	adc2Val := adc2.Get()
	
	adc1Percent := int(float32(adc1Val) * 100.0 / ADC_RESOLUTION)
	adc2Percent := int(float32(adc2Val) * 100.0 / ADC_RESOLUTION)
	
	lowcutScale := int(currentLowcut / 800.0 * 100.0)
	highcutScale := int((currentHighcut - 700.0) / 1400.0 * 100.0)
	
	print("ADC1:")
	print(adc1Percent)
	print(",ADC2:")
	print(adc2Percent)
	print(",LowCut:")
	print(lowcutScale)
	print(",HighCut:")
	print(highcutScale)
	print(",CPULoad:")
	print(int(cpuLoad))
	println()
}

func getMicroseconds() uint64 {
	return uint64(time.Now().UnixNano() / 1000)
}

// 高速数学関数（テーブル参照なし）
func fastAbs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func fastSin(x float32) float32 {
	// 非常に簡易な近似
	x2 := x * x
	return x * (1.0 - x2/6.0)
}

func fastCos(x float32) float32 {
	return fastSin(x + PI/2.0)
}