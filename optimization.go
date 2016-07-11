package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"fmt"
	"time"
	"strconv"
	"math/rand"
	"math"
)

var GDestination string = "LGA"

var GPeople [][2]string = [][2]string{
	[2]string{"Seymour","BOS"},
	[2]string{"Franny","DAL"},
	[2]string{"Zooey","CAK"},
	[2]string{"Walt","MIA"},
	[2]string{"Buddy","ORD"},
	[2]string{"Les","OMA"},
}

var GFlight map[[2]string][][3]string//interface{}里面是[3]string


func getMinutes(t string) int64 {
	tt, err := time.Parse("15:04", t)
	if err != nil {
		fmt.Printf("time.Parse error: %v\n", err)
		return 0
	}
	return int64(tt.Hour() * 60 + tt.Minute())
}

// 从txt文件中解析航班信息
func parseShecdule(fileName string, flight *map[[2]string][][3]string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		//fmt.Println(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		lineList := strings.Split(line, ",")
		origin, dest := lineList[0], lineList[1]
		depart, arrive, price := lineList[2], lineList[3], lineList[4]
		if v, ok := (*flight)[[2]string{origin, dest}]; ok {
			(*flight)[[2]string{origin, dest}] = append(v, [3]string{depart, arrive, price})
		} else {
			(*flight)[[2]string{origin, dest}] = [][3]string{[3]string{depart, arrive, price}}
		}
	}
	return nil
}

//打印选定日程表的航班信息
func printSchedule(schedule []int)  {
	for i := 0; i < len(schedule)/2; i++ {
		name := GPeople[i][0]
		origin := GPeople[i][1]

		out := GFlight[[2]string{origin, GDestination}][schedule[i]]
		ret := GFlight[[2]string{GDestination, origin}][schedule[i+1]]

		fmt.Printf("%10s%10s %5s-%5s $%3s %5s-%5s $%3s\n", name, origin,
			out[0], out[1], out[2], ret[0], ret[1], ret[2])
	}
}

//计算选定日程表的成本
func costSchedule(schedule []int) int64 {
	var totalPrice, totalWait, latestArrive, eariestDepart int64
	eariestDepart = 24*60

	for i := 0; i < len(schedule)/2; i++ {
		origin := GPeople[i][1]
		out := GFlight[[2]string{origin, GDestination}][schedule[i]]
		ret := GFlight[[2]string{GDestination, origin}][schedule[i+1]]

		if price, err := strconv.ParseInt(out[2], 10, 64); err != nil {
			fmt.Printf("strconv.ParseInt error: %v\n", err)
		} else {
			totalPrice += price
		}
		if price, err := strconv.ParseInt(ret[2], 10, 64); err != nil {
			fmt.Printf("strconv.ParseInt error: %v\n", err)
		} else {
			totalPrice += price
		}
		//Track the latest arrival and earliest departure
		if latestArrive < getMinutes(out[1]) {
			latestArrive = getMinutes(out[1])
		}
		if eariestDepart > getMinutes(ret[0]) {
			eariestDepart = getMinutes(ret[0])
		}
	}

	for i := 0; i < len(schedule)/2; i++ {
		origin := GPeople[i][1]
		out := GFlight[[2]string{origin, GDestination}][schedule[i]]
		ret := GFlight[[2]string{GDestination, origin}][schedule[i+1]]

		totalWait += (latestArrive - getMinutes(out[1]))
		totalWait += (getMinutes(ret[1]) - eariestDepart)
	}

	//租车超过一天的额外费用
	if latestArrive < eariestDepart {
		totalPrice += 50
	}

	return totalPrice + totalWait
}

//随机优化
func optimizeRandom(domainList [][2]int, costF func([]int) int64) []int {
	var maxCost int64 = 999999999
	var bestSchedule []int

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 1000; i++ {
		//fmt.Println("i", i)
		schedule := make([]int, 0)
		for _, domain := range domainList {
			//fmt.Println(domain)
			choice := r.Intn(domain[1] - domain[0]) + domain[0]
			schedule = append(schedule, choice)
		}
		//fmt.Println(schedule)

		cost := costF(schedule)
		if cost < maxCost {
			maxCost = cost
			bestSchedule = schedule
		}
	}

	return bestSchedule
}

