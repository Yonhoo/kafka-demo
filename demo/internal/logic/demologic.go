package logic

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"kafka-demo/demo/internal/svc"
	"kafka-demo/demo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	kafka "github.com/segmentio/kafka-go"
)

type DemoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
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

	err = kafkaWriter.WriteMessages(l.ctx, kmsg...)

	if err != nil {
		resp = new(types.Response)
		resp.Message = err.Error()
		logx.Error(resp.Message)
		return
	}

	// Run the web server.
	resp = new(types.Response)
	resp.Message = req.Name
	return
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
