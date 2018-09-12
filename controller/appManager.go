package monitor4sdn

type IFMap map[string]interface{}

type AppManager struct {
	applications []interface{}
}

var appManager *AppManager = newAppManager()

func newAppManager() *AppManager {
	manager := new(AppManager)
	manager.applications = make([]interface{}, 0)
	return manager
}

func GetAppManager() *AppManager {
	return appManager
}

func (manager *AppManager) RegisterApplication(app interface{}) {
	manager.applications = append(manager.applications, app)
}

func (manager *AppManager) GetApplications() []interface{} {
	return manager.applications
}