//登山法优化
func optimizeHillClimb(domainList [][2]int, costF func([]int) int64) []int {
	var bestSchedule []int

	fmt.Println(domainList)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, domain := range domainList {
		choice := r.Intn(domain[1] - domain[0]) + domain[0]
		bestSchedule = append(bestSchedule, choice)
	}

	//Main loop
	for {
		//Create list of neighboring solutions
		var neighbors [][]int
		fmt.Println("start", bestSchedule)

		for i := 0; i < len(domainList); i ++ {
			//One away in each direction
			if bestSchedule[i] > domainList[i][0] {
				newSchedule := make([]int, len(bestSchedule))
				copy(newSchedule, bestSchedule)
				newSchedule[i] = bestSchedule[i]-1
				neighbors = append(neighbors, newSchedule)
			}
			if bestSchedule[i] < domainList[i][1] {
				newSchedule := make([]int, len(bestSchedule))
				copy(newSchedule, bestSchedule)
				newSchedule[i] = bestSchedule[i]+1
				neighbors = append(neighbors, newSchedule)
			}
		}

		//See what the best solution amongst the neighbors is
		current := costF(bestSchedule)
		best := current
		for j := 0; j < len(neighbors); j ++ {
			cost := costF(neighbors[j])
			//fmt.Println("neighbors", neighbors[j], cost)
			if cost < best {
				best = cost
				bestSchedule = neighbors[j]
			}
		}

		//If there's no improvement, then we've reached the top
		if best == current {
			break
		}
		fmt.Println("best", bestSchedule)
	}

	return bestSchedule
}

//退火法优化
func optimizeAnealing(domainList [][2]int, costF func([]int) int64, t, cool float64, step int) []int {
	var bestSchedule []int

	fmt.Println(domainList)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, domain := range domainList {
		choice := r.Intn(domain[1] - domain[0]) + domain[0]
		bestSchedule = append(bestSchedule, choice)
	}

	//Main loop
	for (t > 0.1) {
		//Choose one of the indices
		i := r.Intn(len(domainList))

		//Choose a direction to change it
		dir := r.Intn(2*step) - step

		//Create a new list with one of the values changed
		newSchedule := make([]int, len(bestSchedule))
		copy(newSchedule, bestSchedule)
		newSchedule[i] += dir
		if newSchedule[i] < domainList[i][0] {newSchedule[i] = domainList[i][0]}
		if newSchedule[i] > domainList[i][1] {newSchedule[i] = domainList[i][1]}

		//Calculate the current cost and the new cost
		best := costF(bestSchedule)
		cost := costF(newSchedule)
		p := math.Exp(float64(-(cost-best))/t)
		if cost < best || r.Float64() < p {
			bestSchedule = newSchedule
		}

		t = t*cool
	}

	return bestSchedule
}

func main() {

	GFlight = make(map[[2]string][][3]string)
	err := parseShecdule("/Users/baixiao/Go/src/test/optimization/schedule.txt", &GFlight)
	if err != nil {
		fmt.Printf("parseShecdule error: %v\n", err)
	}

	for k, v := range GFlight {
		fmt.Printf("flight %v: %v\n", k, v)
	}

	//打印看看
	//var schedule []int = []int{1, 4, 3, 2, 7, 3, 6, 3, 2, 4, 5, 3}//依次表示每个人选择的往返航班的编号
	//fmt.Println("origin schedule", schedule)
	//printSchedule(schedule)
	//fmt.Println("origin cost", costSchedule(schedule))

	domainList := make([][2]int, 0)
	for i := 0; i < len(GPeople); i++ {
		domainList = append(domainList, [2]int{0, 9})
		domainList = append(domainList, [2]int{0, 9})
	}

	var bestSchedule []int
	//随机优化
	//bestSchedule = optimizeRandom(domainList, costSchedule)
	//fmt.Println("random schedule", bestSchedule)
	//printSchedule(bestSchedule)
	//fmt.Println("random cost", costSchedule(bestSchedule))

	//登山法优化
	//bestSchedule = optimizeHillClimb(domainList, costSchedule)
	//fmt.Println("hillclimb schedule", bestSchedule)
	//printSchedule(bestSchedule)
	//fmt.Println("hillclimb cost", costSchedule(bestSchedule))

	//退火法优化
	bestSchedule = optimizeAnealing(domainList, costSchedule, 1000, 0.95, 1)
	fmt.Println("anealing schedule", bestSchedule)
	printSchedule(bestSchedule)
	fmt.Println("anealing cost", costSchedule(bestSchedule))

}