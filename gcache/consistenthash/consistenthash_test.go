package consistenthash

import (
	"strconv"
	"testing"
)

// TestHashing 测试哈希算法的一致性和正确性。
// 它创建一个带有自定义哈希函数的哈希环，向环中添加节点，并验证哈希结果是否符合预期。
func TestHashing(t *testing.T) {
    // 初始化一个哈希环，设置副本因子为3，并使用自定义的哈希函数
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

    // 根据上述哈希函数，这将生成具有以下“哈希值”的副本：
    // 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

    // 定义测试用例，键为输入，值为预期输出
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

    // 验证每个测试用例的哈希结果是否符合预期
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("查询 %s 时，应返回 %s", k, v)
		}
	}

    // 添加新的节点
	hash.Add("8")

    // 更新测试用例，27 应该映射到新的节点 8
	testCases["27"] = "8"

    // 再次验证每个测试用例的哈希结果是否符合预期
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("查询 %s 时，应返回 %s", k, v)
		}
	}
}