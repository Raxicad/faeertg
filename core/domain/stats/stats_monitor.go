package stats

type MonitorWatcherStats struct {
	ProductUrl string `json:"productUrl"`
	ProductPic string `json:"productPic"`
	Title      string `json:"title"`
	//Timestamp  time.Time `json:"-"`
}

type ComponentStats struct {
	Stats         string `json:"stats"`
	ComponentType string `json:"componentType"`
	ComponentName string `json:"componentName"`
	Timestamp     string `json:"timestamp"`
}
