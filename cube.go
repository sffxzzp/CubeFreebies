package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

type (
	Config struct {
		UserName string
		Password string
	}
	Cube struct {
		UA       string
		Client   *http.Client
		UserName string
		Password string
	}
	LoginRet struct {
		ResultCode string
		Result     string
	}
	BoxRet struct {
		StatusCode int
		Data       struct {
			GoodID    int
			GoodType  int
			GoodName  string
			GoodValue string
			C_Money   int
		}
	}
	FreeGameRet struct {
		ResultCode int
		Msg        string
		Result     struct {
			List       []FreeGameData
			TotalCount int
		}
	}
	FreeGameData struct {
		GameID    int
		GoodsID   int
		GoodsName string
	}
	OwnedGameRet struct {
		Data  []OwnedGameData
		State int
	}
	OwnedGameData struct {
		M_GameID int
	}
	OrderInfoRet struct {
		Msg    string
		Result struct {
			IconImg string
		}
		ResultCode int
	}
	OrderResultRet struct {
		Msg        string
		Result     string
		ResultCode int
	}
)

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (c *Cube) httpGet(url string) []byte {
	retry := 3
	for retry > 0 {
		res, err := c.Client.Get(url)
		if err != nil {
			retry--
			continue
		}
		defer res.Body.Close()
		rbytes, err := io.ReadAll(res.Body)
		if err != nil {
			retry--
			continue
		}
		return rbytes
	}
	return []byte{}
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func (c *Cube) httpPost(url string, contentType string, body io.Reader, referer string) []byte {
	retry := 3
	for retry > 0 {
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			retry--
			continue
		}
		if referer != "" {
			req.Header.Add("Referer", referer)
		}
		if contentType != "" {
			req.Header.Add("Content-Type", contentType)
		}
		res, err := c.Client.Do(req)
		if err != nil {
			retry--
			continue
		}
		defer res.Body.Close()
		rbytes, err := io.ReadAll(res.Body)
		if err != nil {
			retry--
			continue
		}
		return rbytes
	}
	return []byte{}
}

func NewCube() *Cube {
	cCont, _ := os.ReadFile("config.json")
	var conf Config
	json.Unmarshal(cCont, &conf)
	jar, _ := cookiejar.New(nil)
	return &Cube{
		UserName: conf.UserName,
		Password: conf.Password,
		UA:       "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36 Edg/126.0.0.0",
		Client: &http.Client{
			Jar: jar,
		},
	}
}

func (c *Cube) Login() bool {
	loginUrl := "https://account.cubejoy.com/Handler/Data.ashx"
	postData := fmt.Sprintf("action=mobileloginsubmit&KUserName=s&mobilecode=&KPassword=&PUserName=%s&PPwd=%s&M=%s", url.QueryEscape(c.UserName), c.Password, fmt.Sprintf("%.17f", rand.Float64()))
	res := c.httpPost(loginUrl, "application/x-www-form-urlencoded; charset=UTF-8", bytes.NewBufferString(postData), "https://account.cubejoy.com/html/mobile/login.html")
	if len(res) < 1 {
		return false
	}
	var loginRet LoginRet
	json.Unmarshal(res, &loginRet)
	return loginRet.ResultCode == "1"
}

func (c *Cube) OpenBoxes() {
	fmt.Println("Open cases...")
	c.httpGet("https://me.cubejoy.com/case/indexbox")
	res := c.httpGet("https://me.cubejoy.com/Case/WoodCase")
	var boxRet BoxRet
	json.Unmarshal(res, &boxRet)
	if boxRet.StatusCode == 200 {
		fmt.Printf("Case opened, get %s\n\n", boxRet.Data.GoodName)
	} else {
		fmt.Printf("Case already opened or no cases available\n\n")
	}
}

func (c *Cube) inIntArray(array []int, target int) bool {
	for _, value := range array {
		if value == target {
			return true
		}
	}
	return false
}

func (c *Cube) CheckFreeGames() []FreeGameData {
	fmt.Printf("Getting free games data...")
	freeGameUrl := "https://mine.cubejoy.com/H5/FreeGameJson"
	res := c.httpPost(freeGameUrl, "", nil, "https://mine.cubejoy.com/H5/FreeGame")
	var freeGameRet FreeGameRet
	json.Unmarshal(res, &freeGameRet)
	ownedGameUrl := "https://mine.cubejoy.com/H5/UserOwnGame"
	res = c.httpPost(ownedGameUrl, "", nil, "https://mine.cubejoy.com/H5/FreeGame")
	var ownedGameRet OwnedGameRet
	json.Unmarshal(res, &ownedGameRet)
	var notOwned []FreeGameData
	if freeGameRet.ResultCode == 1 && ownedGameRet.State == 1 {
		var ownedGamesList []int
		for _, game := range ownedGameRet.Data {
			ownedGamesList = append(ownedGamesList, game.M_GameID)
		}
		for _, game := range freeGameRet.Result.List {
			if !c.inIntArray(ownedGamesList, game.GameID) {
				notOwned = append(notOwned, game)
			}
		}
	}
	fmt.Println("Complete!")
	return notOwned
}

func (c *Cube) GetFreeGame(game FreeGameData) bool {
	fmt.Printf("Claiming free game: %s...", game.GoodsName)
	referer := fmt.Sprintf("https://mine.cubejoy.com/h5/FreeOrderGoods?goodsid=%d", game.GoodsID)
	freeGameInfoUrl := "https://mine.cubejoy.com/H5/FreeOrderJson"
	res := c.httpPost(freeGameInfoUrl, "application/json;charset=UTF-8", bytes.NewBufferString(fmt.Sprintf("{\"goodsId\":\"%d\"}", game.GoodsID)), referer)
	var orderInfoRet OrderInfoRet
	json.Unmarshal(res, &orderInfoRet)
	if orderInfoRet.ResultCode == 1 {
		claimUrl := "https://mine.cubejoy.com/H5/GetFreeOrderGoodsJson"
		postData := fmt.Sprintf(`{"gameId":%d,"gameName":"%s","iconImgUrl":"%s"}`, game.GameID, game.GoodsName, orderInfoRet.Result.IconImg)
		res = c.httpPost(claimUrl, "application/json;charset=UTF-8", bytes.NewBufferString(postData), referer)
		var orderResultRet OrderResultRet
		json.Unmarshal(res, &orderResultRet)
		fmt.Println(orderResultRet.Result)
		return orderResultRet.ResultCode == 1
	} else {
		return orderInfoRet.ResultCode == 1
	}
}
