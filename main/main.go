package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gammazero/workerpool"
)

type singleContractSourceCode struct {
	SourceCode   string `json:"SourceCode"`
	ContractName string `json:"ContractName"`
}

type contractSourceCode struct {
	Status  int                        `json:"status"`
	Message string                     `json:"message"`
	Result  []singleContractSourceCode `json:"result"`
}

type contractInfo struct {
	Address      string `json:"address"`
	ContractName string `json:"contractName"`
	Compiler     string `json:"compiler"`
	Balance      string `json:"balance"`
	TxCont       string `json:"txCont"`
	Date         string `json:"date"`
}

var (
	beginPage int64
	endPage   int64

	contractMap map[string]contractInfo

	pool *workerpool.WorkerPool
	m    *sync.RWMutex
)

const (
	apiKey      = "VTUX7DDG1BZS4XW9D1HK3HX5I1YI8F7UIA"
	contractDir = "./contracts/"
)

func httpGetContractListNumber(url string) error {
	// resp, err := http.Get(url)
	// if err != nil {
	// 	return err
	// }

	// defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(body))

	// if err != nil {
	// 	return err
	// }

	// Load the HTML document
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//wrapperSelection := doc.Find(".wrapper")
	wrapperSelection := doc.Find("body").Find("div").First()
	profileSelection := wrapperSelection.Find(".profile")

	pageSelection := profileSelection.Find(".row").Eq(1)
	pageSelection = pageSelection.Find("div").Eq(1)
	pageSelection = pageSelection.Find("p")
	spanSelection := pageSelection.Find("span")
	beginText := spanSelection.Find("b").First().Text()
	endText := spanSelection.Find("b").Next().Text()

	beginPage, _ = strconv.ParseInt(beginText, 10, 32)
	endPage, _ = strconv.ParseInt(endText, 10, 32)

	fmt.Println(beginPage)
	fmt.Println(endPage)

	return nil
}

func httpGetContractList(url string) error {

	// Load the HTML document
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//wrapperSelection := doc.Find(".wrapper")
	wrapperSelection := doc.Find("body").Find("div").First()
	profileSelection := wrapperSelection.Find(".profile")

	pageSelection := profileSelection.Find(".row").Eq(2)
	tableSelection := pageSelection.Find(".table-hover").First()
	tbodySelection := tableSelection.Find("tbody").First()

	addOne := 0
	tbodySelection.Find("tr").Each(func(i int, trSelection *goquery.Selection) {
		var info contractInfo

		trSelection.Find("td").Each(func(i int, tdSelection *goquery.Selection) {
			if i == 0 {
				info.Address = tdSelection.Find("a").Text()
			}

			if i == 1 {
				info.ContractName = tdSelection.Text()
			}

			if i == 2 {
				info.Compiler = tdSelection.Text()
			}

			if i == 3 {
				info.Balance = tdSelection.Text()
			}

			if i == 4 {
				info.TxCont = tdSelection.Text()
			}

			if i == 6 {
				info.Date = tdSelection.Text()
			}
		})

		m.RLock()
		_, ok := contractMap[info.Address]
		m.RUnlock()

		if !ok {
			m.Lock()
			contractMap[info.Address] = info
			m.Unlock()
			addOne++
			fmt.Printf("add contract to list ContractName: %s, Address: %s\n", info.ContractName, info.Address)
		} else {
			fmt.Printf("same contract\n")
		}
	})

	if addOne == 0 {
		return errors.New("get 错误了")
	} else {
		return nil
	}

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreatePath(path string) {
	exist, err := PathExists(path)
	if err != nil {
		return
	}

	if exist {

	} else {
		// 创建文件夹
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			fmt.Printf("mkdir success!\n")
		}
	}
}

func getContractFileName(info contractInfo) string {
	return fmt.Sprintf("%s/%s_%s.sol", contractDir, info.ContractName, info.Address)
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// func httpGetContractCode(url string, info contractInfo) error {
// 	var f *os.File
// 	var err1 error

// 	fileName := getContractFileName(info)

// 	if checkFileIsExist(fileName) { //如果文件存在
// 		return nil
// 	}

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		return err
// 	}

// 	var sourceCode contractSourceCode
// 	json.Unmarshal(body, &sourceCode)
// 	CreatePath(contractDir)

// 	f, err1 = os.Create(fileName)
// 	if err1 != nil {
// 		return err1
// 	}

// 	_, err1 = io.WriteString(f, sourceCode.Result[0].SourceCode) //写入文件(字符串)

// 	return err1
// }

func httpGetContractCode(url string, info contractInfo) error {
	var f *os.File
	var err1 error

	fileName := getContractFileName(info)

	if checkFileIsExist(fileName) { //如果文件存在
		return nil
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//wrapperSelection := doc.Find(".wrapper")
	SourceCode := doc.Find(".js-sourcecopyarea").Text()
	if SourceCode == "" {
		return errors.New("http error")
	}

	CreatePath(contractDir)

	f, err1 = os.Create(fileName)
	if err1 != nil {
		return err1
	}

	_, err1 = io.WriteString(f, SourceCode) //写入文件(字符串)
	fmt.Printf("Get Contract Source Code ContractName: %s, Address: %s\n", info.ContractName, info.Address)

	return err1
}

func main() {
	contractMap = make(map[string]contractInfo)
	pool = workerpool.New(2)
	m = new(sync.RWMutex)

	baseUrl := "https://etherscan.io/contractsVerified/"

	//获得页数
	httpGetContractListNumber(baseUrl)

	fmt.Println("get contract list begin")
	var wg sync.WaitGroup
	wg.Add(int(endPage - beginPage + int64(1)))
	for i := beginPage; i <= endPage; i++ {

		index := i
		pool.Submit(func() {
			for {
				if index == 1 {
					if httpGetContractList(baseUrl) == nil {
						break
					}
				} else {
					url := fmt.Sprintf("%s%d", baseUrl, index)
					if httpGetContractList(url) == nil {
						break
					}
				}
			}

			wg.Done()
		})

	}

	wg.Wait()
	fmt.Println("get contract list end")
	for _, info := range contractMap {
		contractUrl := fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getsourcecode&address=%s&apikey=%s", info.Address, apiKey)
		for {
			err := httpGetContractCode(contractUrl, info)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * 500)
		}

	}

	// wg.Add(len(contractMap))
	// fmt.Println("download source code begin")
	// for _, info := range contractMap {
	// 	contractUrl := fmt.Sprintf("https://etherscan.io/address/%s#code", info.Address)

	// 	myInfo := info
	// 	pool.Submit(func() {
	// 		for {
	// 			err := httpGetContractCode(contractUrl, myInfo)
	// 			if err == nil {
	// 				break
	// 			}
	// 		}

	// 		wg.Done()
	// 	})

	// }
	// wg.Wait()
	fmt.Println("download source code end")
	// var info contractInfo
	// info.Address = "0x061587df81c4d269f5375200dc8d7d8d7d5a0428"
	// info.ContractName = "zXBToken"
	// httpGetContractCode("https://etherscan.io/address/0x061587df81c4d269f5375200dc8d7d8d7d5a0428#code", info)
}
