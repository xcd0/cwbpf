package main

// シンプル割り込み版（TinyGo対応）

import (
	"machine"
	"time"
	"device/rp"
)

const (
	SAMPLE_RATE = 96000
	RING_BUFFER_SIZE = 4096
	OUTPUT_SAMPLES = 50
	
	// システムクロック125MHzベース
	SYSTEM_CLOCK = 125000000
	TIMER_DIVIDER = SYSTEM_CLOCK / SAMPLE_RATE // 96kHz用分周比
)

var (
	adc1 machine.ADC
	ringBuffer [RING_BUFFER_SIZE]uint16
	writeIndex uint32
	samplingActive bool
	
	// 割り込み制御
	interruptFlag bool
)

func init() {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	time.Sleep(time.Millisecond * 500)
	
	machine.InitADC()
	adc1 = machine.ADC{Pin: machine.ADC1}
	adc1.Configure(machine.ADCConfig{})
	
	statusLED := machine.GPIO25
	statusLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	statusLED.High()
	
	writeIndex = 0
	samplingActive = true
	interruptFlag = false
	
	println("# Simple Hardware Interrupt 96kHz System")
}

func main() {
	setupHardwareTimer()
	
	go interruptSamplingLoop()
	go outputLoop()
	
	for {
		time.Sleep(time.Second)
	}
}

func setupHardwareTimer() {
	// RP2040のTimer0を96kHz割り込み用に設定
	
	// Timer0無効化
	rp.TIMER.CTRL.Set(0)
	
	// 分周比設定
	rp.TIMER.LOAD.Set(TIMER_DIVIDER)
	
	// タイマー有効化（周期モード、割り込み有効）
	rp.TIMER.CTRL.Set(rp.TIMER_CTRL_ENABLE | rp.TIMER_CTRL_INTEN)
	
	println("# Hardware timer configured for 96kHz")
}

func interruptSamplingLoop() {
	for samplingActive {
		// タイマー割り込み待機
		if rp.TIMER.RIS.Get() & 1 != 0 {
			// ADCサンプリング
			value := adc1.Get()
			
			// リングバッファ格納
			bufferIndex := writeIndex & (RING_BUFFER_SIZE - 1)
			ringBuffer[bufferIndex] = value
			writeIndex++
			
			// 割り込みクリア
			rp.TIMER.INTCLR.Set(1)
		}
		
		time.Sleep(time.Microsecond * 1)
	}
}

func outputLoop() {
	for {
		time.Sleep(time.Second)
		
		// 最新50サンプル出力
		startPos := writeIndex - OUTPUT_SAMPLES
		if writeIndex < OUTPUT_SAMPLES {
			startPos = 0
		}
		
		for i := uint32(0); i < OUTPUT_SAMPLES; i++ {
			pos := (startPos + i) & (RING_BUFFER_SIZE - 1)
			value := ringBuffer[pos]
			
			normalized := (uint32(value) * 1000) / 4095
			intPart := normalized / 1000
			fracPart := normalized % 1000
			
			print(intPart)
			print(".")
			if fracPart < 100 { print("0") }
			if fracPart < 10 { print("0") }
			print(fracPart)
			println()
		}
	}
}