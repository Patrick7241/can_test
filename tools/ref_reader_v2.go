package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func main() {
	filename := "ref_files/Volkswagen-ID.4 2020-.REF"

	fmt.Printf("========================================\n")
	fmt.Printf("Racelogic .REF 文件分析器 V2\n")
	fmt.Printf("========================================\n\n")

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 读取并显示文本头部
	fmt.Println("文件头信息:")
	fmt.Println("-----------------------------------")

	// 第一行
	line1, _ := reader.ReadString('\n')
	fmt.Printf("%s", line1)

	// 第二行
	line2, _ := reader.ReadString('\n')
	fmt.Printf("%s", line2)

	// 读取剩余的二进制数据
	binaryData, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取二进制数据失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n二进制数据大小: %d 字节\n\n", len(binaryData))

	// 显示完整的十六进制dump
	fmt.Println("完整二进制数据 (十六进制):")
	fmt.Println("-----------------------------------")
	displayHexDump(binaryData)

	// 尝试查找并解压所有zlib块
	fmt.Println("\n\n寻找并解压zlib压缩块:")
	fmt.Println("-----------------------------------")
	decompressAllZlibBlocks(binaryData)

	fmt.Println("\n========================================")
	fmt.Println("分析完成")
	fmt.Println("========================================")
}

// displayHexDump 显示十六进制dump
func displayHexDump(data []byte) {
	for i := 0; i < len(data); i += 16 {
		// 偏移
		fmt.Printf("%04x:  ", i)

		// 十六进制
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

		// ASCII
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

// decompressAllZlibBlocks 查找并解压所有zlib块
func decompressAllZlibBlocks(data []byte) {
	// zlib魔术数字可能的值
	zlibHeaders := []byte{0x78, 0x01, 0x78, 0x9C, 0x78, 0xDA, 0x78, 0x5E}

	offset := 0
	blockNum := 0

	for offset < len(data)-1 {
		// 检查是否是zlib头
		isZlibHeader := false
		for i := 0; i < len(zlibHeaders); i += 2 {
			if data[offset] == zlibHeaders[i] && data[offset+1] == zlibHeaders[i+1] {
				isZlibHeader = true
				break
			}
		}

		if isZlibHeader {
			blockNum++
			fmt.Printf("\n块 #%d (偏移 0x%04x):\n", blockNum, offset)

			// 尝试解压，从当前位置开始尝试不同的长度
			maxTryLen := len(data) - offset
			if maxTryLen > 512 {
				maxTryLen = 512
			}

			decompressed, consumed := tryDecompressZlib(data[offset:])
			if decompressed != nil {
				fmt.Printf("  压缩数据: %d 字节\n", consumed)
				fmt.Printf("  解压后: %d 字节\n", len(decompressed))
				fmt.Printf("  内容: ")

				// 显示内容
				if isPrintable(decompressed) {
					fmt.Printf("\"%s\"\n", string(decompressed))
				} else {
					fmt.Printf("% 02x\n", decompressed)
				}

				offset += consumed
			} else {
				offset++
			}
		} else {
			offset++
		}
	}

	if blockNum == 0 {
		fmt.Println("未找到zlib压缩块")
	}
}

// tryDecompressZlib 尝试解压zlib数据
func tryDecompressZlib(data []byte) ([]byte, int) {
	// 尝试解压
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, 0
	}
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		return nil, 0
	}

	// 估算消耗的字节数（这是个简化的方法）
	// 更精确的方法需要跟踪zlib reader的位置
	consumed := len(data)
	for tryLen := 10; tryLen < len(data); tryLen++ {
		r2, err := zlib.NewReader(bytes.NewReader(data[:tryLen]))
		if err == nil {
			test, err := io.ReadAll(r2)
			r2.Close()
			if err == nil && bytes.Equal(test, decompressed) {
				consumed = tryLen
				break
			}
		}
	}

	return decompressed, consumed
}

// isPrintable 检查数据是否可打印
func isPrintable(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	printableCount := 0
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			printableCount++
		}
	}
	return float64(printableCount)/float64(len(data)) > 0.7
}
