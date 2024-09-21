package discord

const (
	/*
		Accepts a struct with the following parameters (example):
		{
			Players = []string{"PlayerName1", "PlayerName2"}
			ServerAddress = "byond://example.com:1234"
			GitHubLink = "https://github.com/example/exampleRepo"
		}
		Requires functions:
		join - for joining list of strings into a string
		currentUnixTimestamp - for generating a correct discord timestamp
	*/
	statusMessageDescriptionTemplate = `
{{- define "PlayersOnline"}}Players online: {{len .Players}}{{"\n"}}{{end -}}
{{define "PlayerList" -}}
	{{if (gt (len .Players) 0) -}}
		Players: {{ join .Players ", "}}{{"\n" -}}
	{{end -}}
{{end -}}
{{define "RoundTime"}}Round time: {{.RoundTime}}{{"\n"}}{{end -}}
{{define "Map"}}Current map: {{.Map}}{{"\n"}}{{end -}}
{{define "Evac" -}}
	{{if .Evac -}}
		**The station is undergoing evacuation procedures!**{{"\n" -}}
	{{end -}}
{{end -}}
{{define "ServerAddress"}}Server Address: ` + "`" + `{{.ServerAddress}}` + "`" + `{{"\n"}}{{end -}}
{{define "GitHub"}}GitHub: Temporarily Unavailable{{"\n"}}{{end -}}
{{define "LastUpdated"}}Last updated: <t:{{currentUnixTimestamp}}:R>{{"\n"}}{{end -}}

{{block "Description" . -}}
{{template "PlayersOnline" . -}}
{{template "PlayerList" . -}}
{{template "RoundTime" . -}}
{{template "Map" . -}}
{{template "Evac" . -}}
{{template "ServerAddress" . -}}
{{template "GitHub" . -}}
{{template "LastUpdated" . -}}
{{end -}}
`
	statusMessagePayloadTemplate = `
{{- block "Payload" . -}}
{
    "embeds": [{
	    "type": "rich",
	    "title": "{{.Title}}",
	    "description": "{{.Description}}",
	    "color": {{.Color}}
    }]
}
{{- end -}}
`
)

type descriptionPayloadParams struct {
	Players       []string
	ServerAddress string
	RoundTime     string
	Map           string
	Evac          bool
	GitHubLink    string
}

type statusMessagePayloadParams struct {
	Title       string
	Description string
	Color       string
}