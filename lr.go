// AUTHOR: LGT
// START  TIME: 2020-11-22 15:09
// FINISH TIME: 2020-11-25 10:21

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type expMap struct {
	start  rune
	subExp string
}

type source struct {
	from  int
	shift rune
}

type solu struct {
	sources []source
	list    []expMap
	isTran  []bool
}

type place struct {
	x int
	y int
}

type table struct {
	action [][]string
	goTo   [][]int
}

type stack struct {
	state  []int
	symbol []string
}

func push(st stack, sta int, sym string) stack {
	st.state = append(st.state, sta)
	st.symbol = append(st.symbol, sym)
	return st
}

func pop(st stack) stack {
	st.state = st.state[:len(st.state)-1]
	st.symbol = st.symbol[:len(st.symbol)-1]
	return st
}

func peek(st stack) (int, string) {
	return st.state[len(st.state)-1], st.symbol[len(st.state)-1]
}

func isEmpty(st stack) bool {
	if len(st.state) == 0 {
		return true
	} else {
		return false
	}
}

func subPeekSta(st stack, popLen int) int {
	return st.state[len(st.state)-(popLen+1)]
}

type doline struct {
	no int
	st stack
	s  string
	do string
}

var (
	oriBegin    rune
	begin       rune
	beginSubExp string
	vCnt        int
	tCnt        int
	vs          map[int]rune
	ts          map[int]rune
	exps        []expMap
	first       map[rune][]rune
	follow      map[rune][]rune
	flag        map[rune]bool
	solus       []solu
	aTable      table
	inputArr    []string
	cInput      string
	aDoTable    []doline
	errIndex    int
)

func main() {
	var (
		gramName  string
		inputName string
	)
	fmt.Print("请输入记录文法串的文件名（带文件扩展名）：")
	fmt.Scanf("%s", &gramName)

	initialize()

	readGrammar(gramName)

	outputGrammar()

	firstAndFollow()

	getClosure()

	getTable()

	fmt.Print("请输入要进行语法分析的文件名（带文件扩展名）：")
	fmt.Scanf("%s", &inputName)
	readInput(inputName)

	analysis()
}

func analysis() {
	for key, value := range inputArr {
		cInput = value
		err, index := getDoTable(value)
		if err != nil {
			fmt.Printf("第%d个", key+1)
			fmt.Println(err)
			fmt.Println(value)
			for i := 0; i < errIndex; i++ {
				fmt.Print(" ")
			}
			fmt.Println("^")
			aDoTable = aDoTable[index-1:]
		}
	}
}

func getDoTable(input string) (error, int) {
	orInputLen := len(input)
	input = input + "$"
	iDoline := getADoline(input, 1)

	distNum := func(offset int, term string, isNum bool) (error, int) {
		state, _ := peek(iDoline.st)
		if isNum {
			iDoline.do = aTable.action[state][mapShift('n', true)]
		} else {
			iDoline.do = aTable.action[state][mapShift([]rune(term)[0], true)]
		}
		if iDoline.do == " " {
			err := errors.New("待约串不能由此文法推导出来！")
			errIndex = orInputLen - len(input)
			return err, iDoline.no
		}
		iDoline.s = input
		aDoline := getADoline(input, iDoline.no)
		for i := 1; i < len(iDoline.st.state); i++ {
			aDoline.st = push(aDoline.st, iDoline.st.state[i], iDoline.st.symbol[i])
		}
		aDoline.do = iDoline.do
		aDoTable = append(aDoTable, aDoline)
		iDoline.no++
		if iDoline.do != "ACC" {
			sOrR := iDoline.do[:1]
			nextSta, _ := strconv.Atoi(iDoline.do[1:])
			if sOrR == "S" {
				iDoline.st = push(iDoline.st, nextSta, term)
				input = input[offset:]
			} else {
				expNo := nextSta
				nTopSymCh := exps[expNo].start
				popLen := len(exps[expNo].subExp)
				nTopSym := string(nTopSymCh)
				goToLine := subPeekSta(iDoline.st, popLen)
				nTopSta := aTable.goTo[goToLine][mapShift(nTopSymCh, false)]
				for i := 0; i < popLen; i++ {
					iDoline.st = pop(iDoline.st)
				}
				iDoline.st = push(iDoline.st, nTopSta, nTopSym)
			}
		} else {
			for !isEmpty(iDoline.st) {
				iDoline.st = pop(iDoline.st)
			}
		}
		return nil, -1
	}

	for !isEmpty(iDoline.st) {
		var (
			err   error
			index int
		)
		if len(input) > 1 {
			offset, term, kind := cut(input)
			if kind == "num" {
				err, index = distNum(offset, term, true)
			} else {
				err, index = distNum(offset, term, false)
			}
			if err != nil {
				return err, index
			}
		} else {
			err, index = distNum(0, "$", false)
			if err != nil {
				return err, index
			}
		}
	}

	printDoTable()
	fmt.Println()
	return nil, -1
}

