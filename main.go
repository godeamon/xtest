package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"

	"gopkg.in/yaml.v3"
)

var MnemonicList = []string{
	"律 赤 尽 书 仓 传 宣 葱 脑 净 需 即",
	"嘴 籍 示 购 岸 按 曰 订 吴 稍 险 给",
	"歪 幕 杰 版 贵 慌 拆 信 辟 虑 剂 门",
	"响 雷 湿 阴 互 睡 谢 疯 粒 坐 沈 规",
	"株 略 挤 克 牺 食 班 哲 但 抽 冲 治",
	"耗 碳 企 画 巧 舟 站 竞 群 灰 练 门",
	"趣 冻 署 演 眼 抑 偶 牙 弟 隐 赵 南",
	"却 尾 绒 引 伴 卷 液 填 谈 滚 绘 北",
	"乎 东 脏 名 判 霍 仗 目 部 赤 法 南",
	"训 将 趋 留 换 狗 砍 候 敢 汤 耳 队",
	"承 筹 将 致 砖 兰 腰 治 磷 促 宜 给",
	"外 迟 衡 误 齿 浮 速 劝 罪 暗 百 南",
	"观 缘 候 抚 续 换 络 迹 暗 阁 妻 北",
	"箭 件 默 刻 意 邓 浸 词 刊 掉 处 北",
	"常 控 痛 社 缘 雪 元 脑 上 还 厅 队",
	"爷 洲 留 稻 止 毫 唐 继 戴 柬 市 取",
	"气 熔 牧 钢 胺 塔 楚 备 八 零 何 保",
	"签 腹 懂 骑 闪 献 勇 哲 麦 操 谷 色",
	"旨 鼻 壮 题 草 手 撑 湘 翻 仰 标 规",
	"吞 蜡 停 碰 于 简 北 贡 帐 调 格 据",
}

var (
	bankMnemonic = "玉 脸 驱 协 介 跨 尔 籍 杆 伏 愈 即"

	setPreKeys            = "setPreKeys"
	setSuffKeys           = "setSuffKeys"
	setPreKeysWithSender  = "setPreKeysWithSender"
	setSuffKeysWithSender = "setSuffKeysWithSender"

	queryKeys           = "getAllKeys"
	queryKeysWithSender = "getAllKeysWithSender"
)

type Config struct {
	Nodes        []string
	ContractName string
	contractFile string
}

var (
	cfg   *Config
	loger *log.Logger

	node1, node2, node3 string
	// 最好node1、2、3都是矿工
	// 网络中共四个节点，node1 发送所有类型交易
	// node2 只发送写交易
	// node3 只发送只读交易（没有写集）
	// node4 不发送交易
)

func init() {
	value, err := ioutil.ReadFile("./conf/xtest.yaml")
	if err != nil {
		panic(err)
	}
	cfg = &Config{}
	err = yaml.Unmarshal(value, cfg)
	if err != nil {
		panic(err)
	}

	if len(cfg.Nodes) < 3 {
		panic("At least three nodes")
	}
	node1 = cfg.Nodes[0]
	node2 = cfg.Nodes[1]
	node3 = cfg.Nodes[2]

	rand.Seed(101)
	file := "./log/" + time.Now().Format("20180102") + ".log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	loger = log.New(logFile, "[test]", log.LstdFlags|log.Lshortfile|log.LUTC) // 将文件设置为loger作为输出
}

// 程序启动前，先转账给bank:nuSMPvo6UUoTaT8mMQmHbfiRbJNbAymGh 999999999999999999
func main() {

	// a, _ := account.RetrieveAccount(bankMnemonic, 1)
	// fmt.Println(a.Address)
	// as := retrieveAccs()
	// for _, s := range as {
	// 	fmt.Println(s.Address)
	// }

	initDeployContract()
	fmt.Println("init deploy contract done")
	time.Sleep(time.Second * 5)

	initTransfer()
	fmt.Println("init transfer done")
	time.Sleep(time.Second * 5)

	go watch()
	go checkHeight()
	fmt.Println("run checkHeight")

	fmt.Println("run loop!")
	go loopNode2()
	go loopNode3()
	loop()
}

func watch() {
	// 创建节点客户端。
	client, err := xuper.New(node1)
	if err != nil {
		panic(err)
	}

	// 监听时间，返回 Watcher，通过 Watche 中的 channel 获取block。
	watcher, err := client.WatchBlockEvent(xuper.WithSkipEmplyTx())
	if err != nil {
		panic(err)
	}

	defer func() {
		// 关闭监听。
		watcher.Close()
		client.Close()
	}()

	for {
		b, ok := <-watcher.FilteredBlockChan
		if !ok {
			fmt.Println("watch block event channel closed.")
			break
		}
		loger.Println("height:", b.BlockHeight, " txCount:", len(b.Txs))
	}
}

// ndoe2 只发送写交易
func loopNode2() {
	accs := retrieveAccs()

	xclient, err := xuper.New(node2)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	r := rand.Int63n(100000000)
	args := map[string]string{
		"value": strconv.Itoa(int(r)),
	}
	for i, acc := range accs {
		if i >= 8 && i < 10 {
			go do(xclient, acc, setPreKeys, "", args, "node2setPreKeys")
		} else if i >= 10 && i < 12 {
			go do(xclient, acc, setSuffKeys, "", args, "node2setSuffKeys")
		} else if i >= 12 && i < 14 {
			go do(xclient, acc, setPreKeysWithSender, "", args, "node2setPreKeysWithSender")
		} else if i >= 14 && i < 16 {
			go do(xclient, acc, setSuffKeysWithSender, "", args, "node2setSuffKeysWithSender")
		}
	}
	select {}
}

