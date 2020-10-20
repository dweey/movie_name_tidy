package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type RunCommand struct {
	Dir             string
	NameFormat      string
	ManualMode      bool
	RecentFileCount int
	Filename        string
}

func (*RunCommand) Name() string {
	return "run"
}

func (*RunCommand) Synopsis() string {
	return "run name tidy"
}

func (*RunCommand) Usage() string {
	return "run -<option>"
}

func (rc *RunCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&rc.Dir, "dir", "./", "指定目录")
	f.StringVar(&rc.NameFormat, "name_format", "[year][star]title_short", "指定文件名格式")
	f.BoolVar(&rc.ManualMode, "manual_mode", true, "确认模式")
	f.IntVar(&rc.RecentFileCount, "recent_file_count", 0, "只重命名指定数量的最近文件")
	f.StringVar(&rc.Filename, "filename", "", "指定原文件名")

}

func (rc *RunCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	want := []os.FileInfo{}
	fileTime := []int64{}
	files, _ := ioutil.ReadDir(rc.Dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if rc.Filename != "" && f.Name() != rc.Filename {
			continue
		}
		if f.Name() == "movie_name_tidy" || f.Name() == "movie_name_tidy.exe" {
			continue
		}

		want = append(want, f)
		fileTime = append(fileTime, f.ModTime().Unix())
	}

	if len(want) == 0 {
		fmt.Println("没有需要整理的文件")
		return subcommands.ExitSuccess
	}
	if rc.RecentFileCount > 0 {

	}

	for _, f := range want {
		rc.Tide(f.Name())
	}

	return subcommands.ExitSuccess
}

func (rc *RunCommand) Tide(name string) {
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Println("原文件名：" + name)

	params := map[string]interface{}{
		"name":        name,
		"custom_name": rc.NameFormat,
	}

	resp, err := HttpPostWithJson("https://api.rettrue.com/api/movie/query", params, time.Minute*5)
	if nil != err {
		fmt.Println(" 请求接口失败:" + err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		fmt.Println(" 获取结果失败1:" + err.Error())
		return
	}

	result := gjson.Parse(string(body))

	if !result.Get("code").Exists() {
		fmt.Println(" 获取结果失败2:" + string(body))
		return
	}

	if result.Get("code").Int() != 0 {
		fmt.Println(" 获取结果失败3 : " + result.Get("msg").String())
		return
	}

	names := result.Get("data.#.custom_name").Array()
	if len(names) == 1 {
		if rc.ManualMode == false {
			rc.Rename(name, result.Get("data.0.custom_name").String())
			return
		}
	}

	fmt.Println("0 : 取消")
	for k, v := range names {
		fmt.Println(fmt.Sprintf("%v : %v", k+1, v.String()))
	}
	fmt.Println("请选择（回车默认选第一个)：")

	var input string
	var id int
	var newName string
	for {
		_, err = fmt.Scanln(&input)
		if err != nil {
			//直接回车
			input = "1"
		}
		id, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("请输入数字!")
		} else {
			id = id - 1
			if result.Get(fmt.Sprintf("data.%v.custom_name", id)).Exists() {
				newName = result.Get(fmt.Sprintf("data.%v.custom_name", id)).String()
				break
			} else {
				fmt.Println("请输入正确序号!")
			}
		}
	}
	rc.Rename(name, newName)
}

func (rc *RunCommand) Rename(oldName string, newName string) {
	if oldName == newName {
		fmt.Println("新老文件名相同,跳过")
		return
	}

	oldFilePath := fmt.Sprintf("%s/%v", rc.Dir, oldName)
	newFilePath := fmt.Sprintf("%s/%v", rc.Dir, newName)

	is_exist, _ := PathExists(newFilePath)
	if is_exist == true {
		fmt.Println("文件已存在")
		return
	}

	err := os.Rename(oldFilePath, newFilePath)
	if nil != err {
		fmt.Println("重命名失败：" + err.Error())
	} else {
		fmt.Println("重命名成功：" + newName)
	}

}

func HttpPostWithJson(url string, data interface{}, timeout time.Duration) (*http.Response, error) {
	b, err := json.Marshal(data)
	if nil != err {
		return nil, err
	}
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return HttpPostWithHeader(url, header, string(b), timeout)
}

func HttpPostWithHeader(url string, header http.Header, data string, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if nil != err {
		return nil, err
	}
	req.Header = header

	client := &http.Client{Timeout: timeout}
	return client.Do(req)
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
