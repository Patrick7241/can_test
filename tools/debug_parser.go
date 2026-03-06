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

func main() {
	filename := "Volkswagen-ID.4 2020-.REF"

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 跳过文本头
	reader.ReadString('\n')
	reader.ReadString('\n')

	// 读取所有二进制数据
	binaryData, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("二进制数据大小: %d 字节\n\n", len(binaryData))

	// 解析所有块
	offset := 0
	blockNum := 0

	for offset < len(binaryData)-2 {
		// 读取块大小（2字节大端序）
		blockSize := binary.BigEndian.Uint16(binaryData[offset : offset+2])
		offset += 2

		if blockSize == 0 || offset+int(blockSize) > len(binaryData) {
			break
		}

		blockNum++
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("块 #%d\n", blockNum)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("偏移: 0x%04X\n", offset-2)
		fmt.Printf("压缩大小: %d 字节\n", blockSize)

		// 提取块数据
		blockData := binaryData[offset : offset+int(blockSize)]
		offset += int(blockSize)

		// 尝试解压
		if blockSize > 2 && blockData[0] == 0x78 {
			r, err := zlib.NewReader(bytes.NewReader(blockData))
			if err == nil {
				decompressed, err := io.ReadAll(r)
				r.Close()
				if err == nil {
					fmt.Printf("解压后大小: %d 字节\n", len(decompressed))
					fmt.Printf("内容: %s\n\n", string(decompressed))
				} else {
					fmt.Printf("解压失败: %v\n\n", err)
				}
			} else {
				fmt.Printf("zlib初始化失败: %v\n\n", err)
			}
		} else {
			fmt.Printf("非压缩数据: % 02x\n\n", blockData)
		}
	}

	fmt.Printf("总共解析了 %d 个块\n", blockNum)
}
