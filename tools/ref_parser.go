package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CANSignal CAN信号定义
type CANSignal struct {
	Name         string  // 信号名称
	CANID        uint32  // CAN ID
	StartBit     int     // 起始位
	Length       int     // 数据长度（bit）
	MinValue     float64 // 最小值
	Scale        float64 // 缩放因子
	Offset       float64 // 偏移量
	Signed       bool    // 是否有符号
	ByteOrder    string  // 字节序（Motorola/Intel）
	Description  string  // 描述
}

// ParseREFFile 解析Racelogic REF文件
func ParseREFFile(filename string) ([]CANSignal, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 跳过前两行文本头
	reader.ReadString('\n')
	reader.ReadString('\n')

	// 读取所有二进制数据
	binaryData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}

	// 提取所有压缩块的内容
	allText := extractAllZlibBlocks(binaryData)

	// 解析CSV格式的信号定义
	signals, err := parseSignalDefinitions(allText)
	if err != nil {
		return nil, fmt.Errorf("解析信号定义失败: %w", err)
	}

	return signals, nil
}

// extractAllZlibBlocks 提取所有zlib压缩块
func extractAllZlibBlocks(data []byte) string {
	var result strings.Builder
	offset := 0

	for offset < len(data)-1 {
		// 查找zlib头 (0x78)
		if data[offset] == 0x78 {
			// 尝试解压
			r, err := zlib.NewReader(bytes.NewReader(data[offset:]))
			if err == nil {
				decompressed, err := io.ReadAll(r)
				r.Close()
				if err == nil {
					result.Write(decompressed)
				}
			}
		}
		offset++
	}

	return result.String()
}

// parseSignalDefinitions 解析信号定义
func parseSignalDefinitions(text string) ([]CANSignal, error) {
	var signals []CANSignal

	// 使用CSV reader解析
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1 // 允许变长字段

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if len(record) < 10 {
			continue // 跳过不完整的记录
		}

		signal := CANSignal{}
		signal.Name = strings.TrimSpace(record[0])

		// 解析CAN ID
		if canID, err := strconv.ParseUint(record[1], 10, 32); err == nil {
			signal.CANID = uint32(canID)
		}

		// 起始位
		if startBit, err := strconv.Atoi(record[3]); err == nil {
			signal.StartBit = startBit
		}

		// 长度
		if length, err := strconv.Atoi(record[4]); err == nil {
			signal.Length = length
		}

		// 最小值
		if minVal, err := strconv.ParseFloat(record[5], 64); err == nil {
			signal.MinValue = minVal
		}

		// 缩放因子
		if scale, err := strconv.ParseFloat(record[6], 64); err == nil {
			signal.Scale = scale
		}

		// 偏移量
		if offset, err := strconv.ParseFloat(record[7], 64); err == nil {
			signal.Offset = offset
		}

		// 有符号/无符号
		signal.Signed = strings.Contains(record[9], "Signed") && !strings.Contains(record[9], "Unsigned")

		// 字节序
		signal.ByteOrder = strings.TrimSpace(record[10])

		signals = append(signals, signal)
	}

	return signals, nil
}

func main() {
	filename := "Volkswagen-ID.4 2020-.REF"

	fmt.Println("========================================")
	fmt.Println("Racelogic REF 文件解析器")
	fmt.Println("========================================")
	fmt.Printf("\n正在解析: %s\n\n", filename)

	signals, err := ParseREFFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("找到 %d 个CAN信号定义:\n\n", len(signals))

	for i, sig := range signals {
		fmt.Printf("========================================\n")
		fmt.Printf("信号 #%d\n", i+1)
		fmt.Printf("========================================\n")
		fmt.Printf("名称:         %s\n", sig.Name)
		fmt.Printf("CAN ID:       %d (0x%X)\n", sig.CANID, sig.CANID)
		fmt.Printf("起始位:       %d\n", sig.StartBit)
		fmt.Printf("长度:         %d bit\n", sig.Length)
		fmt.Printf("数值范围:     %.3f (最小值)\n", sig.MinValue)
		fmt.Printf("缩放因子:     %.3f\n", sig.Scale)
		fmt.Printf("偏移量:       %.3f\n", sig.Offset)
		fmt.Printf("数据类型:     %s\n", map[bool]string{true: "有符号", false: "无符号"}[sig.Signed])
		fmt.Printf("字节序:       %s\n", sig.ByteOrder)
		fmt.Printf("\n计算公式:     实际值 = 原始值 × %.3f + %.3f\n", sig.Scale, sig.Offset)
		fmt.Printf("\n")

		// 信号含义分析
		fmt.Printf("信号说明:\n")
		analyzeSignal(sig)
		fmt.Printf("\n")
	}

	fmt.Println("========================================")
	fmt.Println("解析完成")
	fmt.Println("========================================")
}

// analyzeSignal 分析信号含义
func analyzeSignal(sig CANSignal) {
	switch sig.Name {
	case "Brake_Position":
		fmt.Printf("  这是刹车踏板位置传感器信号\n")
		fmt.Printf("  - 用于监测驾驶员踩下刹车踏板的深度\n")
		fmt.Printf("  - 4位数据可以表示 0-15 共16个级别\n")
		maxRaw := (1 << uint(sig.Length)) - 1
		fmt.Printf("  - 经过转换后，实际值范围: %.3f ~ %.3f\n",
			sig.Offset, float64(maxRaw)*sig.Scale+sig.Offset)
		fmt.Printf("  - 0 表示未踩刹车，值越大表示踩得越深\n")
		fmt.Printf("  - 这个信号通常用于:\n")
		fmt.Printf("    · 刹车灯控制\n")
		fmt.Printf("    · ABS防抱死系统\n")
		fmt.Printf("    · ESP电子稳定系统\n")
		fmt.Printf("    · 自动驾驶辅助系统\n")
	default:
		fmt.Printf("  未知信号类型\n")
	}
}
