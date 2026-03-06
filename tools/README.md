# REF文件解析工具集

用于解析Racelogic格式CAN信号定义文件(.REF)的工具集。

## 工具列表

### 1. ref_reader_v2.go
**推荐使用** - 最完整的REF文件解析器

功能：
- 自动识别并解压所有zlib压缩块
- 显示完整的十六进制dump
- 提取所有信号的CSV定义
- 详细的输出格式

使用方法：
```bash
cd can_test
go run tools/ref_reader_v2.go
```

输出示例：
```
块 #1 (偏移 0x0002):
  压缩数据: 11 字节
  解压后: 8 字节
  内容: "00000000"

块 #2 (偏移 0x0011):
  压缩数据: 79 字节
  解压后: 76 字节
  内容: "Indicated_Vehicle_Speed_kph,253,km/h,32,16,0,0.01,655.35,0,Unsigned,Intel,8,"
```

### 2. ref_reader.go
早期版本的REF解析器

功能：
- 基础的REF文件读取
- CSV格式信号解析
- 简单的信号含义分析

### 3. ref_parser.go
通用REF解析器（需要指定文件名）

功能：
- 解析单个REF文件
- 结构化的信号定义输出
- 支持批量解析

使用方法：
```bash
# 修改文件名后运行
cd tools
go run ref_parser.go
```

### 4. id4_parser.go
专门用于大众ID.4的解析器

功能：
- 针对ID.4优化的解析逻辑
- 自动信号分类
- 中文信号说明

### 5. debug_parser.go
调试专用解析器

功能：
- 显示每个数据块的详细信息
- 帮助理解REF文件结构
- 用于开发新的解析器

使用方法：
```bash
cd tools
go run debug_parser.go
```

## REF文件格式说明

### 文件结构

```
[文本头部]
Racelogic Can Data File V1a
Unit serial number : 00000000

[二进制数据块]
块1: [2字节长度][zlib压缩数据]
块2: [2字节长度][zlib压缩数据]
...
```

### 块格式

每个块的结构：
- **2字节**: 块大小（大端序，Big-Endian）
- **N字节**: 压缩数据（zlib格式）

### 压缩数据内容

解压后通常是CSV格式的信号定义：
```csv
信号名,CAN_ID,单位,起始位,长度,最小值,缩放,最大值,偏移,有符号,字节序,通道
```

示例：
```csv
Indicated_Vehicle_Speed_kph,253,km/h,32,16,0,0.01,655.35,0,Unsigned,Intel,8
```

字段说明：
1. **信号名**: 信号的唯一标识符
2. **CAN_ID**: CAN消息ID（十进制）
3. **单位**: 物理单位（km/h, °C, %等）
4. **起始位**: 数据在CAN帧中的起始位置
5. **长度**: 数据位数
6. **最小值**: 原始值的最小值
7. **缩放**: 转换缩放因子
8. **最大值**: 实际物理值的最大值
9. **偏移**: 转换偏移量
10. **有符号**: Signed/Unsigned
11. **字节序**: Intel(小端)/Motorola(大端)
12. **通道**: CAN通道号

## 解析流程

```
REF文件
  ↓
读取文本头
  ↓
读取二进制数据
  ↓
按块解析（2字节长度 + 数据）
  ↓
zlib解压缩
  ↓
CSV解析
  ↓
信号定义结构体
```

## 使用示例

### 解析新的REF文件

1. 将REF文件放到合适的位置
2. 使用ref_reader_v2.go查看内容：

```bash
# 修改ref_reader_v2.go中的文件路径
filename := "path/to/your-file.REF"

# 运行
go run tools/ref_reader_v2.go
```

3. 根据输出创建信号定义

### 添加新车型

1. 解析REF文件获取所有信号
2. 在`vehicles/`下创建新目录，如`vehicles/tesla_model3/`
3. 创建`signals.go`定义信号结构
4. 创建`demo.go`演示使用
5. 添加README文档

## zlib压缩

REF文件使用zlib压缩来节省空间。常见的zlib头部字节：
- `78 01` - 无压缩/低压缩
- `78 5E` - 快速压缩
- `78 9C` - 默认压缩
- `78 DA` - 最佳压缩

## 常见问题

### Q: 解析失败怎么办？
A:
1. 检查文件是否完整
2. 确认文件格式是Racelogic V1a
3. 使用debug_parser.go查看详细信息

### Q: 如何确定字节序？
A: REF文件中会明确标注Intel或Motorola

### Q: 信号ID冲突怎么办？
A: 同一个CAN ID下可能有多个信号，需要根据起始位和长度区分

### Q: 转换公式不准确？
A:
1. 检查Scale和Offset是否正确
2. 注意有符号/无符号的差异
3. 参考车辆技术文档验证

## 开发新解析器

如果需要支持其他格式的CAN数据库文件：

1. 研究目标格式的文件结构
2. 参考ref_reader_v2.go的实现
3. 实现解析逻辑
4. 添加测试用例
5. 更新文档

支持的格式扩展方向：
- [ ] DBC格式（Vector CANdb++）
- [ ] ARXML格式（AUTOSAR）
- [ ] JSON格式（自定义）
- [ ] XML格式（通用）

## 贡献

欢迎贡献新的解析器或改进现有工具！

提交前请确保：
1. 代码格式化（`go fmt`）
2. 添加必要的注释
3. 更新相关文档
4. 提供测试用例

## 许可

本工具集仅用于学习和研究目的。
