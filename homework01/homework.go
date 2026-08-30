package homework01

import (
	"sort"
	"strings"
)

// 1. 只出现一次的数字
// 给定一个非空整数数组，除了某个元素只出现一次以外，其余每个元素均出现两次找出那个只出现了一次的元素
func SingleNumber(nums []int) int {
	// 相同数字异或后为 0，0 与只出现一次的数字异或后仍为该数字
	// 因此所有成对数字会相互抵消，最后留下只出现一次的数字
	result := 0
	for _, num := range nums {
		result ^= num
	}
	return result
}

// 2. 回文数
// 判断一个整数是否是回文数
func IsPalindrome(x int) bool {
	// 负数和末尾为 0 的非零数不可能是回文数
	if x < 0 || (x%10 == 0 && x != 0) {
		return false
	}
	// 只反转后一半数字，既减少计算，也避免完整反转可能发生的整数溢出
	reversed := 0
	for x > reversed {
		reversed = reversed*10 + x%10
		x /= 10
	}
	// 偶数位时两半应完全相等；奇数位时忽略反转部分的中间数字
	return x == reversed || x == reversed/10
}

// 3. 有效的括号
// 给定一个只包括 '(', ')', '{', '}', '[', ']' 的字符串，判断字符串是否有效
func IsValid(s string) bool {
	// 使用栈保存左括号，遇到右括号时检查栈顶是否为对应的左括号
	stack := make([]rune, 0)
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	for _, char := range s {
		switch char {
		case '(', '[', '{':
			stack = append(stack, char)
		case ')', ']', '}':
			if len(stack) == 0 || stack[len(stack)-1] != pairs[char] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	return len(stack) == 0
}

// 4. 最长公共前缀
// 查找字符串数组中的最长公共前缀
func LongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	// 先假设第一个字符串是公共前缀，再逐个缩短它
	prefix := strs[0]
	for _, str := range strs[1:] {
		for !strings.HasPrefix(str, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

// 5. 加一
// 给定一个由整数组成的非空数组所表示的非负整数，在该数的基础上加一
func PlusOne(digits []int) []int {
	// 从最低位开始加一，当前位小于 9 时不再产生进位
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] < 9 {
			digits[i]++
			return digits
		}
		digits[i] = 0
	}
	// 所有位都是 9 时，最高位需要新增一个 1
	return append([]int{1}, digits...)
}

// 6. 删除有序数组中的重复项
// 给你一个有序数组 nums ，请你原地删除重复出现的元素，使每个元素只出现一次，返回删除后数组的新长度
// 不要使用额外的数组空间，你必须在原地修改输入数组并在使用 O(1) 额外空间的条件下完成
func RemoveDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	// slow 指向下一个不重复元素应该写入的位置
	slow := 1
	for fast := 1; fast < len(nums); fast++ {
		if nums[fast] != nums[slow-1] {
			nums[slow] = nums[fast]
			slow++
		}
	}
	return slow
}

// 7. 合并区间
// 以数组 intervals 表示若干个区间的集合，其中单个区间为 intervals[i] = [starti, endi]
// 请你合并所有重叠的区间，并返回一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间
func Merge(intervals [][]int) [][]int {
	if len(intervals) == 0 {
		return [][]int{}
	}
	// 先按区间起点排序，这样只需要比较当前区间和结果末尾区间
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	merged := make([][]int, 0, len(intervals))
	merged = append(merged, []int{intervals[0][0], intervals[0][1]})
	for _, interval := range intervals[1:] {
		last := merged[len(merged)-1]
		if interval[0] <= last[1] {
			if interval[1] > last[1] {
				last[1] = interval[1]
			}
		} else {
			merged = append(merged, []int{interval[0], interval[1]})
		}
	}
	return merged
}

// 8. 两数之和
// 给定一个整数数组 nums 和一个目标值 target，请你在该数组中找出和为目标值的那两个整数
func TwoSum(nums []int, target int) []int {
	// 逐一尝试两个不同下标，找到满足条件的一对后立即返回
	for i := 0; i < len(nums); i++ {
		for j := i + 1; j < len(nums); j++ {
			if nums[i]+nums[j] == target {
				return []int{i, j}
			}
		}
	}
	// 题目通常保证存在答案；没有答案时返回 nil
	return nil
}
