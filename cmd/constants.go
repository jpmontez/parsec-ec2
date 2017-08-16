package cmd


var tfCommands = struct {
	apply string
	destroy string
	init string
	output string
	plan string
	refresh string
}{
	"apply",
	"destroy",
	"init",
	"output",
	"plan",
	"refresh",
}

const (
	windows          = "Windows"
	currentSession   = "currentSession.json"
	terraform        = "terraform"
	parsecTemplate   = "parsec.tf"
	userDataTemplate = "user_data.tmpl"
	force            = "-force"
	spotInstanceID   = "spot_instance_id"
	ok               = "ok"
)
