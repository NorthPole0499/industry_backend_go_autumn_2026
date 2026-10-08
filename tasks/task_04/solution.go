package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	var answer Stats
	if (len(nums) < 2 ) {
		return Stats{}
	}

	 
	for i := 1; i < len(nums); i++ {
		diff := nums[i] - nums[i - 1]

		answer.Count++
		answer.Sum += diff

		if diff < answer.Min || i == 1 {
			answer.Min = diff
		}
		if diff > answer.Max || i == 1 {
			answer.Max = diff
		}
	}
	return answer
}
