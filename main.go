package main

import (
	"bytes"
	"context"
	"draft/l1_events"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var (
	server_address 		= flag.String("ip", "0.0.0.0:12000", "服务器地址")
	target_id 			= flag.Int("target_id", 30000, "自定义字段，以账号id为例")
	coroutines_count 	= flag.Int("users", 2, "同时运行的最大协程数")
	engine_mode 		= flag.String("mode", "amount_mode", "引擎模式")
	runtime 			= flag.Int("timeout", 30, "time_mode下的运行时间")
	amount 				= flag.Int("amount", 2500, "amount_mode下的总数")
	logger 				log.Logger
	// 用于生成随机字符串，测试用
	letters				= []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	succeed_count		int
	failed_count		int
)

// 使用引擎方式启动的结构体
type engine struct {
	server_address		string
	start_id 			int
	coroutines_count	int
}

type msgError struct {
	when	int64
	data 	[]byte
}

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 自行定义的错误类型
func (e msgError) Error() string {
	if e.data == nil {
		return fmt.Sprintf("%v: 返回数据为空", e.when)
	}
	return fmt.Sprintf("error data in %v: %v", e.when, e.data)
}

func NewmsgError(d []byte) error {
	return &msgError{time.Now().Unix(), d}
}

// 这个函数用于接收一个buffer， 提取buffer中指定id的包，返回指定包的pb数据
func createSock(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	return conn, err
}

func main_event(address string, ID int) error{
	// 网关服流程
	conn, err := createSock(address)
	if err != nil {return err}

	// 连接成功后创建属于该协程的全局rsp_pag_buffer，用于收包，对包内不完整的包做缓存处理
	rsp_pag_buffer := bytes.NewBuffer([]byte{})
	gate_c := l1_events.NewGateClient(conn, *rsp_pag_buffer, uint8(0))
	err = gate_c.GateLoginEvent(ID)
	conn.Close()
	if err != nil {
		logger.Output(2, fmt.Sprintf("网关服登录失败, error: %s", err.Error()))
		return err}
	logger.Output(2, "网关服登录成功")

	// 游戏服流程
	conn2, err := createSock(fmt.Sprintf("%s:%d", gate_c.Login_rst.Ip, gate_c.Login_rst.Port))
	defer conn2.Close()
	if err != nil {return err}
	rsp_pag_buffer_game := bytes.NewBuffer([]byte{})
	base_c := l1_events.NewGameClient(conn2, *rsp_pag_buffer_game, uint8(0))
	// 登录
	err = base_c.BaseLoginEvent(gate_c.Login_rst.ActorId, gate_c.Login_rst.ServerId)
	if err != nil {
		logger.Output(2, fmt.Sprintf("登录失败, error: %s", err.Error()))
		return err}
	logger.Output(2, "登录成功")
	return nil
}

// 以时间为基础的运行模式
func (e *engine) time_mode(seconds int) {
	// logger.SetPrefix(fmt.Sprintf("iggID:%d ", start_id))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds) * time.Second)
	for i := 0; i < e.coroutines_count; i++ {
		time.Sleep(100 * time.Millisecond)
		go func(ctx context.Context) {
			ticker := time.NewTicker(1 * time.Second)
			for _ = range ticker.C {
				select {
					case <-ctx.Done():
						fmt.Println("child process interrupt...")
						return
					default:
						// 运行main_event并统计结果，按结果是否成功来进行判断
						st := time.Now()
						e.start_id++
						err := main_event(e.server_address, e.start_id)
						if err != nil {
							failed_count++
							logger.Output(2, fmt.Sprintf("运行失败%s" , err.Error()))
						} else {
							succeed_count++
							logger.Output(2, fmt.Sprintf("运行成功, 运行耗时: %d", time.Since(st)))
						}
				}
			}
		}(ctx)
	}
	defer cancel()
	select {
	case <-ctx.Done():
		time.Sleep(1 * time.Second)
		fmt.Println("main process exit!")
	}
}

// 以总数为基础的运行模式
func (e *engine) amount_mode(amount int) {
	gen := func(ctx context.Context) <- chan int {
		count_chan := make(chan int)
		count := 0
		runner := func() {
			for {
				select {
					case <-ctx.Done():
						return
					case count_chan <-count:
						st := time.Now()
						e.start_id++
						err := main_event(e.server_address, e.start_id)
						if err != nil {
							failed_count++
							logger.Output(2, fmt.Sprintf("运行失败%s" , err.Error()))
						} else {
							succeed_count++
							logger.Output(2, fmt.Sprintf("运行成功, 运行耗时: %d", time.Since(st)))
						}
						count++
				}
			}
		}
		for i:=0 ; i <e.coroutines_count; i++ {
			go runner()
			e.start_id++
		}
		return count_chan
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for n := range gen(ctx) {
		if n == amount {
			break
		}
	}
}

// init过程建立运行日志
func init() {
	logFile, err := os.Create("./" + time.Now().Format("20060102") + ".txt")
	if err != nil {
		fmt.Println(err)
	}
	logger = *log.New(logFile, "L1_AOW ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	flag.Parse()
	logger.Output(2, "-------------------l1压测-------------------")
	logger.Output(2, "压测内容： 新用户选洲注册")
	logger.Output(2, "具体参数：")
	logger.Output(2, " server_address: " + *server_address)
	logger.Output(2, " mode: " + *engine_mode)
	logger.Output(2, fmt.Sprintf(" coroutines_count: %d", *coroutines_count))
	defer logger.Output(2, "引擎运行完毕")
	start_id := *target_id
	engine := engine{server_address: *server_address, start_id: start_id,coroutines_count: *coroutines_count}
	switch *engine_mode{
	case "time_mode" :
		logger.Output(2, "进入时间运行模式")
		engine.time_mode(*runtime)
	case "amount_mode":
		logger.Output(2, "进入总数运行模式")
		engine.amount_mode(*amount)
	}
	defer logger.Output(2, fmt.Sprintf("运行总数: %d", (succeed_count + failed_count)))
	defer logger.Output(2, fmt.Sprintf("运行成功数: %d", succeed_count))
	defer logger.Output(2, fmt.Sprintf("运行成功率: %f%%", 100 * float32(succeed_count/(succeed_count + failed_count))))

}