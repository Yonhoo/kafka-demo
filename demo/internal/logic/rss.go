package logic

// 定义 XML 结构体
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Description CDATA  `xml:"description"`
	Title       string `xml:"title"`
	PubDate     string `xml:"pubDate"`
	GUID        GUID   `xml:"guid"`
	Link        string `xml:"link"`
	Author      string `xml:"author"`
}

type GUID struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type CDATA struct {
	Text string `xml:",cdata"`
}
