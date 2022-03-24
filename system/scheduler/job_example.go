package scheduler

import "time"

func getExampleJobCMD() {
	//fmt.Println("")
}

func getExampleJob1() Job {
	var j = Job{
		Name:     "j1",
		Interval: 1 * time.Second,
		CMD:      getExampleJobCMD,
	}
	return j
}

func getExampleJob2() Job {
	var j = Job{
		Name:     "j2",
		Interval: 2 * time.Second,
		CMD:      getExampleJobCMD,
	}
	return j
}

func getExampleJob3() Job {
	var j = Job{
		Name:     "j3",
		Interval: 3 * time.Second,
		CMD:      getExampleJobCMD,
	}
	return j
}