// ndoe3 只发送只读交易，没有写集。
func loopNode3() {
	accs := retrieveAccs()

	xclient, err := xuper.New(node3)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	r := rand.Int63n(100000000)
	args := map[string]string{
		"value": strconv.Itoa(int(r)),
	}
	for i, acc := range accs {
		if i >= 16 && i < 18 {
			go do(xclient, acc, "", queryKeys, args, "node3getAllKeys")
		} else if i >= 18 && i < 21 {
			go do(xclient, acc, "", queryKeys, args, "node3getAllKeys")
		} else if i >= 4 && i < 8 {
			go do(xclient, acc, "", queryKeysWithSender, args, "node3getAllKeysWithSender")
		} else if i >= 8 && i < 16 {
			go do(xclient, acc, "", queryKeysWithSender, args, "node3getAllKeysWithSender")
		}
	}
	select {}
}

// node1
func loop() {
	accs := retrieveAccs()

	xclient, err := xuper.New(node1)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()
	r := rand.Int63n(100000000)
	args := map[string]string{
		"value": strconv.Itoa(int(r)),
	}
	for i, acc := range accs {
		if i >= 0 && i < 2 {
			go do(xclient, acc, setPreKeys, queryKeys, args, "node1getAllKeys")
		} else if i >= 2 && i < 4 {
			go do(xclient, acc, setSuffKeys, queryKeys, args, "node1getAllKeys")
		} else if i >= 4 && i < 6 {
			go do(xclient, acc, setPreKeysWithSender, queryKeysWithSender, args, "node1getAllKeysWithSender")
		} else if i >= 6 && i < 8 {
			go do(xclient, acc, setSuffKeysWithSender, queryKeysWithSender, args, "node1getAllKeysWithSender")
		}
	}

	select {}
}

// 如果某个节点30s高度还不变，则程序退出。
func checkHeight() {
	begin := time.Now()
	xclient, err := xuper.New(node1)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	lastHeight := int64(-1)

	for {
		time.Sleep(time.Second * 30) // 30s 检查一次
		bcs, err := xclient.QuerySystemStatus()
		if err != nil {
			panic(err)
		}
		for _, status := range bcs.SystemsStatus.BcsStatus {
			if status.Bcname == "xuper" {
				height := status.Block.Height
				if height == lastHeight {
					fmt.Println("total time: ", time.Since(begin))
					panic("check height failed")
				} else {
					lastHeight = height
				}
			}
		}
	}
}

func do(xclient *xuper.XClient, acc *account.Account, method, queryMethod string, args map[string]string, desc string) {
	for {
		r := rand.Int63n(100000)
		args1 := map[string]string{
			"value": strconv.Itoa(int(r)),
		}
		if method != "" {
			_, err := xclient.InvokeWasmContract(acc, cfg.ContractName, method, args1)
			if err != nil {
				// fmt.Println("Invoke contract error:", err.Error(), "desc: ", desc)
			} else {
				// fmt.Println("Invoke contract succ:", method, "desc: ", desc)
			}
			time.Sleep(time.Millisecond * 100)
		}

		if queryMethod != "" {
			_, err := xclient.InvokeWasmContract(acc, cfg.ContractName, queryMethod, args1)
			if err != nil {
				// fmt.Println("Invoke contract error:", err.Error(), "desc: ", desc)
			} else {
				// fmt.Println("Invoke contract succ:", method, "desc: ", desc)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func retrieveAccs() []*account.Account {
	result := make([]*account.Account, 0, 20)
	for _, m := range MnemonicList {
		acc, _ := account.RetrieveAccount(m, 1)
		result = append(result, acc)
	}
	return result
}

func initTransfer() {
	bank, _ := account.RetrieveAccount(bankMnemonic, 1)
	xclient, err := xuper.New(node1)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	accs := retrieveAccs()

	for _, acc := range accs {
		_, err := xclient.Transfer(bank, acc.Address, "90000000000")
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
		fmt.Println("transfer done, address: ", acc.Address)
	}
	for _, acc := range accs {
		bal, err := xclient.QueryBalance(acc.Address)
		if err != nil {
			panic(err)
		}
		fmt.Println("address: ", acc.Address, " balance: ", bal)
	}
}

func initDeployContract() {
	contractAcc := "XC1111111199999999@xuper"

	bank, _ := account.RetrieveAccount(bankMnemonic, 1)
	xclient, err := xuper.New(node1)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	_, err = xclient.CreateContractAccount(bank, contractAcc)
	if err != nil {
		fmt.Println("create contract account failed:", err)
	}

	time.Sleep(time.Second * 3)

	_, err = xclient.Transfer(bank, contractAcc, "1000000000")
	if err != nil {
		fmt.Println("transfer contract account failed:", err)
	}

	code, err := ioutil.ReadFile(cfg.contractFile)
	if err != nil {
		panic(err)
	}

	args := map[string]string{
		"creator": "test",
		"key":     "test",
	}

	bank.SetContractAccount(contractAcc)
	tx, err := xclient.DeployWasmContract(bank, cfg.ContractName, code, args)
	if err != nil {
		fmt.Println("deploy contract failed:", err)
	} else {
		fmt.Printf("Deploy wasm Success!TxID:%x\n", tx.Tx.Txid)
	}
}
