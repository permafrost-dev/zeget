package globals

const (
	ApplicationName       = "eget"
	ApplicationRepository = "permafrost-dev/" + ApplicationName
)

func GetApplicationName() string {
	return ApplicationName
}
