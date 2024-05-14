package globals

const (
	ApplicationName       = "zeget"
	ApplicationRepository = "permafrost-dev/" + ApplicationName
)

func GetApplicationName() string {
	return ApplicationName
}
