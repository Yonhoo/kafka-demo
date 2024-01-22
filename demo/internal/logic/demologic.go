package logic

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"kafka-demo/demo/internal/svc"
	"kafka-demo/demo/internal/types"

	cron "github.com/robfig/cron"
	"github.com/zeromicro/go-zero/core/logx"

	kafka "github.com/segmentio/kafka-go"
)

type DemoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type CronMetaData struct {
	invoked bool
	crr     *cron.Cron
}

var crrData = &CronMetaData{
	invoked: true,
	crr:     cron.New(),
}

func NewDemoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoLogic {
	return &DemoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoLogic) Demo(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	if req.Name == "cancel" && crrData.invoked == true {
		crrData.invoked = false
		crrData.crr.Stop()
		logx.Info("cancel cron job")
	}

	if req.Name == "cron" && crrData.invoked == false {
		crrData.invoked = true
		crrData.crr.AddFunc("0 */2 * * *", func() {
			fetchRSSContent(l.ctx)
			logx.Info("fetch rss per 2 hour")
		})

		crrData.crr.Start()

		logx.Info("start cron job")

		time.Sleep(time.Second * 3)
	}

	if req.Name == "once" {
		err = fetchRSSContent(l.ctx)
	}

	// Run the web server.
	resp = new(types.Response)
	if err != nil {
		resp.Message = err.Error()
		return
	}
	resp.Message = req.Name
	return
}

func fetchRSSContent(ctx context.Context) (err error) {
	logx.Info("fetch rss content")
	topic := "kafka-demo-topic"
	kafkaWriter := getKafkaWriter(topic)

	defer kafkaWriter.Close()

	rss, err := getRSSContent("https://rsshub.app/v2ex/topics/latest")

	logx.Infof("rss size: %d", len(rss.Channel.Items))

	var kmsg []kafka.Message
	for _, itemDec := range rss.Channel.Items {

		re := regexp.MustCompile(`/t/(\d+)`)

		matches := re.FindStringSubmatch(itemDec.GUID.Value)

		kmsg = append(kmsg, kafka.Message{
			Key:   []byte(matches[1]),
			Value: []byte(itemDec.Description.Text),
		})

	}

	err = kafkaWriter.WriteMessages(ctx, kmsg...)

	if err != nil {
		logx.Error("error occured : ", err)
		return err
	}
	return nil
}

func getKafkaWriter(topic string) *kafka.Writer {

	return &kafka.Writer{
		Addr:         kafka.TCP("127.0.0.1:9092", "127.0.0.1:9093", "127.0.0.1:9094"),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{}, // 指定分区的balancer模式为最小字节分布
		RequiredAcks: kafka.RequireAll,    // ack模式
		Async:        true,                // 异步
	}
}

func getRSSContent(url string) (RSS, error) {
	// 使用 go-zero/rest 发起 GET 请求
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return RSS{}, err
	}
	defer response.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return RSS{}, err
	}

	// 解析 XML
	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return RSS{}, err
	}

	// 提取 <description> 内容

	// 返回请求结果
	return rss, nil
}
