package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

const DING_ERROR_URL = "https://oapi.dingtalk.com/robot/send?access_token=d5e507b4ef9122b60af42e7418050a67eb43d797315a2504469de01570684ba9"
const ERROR_WORD = "提醒 "

func dingInfo1(msg string, bAll bool) error {
	client := &http.Client{}
	type sendDing struct {
		Msgtype string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
		At struct {
			IsAtAll bool `json:"isAtAll"`
		} `json:"at"`
	}
	var cmd sendDing
	cmd.Msgtype = "text"
	cmd.Text.Content = ERROR_WORD + msg
	cmd.At.IsAtAll = bAll
	a, _ := json.Marshal(cmd)
	req, err := http.NewRequest("POST", DING_ERROR_URL,
		strings.NewReader(string(a)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	res := struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}
	if res.Errmsg == "ok" {
		return nil
	} else {
		return xerrors.New(res.Errmsg)
	}
}

type PRICE_INFO struct {
	Price string `json:"price"`
}

const SPOT_URL = "https://api1.binance.com/api/v3/ticker/price?symbol=FTMUSDT"
const UST_FUTURES_URL = "https://fapi.binance.com/fapi/v1/ticker/price?symbol=FTMUSDT"

//const COIN_FUTURES_URL = "https://dapi.binance.com/dapi/v1/ticker/price?symbol=FTMUSD_PERP" json结构不一样 不解析了

func getFMTBinancePrice(url string) (float64, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	var priceInfo PRICE_INFO
	err = json.Unmarshal(body, &priceInfo)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	price, err := strconv.ParseFloat(priceInfo.Price, 64)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return price, nil
}

type OK_PRICE_INFO_DATA struct {
	Last string `json:"last"`
}

type OK_PRICE_INFO struct {
	Code string               `json:"code"`
	Msg  string               `json:"msg"`
	Data []OK_PRICE_INFO_DATA `json:"data"`
}

func getFMTOkPrice() (float64, error) {
	resp, err := http.Get("https://www.okex.com/api/v5/market/ticker?instId=FTM-USDT-SWAP")
	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	//log.Print("Response: ", string(body))
	var priceInfo OK_PRICE_INFO
	err = json.Unmarshal(body, &priceInfo)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if priceInfo.Code == "0" {
		data := priceInfo.Data[0]
		price, err := strconv.ParseFloat(data.Last, 64)
		if err != nil {
			log.Println(err)
			return 0, err
		}
		return price, nil
	} else {
		return 0, xerrors.New(priceInfo.Msg)
	}
}

func main() {
	for {
		time.Sleep(time.Duration(10 * time.Second))
		price1, err := getFMTBinancePrice(SPOT_URL)
		if err != nil {
			log.Println(err)
			continue
		}

		price2, err := getFMTBinancePrice(SPOT_URL)
		if err != nil {
			log.Println(err)
			continue
		}

		price3, err := getFMTOkPrice()
		if err != nil {
			log.Println(err)
			continue
		}

		lowPrice := price1
		highPrice := price1
		if price2 < price1 {
			lowPrice = price2
			highPrice = price1
		}
		if price3 < price2 {
			lowPrice = price3
			highPrice = price2
		}
		if price3 < price1 {
			lowPrice = price3
			highPrice = price1
		}
		log.Println(lowPrice)
		log.Println(highPrice)
		if highPrice > lowPrice*1.05 {
			dingInfo1("交易所ftm存在溢价", true)
		}
	}
}
