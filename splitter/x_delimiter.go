package splitter

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

const (
	DataBytesDelimiter = byte('\n') // TCP通讯消息包的分割符号
)

func WrapDelimiter(delimiter byte, bytes []byte) []byte {
	return append(bytes, delimiter)
}

func WrapDelimiterDefault(bytes []byte) []byte {
	return WrapDelimiter(DataBytesDelimiter, bytes)
}

func UnwrapDelimiter(delimiter byte, bytes []byte) ([]byte, int) {
	size := len(bytes)
	if size > 0 {
		lstIdx := size - 1
		if delimiter == bytes[lstIdx] {
			// remove: '\n'
			return bytes[:lstIdx], lstIdx
		} else {
			return bytes, size
		}
	}

	return bytes, 0
}

func UnwrapDelimiterDefault(bytes []byte) []byte {
	bs, _ := UnwrapDelimiter(DataBytesDelimiter, bytes)
	return bs
}
