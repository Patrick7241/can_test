package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// RacelogicREFReader 读取Racelogic .REF文件
type RacelogicREFReader struct {
	filename     string
	serialNumber string
	version      string
}

// NewRacelogicREFReader 创建新的REF文件读取器
func NewRacelogicREFReader(filename string) *RacelogicREFReader {
	return &RacelogicREFReader{
		filename: filename,
	}
}

// Parse 解析REF文件
func (r *RacelogicREFReader) Parse() error {
	file, err := os.Open(r.filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 读取文本头部
	reader := bufio.NewReader(file)

	// 第一行：版本信息
	versionLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取版本行失败: %w", err)
	}
	r.version = versionLine
	fmt.Printf("版本: %s", versionLine)

	// 第二行：序列号
	serialLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取序列号行失败: %w", err)
	}
	r.serialNumber = serialLine
	fmt.Printf("序列号: %s", serialLine)

	// 读取剩余的二进制数据
	compressedData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("读取压缩数据失败: %w", err)
	}

	fmt.Printf("\n原始二进制数据大小: %d 字节\n", len(compressedData))
	fmt.Printf("前32字节: % 02x\n", compressedData[:min(32, len(compressedData))])

	// 解析多个压缩块
	return r.parseBlocks(compressedData)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseBlocks 解析多个数据块
func (r *RacelogicREFReader) parseBlocks(data []byte) error {
	offset := 0
	blockNum := 0
	var allDecompressedData []byte

	fmt.Println("\n解析数据块:")
	fmt.Println("-----------------------------------")

	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}

		// 读取块大小（2字节，大端序）
		blockSize := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if blockSize == 0 {
			fmt.Printf("偏移=%d, 遇到零长度块，跳过\n", offset-2)
			continue
		}

		if offset+int(blockSize) > len(data) {
			fmt.Printf("偏移=%d, 块大小=%d, 超出范围 (剩余=%d)，尝试作为单字节长度\n",
				offset-2, blockSize, len(data)-offset)
			// 可能前面的解析有误，尝试回退并使用单字节长度
			offset -= 1
			blockSize = uint16(data[offset-1])
			if offset+int(blockSize) > len(data) {
				fmt.Printf("单字节长度也失败，停止解析\n")
				break
			}
			fmt.Printf("使用单字节长度: %d\n", blockSize)
		}

		blockNum++
		fmt.Printf("块 #%d: 大小=%d 字节, ", blockNum, blockSize)

		// 提取块数据
		blockData := data[offset : offset+int(blockSize)]
		offset += int(blockSize)

		// 检查是否是zlib压缩数据
		if blockSize > 2 && blockData[0] == 0x78 {
			fmt.Print("(zlib压缩) ")

			// 解压缩
			zlibReader, err := zlib.NewReader(bytes.NewReader(blockData))
			if err != nil {
				fmt.Printf("解压失败: %v\n", err)
				continue
			}

			decompressed, err := io.ReadAll(zlibReader)
			zlibReader.Close()

			if err != nil {
				fmt.Printf("读取失败: %v\n", err)
				continue
			}

			fmt.Printf("-> 解压后 %d 字节\n", len(decompressed))
			allDecompressedData = append(allDecompressedData, decompressed...)
		} else {
			fmt.Printf("(原始数据): % 02x\n", blockData)
		}
	}

	fmt.Printf("\n总共解析 %d 个块\n", blockNum)
	fmt.Printf("解压后总数据大小: %d 字节\n\n", len(allDecompressedData))

	if len(allDecompressedData) > 0 {
		// 显示解压后的数据
		fmt.Println("解压后的数据内容:")
		fmt.Println("-----------------------------------")
		displayHex(allDecompressedData, 512)

		// 尝试解析CAN消息
		fmt.Println("\n-----------------------------------")
		fmt.Println("解压后的可读文本内容:")
		fmt.Println("-----------------------------------")
		r.displayReadableText(allDecompressedData)
	}

	return nil
}

// displayReadableText 显示可读的文本内容
func (r *RacelogicREFReader) displayReadableText(data []byte) {
	// 将数据转换为字符串，只显示可打印字符
	readable := make([]byte, 0, len(data))
	for _, b := range data {
		if b >= 32 && b <= 126 || b == '\n' || b == '\r' || b == '\t' {
			readable = append(readable, b)
		}
	}

	if len(readable) > 0 {
		fmt.Println(string(readable))
	} else {
		fmt.Println("(没有可读的文本内容，可能是纯二进制CAN数据)")
	}
}

// decompressAndParse 解压缩并解析CAN数据
func (r *RacelogicREFReader) decompressAndParse(compressedData []byte) error {
	// 创建zlib读取器
	zlibReader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("创建zlib读取器失败: %w", err)
	}
	defer zlibReader.Close()

	// 解压缩
	decompressedData, err := io.ReadAll(zlibReader)
	if err != nil {
		return fmt.Errorf("解压缩失败: %w", err)
	}

	fmt.Printf("解压后数据大小: %d 字节\n\n", len(decompressedData))

	// 显示前256字节的十六进制和ASCII
	fmt.Println("解压后的数据内容（前256字节）:")
	fmt.Println("-----------------------------------")
	displayHex(decompressedData, 256)

	// 尝试解析CAN消息
	fmt.Println("\n-----------------------------------")
	fmt.Println("尝试解析CAN消息:")
	r.parseCANMessages(decompressedData)

	return nil
}

// parseCANMessages 解析CAN消息
func (r *RacelogicREFReader) parseCANMessages(data []byte) {
	buf := bytes.NewReader(data)
	messageCount := 0

	// 这是一个简单的尝试解析，实际格式可能需要根据数据调整
	for buf.Len() > 0 && messageCount < 20 {
		// 尝试读取可能的CAN消息结构
		var msgSize byte
		if err := binary.Read(buf, binary.LittleEndian, &msgSize); err != nil {
			break
		}

		if msgSize == 0 || msgSize > 64 {
			// 可能不是消息大小，跳过
			continue
		}

		messageCount++
		fmt.Printf("消息 #%d: 大小=%d ", messageCount, msgSize)

		// 读取消息内容
		msgData := make([]byte, msgSize)
		n, err := buf.Read(msgData)
		if err != nil || n != int(msgSize) {
			fmt.Println("读取失败")
			break
		}

		fmt.Printf("数据: % 02x\n", msgData)

		if messageCount >= 20 {
			fmt.Println("... (显示前20条消息)")
			break
		}
	}

	if messageCount == 0 {
		fmt.Println("未能识别标准CAN消息格式")
		fmt.Println("这可能需要Racelogic的专有解析工具或文档")
	}
}

// displayHex 以十六进制和ASCII格式显示数据
func displayHex(data []byte, maxBytes int) {
	if len(data) > maxBytes {
		data = data[:maxBytes]
	}

	for i := 0; i < len(data); i += 16 {
		// 显示偏移
		fmt.Printf("%08x  ", i)

		// 显示十六进制
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Print("   ")
			}
			if j == 7 {
				fmt.Print(" ")
			}
		}

		// 显示ASCII
		fmt.Print(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}
}

func main() {
	filename := "Volkswagen-Up!.REF"

	fmt.Printf("========================================\n")
	fmt.Printf("Racelogic .REF 文件读取器\n")
	fmt.Printf("========================================\n\n")
	fmt.Printf("正在读取文件: %s\n\n", filename)

	reader := NewRacelogicREFReader(filename)
	if err := reader.Parse(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n========================================")
	fmt.Println("解析完成")
	fmt.Println("========================================")
}