func getADoline(input string, no int) doline {
	var aDoLine doline
	aDoLine.st.state = make([]int, 0)
	aDoLine.st.symbol = make([]string, 0)
	aDoLine.st = push(aDoLine.st, 0, "$")
	aDoLine.no = no
	aDoLine.s = input
	return aDoLine
}

func cut(s string) (int, string, string) {
	var (
		offset int
		res    string
		kind   string
	)
	chars := []rune(s)
	c := chars[0]
	if unicode.IsDigit(c) {
		if len(s) != 1 {
			for offset = 1; unicode.IsDigit(chars[offset]) || chars[offset] == '.'; offset++ {
			}
		} else {
			offset = 1
		}
		kind = "num"
		res = s[:offset]
	} else if unicode.IsLower(c) {
		offset = 1
		kind = "alpha"
		res = s[:offset]
	} else if unicode.IsSymbol(c) || c == '-' || c == '(' || c == ')' || c == '*' || c == '/' {
		offset = 1
		kind = "symbol"
		res = s[:offset]
	} else {
		offset = -1
		kind = "unknown"
		res = s
	}
	return offset, res, kind
}

func readInput(fName string) {
	file, err := os.Open(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		input := scanner.Text()
		inputArr = append(inputArr, input)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func initialize() {
	exps = make([]expMap, 0)
	solus = make([]solu, 0)
	vs = make(map[int]rune)
	ts = make(map[int]rune)
	aTable.action = make([][]string, 0)
	aTable.goTo = make([][]int, 0)
	inputArr = make([]string, 0)
}

func readGrammar(fName string) {
	file, err := os.Open(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		iExp := scanner.Text()
		if i == 0 {
			begin = '@'
			oriBegin = []rune(strings.Split(iExp, "->")[0])[0]
			beginSubExp = string(oriBegin)
			beginExp := make([]rune, 0)
			beginExp = append(beginExp, begin)
			beginExp = append(beginExp, '-')
			beginExp = append(beginExp, '>')
			beginExp = append(beginExp, oriBegin)
			getVT(string(beginExp))
		}
		i++
		getVT(iExp)
	}
	ts[tCnt] = '$'
	tCnt++

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func outputGrammar() {
	fmt.Println("非终结符：")
	printChar(vs)
	fmt.Println("终结符：")
	printChar(ts)
	fmt.Println("输入文法串：")
	printExpMap(exps)
}

func getTable() {
	for i := 0; i < len(solus); i++ {
		lineA := make([]string, 0)
		lineG := make([]int, 0)
		aTable.action = append(aTable.action, lineA)
		aTable.goTo = append(aTable.goTo, lineG)
	}
	for i := 0; i < len(solus); i++ {
		for j := 0; j < tCnt; j++ {
			aTable.action[i] = append(aTable.action[i], " ")
		}
		for j := 0; j < vCnt; j++ {
			aTable.goTo[i] = append(aTable.goTo[i], 0)
		}
	}
	for {
		isE, p := isEnd(true)
		if !isE {
			break
		} else {
			from := p.x
			solus[p.x].isTran[p.y] = false
			sE := solus[p.x].list[p.y].start
			subExp := solus[p.x].list[p.y].subExp
			chars := []rune(subExp)
			index := strings.Index(subExp, ".")
			if index == len(subExp)-1 {
				if chars[index-1] == oriBegin {
					aTable.action[from][mapShift('$', true)] = "ACC"
				} else {
					aTable.action[from][mapShift('$', true)] = ("R" + strconv.Itoa(mapSubExp(subExp)))
					for _, value := range follow[sE] {
						aTable.action[from][mapShift(value, true)] = ("R" + strconv.Itoa(mapSubExp(subExp)))
					}
				}
			} else {
				shift := []rune(subExp)[index+1]
				to := findTo(from, shift)
				if isExist(shift, ts) {
					aTable.action[from][mapShift(shift, true)] = ("S" + strconv.Itoa(to))
				} else {
					aTable.goTo[from][mapShift(shift, false)] = to
				}
			}
		}
	}
	printTable()
}

func mapSubExp(subExp string) int {
	var res int
	subExp = subExp[:len(subExp)-1]
	for key, value := range exps {
		if value.subExp == subExp {
			res = key
			break
		}
	}
	return res
}

func mapShift(shift rune, isT bool) int {
	var res int
	if isT {
		for key, value := range ts {
			if value == shift {
				res = key
				break
			}
		}
	} else {
		for key, value := range vs {
			if value == shift {
				res = key
				break
			}
		}
	}
	return res
}

func findTo(from int, shift rune) int {
	var res int
	for key, aSolu := range solus {
		for _, aSource := range aSolu.sources {
			if from == aSource.from && shift == aSource.shift {
				res = key
			}
		}
	}
	return res
}

func getClosure() {
	iMap := expMap{begin, addDot(0, beginSubExp)}
	iSolu := closure(iMap)
	signs := make([]bool, 0)
	for i := 0; i < len(iSolu); i++ {
		signs = append(signs, false)
	}
	iSource := make([]source, 0)
	iSource = append(iSource, source{-1, '@'})
	solus = append(solus, solu{iSource, iSolu, signs})
	for {
		isE, p := isEnd(false)
		if !isE {
			break
		} else {
			subExp := solus[p.x].list[p.y].subExp
			index := strings.Index(subExp, ".")
			if index == len(subExp)-1 {
				solus[p.x].isTran[p.y] = true
			} else {
				shift := []rune(subExp)[index+1]
				from := p.x
				list := goShift(shift, &solus[p.x])
				isEx, i := isSoluExist(list)
				if !isEx {
					flags := make([]bool, 0)
					for range list {
						flags = append(flags, false)
					}
					cSource := make([]source, 0)
					cSource = append(cSource, source{from, shift})
					solus = append(solus, solu{cSource, list, flags})
				} else {
					solus[i].sources = append(solus[i].sources, source{from, shift})
				}
			}
		}
	}
	printClosure()
}

func isSoluExist(list []expMap) (bool, int) {
	var (
		res   bool
		index = -1
	)
	for key, value := range solus {
		if len(value.list) == len(list) {
			cnt := 0
			for key, exp := range value.list {
				if exp == list[key] {
					cnt++
				}
			}
			if cnt == len(list) {
				res = true
				index = key
				break
			}
		}
	}
	return res, index
}

func isEnd(isBack bool) (bool, place) {
	var (
		res1 bool
		res2 place
	)
	for x, aSolu := range solus {
		for y, value := range aSolu.isTran {
			if isBack {
				if value {
					res1 = true
					res2.x = x
					res2.y = y
					return res1, res2
				}
			} else {
				if !value {
					res1 = true
					res2.x = x
					res2.y = y
					return res1, res2
				}
			}
		}
	}
	return res1, res2
}

func goShift(sE rune, pSolu *solu) []expMap {
	var tmp []expMap
	for key, value := range pSolu.list {
		subExp := value.subExp
		chars := []rune(subExp)
		index := strings.Index(subExp, ".")
		if index < len(subExp)-1 {
			if chars[index+1] == sE {
				moved := moveDot(index, subExp)
				pSolu.isTran[key] = true
				tmp = append(tmp, expMap{value.start, moved})
			}
		}
	}
	var res []expMap
	for _, value := range tmp {
		res = append(res, closure(value)...)
	}
	return res
}

func closure(iMap expMap) []expMap {
	res := make([]expMap, 0)
	res = append(res, iMap)
	cExp := iMap.subExp
	i := strings.Index(cExp, ".")
	if i < len(cExp)-1 {
		nextMap := getNextMap([]rune(cExp)[i+1])
		for key, value := range nextMap {
			nextMap[key].subExp = addDot(0, value.subExp)
		}
		// 遍历产生式集合确定是否需要继续添加
		for _, value := range nextMap {
			subExp := value.subExp
			index := strings.Index(subExp, ".")
			chars := []rune(subExp)
			if index < len(subExp)-1 {
				if !isExist(chars[index+1], ts) {
					if chars[index+1] != value.start {
						res = append(res, closure(value)...)
					} else {
						res = append(res, value)
					}
				} else {
					res = append(res, value)
				}
			}
		}
	}
	return res
}

func getNextMap(start rune) []expMap {
	res := make([]expMap, 0)
	for _, value := range exps {
		if start == value.start {
			res = append(res, value)
		}
	}
	return res
}

func addDot(p int, oriExp string) string {
	return (oriExp[:p] + "." + oriExp[p:])
}

func moveDot(p int, oriExp string) string {
	var tmp, res string
	if p == 0 {
		tmp = oriExp[p+1:]
		if len(tmp) == 1 {
			res = tmp + "."
		} else {
			res = tmp[:p+1] + "." + tmp[p+1:]
		}
	} else if p == len(oriExp)-1 {
		res = oriExp
	} else {
		tmp = oriExp[:p] + oriExp[p+1:]
		res = tmp[:p+1] + "." + tmp[p+1:]
	}
	return res
}

func firstAndFollow() {
	first = make(map[rune][]rune)
	follow = make(map[rune][]rune)
	flag = make(map[rune]bool)

	for _, value := range vs {
		first[value] = getFirst(value)
	}
	for _, value := range vs {
		follow[value] = getFollow(value)
	}
	fmt.Println("First集：")
	printF(first)
	fmt.Println("Follow集：")
	printF(follow)
}

func getFirst(start rune) []rune {
	var res []rune
	if isExist(start, ts) {
		res = append(res, start)
	} else {
		for _, value := range exps {
			if value.start == start {
				next := []rune(value.subExp)[0]
				if next != start {
					res = append(res, getFirst(next)...)
				}
			}
		}
	}
	return res
}

func getFollow(start rune) []rune {
	var res []rune
	if begin == start {
		res = append(res, '$')
	}
	for _, value := range exps {
		subExp := []rune(value.subExp)
		for index, char := range subExp {
			if char == start {
				if len(subExp)-index == 1 {
					if len(follow[value.start]) != 0 {
						if !flag[char] {
							res = append(res, follow[value.start]...)
							flag[char] = true
						}
					} else {
						res = append(res, getFollow(value.start)...)
					}
				} else {
					if isExist(subExp[index+1], ts) {
						res = append(res, subExp[index+1])
					} else {
						res = append(res, getFollow(subExp[index+1])...)
					}
				}
			}
		}
	}
	return res
}

func isExist(c rune, cArr map[int]rune) bool {
	var res bool
	for _, value := range cArr {
		if value == c {
			res = true
			break
		}
	}
	return res
}

func getVT(iStr string) {
	strArr := strings.Split(iStr, "->")
	left, right := strArr[0], strArr[1]
	v := []rune(left)[0]
	exp := expMap{v, right}
	exps = append(exps, exp)
	if !isExist(v, vs) {
		vs[vCnt] = v
		vCnt++
	}
	for _, value := range right {
		if !unicode.IsUpper(value) && !isExist(value, ts) {
			ts[tCnt] = value
			tCnt++
		}
	}
}

func printChar(charMap map[int]rune) {
	for _, value := range charMap {
		fmt.Printf("%c ", value)
	}
	fmt.Println()
}

func printStr(strArr []string) {
	for _, value := range strArr {
		fmt.Println(value)
	}
}

func printExpMap(expArr []expMap) {
	for _, value := range expArr {
		fmt.Printf("开始元素：%c, 产生式：%s\n", value.start, value.subExp)
	}
}

func printF(f map[rune][]rune) {
	for key, value1 := range f {
		fmt.Printf("%c: ", key)
		for _, value2 := range value1 {
			fmt.Printf("%c ", value2)
		}
		fmt.Println()
	}
}

func printClosure() {
	for key, value := range solus {
		fmt.Printf("I%d:\n", key)
		for _, value1 := range value.sources {
			if value1.from != -1 {
				fmt.Printf("from: I%d, shift: %c\n", value1.from, value1.shift)
			}
		}
		for _, value1 := range value.list {
			fmt.Printf("%c->%s\n", value1.start, value1.subExp)
		}
	}
}

func printTable() {
	fmt.Printf("\nSLR(1)文法分析表：\n")
	// exps
	for key, value := range exps {
		fmt.Printf("(%d)%c->%s ", key, value.start, value.subExp)
	}
	// head
	fmt.Print("\nState\t")
	for i := 0; i < tCnt/2-1; i++ {
		fmt.Print("\t")
	}
	fmt.Print("action")
	for i := 0; i < tCnt/2+vCnt/2; i++ {
		fmt.Print("\t")
	}
	fmt.Print("goto")
	for i := 0; i < vCnt/2; i++ {
		fmt.Print("\t")
	}
	// sub-head
	fmt.Print("\n\t")
	for i := 0; i < tCnt; i++ {
		fmt.Printf("%c\t", ts[i])
	}
	for i := 1; i < vCnt; i++ {
		fmt.Printf("%c\t", vs[i])
	}
	fmt.Println()
	// body
	for i := 0; i < len(aTable.goTo); i++ {
		fmt.Printf("I%d\t", i)
		for j := 0; j < tCnt; j++ {
			fmt.Printf("%s\t", aTable.action[i][j])
		}
		for j := 1; j < vCnt; j++ {
			if aTable.goTo[i][j] != 0 {
				fmt.Printf("%d\t", aTable.goTo[i][j])
			} else {
				fmt.Print(" \t")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func getMaxStLen() int {
	var res int
	for _, aDoline := range aDoTable {
		if len(aDoline.st.state) > res {
			res = len(aDoline.st.state)
		}
	}
	return res
}

func printDoTable() {
	stackLen := getMaxStLen() + 3
	for key, value := range exps {
		fmt.Printf("(%d)%c->%s ", key, value.start, value.subExp)
	}
	fmt.Printf("\n步骤")
	for i := 0; i < stackLen/2; i += 2 {
		fmt.Print("\t")
	}
	fmt.Print("栈")
	for i := stackLen / 2; i < stackLen; i += 2 {
		fmt.Print("\t")
	}
	fmt.Print("输入串")
	for i := 0; i < len(cInput); i += 3 {
		fmt.Print("\t")
	}
	fmt.Print("动作\n")

	for _, value := range aDoTable {
		fmt.Printf("%d\t", value.no)
		fmt.Printf("State:  ")
		offset := 0
		for _, sta := range value.st.state {
			fmt.Printf("%d ", sta)
			offset++
		}
		if offset == 3 || offset == 5 || offset == 6 {
			fmt.Print("\t")
		}
		if offset == 7 {
			fmt.Print("\t\t")
		}
		for i := 0; i < (stackLen-offset)/2; i++ {
			fmt.Print("\t")
		}
		fmt.Printf("%s", value.s)
		for i := 3; i < len(cInput); i += 3 {
			fmt.Print("\t")
		}
		if len(value.s) <= 7 {
			fmt.Print("\t")
		}
		fmt.Printf("%s", value.do)
		fmt.Print("\n\tSymbol: ")
		for key, sym := range value.st.symbol {
			fmt.Printf("%s", sym)
			for i := 0; i < len(strconv.Itoa(value.st.state[key])); i++ {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n\n")
	}
}
