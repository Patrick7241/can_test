package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CANSignal CAN信号定义
type CANSignal struct {
	Name        string  // 信号名称
	CANID       uint32  // CAN ID
	StartBit    int     // 起始位
	BitLength   int     // 数据长度（bit）
	MinValue    float64 // 最小值
	Scale       float64 // 缩放因子
	Offset      float64 // 偏移量
	Signed      bool    // 是否有符号
	ByteOrder   string  // 字节序（Motorola/Intel）
	Description string  // 描述
}

func main() {
	filename := "Volkswagen-ID.4 2020-.REF"

	fmt.Println("========================================")
	fmt.Println("大众ID.4 REF文件完整解析器")
	fmt.Println("========================================\n")

	signals, err := parseREFFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 成功解析出 %d 个CAN信号定义\n\n", len(signals))

	for i, sig := range signals {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("信号 #%d\n", i+1)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("名称:         %s\n", sig.Name)
		fmt.Printf("CAN ID:       %d (0x%X)\n", sig.CANID, sig.CANID)
		fmt.Printf("起始位:       %d\n", sig.StartBit)
		fmt.Printf("长度:         %d bit\n", sig.BitLength)
		fmt.Printf("缩放因子:     %.6f\n", sig.Scale)
		fmt.Printf("偏移量:       %.6f\n", sig.Offset)
		fmt.Printf("数据类型:     %s\n", map[bool]string{true: "有符号", false: "无符号"}[sig.Signed])
		fmt.Printf("字节序:       %s\n", sig.ByteOrder)
		fmt.Printf("计算公式:     实际值 = 原始值 × %.6f + %.6f\n", sig.Scale, sig.Offset)

		// 计算值范围
		maxRaw := (1 << uint(sig.BitLength)) - 1
		minActual := sig.Offset
		maxActual := float64(maxRaw)*sig.Scale + sig.Offset
		fmt.Printf("值范围:       %.3f ~ %.3f\n", minActual, maxActual)

		// 信号说明
		desc := analyzeSignal(sig)
		if desc != "" {
			fmt.Printf("\n信号说明:\n%s\n", desc)
		}
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Println("解析完成")
	fmt.Println("========================================")
}

// parseREFFile 解析REF文件
func parseREFFile(filename string) ([]CANSignal, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 跳过文本头
	reader.ReadString('\n')
	reader.ReadString('\n')

	// 读取所有二进制数据
	binaryData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// 解析所有块并提取信号
	signals := extractAllSignals(binaryData)
	return signals, nil
}

// extractAllSignals 提取所有信号
func extractAllSignals(data []byte) []CANSignal {
	var signals []CANSignal
	offset := 0

	for offset < len(data)-2 {
		// 读取块大小（2字节大端序）
		blockSize := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if blockSize == 0 || offset+int(blockSize) > len(data) {
			break
		}

		// 提取块数据
		blockData := data[offset : offset+int(blockSize)]
		offset += int(blockSize)

		// 尝试解压
		if blockSize > 2 && blockData[0] == 0x78 {
			r, err := zlib.NewReader(bytes.NewReader(blockData))
			if err == nil {
				decompressed, err := io.ReadAll(r)
				r.Close()
				if err == nil {
					// 解析信号
					sig := parseSignal(string(decompressed))
					if sig != nil {
						signals = append(signals, *sig)
					}
				}
			}
		}
	}

	return signals
}

// parseSignal 解析单个信号
func parseSignal(text string) *CANSignal {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// 使用CSV reader解析
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}

	record := records[0]
	if len(record) < 10 {
		return nil
	}

	sig := &CANSignal{}
	sig.Name = strings.TrimSpace(record[0])

	// 解析CAN ID
	if canID, err := strconv.ParseUint(record[1], 10, 32); err == nil {
		sig.CANID = uint32(canID)
	}

	// 起始位
	if startBit, err := strconv.Atoi(record[3]); err == nil {
		sig.StartBit = startBit
	}

	// 长度
	if length, err := strconv.Atoi(record[4]); err == nil {
		sig.BitLength = length
	}

	// 最小值
	if minVal, err := strconv.ParseFloat(record[5], 64); err == nil {
		sig.MinValue = minVal
	}

	// 缩放因子
	if scale, err := strconv.ParseFloat(record[6], 64); err == nil {
		sig.Scale = scale
	}

	// 偏移量
	if offset, err := strconv.ParseFloat(record[7], 64); err == nil {
		sig.Offset = offset
	}

	// 有符号/无符号
	if len(record) > 9 {
		sig.Signed = strings.Contains(record[9], "Signed") && !strings.Contains(record[9], "Unsigned")
	}

	// 字节序
	if len(record) > 10 {
		sig.ByteOrder = strings.TrimSpace(record[10])
	}

	return sig
}

// analyzeSignal 分析信号含义
func analyzeSignal(sig CANSignal) string {
	name := strings.ToLower(sig.Name)

	if strings.Contains(name, "speed") || strings.Contains(name, "vehicle") {
		return "  车速信号 - 指示当前车辆行驶速度"
	} else if strings.Contains(name, "brake") {
		return "  刹车系统信号 - 监测刹车状态或力度"
	} else if strings.Contains(name, "switch") {
		return "  开关信号 - 二进制状态指示"
	} else if strings.Contains(name, "indicated") || strings.Contains(name, "display") {
		return "  仪表显示信号 - 用于驾驶员界面显示"
	} else if strings.Contains(name, "battery") {
		return "  电池相关信号 - 电动车电池管理"
	} else if strings.Contains(name, "temperature") || strings.Contains(name, "temp") {
		return "  温度传感器信号"
	} else if strings.Contains(name, "rate") {
		return "  速率/频率信号"
	} else if strings.Contains(name, "direction") {
		return "  方向信号"
	} else if strings.Contains(name, "angle") {
		return "  角度传感器信号"
	} else if strings.Contains(name, "accelerat") {
		return "  加速度/油门信号"
	} else if strings.Contains(name, "pedal") {
		return "  踏板位置信号"
	} else if strings.Contains(name, "position") {
		return "  位置传感器信号"
	}

	return ""
}
