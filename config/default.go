package config

// Default config. Overridden from script file or environment
var defaultConfig = `
version: '1'

config:
	verbosity: 4
	threads: 100
	loop: 1
	task_timeout: 15s
default:
  http: &http
	follow_redirects: false
	headers:
		user-agent: Mozilla/5.0 (Windows NT 6.1; rv:45.0) Gecko/20100101 Firefox/45.0
		accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
		accept-language: en-US,en;q=0.5
		accept-encoding: gzip, deflate
	timeout: 15s
`
